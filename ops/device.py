from invoke import task, Collection
import asyncio
import json
import os

@task
def boot(c, id="ghost-01", capabilities="gpio,camera"):
    """
    Simulate a device boot by publishing an announcement to NATS.
    Usage: inv device.boot --id=drone-alpha --capabilities=motor,gps
    """
    try:
        from nats.aio.client import Client as NATS
    except ImportError:
        print("Error: 'nats-py' is required. Run 'uv pip install nats-py'")
        return

    async def run():
        nc = NATS()
        nats_url = os.getenv("NATS_URL", "nats://localhost:4222")
        print(f"Connecting to NATS at {nats_url}...")
        
        try:
            await nc.connect(nats_url)
        except Exception as e:
            print(f"Failed to connect to NATS: {e}")
            return

        caps_list = [c.strip() for c in capabilities.split(",")]
        payload = {
            "id": id,
            "capabilities": caps_list
        }
        
        subject = "swarm.bootstrap.announce"
        await nc.publish(subject, json.dumps(payload).encode())
        print(f"Device '{id}' announced on '{subject}'")
        print(f"   Capabilities: {caps_list}")
        
        await nc.drain()

    asyncio.run(run())

ns = Collection("device")
ns.add_task(boot)
