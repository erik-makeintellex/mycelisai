from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import StreamingResponse
from shared.schemas import Channel, ChannelCreate, EventMessage
from nats.js import JetStreamContext
import asyncio
from shared.logger import get_logger
import json
import uuid
from datetime import datetime


log = get_logger("api.routers.channels")

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
        log.error("list_streams_error", error=str(e))
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
        # Immediate visible ping to flush headers/proxy buffers
        yield f"data: {json.dumps({'type': 'system', 'status': 'connected', 'id': 'ping-' + uuid.uuid4().hex})}\n\n"
        
        async def persist_reply(data):
            try:
                # Derive conversation_id from source agent (Simple Isolation)
                # Ideally this comes from the message payload itself if part of a specific thread
                source_agent = data.get("source_agent_id", "agent")
                conversation_id = data.get("conversation_id", f"session-{source_agent}")
                 
                async for session in get_db():
                     db_msg = MessageDB(
                         id=data.get("id", f"msg-{uuid.uuid4().hex}"),
                         conversation_id=conversation_id,
                         sender=source_agent,
                         role="agent",
                         content=data.get("content") or data.get("payload", {}).get("content", ""),
                         type="text"
                     )
                     session.add(db_msg)
                     await session.commit()
                     break
            except Exception as e:
                # Basic idempotency: If duplicates occur (race condition from multiple clients), ignore it.
                if "IntegrityError" in str(type(e)) or "UniqueViolationError" in str(e):
                    # log.debug("message_already_persisted", id=data.get("id"))
                    pass
                else:
                    log.error("persistence_failed", error=str(e))

        async def cb(msg):
            log.info("nats_message_received", channel=channel, subject=msg.subject)
            
            # Put on queue immediately for UI
            await queue.put(msg)
            
            # Persist in background (Fire and Forget)
            try:
                data = json.loads(msg.data.decode())
                if data.get("type") == "text" or "content" in data:
                    asyncio.create_task(persist_reply(data))
            except Exception:
                pass

        # Subscribe to NATS (wildcard for all events in channel)
        # We use the shared NATS connection from app.state
        try:
            # Ensure stream exists to avoid subscription errors
            # Ensure stream exists to avoid subscription errors
            stream_name = channel.replace(".", "-")
            subjects = [channel, f"{channel}.>"]
            try:
                await request.app.state.js.add_stream(name=stream_name, subjects=subjects)
                log.info("stream_provisioned", stream=stream_name, channel=channel)
            except Exception as e:
                # Add failed (maybe exists, maybe collision)
                try:
                    await request.app.state.js.update_stream(name=stream_name, subjects=subjects)
                    log.info("stream_updated", stream=stream_name, channel=channel)
                except Exception as e2:
                    # Update failed (maybe stream name mismatch / zombie stream owns subject)
                    try:
                        # Find which stream actually owns this subject
                        actual_stream = await request.app.state.js.find_stream_name_by_subject(channel)
                        await request.app.state.js.update_stream(name=actual_stream, subjects=subjects)
                        log.info("stream_updated_via_lookup", stream=actual_stream, original_target=stream_name)
                    except Exception as e3:
                        log.warning("stream_provision_failed", error=str(e3), prev_error=str(e), stream=stream_name)

            # Use JetStream subscription
            # manual_ack=True is required to avoid double-ack errors if we ack manually
            # Subscribe to NATS (Exact Match OR Wildcard)
            # We need to catch 'task.assignment' AND 'task.assignment.updated'
            # NATS doesn't support "foo + foo.>" in one subscription easily without potentially duplicate if we just use ">" at root.
            # But we can make two subscriptions or use a subject list if supported by py-nats, 
            # or just assume the UI passes the root and we want everything under it.
            # Safest: Subscribe to both distinct subjects.
            
            sub1 = await request.app.state.js.subscribe(channel, cb=cb, manual_ack=True)
            sub2 = await request.app.state.js.subscribe(f"{channel}.>", cb=cb, manual_ack=True)
        except Exception as e:
            log.error("subscription_failed", channel=channel, error=str(e))
            yield f"data: error: {str(e)}\n\n"
            return

        try:
            while True:
                try:
                    # Wait for message with timeout to send heartbeat
                    msg = await asyncio.wait_for(queue.get(), timeout=15.0)
                    log.info("yielding_message_to_client", channel=channel, id=msg.header.get("Nats-Msg-Id", "unknown") if msg.header else "unknown")
                    yield f"data: {msg.data.decode()}\n\n"
                    try:
                        await msg.ack()
                    except Exception:
                        # Ignore ack errors (e.g. already acked)
                        pass
                except asyncio.TimeoutError:
                    # Send heartbeat to keep SSE alive
                    yield f"data: {json.dumps({'type': 'ping', 'timestamp': str(datetime.utcnow())})}\n\n"
        except asyncio.CancelledError:
            await sub1.unsubscribe()
            await sub2.unsubscribe()
            log.info("sse_client_disconnected", channel=channel)
        finally:
            # Ensure unsubscribe happens if generator exits
            try:
                await sub.unsubscribe()
            except Exception:
                pass

    return StreamingResponse(
        event_generator(), 
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "Content-Type": "text/event-stream",
            "X-Accel-Buffering": "no"
        }
    )
