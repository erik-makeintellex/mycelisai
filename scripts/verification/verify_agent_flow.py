import asyncio
import nats
from nats.errors import ConnectionClosedError, TimeoutError, NoRespondersError
import json
import os
from datetime import datetime

async def run():
    # Connect to NATS
    nc = await nats.connect("nats://localhost:4222")
    js = nc.jetstream()

    print("Connected to NATS")

    # Ensure stream exists
    try:
        # Use a specific stream for agent chat to avoid overlap with chat-user-user
        await js.add_stream(name="chat-agent", subjects=["chat.agent.>"])
        print("Stream 'chat-agent' created/ensured.")
    except Exception as e:
        print(f"Stream creation warning: {e}")
        try:
            await js.update_stream(name="chat-agent", subjects=["chat.agent.>"])
            print("Stream 'chat-agent' updated.")
        except Exception as e2:
             print(f"Stream update failed: {e2}")

    # Ensure reply stream exists
    try:
        await js.add_stream(name="chat-user-user", subjects=["chat.user.user.>"])
        print("Stream 'chat-user-user' created/ensured.")
    except Exception:
        pass

    # Subscribe to agent topic (wildcard to catch .event suffix from ingest)
    agent_name = "test-agent"
    subject = f"chat.agent.{agent_name}.>"
    
    async def message_handler(msg):
        print(f"\n[Agent] Received message on {msg.subject}:")
        data = json.loads(msg.data.decode())
        print(json.dumps(data, indent=2))
        
        # Extract content from payload if present (EventMessage) or direct (TextMessage)
        content = data.get('content')
        if not content and 'payload' in data:
            if isinstance(data['payload'], dict):
                content = data['payload'].get('content')
            else:
                content = str(data['payload'])
        
        if not content:
            content = "no content"

        # Simulate processing
        print("[Agent] Processing...")
        await asyncio.sleep(1)
        
        # Reply
        reply_content = f"Echo: {content}"
        reply_msg = {
            "id": f"reply-{datetime.utcnow().timestamp()}",
            "source_agent_id": agent_name,
            "sender": agent_name,
            "recipient": "user",
            "content": reply_content,
            "type": "text",
            "timestamp": datetime.utcnow().isoformat()
        }
        
        reply_subject = "chat.user.user"
        await js.publish(reply_subject, json.dumps(reply_msg).encode())
        print(f"[Agent] Sent reply to {reply_subject}: {reply_content}")

    # Create consumer
    print(f"Subscribing to {subject}...")
    sub = await js.subscribe(subject, cb=message_handler)

    print("Mock Agent Running. Waiting for messages... (Press Ctrl+C to stop)")
    
    # Keep running
    while True:
        await asyncio.sleep(1)

if __name__ == '__main__':
    try:
        asyncio.run(run())
    except KeyboardInterrupt:
        print("Stopping...")
