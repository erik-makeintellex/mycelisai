import httpx
import asyncio

API_URL = "http://localhost/api"

AGENTS = [
    {
        "name": "a1",
        "languages": ["python", "english"],
        "prompt_config": {
            "system": "You are a helpful AI assistant tasked with executing general requests."
        },
        "role": "standard",
        "assignment": "General Assistant",
        "backend": "ollama",
        "host": "192.168.50.156",
        "capabilities": ["general"],
        "messaging": {
            "inputs": ["agent.a1.input"],
            "outputs": ["agent.a1.output"]
        }
    },
    {
        "name": "architect",
        "languages": ["python", "english"],
        "prompt_config": {
            "system": "You are a Senior Software Architect. You design systems, review code, and ensure scalability and maintainability."
        },
        "role": "architect",
        "assignment": "System Architect",
        "backend": "ollama",
        "host": "192.168.50.156",
        "capabilities": ["architecture", "code-review"],
        "messaging": {
            "inputs": ["agent.architect.input"],
            "outputs": ["agent.architect.output"]
        }
    }
]

async def register_agents():
    async with httpx.AsyncClient() as client:
        # Check API health first
        try:
            resp = await client.get(f"{API_URL}/health")
            resp.raise_for_status()
            print("API is healthy.")
        except Exception as e:
            print(f"Failed to connect to API: {e}")
            return

        for agent in AGENTS:
            print(f"Registering agent: {agent['name']}...")
            try:
                resp = await client.post(f"{API_URL}/agents/register", json=agent)
                if resp.status_code in [200, 201]:
                    print(f"Successfully registered {agent['name']}")
                else:
                    print(f"Failed to register {agent['name']}: {resp.text}")
                
                # Create Channels
                inputs = agent.get("messaging", {}).get("inputs", [])
                outputs = agent.get("messaging", {}).get("outputs", [])
                for channel in inputs + outputs:
                     # Simple logic: Name = channel with dots to dashes
                     stream_name = channel.replace(".", "-")
                     payload = {
                         "name": stream_name,
                         "subject": f"{channel}.>", # Wildcard to capture everything under it
                         "description": f"Channel for {channel}"
                     }
                     # Subject must match expectation. The error was likely due to missing stream.
                     # Actually, the agent probably subscribes to specific subjects.
                     # If I use `channel` as subject, it covers exact match.
                     # But key usage: creating a stream named X with subject Y.
                     # Let's use the safer wildcard.
                     
                     print(f"Creating channel: {channel}...")
                     try:
                        c_resp = await client.post(f"{API_URL}/channels", json=payload)
                        if c_resp.status_code in [200, 201]:
                            print(f"Created channel {channel}")
                        else:
                             # Ignore if exists (500 usually if dup)
                             print(f"Channel creation note for {channel}: {c_resp.text}")
                     except Exception as e:
                         print(f"Error creating channel {channel}: {e}")

            except Exception as e:
                print(f"Error registering {agent['name']}: {e}")

if __name__ == "__main__":
    asyncio.run(register_agents())
