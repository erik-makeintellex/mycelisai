import asyncio
import os
import json
import nats
from nats.errors import ConnectionClosedError, TimeoutError, NoRespondersError
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.future import select
import httpx

# Import shared schemas and db models
# We need to ensure the python path includes the root directory
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from shared.db import AgentDB, DATABASE_URL
from shared.schemas import AgentConfig

class AgentRunner:
    def __init__(self, agent_config: AgentConfig, nc):
        self.config = agent_config
        self.nc = nc
        self.js = nc.jetstream()
        self.subscription = None
        self.running = False

    async def start(self):
        if self.running:
            return
        
        self.running = True
        input_channel = self.config.messaging.inputs[0] if self.config.messaging.inputs else f"chat.agent.{self.config.name}"
        
        # Ensure input stream exists (best effort)
        try:
            stream_name = input_channel.replace(".", "-")
            await self.js.add_stream(name=stream_name, subjects=[f"{input_channel}.>"])
        except Exception:
            pass

        print(f"[{self.config.name}] Starting runner. Listening on {input_channel}.> using model {self.config.backend}")
        
        # Subscribe to input channel
        self.subscription = await self.js.subscribe(f"{input_channel}.>", cb=self.handle_message)

    async def stop(self):
        self.running = False
        if self.subscription:
            await self.subscription.unsubscribe()
            self.subscription = None
        print(f"[{self.config.name}] Stopped runner.")

    async def handle_message(self, msg):
        try:
            data = json.loads(msg.data.decode())
            
            # Extract content
            content = data.get('content')
            if not content and 'payload' in data:
                if isinstance(data['payload'], dict):
                    content = data['payload'].get('content')
                else:
                    content = str(data['payload'])
            
            if not content:
                print(f"[{self.config.name}] Received empty message.")
                return

            print(f"[{self.config.name}] Processing: {content[:50]}...")

            # Call LLM
            response_text = await self.call_llm(content)

            # Publish Response
            output_channel = self.config.messaging.outputs[0] if self.config.messaging.outputs else f"chat.agent.{self.config.name}.reply"
            
            # Ensure output stream exists
            try:
                stream_name = output_channel.replace(".", "-")
                await self.js.add_stream(name=stream_name, subjects=[f"{output_channel}.>"])
            except Exception:
                pass

            reply_msg = {
                "id": f"reply-{os.urandom(4).hex()}",
                "source_agent_id": self.config.name,
                "sender": self.config.name,
                "recipient": data.get("source_agent_id", "user"),
                "content": response_text,
                "type": "text",
                "timestamp": str(asyncio.get_event_loop().time()),
                "payload": {
                    "content": response_text
                }
            }

            await self.js.publish(f"{output_channel}.event", json.dumps(reply_msg).encode())
            print(f"[{self.config.name}] Replied to {output_channel}.event")

        except Exception as e:
            print(f"[{self.config.name}] Error handling message: {e}")

    async def call_llm(self, prompt):
        # Determine backend
        model_id = self.config.backend
        system_prompt = self.config.prompt_config.get("system_prompt", "You are a helpful assistant.")
        
        # Default to Ollama for now
        ollama_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        
        try:
            async with httpx.AsyncClient() as client:
                payload = {
                    "model": model_id,
                    "messages": [
                        {"role": "system", "content": system_prompt},
                        {"role": "user", "content": prompt}
                    ],
                    "stream": False
                }
                
                resp = await client.post(f"{ollama_url}/api/chat", json=payload, timeout=60.0)
                if resp.status_code == 200:
                    result = resp.json()
                    return result.get("message", {}).get("content", "")
                else:
                    return f"Error from LLM: {resp.status_code} - {resp.text}"
        except Exception as e:
            return f"LLM Call Failed: {str(e)}"

async def main():
    # Database Connection
    engine = create_async_engine(DATABASE_URL)
    async_session = sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)

    # NATS Connection
    nc = await nats.connect(os.getenv("NATS_URL", "nats://localhost:4222"))
    
    runners = {}

    print("Runner Service Started. Watching for agents...")
    print(f"Configuration: OLLAMA_BASE_URL={os.getenv('OLLAMA_BASE_URL', 'http://localhost:11434')}")

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
                        # Check if config updated? (Simplified: ignore updates for now)
                        pass
                
                # Stop removed agents
                for name in list(runners.keys()):
                    if name not in active_agent_names:
                        await runners[name].stop()
                        del runners[name]

        except Exception as e:
            print(f"Error in watcher loop: {e}")
        
        await asyncio.sleep(5)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        pass
