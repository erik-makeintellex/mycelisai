import asyncio
import nats
import json
import logging
from datetime import datetime

# Configuration
NATS_URL = "nats://localhost:4222"

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

async def verify_mcp():
    logger.info("Starting MCP Verification...")
    
    nc = await nats.connect(NATS_URL)
    js = nc.jetstream()
    
    # 1. Subscribe to Result
    agent_id = "verifier-agent"
    result_sub = await js.subscribe(f"mcp.result.{agent_id}")
    
    # 2. Publish Tool Call
    call_id = f"call-{datetime.utcnow().timestamp()}"
    tool_call = {
        "id": f"msg-{datetime.utcnow().timestamp()}",
        "source_agent_id": agent_id,
        "type": "text", # Using TEXT for now as per schema
        "tool_name": "mqtt_publish",
        "arguments": {
            "topic": "test/topic",
            "payload": "Hello MCP"
        },
        "call_id": call_id
    }
    
    logger.info(f"Sending Tool Call: {tool_call['tool_name']}...")
    await js.publish("mcp.call.mqtt_publish", json.dumps(tool_call).encode())
    
    # 3. Wait for Result
    try:
        msg = await result_sub.next_msg(timeout=5)
        data = json.loads(msg.data.decode())
        
        logger.info(f"Received Result: {data}")
        
        if data.get("call_id") == call_id and not data.get("is_error"):
            logger.info("✅ MCP Tool Execution Successful")
            if "Successfully published" in str(data.get("result")):
                 logger.info("✅ Output matches expected response")
            else:
                 logger.warning("⚠️ Output format unexpected")
        else:
            logger.error(f"❌ MCP Execution Failed: {data}")
            
    except asyncio.TimeoutError:
        logger.error("❌ Timeout waiting for MCP result")
    except Exception as e:
        logger.error(f"❌ Error: {e}")
        
    await nc.close()

if __name__ == "__main__":
    asyncio.run(verify_mcp())
