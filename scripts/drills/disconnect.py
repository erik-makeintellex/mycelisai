
import asyncio
import logging
import sys
import os
import time

# Verify SDK Path
SDK_PATH = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../sdk/python/src'))
sys.path.append(SDK_PATH)

from relay.client import RelayClient

# Configure Logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("chaos.drill")

async def run_drill():
    logger.info("🔥 Igniting Chaos Drill...")
    
    client = RelayClient(agent_id="chaos-agent", team_id="red-team")
    
    # 1. Ignition
    try:
        await client.connect()
    except Exception as e:
        logger.error(f"Ignition Failed: {e}. Ensure Bridge is OPEN initially.")
        return

    logger.info("✅ Connected. Initiating Pulse...")
    
    # 2. Impulse Loop
    for i in range(60):
        try:
            msg = f"Pulse #{i}"
            # We use send_text which wraps _publish -> save_impulse
            await client.send_text(msg)
            # logger.info(f"Emitted {msg}") # Too spammy, rely on client logs
        except Exception as e:
            logger.error(f"Unexpected Error: {e}")
        
        await asyncio.sleep(0.5) # 30 seconds total run time
        
        if i == 10:
             logger.info("🔻 >>> SEVER CONNECTION NOW <<< (Kill Bridge)")
        if i == 40:
             logger.info("🔺 >>> RESTORE CONNECTION NOW <<< (Start Bridge)")

    await client.close()
    logger.info("🏁 Drill Complete.")

if __name__ == "__main__":
    try:
        asyncio.run(run_drill())
    except KeyboardInterrupt:
        pass
