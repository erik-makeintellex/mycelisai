import asyncio
import os
import json
import logging
from typing import Optional, Dict, Any, List
# import nats # Removed raw nats usage
from nats.errors import ConnectionClosedError, TimeoutError, NoRespondersError
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.future import select
import httpx
import uuid
import structlog
import time
from datetime import datetime

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

# Import Relay SDK
from relay.client import RelayClient
from relay.proto.swarm.v1 import swarm_pb2
from google.protobuf.json_format import MessageToDict

class AgentRunner:
    def __init__(self, agent_config: AgentConfig, nats_url: str):
        self.config = agent_config
        self.nats_url = nats_url
        
        # Setup Logger
        self.log = get_logger(f"runner.{self.config.name}")
        self.log.info("agent_runner_initialized", agent=self.config.name, role=self.config.role)
        
        # Initialize Relay Client
        # Using config.name as agent_id. Team ID defaults to "default" or should be in config?
        # Assuming config.role maps to team or fixed "legacy-migration" team for now.
        self.client = RelayClient(
            agent_id=self.config.name,
            team_id=self.config.role or "legacy",
            source_uri="swarm:python:runner",
            nats_url=self.nats_url
        )
        
        # Context Management
        self.context = AgentContext(
            max_history=10, 
            system_prompt=self.config.prompt_config.get("system_prompt")
        )
        
        # LLM Client Setup
        base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        self.llm_client = OllamaClient(base_url=base_url)
        self.log.info("llm_client_ready", backend="ollama", url=base_url)

        # Tool Registry
        self.registry = ToolRegistry()
        self.registry.register(remember)
        self.registry.register(recall)
        self.log.info("tools_registered", count=len(self.registry._tools))

    async def start(self):
        # Connect to NATS via Relay
        await self.client.connect()
        self.log.info("relay_client_connected")
        
        # Input Stream (Work Queue)
        input_subjects = self.config.messaging.inputs
        if not input_subjects:
            self.log.warning("no_input_subjects_defined")
            # We still stay alive
        else:
            for subject in input_subjects:
                # Ensure stream exists (Best Effort via raw JS access if needed)
                # Ideally RelayClient handles this, but for legacy compatibility we might need to rely on 
                # pre-existing streams or explicit creation.
                # RelayClient exposes `js`? No, creates it internally.
                # We can access it via `self.client.js`.
                
                try:
                    stream_name = f"stream_{subject.replace('.', '_').replace('*', 'all').replace('>', 'all')}"
                    stream_name = stream_name.replace(">", "all").replace("*", "all")
                    
                    # Ensure JS is available
                    if self.client.js:
                        await self.client.js.add_stream(name=stream_name, subjects=[subject])
                        self.log.info("stream_ensured", stream=stream_name)
                except Exception as e:
                    self.log.warning("stream_creation_note", subject=subject, error=str(e))

                self.log.info("subscribing_to_input", subject=subject)
                # Subscribe delegating to handle_message
                await self.client.subscribe(subject, self.handle_message)

        # Keep alive loop or wait
        # Main loop manages lifecycle, but here we can just wait?
        # The main runner loop calls start() and stores runner.
        # It does NOT wait on runner.start().
        # Wait, previous start() had:
        # await self.js.subscribe(...)
        # await asyncio.Future() -> This blocked!
        # Incorrect. runner/main.py called `await runner.start()`.
        # If `start()` blocks, the main loop stops iterating other agents!
        # Previous code:
        #     await self.js.subscribe(...)
        #     # Keep alive
        #     try: await asyncio.Future() ...
        # THAT WAS A BUG in my previous reading or the previous code.
        # If `runner.start()` blocks, then `main()` loop hangs on the first agent.
        # Inspecting previous code:
        #     await runner.start() 
        # Line 92: await asyncio.Future()
        # YES, the previous code was blocking! It could only run ONE agent.
        # FIX: We should NOT block in start(). We should start background tasks.
        
        # Since we use callbacks with NATS, we don't need to block `start()`.
        # We just set up subscriptions and return.
        pass

    async def stop(self):
        await self.client.close()

    async def handle_message(self, envelope: swarm_pb2.MsgEnvelope):
        """
        Handle incoming Mycelis Protocol V2 Messages.
        """
        try:
            self.log.info("message_received", id=envelope.id, type=envelope.type)
            
            # Extract Content
            content = ""
            recipient = ""
            
            if envelope.type == swarm_pb2.MESSAGE_TYPE_TEXT:
                content = envelope.text.content
                recipient = envelope.text.recipient_id
            elif envelope.type == swarm_pb2.MESSAGE_TYPE_EVENT:
                # Fallback: try to extract 'content' from data struct
                # This depends on event convention.
                # Converts struct to dict
                data_dict = MessageToDict(envelope.event.data)
                content = data_dict.get("content", "")
            
            # 1. Update Context
            if content:
                self.context.add_message("user", content)
                self.log.info("processing_content", content=content[:50])

                # Broadcast Context for Inspector (Debug)
                # ... (Skipped for brevity/MVP)

                # 2. ReAct Loop
                max_turns = 3
                current_turn = 0
                final_response = None
                
                while current_turn < max_turns:
                    # Construct messages
                    messages = [{"role": m.role, "content": m.content} for m in self.context.history]
                    
                    # Generate
                    response = await self.llm_client.generate(
                        model="llama3", 
                        messages=messages,
                        system_prompt=self.config.prompt_config.get("system_prompt"),
                        tools=self.registry.definitions
                    )
                    
                    if response.tool_calls:
                        self.log.info("tool_calls_detected", count=len(response.tool_calls))
                        for tool_call in response.tool_calls:
                            result = await self.registry.execute(tool_call.name, tool_call.arguments)
                            self.context.add_message("system", f"Tool '{tool_call.name}' Output: {result}")
                        current_turn += 1
                        continue
                        
                    elif response.content:
                        final_response = response.content
                        break
                    else:
                        break

                # 3. Publish Response
                if final_response:
                    self.context.add_message("assistant", final_response)
                    
                    # Determine Output Channel
                    # If legacy config provided output lists, use first.
                    # Else default to recipient or broadcast.
                    output_channel = self.config.messaging.outputs[0] if self.config.messaging.outputs else None
                    
                    # We use RelayClient's semantic sending
                    # If output_channel represents a specific topic, we might need manual publish?
                    # RelayClient.send_text publishes to:
                    # swarm.team.{team}.agent.{id}.output
                    # AND handles target_team/recipient routing.
                    
                    # If legacy config expects a specific NATS subject like "chat.reply",
                    # RelayClient semantic routing might not match.
                    # Migration Strategy: Use semantic routing (the future).
                    # Send back to the source agent?
                    # envelope.source_agent_id
                    
                    await self.client.send_text(
                        content=final_response,
                        recipient_id=envelope.source_agent_id,
                        intent="reply",
                        context={"thread": envelope.trace_id}
                    )
                    self.log.info("reply_sent", recipient=envelope.source_agent_id)

        except Exception as e:
            self.log.error("message_handling_error", error=str(e))


async def main():
    setup_logger("runner", level=os.getenv("LOG_LEVEL", "INFO"))
    
    engine = create_async_engine(DATABASE_URL)
    async_session = sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)

    nats_url = os.getenv("NATS_URL", "nats://localhost:4222")
    runners = {}

    structlog.get_logger("runner.main").info("service_started_hybrid_mode")

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
                        # Start new runner with Relay
                        runner = AgentRunner(config, nats_url)
                        await runner.start() # Non-blocking now
                        runners[config.name] = runner
                    else:
                        # Check config diff
                        current_runner = runners[config.name]
                        if current_runner.config.model_dump() != config.model_dump():
                             await current_runner.stop()
                             new_runner = AgentRunner(config, nats_url)
                             await new_runner.start()
                             runners[config.name] = new_runner
                
                # Stop removed
                for name in list(runners.keys()):
                    if name not in active_agent_names:
                        await runners[name].stop()
                        del runners[name]

        except Exception as e:
            structlog.get_logger("runner.watcher").error("watcher_loop_error", error=str(e))
        
        await asyncio.sleep(5)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        pass
