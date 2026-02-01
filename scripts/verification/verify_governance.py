import asyncio
import aiohttp
import json
import logging
import sys
import os
import random

# Path setup for SDK
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../../sdk/python/src")))

from relay.client import RelayClient
from relay.proto.swarm.v1 import swarm_pb2

# Config
CORE_API = "http://localhost:8080"
AGENT_NAME = "payment-processor" # Matches policy target
TEAM_ID = "marketing" # Matches policy target
NATS_URL = "nats://127.0.0.1:4222"

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("gov-test")

async def main():
    logger.info("🚀 Starting Governance Verification Smoke Test")

    # 1. Setup Agent
    client = RelayClient(agent_id=AGENT_NAME, team_id=TEAM_ID, nats_url=NATS_URL)
    received_approved_msg = asyncio.Event()

    async def on_message(msg):
        # We expect to receive our own message back after approval
        # Or a system notification.
        logger.info(f"📨 Received Message: {msg.id} Type: {msg.type}")
        
        # Check if it is our approved transaction
        if msg.event and msg.event.event_type == "process_payment": # Matched intent
            logger.info("✅ Verified: Payment Event received back from bus!")
            received_approved_msg.set()

    try:
        await client.connect()
        
        # Subscribe to our own output channel to see the re-published message
        sub_topic = f"swarm.team.{TEAM_ID}.agent.{AGENT_NAME}.output"
        await client.subscribe(sub_topic, on_message)
        logger.info(f"👂 Listening on {sub_topic}")

        # 2. Trigger High-Risk Event
        amount = 150
        logger.info(f"💸 Triggering High-Value Payment (Amount: {amount})...")
        
        await client.send_event(
            event_type="process_payment", # Matches policy
            data={"amount": amount, "currency": "USD"},
            context={"mode": "test"},
            target_team="bank" 
        )
        
        logger.info("⏳ Waiting for Gatekeeper interception (2s)...")
        await asyncio.sleep(2)
        
        # If intercepted, we should NOT have received it yet.
        if received_approved_msg.is_set():
            logger.error("❌ FAILURE: Message was NOT intercepted! Check Policy.")
            sys.exit(1)

        # 3. Check Pending
        logger.info("🔍 Checking Admin API for pending requests...")
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{CORE_API}/admin/approvals") as resp:
                if resp.status != 200:
                    logger.error(f"API Error: {resp.status}")
                    sys.exit(1)
                
                pending = await resp.json()
                logger.info(f"📋 Pending Requests: {len(pending)}")
                
                target_req_id = None
                for req in pending:
                    # Check if this is our request (can check reason or original msg)
                    # For simplicity, take the first one or matches trace?
                    # We don't have easy trace match here, just assume it's the one.
                    req_id = req.get("request_id")
                    logger.info(f"   - Found Request: {req_id} (Reason: {req.get('reason')})")
                    target_req_id = req_id
                
                if not target_req_id:
                    logger.error("❌ FAILURE: No pending request found in Gatekeeper.")
                    sys.exit(1)

                # 4. Approve
                logger.info(f"👮 Approving Request {target_req_id}...")
                approve_payload = {"action": "APPROVE"} 
                # Note: Core implementation expects {"action": "APPROVE"}
                
                async with session.post(f"{CORE_API}/admin/approvals/{target_req_id}", json=approve_payload) as app_resp:
                    if app_resp.status != 200:
                        logger.error(f"Approval Failed: {app_resp.status}")
                        sys.exit(1)
                    logger.info("✅ Approval Sent.")

        # 5. Wait for Execution
        logger.info("⏳ Waiting for message release...")
        try:
            await asyncio.wait_for(received_approved_msg.wait(), timeout=5.0)
            logger.info("🎉 SUCCESS: Governance Loop Verified!")
        except asyncio.TimeoutError:
            logger.error("❌ FAILURE: Approved message was not received back on the bus.")
            sys.exit(1)

    except Exception as e:
        logger.error(f"❌ Exception: {e}")
        sys.exit(1)
    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())
