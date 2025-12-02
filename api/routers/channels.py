from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import StreamingResponse
from shared.schemas import Channel, ChannelCreate, EventMessage
from nats.js import JetStreamContext
import asyncio

router = APIRouter(tags=["channels"])

@router.get("/channels")
async def list_channels(request: Request):
    """List active NATS streams/channels."""
    js: JetStreamContext = request.app.state.js
    channels = []
    try:
        # List all streams
        streams = await js.streams_info()
        for stream in streams:
            # For each stream, create a Channel object
            # This is a simplification; a stream can have multiple subjects
            for subject in stream.config.subjects:
                channels.append(Channel(
                    name=stream.config.name,
                    subject=subject,
                    stream=stream.config.name,
                    description=stream.config.description,
                    created_at=stream.created
                ))
    except Exception as e:
        # If no streams or error, return empty list or handle gracefully
        print(f"Error listing streams: {e}")
        pass
        
    return channels

@router.post("/channels")
async def create_channel(channel: ChannelCreate, request: Request):
    """Create a new NATS stream/channel."""
    js: JetStreamContext = request.app.state.js
    try:
        # Create a stream with the given name and subject
        # We use the name as the stream name
        await js.add_stream(name=channel.name, subjects=[channel.subject], description=channel.description)
        
        return Channel(
            name=channel.name,
            subject=channel.subject,
            stream=channel.name,
            description=channel.description
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to create channel: {str(e)}")

@router.post("/ingest/{channel}")
async def ingest_event(channel: str, event: EventMessage, request: Request):
    js: JetStreamContext = request.app.state.js
    try:
        # Ensure the stream exists (simple auto-provisioning for demo)
        stream_name = channel.replace(".", "-")
        try:
            await js.add_stream(name=stream_name, subjects=[f"{channel}.>"])
        except Exception:
            pass # Stream might already exist

        await js.publish(f"{channel}.event", event.model_dump_json().encode())
        return {"status": "published", "id": event.id}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/stream/{channel}")
async def stream_events(channel: str, request: Request):
    async def event_generator():
        queue = asyncio.Queue()
        
        async def cb(msg):
            print(f"DEBUG: cb received message: {msg.data.decode()[:50]}...")
            await queue.put(msg)

        # Subscribe to NATS (wildcard for all events in channel)
        # We use the shared NATS connection from app.state
        try:
            # Ensure stream exists to avoid subscription errors
            stream_name = channel.replace(".", "-")
            try:
                await request.app.state.js.add_stream(name=stream_name, subjects=[f"{channel}.>"])
                print(f"DEBUG: Auto-provisioned stream {stream_name} for channel {channel}")
            except Exception as e:
                # Stream might already exist
                print(f"DEBUG: Stream provisioning check for {channel}: {e}")
                pass

            # Use JetStream subscription
            # manual_ack=True is required to avoid double-ack errors if we ack manually
            sub = await request.app.state.js.subscribe(f"{channel}.>", cb=cb, manual_ack=True)
        except Exception as e:
            print(f"ERROR: Failed to subscribe to {channel}: {e}")
            yield f"data: error: {str(e)}\n\n"
            return

        try:
            while True:
                try:
                    # Wait for message with timeout to send heartbeat
                    msg = await asyncio.wait_for(queue.get(), timeout=15.0)
                    yield f"data: {msg.data.decode()}\n\n"
                    try:
                        await msg.ack()
                    except Exception:
                        # Ignore ack errors (e.g. already acked)
                        pass
                except asyncio.TimeoutError:
                    # Send heartbeat (comment) to keep connection alive
                    yield ": keepalive\n\n"
        except asyncio.CancelledError:
            await sub.unsubscribe()
        finally:
            # Ensure unsubscribe happens if generator exits
            try:
                await sub.unsubscribe()
            except Exception:
                pass

    return StreamingResponse(event_generator(), media_type="text/event-stream")
