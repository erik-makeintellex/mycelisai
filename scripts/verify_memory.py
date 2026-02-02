
import asyncio
import json
import nats
import aiohttp
import time
from datetime import datetime
import sys
import os

# Add SDK to path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../sdk/python/src')))

# Try importing protobufs. If this fails, the SDK might not be generated or paths are wrong.
try:
    from relay.proto.swarm.v1 import swarm_pb2
    from google.protobuf.timestamp_pb2 import Timestamp
    from google.protobuf.struct_pb2 import Struct
except ImportError:
    print("❌ Failed to import Protobuf definitions. Ensure 'sdk/python' is built/generated.")
    sys.exit(1)

# Configuration
NATS_URL = "nats://localhost:4222"
API_URL = "http://localhost:8080/api/v1/memory/stream"

async def main():
    print(f"⚡ Connecting to NATS at {NATS_URL}...")
    try:
        nc = await nats.connect(NATS_URL)
    except Exception as e:
        print(f"❌ Failed to connect to NATS: {e}")
        print("   (Ensure 'inv k8s.bridge' is running!)")
        return

    # Create Proto Envelope
    envelope = swarm_pb2.MsgEnvelope()
    envelope.id = f"test-{int(time.time())}"
    envelope.source_agent_id = "verifier"
    envelope.trace_id = f"trace-{int(time.time())}"
    envelope.type = swarm_pb2.MESSAGE_TYPE_EVENT
    envelope.team_id = "verification"
    
    # Timestamp
    ts = Timestamp()
    ts.GetCurrentTime()
    envelope.timestamp.CopyFrom(ts)
    
    # Payload
    event_payload = swarm_pb2.EventPayload()
    event_payload.event_type = "verification.memory.test"
    
    # Data Struct
    data_struct = Struct()
    data_struct.update({"message": "I remember."})
    event_payload.data.CopyFrom(data_struct)
    
    envelope.event.CopyFrom(event_payload)
    
    # Serialize
    data = envelope.SerializeToString()
    
    # Publish
    subject = "swarm.agent.verifier.output"
    await nc.publish(subject, data)
    print(f"🚀 Published Event to {subject}")
    
    await nc.flush()
    await nc.close()
    
    # Wait for processing
    print("⏳ Waiting for persistence (2s)...")
    await asyncio.sleep(2)
    
    # Check Read API
    async with aiohttp.ClientSession() as session:
        try:
            async with session.get(API_URL) as response:
                if response.status == 200:
                    data = await response.json()
                    print(f"✅ API Reachable. Log Count: {len(data)}")
                    found = False
                    for log in data:
                        # Check for our message
                        # JSON keys mismatch? Archivist.LogEntry uses JSON tags?
                        # LogEntry: {trace_id, timestamp, level, source, intent, message, context}
                        if log.get("message") == "Event: verification.memory.test":
                             found = True
                             print(f"🎉 FOUND VERIFICATION LOG: {log['trace_id']}")
                             break
                    
                    if not found:
                        print("⚠️  Log not found in recent stream (Async delay?). Dumping last log:")
                        if data:
                            print(json.dumps(data[0], indent=2))
                else:
                    print(f"❌ API Error: {response.status}")
        except Exception as e:
             print(f"❌ API Connect Fail: {e}")

if __name__ == "__main__":
    asyncio.run(main())
