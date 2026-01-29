import asyncio
import os
import json
import nats
from nats.errors import ConnectionClosedError, TimeoutError, NoRespondersError
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.future import select
import httpx
import uuid
import structlog
import time

# Import shared schemas and db models
# We need to ensure the python path includes the root directory
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from shared.schemas import AgentConfig
from shared.db import AgentDB, DATABASE_URL
from shared.logger import get_logger, setup_logger
from context import AgentContext
from llm.ollama import OllamaClient
from tools import ToolRegistry
# Import tools to register
from services.memory_mcp.tools import remember, recall

class AgentRunner:
    def __init__(self, agent_config: AgentConfig, nc):
        self.config = agent_config
        self.nc = nc
        self.js = nc.jetstream()
        
        # Setup Logger
        self.log = get_logger(f"runner.{self.config.name}")
        self.log.info("agent_runner_initialized", agent=self.config.name, role=self.config.role)
        
        # Context Management
        self.context = AgentContext(
            max_history=10, 
            system_prompt=self.config.prompt_config.get("system_prompt")
        )
        
        # LLM Client Setup
        # V1: Assume Ollama for now, extend later based on config.backend
        base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        self.llm_client = OllamaClient(base_url=base_url)
        self.log.info("llm_client_ready", backend="ollama", url=base_url)

        # Tool Registry
        self.registry = ToolRegistry()
        self.registry.register(remember)
        self.registry.register(recall)
        self.log.info("tools_registered", count=len(self.registry._tools))

    # async def connect(self): -> Handled by main loop


    async def start(self):
        # await self.connect() -> Already connected in main

        
        # Input Stream (Work Queue)
        # We listen to specific inputs defined in config
        input_subjects = self.config.messaging.inputs
        if not input_subjects:
            self.log.warning("no_input_subjects_defined")
            return

        for subject in input_subjects:
            # Subscribe with Queue Group for load balancing if multiple instances
            sub_subject = subject
            if not subject.endswith(">") and not subject.endswith("*"):
                 # Robustness: ensure we catch specific messages
                 pass
            
            # Ensure stream exists (idempotent)
            try:
                stream_name = f"stream_{subject.replace('.', '_').replace('*', 'all').replace('>', 'all')}"
                # sanitize stream name
                stream_name = stream_name.replace(">", "all").replace("*", "all")
                
                await self.js.add_stream(name=stream_name, subjects=[subject])
            except Exception as e:
                self.log.warning("stream_creation_note", subject=subject, error=str(e))

            self.log.info("subscribing_to_input", subject=subject)
            await self.js.subscribe(subject, cb=self.handle_message, manual_ack=True)

        # Keep alive
        try:
            await asyncio.Future()
        except asyncio.CancelledError:
            await self.nc.close()

    async def handle_message(self, msg):
        try:
            data = json.loads(msg.data.decode())
            self.log.info("message_received", subject=msg.subject)
            
            # 1. Update Context (User Input)
            # Only add if it's a 'user' message or pertinent event
            content = data.get("payload", {}).get("content") or data.get("content")
            
            # Hydrate Context from Persistent History (Context Layering) (Simplified for brevity in diff)
            # ... (Assuming hydration logic remains or is called once at startup in real app)
            # For this loop, we just process.
            
            if content:
                self.context.add_message("user", content)

                # Broadcast Context for Inspector
                try:
                    debug_payload = {
                        "id": f"debug-{uuid.uuid4().hex[:8]}",
                        "type": "event",
                        "source_agent_id": self.config.name,
                        "sender": self.config.name,
                        "recipient": "user", 
                        "content": "Debug Context State",
                        "payload": {
                            "type": "agent_context_state",
                            "system_prompt": self.config.prompt_config.get("system_prompt"),
                            "history": [{"role": m.role, "content": m.content} for m in self.context.history]
                        }
                    }
                    output_channel = self.config.messaging.outputs[0] if self.config.messaging.outputs else f"chat.agent.{self.config.name}.reply"
                    await self.js.publish(output_channel, json.dumps(debug_payload).encode())
                except Exception as e:
                    pass

                # 2. ReAct Loop (Think -> Tool -> Think)
                # We loop until the LLM produces a final answer (content)
                max_turns = 5
                current_turn = 0
                
                final_response = None
                
                while current_turn < max_turns:
                    # Construct messages for API
                    messages = [{"role": m.role, "content": m.content} for m in self.context.history]
                    
                    # Generate
                    response = await self.llm_client.generate(
                        model="llama3", # TODO: Get from config
                        messages=messages,
                        system_prompt=self.config.prompt_config.get("system_prompt"),
                        tools=self.registry.definitions
                    )
                    
                    # Check for Tool Calls
                    if response.tool_calls:
                        self.log.info("tool_calls_detected", count=len(response.tool_calls))
                        
                        # Add assistant message with tool calls (if API requires it, 
                        # usually we add the tool results directly or an assistant msg first)
                        # For Ollama/Basic ReAct: We append tool results as 'system' or 'user' with OBSERVATION.
                        # For Native: We follow the role schema.
                        
                        # Let's assume standard ChatML roles:
                        # Assistant: "Use tool X"
                        # Tool: "Result"
                        
                        # We don't have a native 'tool' role in our simple Context class yet,
                        # so we'll simulate it or extend context.
                        # For now, we just act.
                        
                        for tool_call in response.tool_calls:
                            self.log.info("executing_tool", tool=tool_call.name, args=tool_call.arguments)
                            
                            result = await self.registry.execute(tool_call.name, tool_call.arguments)
                            
                            # Add observation to context
                            # "Observation: [Result]"
                            self.context.add_message("system", f"Tool '{tool_call.name}' Output: {result}")
                            self.log.info("tool_execution_complete", result=result)
                            
                        current_turn += 1
                        continue # Loop again to let LLM see result
                        
                    elif response.content:
                        # Final Answer
                        final_response = response.content
                        break
                    else:
                        # No content, no tools? 
                        final_response = "I am confused."
                        break

                # 3. Publish Response
                self.context.add_message("assistant", final_response)
                
                output_channel = self.config.messaging.outputs[0] if self.config.messaging.outputs else f"chat.agent.{self.config.name}.reply"
                
                # Ensure output stream exists (Best Effort)
                stream_name = output_channel.replace(".", "-").replace(">", "all").replace("*", "all")
                subjects = [output_channel, f"{output_channel}.>"]
                try:
                    await self.js.add_stream(name=stream_name, subjects=subjects)
                except Exception:
                    pass

                reply_msg = {
                    "id": f"choice-{uuid.uuid4().hex}", # Use 'choice' ID style or standard
                    "source_agent_id": self.config.name,
                    "sender": self.config.name,
                    "recipient": data.get("source_agent_id", "user"),
                    "content": final_response,
                    "type": "text",
                    "timestamp": datetime.utcnow().timestamp(),
                    "payload": {
                        "content": final_response
                    }
                }

                await self.js.publish(f"{output_channel}.event", json.dumps(reply_msg).encode())
                self.log.info("reply_published", channel=f"{output_channel}.event")
                
                await msg.ack()

        except Exception as e:
            self.log.error("message_handling_error", error=str(e))
        finally:
            # Clean up context
            # reset_trace_id(trace_token)
            pass


async def main():
    # Setup Logger
    setup_logger("runner", level=os.getenv("LOG_LEVEL", "INFO"))
    
    # Database Connection
    engine = create_async_engine(DATABASE_URL)
    async_session = sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)

    # NATS Connection
    nc = await nats.connect(os.getenv("NATS_URL", "nats://localhost:4222"))
    
    runners = {}

    structlog.get_logger("runner.main").info("service_started", config={"ollama": os.getenv('OLLAMA_BASE_URL', 'http://localhost:11434')})

    while True:
        try:
            async with async_session() as session:
                result = await session.execute(select(AgentDB))
                agents = result.scalars().all()
                
                active_agent_names = set()

                for agent_db in agents:
                    config = AgentConfig(**agent_db.config)
                    active_agent_names.add(config.name)
                    
                    if config.name not in runners:
                        # Start new runner
                        runner = AgentRunner(config, nc)
                        await runner.start()
                        runners[config.name] = runner
                    else:
                        # Check if config updated
                        current_runner = runners[config.name]
                        # Simple comparison of config dicts
                        # We compare the model_dump of the new config with the stored config
                        if current_runner.config.model_dump() != config.model_dump():
                            structlog.get_logger("runner.watcher").info("config_changed", message=f"Configuration changed for {config.name}. Restarting runner...")
                            await current_runner.stop()
                            
                            # Start new runner with updated config
                            new_runner = AgentRunner(config, nc)
                            await new_runner.start()
                            runners[config.name] = new_runner
                
                # Stop removed agents
                for name in list(runners.keys()):
                    if name not in active_agent_names:
                        await runners[name].stop()
                        del runners[name]

        except Exception as e:
            # We don't have a top-level logger instance easily accessible here without global ref or creating one
            # But we can just use print for fatal watcher error or get a temporary logger
            structlog.get_logger("runner.watcher").error("watcher_loop_error", error=str(e))
        
        await asyncio.sleep(5)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        pass
