import httpx
import asyncio
import uuid

API_URL = "http://localhost/api"
AGENT_NAME = f"temp-agent-{uuid.uuid4().hex[:4]}"
SESSION_ID = f"session-{AGENT_NAME}"

async def reproduce():
    async with httpx.AsyncClient() as client:
        print(f"1. Creating agent {AGENT_NAME}...")
        payload = {
            "name": AGENT_NAME,
            "languages": ["english"],
            "prompt_config": {},
            "capabilities": []
        }
        await client.post(f"{API_URL}/agents/register", json=payload)
        
        print("2. Sending a message...")
        msg_payload = {
            "id": f"msg-{uuid.uuid4()}",
            "source_agent_id": "user",
            "type": "text",
            "content": "Hello, do you remember me?"
        }
        resp = await client.post(f"{API_URL}/agents/{AGENT_NAME}/chat", json=msg_payload)
        if resp.status_code != 200:
             print(f"Error sending message: {resp.status_code} {resp.text}")
             return
        
        print("3. Checking history...")
        resp = await client.get(f"{API_URL}/conversations/{SESSION_ID}/history")
        if resp.status_code != 200:
            print(f"Error checking history: {resp.status_code} {resp.text}")
            return
        history = resp.json()
        print(f"History count: {len(history)}")
        if len(history) == 0:
            print("ERROR: History setup failed.")
            return

        print("4. Deleting Agent...")
        resp = await client.delete(f"{API_URL}/agents/{AGENT_NAME}")
        print(f"Delete status: {resp.status_code}")
        
        print("5. Checking history again (Should be empty)...")
        resp = await client.get(f"{API_URL}/conversations/{SESSION_ID}/history")
        history = resp.json()
        print(f"History count post-delete: {len(history)}")
        
        if len(history) > 0:
            print("FAIL: History persisted after deletion!")
        else:
            print("SUCCESS: History cleared.")

if __name__ == "__main__":
    asyncio.run(reproduce())
