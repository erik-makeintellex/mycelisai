import asyncio
import logging
import os
import sys
import random
from typing import Dict, Any

# Path hacking for local dev
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../../sdk/python/src")))

from relay.client import RelayClient
from relay.proto.swarm.v1 import swarm_pb2

# Configuration
AGENT_NAME = "sensor-manager-01"
TEAM_ID = "sensors"
SOURCE_URI = "swarm:sensor:simulated"
NATS_URL = os.getenv("NATS_URL", "nats://localhost:4222")
POLLING_RATE = 2.0

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(name)s: %(message)s")
logger = logging.getLogger(AGENT_NAME)

async def handle_config_update(envelope: swarm_pb2.MsgEnvelope):
    """Handle dynamic configuration updates."""
    global POLLING_RATE
    logger.info(f"⚙️ Config Update Received: {envelope.event.event_type}")
    
    # Check if we should change polling rate
    # Simplified logic: just check data
    if envelope.event.data:
        if "polling_rate" in envelope.event.data:
            new_rate = envelope.event.data["polling_rate"]
            POLLING_RATE = float(new_rate)
            logger.info(f"Updated Polling Rate to {POLLING_RATE}s")

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
        
        # Subscribe to configuration updates
        # Topic: swarm.team.sensors.config
        await client.subscribe(f"swarm.team.{TEAM_ID}.config", handle_config_update)
        
        # Main Telemetry Loop
        while True:
            # Simulate Reading Hardware
            temp = 45.0 + random.uniform(-1.0, 1.0)
            battery = 98.0 - (random.random() * 0.1)
            
            telemetry = {
                "temp": round(temp, 2),
                "battery": round(battery, 1)
            }
            
            # Publish as LogEntry (via Event for now as per RelayClient V1)
            # Intent: "telemetry"
            await client.send_event(
                event_type="telemetry",
                data=telemetry,
                context={"mode": "active", "sensor_id": "sim-01"}
            )
            
            logger.info(f"📡 Telemetry Sent: {telemetry}")
            await asyncio.sleep(POLLING_RATE)
            
    except KeyboardInterrupt:
        logger.info("🛑 Stopping sensors...")
    except Exception as e:
        logger.error(f"❌ Error: {e}")
    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())
