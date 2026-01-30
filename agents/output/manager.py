import asyncio
import logging
import os
import sys
from typing import Dict, Any

# Path hacking for local dev
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../../sdk/python/src")))

from relay.client import RelayClient
from relay.proto.swarm.v1 import swarm_pb2

# Configuration
AGENT_NAME = "display-manager-01"
TEAM_ID = "user_output"
SOURCE_URI = "swarm:display:cli"
NATS_URL = os.getenv("NATS_URL", "nats://localhost:4222")

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(name)s: %(message)s")
logger = logging.getLogger(AGENT_NAME)

async def handle_alert(envelope: swarm_pb2.MsgEnvelope):
    """
    Handles 'alert' events intended for the user.
    """
    # Extract message from event data or text payload
    message = ""
    
    if envelope.type == swarm_pb2.MESSAGE_TYPE_TEXT:
        message = envelope.text.content
    elif envelope.type == swarm_pb2.MESSAGE_TYPE_EVENT:
        if "message" in envelope.event.data:
            message = envelope.event.data["message"]
        else:
            message = f"Event: {envelope.event.event_type}"
            
    # VISUAL OUTPUT
    print(f"\n[DISPLAY] >>> ALERT: {message}\n")
    
    # Acknowledge via Log/Event?
    # For now just log internally
    logger.info(f"Distributed Alert Displayed: {message} (Source: {envelope.source_agent_id})")

async def main():
    logger.info(f"🚀 Starting {AGENT_NAME} ({SOURCE_URI})...")
    
    client = RelayClient(
        agent_id=AGENT_NAME,
        team_id=TEAM_ID,
        source_uri=SOURCE_URI,
        nats_url=NATS_URL
    )
    
    try:
        await client.connect()
        
        # Subscribe to Alerts
        # Topic: swarm.team.user_output.alert
        await client.subscribe(f"swarm.team.{TEAM_ID}.alert", handle_alert)
        logger.info(f"👀 Watching channel: swarm.team.{TEAM_ID}.alert")
        
        # Keep alive
        while True:
            await asyncio.sleep(1)
            
    except KeyboardInterrupt:
        logger.info("🛑 Stopping display...")
    except Exception as e:
        logger.error(f"❌ Error: {e}")
    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())
