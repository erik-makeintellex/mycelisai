import asyncio
import nats
import json
from datetime import datetime

async def run():
    nc = await nats.connect("nats://localhost:4222")
    js = nc.jetstream()
    
    agent_name = "researcher-v1"
    input_channel = "research.request"
    output_channel = "research.summary"

    print(f"Starting {agent_name}...")
    print(f"Listening on: {input_channel}.>")
    print(f"Replying to: {output_channel}")

    # Ensure output stream exists
    try:
        stream_name = output_channel.replace(".", "-")
        await js.add_stream(name=stream_name, subjects=[f"{output_channel}.>"])
        print(f"Stream '{stream_name}' ensured.")
    except Exception:
        pass

    async def message_handler(msg):
        print(f"\n[Agent] Received on {msg.subject}:")
        data = json.loads(msg.data.decode())
        
        # Extract content
        content = data.get('content')
        if not content and 'payload' in data:
            if isinstance(data['payload'], dict):
                content = data['payload'].get('content')
            else:
                content = str(data['payload'])
        
        print(f"Content: {content}")
        
        # Simulate work
        await asyncio.sleep(1)
        
        # Reply
        reply_content = f"Research Results for: {content}\n\n1. Source A confirms '{content}' is interesting.\n2. Source B agrees."
        
        reply_msg = {
            "id": f"reply-{datetime.utcnow().timestamp()}",
            "source_agent_id": agent_name,
            "sender": agent_name,
            "recipient": "user",
            "content": reply_content,
            "type": "text",
            "timestamp": datetime.utcnow().isoformat(),
            "payload": { # Also include in payload for consistency
                "content": reply_content
            }
        }
        
        await js.publish(f"{output_channel}.event", json.dumps(reply_msg).encode())
        print(f"[Agent] Sent reply to {output_channel}.event")

    # Subscribe
    await js.subscribe(f"{input_channel}.>", cb=message_handler)
    
    # Keep running
    while True:
        await asyncio.sleep(1)

if __name__ == '__main__':
    try:
        asyncio.run(run())
    except KeyboardInterrupt:
        print("Stopping...")
