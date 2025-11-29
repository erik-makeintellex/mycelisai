import asyncio
import httpx
import json

async def run():
    url = "http://localhost:8000/stream/chat.user.user"
    print(f"Connecting to SSE: {url}")
    
    async with httpx.AsyncClient() as client:
        async with client.stream("GET", url, timeout=None) as response:
            if response.status_code != 200:
                print(f"Failed to connect: {response.status_code}")
                return

            print("Connected! Waiting for events...")
            async for line in response.aiter_lines():
                if line:
                    print(f"Received: {line}")
                    if "keepalive" in line:
                        print("Keepalive received. Success.")
                        break
                    if line.startswith("data:"):
                        print("Data received.")

if __name__ == '__main__':
    try:
        asyncio.run(run())
    except KeyboardInterrupt:
        print("Stopping...")
    except Exception as e:
        print(f"Error: {e}")
