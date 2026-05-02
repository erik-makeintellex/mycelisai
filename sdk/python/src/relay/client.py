import asyncio
import uuid
import time
import logging
from typing import Any, Callable

import nats
from nats.errors import ConnectionClosedError, TimeoutError

from google.protobuf.timestamp_pb2 import Timestamp
from google.protobuf.struct_pb2 import Struct

from .proto.swarm.v1 import swarm_pb2

from . import persistence


logger = logging.getLogger("relay.client")

class RelayClient:
    """
    The RelayClient bridges Python agents/scripts to the Mycelis Swarm.
    It enforces swarm.proto schemas (v1) and manages the NATS connection.
    
    Features:
    - Team Routing (swarm.team.{id})
    - Context Propagation (AG2 Pattern)
    - Auto-Subscription to Inbox
    """
    def __init__(self, 
                 agent_id: str, 
                 team_id: str, 
                 source_uri: str = "swarm:base", 
                 nats_url: str = "nats://localhost:4222"):
        
        self.agent_id = agent_id or f"python-relay-{uuid.uuid4().hex[:8]}"
        self.team_id = team_id
        self.source_uri = source_uri
        self.nats_url = nats_url
        
        self.nc = None
        self.js = None
        self._connected = False
        
        self._handlers: dict[str, Callable[[swarm_pb2.MsgEnvelope], None]] = {}
        self._subscriptions = []
        self._heartbeat_task = None

    async def connect(self):
        """Connects to NATS and sets up inbox subscriptions."""
        if self._connected:
            return

        try:
            self.nc = await nats.connect(
                self.nats_url,
                reconnected_cb=self._on_reconnect
            )
            self.js = self.nc.jetstream()
            self._connected = True
            logger.info(f"Relay connected to {self.nats_url} as {self.agent_id} (Team: {self.team_id})")
            
            persistence.init_db()

            await self.subscribe(f"swarm.agent.{self.agent_id}.>", self._default_handler)
            await self.subscribe(f"swarm.team.{self.team_id}.>", self._default_handler)

            self._heartbeat_task = asyncio.create_task(self._heartbeat_loop())
            
        except Exception as e:
            logger.error(f"Failed to connect to NATS: {e}")
            raise

    async def close(self):
        if self._heartbeat_task:
            self._heartbeat_task.cancel()
            try:
                await self._heartbeat_task
            except asyncio.CancelledError:
                pass

        if self.nc:
            await self.nc.drain()
            await self.nc.close()
            self._connected = False
            logger.info("Relay connection drained.")

    async def _heartbeat_loop(self):
        """Emits a heartbeat every 5 seconds."""
        logger.info("💓 Heartbeat Loop Started")
        while True:
            try:
                await asyncio.sleep(5)
                if not self._connected:
                    continue
                
                # Payload matching CloudEvent or simple JSON
                hb_data = {
                    "agent_id": self.agent_id,
                    "team_id": self.team_id,
                    "status": "alive",
                    "timestamp": time.time(),
                    "meta": {"sdk": "python-relay-v1"}
                }
                await self.send_event("agent.heartbeat", hb_data)

            except asyncio.CancelledError:
                logger.info("💓 Heartbeat Loop Stopped")
                break
            except Exception as e:
                logger.error(f"Heartbeat Error: {e}")


    async def subscribe(self, subject: str, callback: Callable[[swarm_pb2.MsgEnvelope], None]):
        """
        Subscribe to a NATS subject.
        The callback receives a valid MsgEnvelope.
        """
        if not self._connected:
            raise ConnectionError("Not connected to NATS")

        logger.info(f"Subscribing to {subject}")
        
        # We wrap the callback to deserialize Protocol Buffer
        async def _nats_callback(msg):
            try:
                envelope = swarm_pb2.MsgEnvelope()
                envelope.ParseFromString(msg.data)
                
                # Invoke User Callback
                if asyncio.iscoroutinefunction(callback):
                    await callback(envelope)
                else:
                    callback(envelope)
                    
            except Exception as e:
                logger.error(f"Error handling message on {subject}: {e}")

        # NATS Subscription
        sub = await self.nc.subscribe(subject, cb=_nats_callback)
        self._subscriptions.append(sub)
        self._handlers[subject] = callback

    async def _default_handler(self, envelope: swarm_pb2.MsgEnvelope):
        """Default handler for messages without a specific override."""
        # Just log for now, or user can override
        logger.debug(f"Received msg {envelope.id} type={envelope.type}")

    # -- Sending Methods --

    async def send_text(self, 
                        content: str, 
                        recipient_id: str = None, 
                        context: dict[str, Any] = None,
                        target_team: str = None):
        """Send a standard TextPayload."""
        payload = swarm_pb2.TextPayload(
            content=content,
            recipient_id=recipient_id or "",
            intent="talk"
        )
        await self._publish("text", payload, context=context, recipient_id=recipient_id, target_team=target_team)

    async def send_event(self, 
                         event_type: str, 
                         data: dict[str, Any],
                         context: dict[str, Any] = None,
                         target_team: str = None):
        """Send a standard EventPayload."""
        # Convert dict to Struct
        data_struct = Struct()
        data_struct.update(data)
        
        payload = swarm_pb2.EventPayload(
            event_type=event_type,
            data=data_struct
        )
        await self._publish("event", payload, context=context, target_team=target_team)

    async def _publish(self, 
                       payload_key: str, 
                       payload_obj, 
                       context: dict[str, Any] = None,
                       recipient_id: str = None,
                       target_team: str = None):
        
        if not self._connected:
            raise ConnectionError("Relay not connected")

        envelope = self._create_envelope(payload_key, payload_obj, context)
        
        # Determine Topic
        # Default: swarm.team.{my_team}.agent.{my_id}.output
        # Or if routed: swarm.agent.{recipient}.input
        
        topic = f"swarm.team.{self.team_id}.agent.{self.agent_id}.output"
        
        # Routing Logic
        if payload_key == "event" and payload_obj.event_type == "agent.heartbeat":
             topic = "swarm.global.heartbeat"
        elif recipient_id:
            topic = f"swarm.agent.{recipient_id}.input"
        elif target_team:
            topic = f"swarm.team.{target_team}.input"
            
        data = envelope.SerializeToString()
        
        # Use JetStream publish if stream likely exists, else Core NATS?
        # Core NATS is faster for ephemeral.
        # We'll use Core Publish for now as we didn't setup streams in this client.
        try:
            await self.nc.publish(topic, data)
            logger.debug(f"Published {payload_key} to {topic}")
        except (TimeoutError, ConnectionClosedError, ConnectionError) as e:
            logger.warning(f"⚠️  Network Severed ({e}). Buffering {payload_key} to {topic}...")
            persistence.save_impulse(topic, data)

    async def _on_reconnect(self):
        """Replay buffered events on reconnection."""
        logger.info("♻️  Connection Restored. Replaying Black Box Buffer...")
        count = 0
        try:
            for row_id, topic, payload, ts in persistence.drain_buffer():
                try:
                    await self.nc.publish(topic, payload)
                    logger.info(f"   Replayed [{ts}] {topic}")
                    count += 1
                except Exception as e:
                    logger.error(f"   Failed to replay {topic}: {e} (Aborting Drain)")
                    break # Delete wont happen for this row, preserving it.
        except Exception as e:
             logger.error(f"Error checking buffer: {e}")
        
        if count > 0:
            logger.info(f"✅ Replayed {count} buffered events.")

    def _create_envelope(self, payload_key: str, payload_obj, context: dict = None) -> swarm_pb2.MsgEnvelope:
        ts = Timestamp()
        ts.GetCurrentTime()
        
        envelope = swarm_pb2.MsgEnvelope(
            id=f"msg-{uuid.uuid4().hex}",
            timestamp=ts,
            source_agent_id=self.agent_id,
            team_id=self.team_id,
            trace_id=uuid.uuid4().hex, # Trace propagation would go here
            type=swarm_pb2.MESSAGE_TYPE_TEXT if payload_key == "text" else swarm_pb2.MESSAGE_TYPE_EVENT
        )
        
        # Attach Payload (OneOf)
        if payload_key == "text":
            envelope.text.CopyFrom(payload_obj)
        elif payload_key == "event":
            envelope.event.CopyFrom(payload_obj)
        elif payload_key == "tool_call":
            envelope.tool_call.CopyFrom(payload_obj)
        elif payload_key == "tool_result":
            envelope.tool_result.CopyFrom(payload_obj)
            
        # Attach AG2 Context
        if context:
            struct_ctx = Struct()
            struct_ctx.update(context)
            envelope.swarm_context.CopyFrom(struct_ctx)
            
        return envelope
