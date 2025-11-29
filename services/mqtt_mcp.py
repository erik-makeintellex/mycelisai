import sys
import json
import logging

# Configure logging to stderr so it doesn't interfere with stdout JSON-RPC
logging.basicConfig(stream=sys.stderr, level=logging.INFO, format='[MQTT-MCP] %(message)s')
logger = logging.getLogger(__name__)

TOOLS = [
    {
        "name": "mqtt_publish",
        "description": "Publish a message to an MQTT topic",
        "inputSchema": {
            "type": "object",
            "properties": {
                "topic": {"type": "string"},
                "payload": {"type": "string"}
            },
            "required": ["topic", "payload"]
        }
    }
]

def handle_request(request):
    method = request.get("method")
    params = request.get("params", {})
    req_id = request.get("id")

    if method == "tools/list":
        return {
            "jsonrpc": "2.0",
            "id": req_id,
            "result": {
                "tools": TOOLS
            }
        }
    
    if method == "tools/call":
        name = params.get("name")
        args = params.get("arguments", {})
        
        if name == "mqtt_publish":
            topic = args.get("topic")
            payload = args.get("payload")
            
            # In a real implementation, we would use paho-mqtt here.
            # For this demo, we'll just log it and pretend.
            logger.info(f"PUBLISHING: {topic} -> {payload}")
            
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {
                    "content": [
                        {
                            "type": "text",
                            "text": f"Successfully published to {topic}"
                        }
                    ]
                }
            }
        
        return {
            "jsonrpc": "2.0",
            "id": req_id,
            "error": {"code": -32601, "message": "Method not found"}
        }

    return None

def main():
    logger.info("Starting MQTT MCP Server (stdio)...")
    for line in sys.stdin:
        try:
            request = json.loads(line)
            response = handle_request(request)
            if response:
                print(json.dumps(response))
                sys.stdout.flush()
        except Exception as e:
            logger.error(f"Error processing request: {e}")

if __name__ == "__main__":
    main()
