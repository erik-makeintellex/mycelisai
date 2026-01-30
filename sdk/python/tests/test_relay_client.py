import pytest
import sys
import os
import asyncio
from unittest.mock import AsyncMock, MagicMock

# Path hacking for local dev without install
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../src")))

from relay.client import RelayClient
try:
    from relay.proto.swarm.v1 import swarm_pb2
except ImportError:
    import swarm_pb2

@pytest.mark.asyncio
async def test_relay_defaults():
    """Verify RelayClient defaults to 'swarm:base' if no source provided."""
    client = RelayClient(agent_id="test-agent", team_id="testing")
    
    assert client.agent_id == "test-agent"
    assert client.team_id == "testing"
    assert client.source_uri == "swarm:base"

@pytest.mark.asyncio
async def test_relay_custom_source():
    """Verify Source URI override."""
    client = RelayClient(agent_id="ros-node", team_id="robotics", source_uri="ros2:turtlebot")
    assert client.source_uri == "ros2:turtlebot"

@pytest.mark.asyncio
async def test_context_packing():
    """Verify AG2 Context dict is correctly packed into Protobuf Struct."""
    client = RelayClient(agent_id="ctx-test", team_id="qa")
    
    # Mock NATS connection to avoid real network
    client.nc = MagicMock()
    client.nc.publish = AsyncMock()
    client._connected = True
    
    context_data = {
        "user_name": "Alice",
        "has_plan": True,
        "step": 1
    }
    
    # We inspect the generated envelope before publish
    # Since _publish calls SerializeToString, we can check the call args but harder to deserialize without knowing structure
    # Better: Inspect _create_envelope output
    
    payload = swarm_pb2.TextPayload(content="test")
    envelope = client._create_envelope("text", payload, context=context_data)
    
    assert envelope.swarm_context is not None
    assert envelope.swarm_context["user_name"] == "Alice"
    assert envelope.swarm_context["has_plan"] == True
    assert envelope.swarm_context["step"] == 1.0 # Protobuf Struct numbers are doubles usually
