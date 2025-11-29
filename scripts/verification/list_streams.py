import asyncio
import nats
import json

async def run():
    nc = await nats.connect("nats://localhost:4222")
    js = nc.jetstream()
    
    try:
        streams = await js.streams_info()
        for stream in streams:
            print(f"Stream: {stream.config.name}")
            print(f"Subjects: {stream.config.subjects}")
    except Exception as e:
        print(f"Error listing streams: {e}")

if __name__ == '__main__':
    asyncio.run(run())
