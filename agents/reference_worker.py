import asyncio
import logging
import os
import sys
from typing import Dict, Any

# Ensure we can import the relay sdk
# In production, this would be installed via pip/uv
# For this script, we assume PYTHONPATH is set, or we hack it
sys.path.append(os.path.join(os.getcwd(), "sdk/python/src"))

from relay.client import RelayClient
from relay.proto.swarm.v1 import swarm_pb2

# Configuration
AGENT_NAME = "reference-worker-01"
TEAM_ID = "marketing"
SOURCE_URI = "swarm:reference" # The "Standard" implementation
NATS_URL = os.getenv("NATS_URL", "nats://localhost:4222")

# Logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s"
)
logger = logging.getLogger(AGENT_NAME)

async def handle_marketing_request(envelope: swarm_pb2.MsgEnvelope):
    """
    Handles 'marketing.request' events.
    In a real agent, this might call an LLM.
    """
    logger.info(f"📨 Received Task: {envelope.event.event_type}")
    
    # Simulate work
    await asyncio.sleep(1)
    
    # Create Result Data
    result_data = {
        "status": "success",
        "generated_content": "This is a reference marketing copy.",
        "original_task_id": envelope.id
    }
    
    # We ideally emit back to the source or a result topic
    # For now, we emit a 'task.complete' event
    logger.info("✅ Task Complete. Sending result...")
    
    client = get_global_client() # Helper to get client instance
    if client:
        await client.send_event(
            event_type="task.complete",
            data=result_data,
            context={"processor": AGENT_NAME, "version": "1.0"},
            target_team=envelope.team_id # Reply to the team?
        )

# Global helper for callback access (simplified for demo)
_client: RelayClient = None

def get_global_client():
    return _client

async def main():
    global _client
    
    logger.info(f"🚀 Starting {AGENT_NAME}...")
    
    # 1. Initialize Relay
    client = RelayClient(
        agent_id=AGENT_NAME,
        team_id=TEAM_ID,
        source_uri=SOURCE_URI,
        nats_url=NATS_URL
    )
    _client = client
    
    try:
        # 2. Connect
        await client.connect()
        
        # 3. Subscribe to Team Tasks
        # Topic: swarm.team.marketing.request (> wildcard for sub-topics)
        topic = f"swarm.team.{TEAM_ID}.request.>"
        await client.subscribe(topic, handle_marketing_request)
        logger.info(f"👂 Listening on {topic}")
        
        # 4. Subscribe to Direct Instructions
        await client.subscribe(f"swarm.agent.{AGENT_NAME}.input", handle_marketing_request)

        # 5. Heartbeat Loop
        while True:
            # Send Heartbeat (LogEntry simulation via Event)
            # In Phase 3, this would be a specific LogEntry message type if supported,
            # currently mapping to an event for observability.
            await client.send_event(
                event_type="agent.heartbeat",
                data={"status": "alive", "load": 0.1},
                context={"uptime": "forever"}
            )
            logger.info("💓 Heartbeat sent")
            await asyncio.sleep(5)
            
    except KeyboardInterrupt:
        logger.info("🛑 Stopping agent...")
    except Exception as e:
        logger.error(f"❌ Error: {e}")
    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())
