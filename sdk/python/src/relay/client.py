import asyncio
import uuid
import time
import logging
import nats
from nats.errors import ConnectionClosedError, TimeoutError, NoRespondersError
from google.protobuf.timestamp_pb2 import Timestamp
from google.protobuf import struct_pb2

# Adjust import based on generation output structure
try:
    from .proto.swarm.v1 import swarm_pb2
except ImportError:
    # Fallback for local testing if path varies
    import swarm_pb2

logger = logging.getLogger("relay")

class RelayClient:
    """
    The RelayClient bridges Python agents/scripts to the Mycelis Swarm.
    It enforces swarm.proto schemas and manages the NATS connection.
    """
    def __init__(self, agent_id: str, team_id: str, source_uri: str = "swarm:base", nats_url: str = "nats://localhost:4222"):
        self.nats_url = nats_url
        self.agent_id = agent_id or f"python-relay-{uuid.uuid4().hex[:8]}"
        self.team_id = team_id
        self.source_uri = source_uri
        self.nc = None
        self.js = None
        self._connected = False

    async def connect(self):
        """Connects to the NATS Nervous System."""
        try:
            self.nc = await nats.connect(self.nats_url)
            self.js = self.nc.jetstream()
            self._connected = True
            logger.info(f"Relay connected to {self.nats_url} as {self.agent_id}")
        except Exception as e:
            logger.error(f"Failed to connect to NATS: {e}")
            raise

    async def close(self):
        """Closes the connection."""
        if self.nc:
            await self.nc.drain()
            self._connected = False
            logger.info("Relay connection drained.")

    def _create_envelope(self, payload_key: str, payload_obj, context: dict = None, intent: str = None) -> swarm_pb2.MsgEnvelope:
        """Wraps a payload in a standard MsgEnvelope with Swarm Context."""
        
        # 1. Timestamp
        now = time.time()
        ts = Timestamp()
        ts.FromSeconds(int(now))

        # 2. Build Envelope
        envelope = swarm_pb2.MsgEnvelope(
            id=f"msg-{uuid.uuid4()}",
            timestamp=ts,
            source_agent_id=self.agent_id,
            team_id=self.team_id,
            type=swarm_pb2.MESSAGE_TYPE_TEXT if payload_key == "text" else swarm_pb2.MESSAGE_TYPE_EVENT, # Default logic
            trace_id=uuid.uuid4().hex # Simple trace generation
        )

        # 3. Attach Payload (OneOf)
        # We use getattr/setattr or the generated field setter
        if payload_key == "text":
            envelope.text.CopyFrom(payload_obj)
            envelope.type = swarm_pb2.MESSAGE_TYPE_TEXT
        elif payload_key == "event":
            envelope.event.CopyFrom(payload_obj)
            envelope.type = swarm_pb2.MESSAGE_TYPE_EVENT
        elif payload_key == "tool_call":
            envelope.tool_call.CopyFrom(payload_obj)
            envelope.type = swarm_pb2.MESSAGE_TYPE_TOOL_CALL
        elif payload_key == "tool_result":
            envelope.tool_result.CopyFrom(payload_obj)
            envelope.type = swarm_pb2.MESSAGE_TYPE_TOOL_RESULT

        # 4. Attach Context (AG2 Pattern)
        if context:
            # Convert dict to Struct
            s = struct_pb2.Struct()
            s.update(context)
            envelope.swarm_context.CopyFrom(s)

        return envelope

    async def send_text(self, content: str, recipient: str = None, intent: str = "inform", context: dict = None):
        """Sends a standard Human/Chat message."""
        payload = swarm_pb2.TextPayload(
            content=content,
            recipient_id=recipient or "",
            intent=intent,
            context_id="" 
        )
        envelope = self._create_envelope("text", payload, context)
        await self._publish(envelope)

    async def send_event(self, event_type: str, data: dict, stream_id: str = "main", context: dict = None):
        """Sends a Helper/System event (e.g. Telemetry)."""
        # Dict to Struct
        s_data = struct_pb2.Struct()
        s_data.update(data)

        payload = swarm_pb2.EventPayload(
            event_type=event_type,
            stream_id=stream_id,
            data=s_data
        )
        envelope = self._create_envelope("event", payload, context)
        await self._publish(envelope)

    async def _publish(self, envelope: swarm_pb2.MsgEnvelope):
        """Serializes and publishes to the bus."""
        if not self._connected:
            raise ConnectionError("Relay not connected")

        # Subject Logic: swarm.prod.agent.{id}.{type}
        # Simplified for now: defaults to 'event' or specific intent
        subject = f"swarm.prod.agent.{self.agent_id}.message" 
        
        # Override for heartbeats or specific types if needed, keeping it simple
        if envelope.type == swarm_pb2.MESSAGE_TYPE_EVENT and envelope.event.event_type == "heartbeat":
            subject = f"swarm.prod.agent.{self.agent_id}.heartbeat"

        data = envelope.SerializeToString()
        await self.nc.publish(subject, data)
        # logger.debug(f"Published to {subject}: {envelope.id}")
