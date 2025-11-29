import asyncio
import argparse
import json
import nats
from datetime import datetime

async def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("channel", help="Channel to publish to")
    parser.add_argument("message", help="Message content")
    parser.add_argument("--sender", default="user", help="Sender name")
    args = parser.parse_args()

    nc = await nats.connect("nats://localhost:4222")
    js = nc.jetstream()

    msg = {
        "id": f"msg-{datetime.utcnow().timestamp()}",
        "timestamp": datetime.utcnow().isoformat(),
        "source_agent_id": args.sender,
        "type": "text",
        "sender": args.sender,
        "recipient": "all",
        "content": args.message,
        "intent": "inform"
    }

    await js.publish(args.channel, json.dumps(msg).encode())
    print(f"Published to {args.channel}: {args.message}")

    await nc.close()

if __name__ == "__main__":
    asyncio.run(main())
