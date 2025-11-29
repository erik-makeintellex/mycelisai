import asyncio
import aiohttp
import nats
import json
import logging
from datetime import datetime

# Configuration
API_URL = "http://localhost:8000"
NATS_URL = "nats://localhost:4222"
TEST_CHANNEL = "test.channel"

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

async def check_api():
    logger.info(f"Checking API at {API_URL}...")
    try:
        async with aiohttp.ClientSession() as session:
            # Check Health (if endpoint exists) or Root
            async with session.get(f"{API_URL}/") as response:
                if response.status == 200:
                    logger.info("✅ API Root is accessible")
                else:
                    logger.error(f"❌ API Root returned {response.status}")
                    return False
            
            # Check Agents List
            async with session.get(f"{API_URL}/agents") as response:
                if response.status == 200:
                    data = await response.json()
                    logger.info(f"✅ API Agents Endpoint accessible. Found {len(data)} agents.")
                else:
                    logger.error(f"❌ API Agents Endpoint returned {response.status}")
                    return False
        return True
    except Exception as e:
        logger.error(f"❌ API Check Failed: {e}")
        return False

async def check_nats_and_flow():
    logger.info(f"Checking NATS at {NATS_URL}...")
    try:
        nc = await nats.connect(NATS_URL)
        js = nc.jetstream()
        logger.info("✅ NATS Connection Successful")

        # 1. Create Stream (if not exists)
        try:
            await js.add_stream(name="test-stream", subjects=[f"{TEST_CHANNEL}.>"])
            logger.info("✅ Test Stream created/verified")
        except Exception as e:
            logger.warning(f"Stream creation note: {e}")

        # 2. Publish Message via API (Ingest)
        logger.info("Testing Ingest API -> NATS Flow...")
        
        # Construct valid EventMessage
        payload = {
            "id": f"evt-{datetime.utcnow().timestamp()}",
            "source_agent_id": "verifier",
            "type": "event",
            "payload": {
                "content": "Hello World"
            },
            "timestamp": datetime.utcnow().isoformat()
        }
        
        async with aiohttp.ClientSession() as session:
            async with session.post(f"{API_URL}/ingest/{TEST_CHANNEL}", json=payload) as response:
                if response.status == 200:
                    logger.info("✅ Ingest API accepted message")
                else:
                    text = await response.text()
                    logger.error(f"❌ Ingest API failed: {response.status} - {text}")
                    await nc.close()
                    return False

        # 3. Verify Message in NATS
        # We'll subscribe and wait for a message (or just check stream info)
        sub = await js.subscribe(f"{TEST_CHANNEL}.>")
        try:
            msg = await sub.next_msg(timeout=2)
            data = json.loads(msg.data.decode())
            if data.get("payload", {}).get("content") == "Hello World":
                logger.info("✅ Message received from NATS (End-to-End Flow Confirmed)")
            else:
                logger.error(f"❌ Message content mismatch: {data}")
                await nc.close()
                return False
        except Exception as e:
            logger.error(f"❌ Failed to receive message from NATS: {e}")
            await nc.close()
            return False

        await nc.close()
        return True

    except Exception as e:
        logger.error(f"❌ NATS Check Failed: {e}")
        return False

async def main():
    logger.info("Starting System Verification...")
    
    api_ok = await check_api()
    if not api_ok:
        logger.error("API Verification Failed. Aborting.")
        return

    nats_ok = await check_nats_and_flow()
    if not nats_ok:
        logger.error("NATS/Flow Verification Failed.")
        return

    logger.info("🎉 ALL SYSTEMS GO! Stack is functioning correctly.")

if __name__ == "__main__":
    asyncio.run(main())
