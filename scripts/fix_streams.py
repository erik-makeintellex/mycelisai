import asyncio
import nats
from nats.js.api import StreamConfig

async def fix_streams():
    nc = await nats.connect("nats://localhost:4222")
    js = nc.jetstream()

    # List streams
    streams = await js.streams_info()
    for stream in streams:
        print(f"Checking stream: {stream.config.name}")
        # Logic: If name matches agent-*-input/output, ensure subjects are correct
        # Our naming convention in previous script was replace(".", "-")
        # e.g. agent.a1.input -> agent-a1-input
        
        # We want to support the original dot-notation as exact subject, plus wildcard
        # Infer original channel name from stream name? 
        # approximate: replace "-" with "."? 
        # No, "agent-a1-input" -> "agent.a1.input" works.
        
        if stream.config.name.startswith("agent-"):
            derived_subject = stream.config.name.replace("-", ".")
            
            # Desired subjects: [derived_subject, derived_subject + ".>"]
            desired_subjects = [derived_subject, f"{derived_subject}.>"]
            
            # Check if current matches
            current_subjects = stream.config.subjects
            # If not exact match set-wise
            if set(current_subjects) != set(desired_subjects):
                print(f"Updating {stream.config.name}: {current_subjects} -> {desired_subjects}")
                cfg = stream.config
                cfg.subjects = desired_subjects
                try:
                    await js.update_stream(cfg)
                    print(f"Successfully updated {stream.config.name}")
                except Exception as e:
                    print(f"Failed to update {stream.config.name}: {e}")
            else:
                print(f"Stream {stream.config.name} already correct.")

    # Also, we need to ensure the other channels mentioned in agent config exist.
    # chat.agent.{name} -> stream chat-agent-{name}
    # mcp.result.{name} -> stream mcp-result-{name}
    # My previous script didn't create these!
    # Agents try to subscribe:
    # await self.subscribe(f"chat.agent.{self.config.name}", ...)
    # await self.subscribe(f"mcp.result.{self.config.name}", ...)
    
    # We must create these streams too.
    agents = ["a1", "architect"]
    for agent in agents:
        channels_to_ensure = [
            f"chat.agent.{agent}",
            f"mcp.result.{agent}"
        ]
        
        for channel in channels_to_ensure:
             stream_name = channel.replace(".", "-")
             desired_subjects = [channel, f"{channel}.>"]
             
             try:
                 await js.add_stream(name=stream_name, subjects=desired_subjects)
                 print(f"Created missing stream {stream_name}")
             except Exception as e:
                 # If likely exists, try update
                 try:
                     await js.update_stream(name=stream_name, subjects=desired_subjects)
                     print(f"Updated existing stream {stream_name}")
                 except Exception as e2:
                     print(f"Could not provision {stream_name}: {e2}")

    await nc.close()

if __name__ == "__main__":
    asyncio.run(fix_streams())
