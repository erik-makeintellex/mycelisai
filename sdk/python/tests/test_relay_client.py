import pytest
import sys
import os
import asyncio
from unittest.mock import AsyncMock, MagicMock
from google.protobuf import timestamp_pb2

# Path hacking for local dev without install
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../src")))

from relay.client import RelayClient
try:
    from relay.proto.swarm.v1 import swarm_pb2
except ImportError:
    # Fallback if locally generated in a weird spot
    import swarm_pb2

@pytest.mark.asyncio
async def test_relay_defaults():
    """Verify RelayClient defaults."""
    client = RelayClient(agent_id="test-agent", team_id="testing")
    assert client.agent_id == "test-agent"
    assert client.team_id == "testing"
    assert client.source_uri == "swarm:base"

@pytest.mark.asyncio
async def test_context_packing():
    """Verify AG2 Context dict is correctly packed into Protobuf Struct."""
    client = RelayClient(agent_id="ctx-test", team_id="qa")
    client.nc = MagicMock()
    client._connected = True
    
    context_data = {"user": "Alice", "score": 100}
    payload = swarm_pb2.TextPayload(content="test")
    envelope = client._create_envelope("text", payload, context=context_data)
    
    assert envelope.swarm_context["user"] == "Alice"
    assert envelope.swarm_context["score"] == 100.0

@pytest.mark.asyncio
async def test_send_event_structure():
    """Verify send_event creates correct Envelope structure."""
    client = RelayClient(agent_id="sender", team_id="evt")
    client.nc = MagicMock()
    client.nc.publish = AsyncMock()
    client._connected = True

    await client.send_event(
        event_type="sensor.update",
        data={"temp": 23.5},
        context={"loc": "lab"}
    )

    # Capture the call
    args, _ = client.nc.publish.call_args
    topic, data = args
    
    # helper to decode
    envelope = swarm_pb2.MsgEnvelope()
    envelope.ParseFromString(data)

    assert envelope.type == swarm_pb2.MESSAGE_TYPE_EVENT
    assert envelope.event.event_type == "sensor.update"
    assert envelope.event.data["temp"] == 23.5
    assert envelope.source_agent_id == "sender"
    assert envelope.team_id == "evt"

@pytest.mark.asyncio
async def test_subscription_wrapper():
    """Verify subscribe wraps the NATS msg into an Envelope."""
    client = RelayClient(agent_id="sub", team_id="evt")
    client.nc = MagicMock()
    client.nc.subscribe = AsyncMock()
    client._connected = True

    # Mock user callback
    user_cb = AsyncMock()
    await client.subscribe("test.topic", user_cb)

    # Get the wrapper callback that RelayClient passed to nc.subscribe
    args, kwargs = client.nc.subscribe.call_args
    subj = args[0]
    wrapper_cb = kwargs['cb']

    # Simulate NATS message
    fake_env = swarm_pb2.MsgEnvelope(
        id="123", 
        type=swarm_pb2.MESSAGE_TYPE_TEXT,
        text=swarm_pb2.TextPayload(content="Hello")
    )
    nats_msg = MagicMock()
    nats_msg.data = fake_env.SerializeToString()

    # Call wrapper
    await wrapper_cb(nats_msg)

    # Verify user_cb was called with Envelope
    user_cb.assert_called_once()
    called_env = user_cb.call_args[0][0]
    assert isinstance(called_env, swarm_pb2.MsgEnvelope)
    assert called_env.id == "123"
    assert called_env.text.content == "Hello"
