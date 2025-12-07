import asyncio
import json
import logging
import os
import nats
from nats.aio.client import Client as NATS
from datetime import datetime

# Import schemas (assuming running from project root)
import sys
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
from shared.schemas import ToolCallMessage, ToolResultMessage, MessageType

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class MCPBridge:
    def __init__(self, nats_url: str = "nats://localhost:4222"):
        self.nats_url = nats_url
        self.nc = None
        self.js = None

    async def connect(self):
        self.nc = await nats.connect(self.nats_url)
        self.js = self.nc.jetstream()
        logger.info(f"MCP Bridge connected to NATS at {self.nats_url}")

    async def start(self):
        await self.connect()
        
        # Auto-provision MCP stream
        try:
            await self.js.add_stream(name="mcp", subjects=["mcp.call.>", "mcp.result.>"])
            logger.info("Provisioned 'mcp' stream")
        except Exception as e:
            logger.warning(f"Stream provisioning check: {e}")

        # Subscribe to all tool calls
        await self.js.subscribe("mcp.call.>", cb=self.handle_tool_call)
        logger.info("MCP Bridge listening on mcp.call.>")

        # Keep alive
        while True:
            await asyncio.sleep(1)

    async def handle_tool_call(self, msg):
        try:
            data = json.loads(msg.data.decode())
            tool_call = ToolCallMessage(**data)
            logger.info(f"Received tool call: {tool_call.tool_name} [{tool_call.call_id}]")

            # Execute Tool (Mocking MQTT for now)
            result = await self.execute_tool(tool_call.tool_name, tool_call.arguments)

            # Publish Result
            response = ToolResultMessage(
                id=f"res-{datetime.utcnow().timestamp()}",
                source_agent_id="mcp-bridge",
                call_id=tool_call.call_id,
                result=result,
                is_error=False
            )
            
            # Publish to a result subject (e.g., mcp.result.<call_id>)
            # Agents should subscribe to mcp.result.<call_id> or mcp.result.<agent_id>
            # For simplicity, let's publish to mcp.result.<source_agent_id>
            subject = f"mcp.result.{tool_call.source_agent_id}"
            await self.js.publish(subject, response.model_dump_json().encode())
            logger.info(f"Published result to {subject}")

            await msg.ack()

        except Exception as e:
            logger.error(f"Error handling tool call: {e}")
            # Optionally publish error result

    async def execute_tool(self, name: str, args: dict):
        """
        Execute tool via local MCP Server (stdio).
        """
        try:
            # Spawn the MCP server process (lazy loading or persistent?)
            # For simplicity, we'll spawn a new one for each call (inefficient but safe)
            # Or better, keep a persistent connection. Let's spawn for now to avoid complex state management.
            
            cmd = [sys.executable, "services/mqtt_mcp.py"]
            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # Construct JSON-RPC Request
            request = {
                "jsonrpc": "2.0",
                "id": 1,
                "method": "tools/call",
                "params": {
                    "name": name,
                    "arguments": args
                }
            }

            stdout, stderr = await process.communicate(input=json.dumps(request).encode() + b"\n")
            
            if stderr:
                logger.info(f"MCP Server Log: {stderr.decode()}")

            if stdout:
                response = json.loads(stdout.decode())
                if "error" in response:
                    return {"status": "error", "message": response["error"]["message"]}
                
                # Extract text content
                content = response["result"]["content"][0]["text"]
                return {"status": "success", "message": content}
            
            return {"status": "error", "message": "No response from MCP Server"}

        except Exception as e:
            logger.error(f"MCP Execution Error: {e}")
            return {"status": "error", "message": str(e)}

if __name__ == "__main__":
    nats_url = os.getenv("NATS_URL", "nats://localhost:4222")
    bridge = MCPBridge(nats_url=nats_url)
    asyncio.run(bridge.start())
