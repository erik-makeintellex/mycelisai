import pytest
import sys
import os
import asyncio
from unittest.mock import MagicMock, patch

# Path hacking to find agents
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../..")))

# Import the modules under test
# We use full paths assuming agents is a package now
from agents.sensors import manager as sensor_manager
from agents.output import manager as output_manager
from relay.proto.swarm.v1 import swarm_pb2

@pytest.mark.asyncio
async def test_sensor_config_update():
    """Verify sensor manager updates polling rate on config event."""
    # Create a fake envelope
    data = {"polling_rate": 0.5}
    envelope = swarm_pb2.MsgEnvelope(
        type=swarm_pb2.MESSAGE_TYPE_EVENT,
        event=swarm_pb2.EventPayload(
            event_type="config_update",
        )
    )
    # Pack data manually since it's a struct in proto usually, but RelayClient V1 helper logic
    # In the manager we access envelope.event.data which is mapped by protobuf as a Map or Struct?
    # correct: envelope.event.data is a google.protobuf.Struct. 
    # Providing a python dict to the constructor doesn't automatically convert to Struct for some fields in pure python protobuf unless assignment is used.
    # We'll use the proper struct wrapper or just mock the access if the code assumes dict-like interface (which protobuf structs do in python).
    
    from google.protobuf import struct_pb2
    s = struct_pb2.Struct()
    s.update(data)
    envelope.event.data.CopyFrom(s)

    # Capture initial state
    initial_rate = sensor_manager.POLLING_RATE
    
    # Run handler
    await sensor_manager.handle_config_update(envelope)
    
    # Verify change
    assert sensor_manager.POLLING_RATE == 0.5
    assert sensor_manager.POLLING_RATE != initial_rate

@pytest.mark.asyncio
async def test_output_handler(capsys):
    """Verify output manager prints alerts to stdout."""
    # Create fake envelope
    envelope = swarm_pb2.MsgEnvelope(
        type=swarm_pb2.MESSAGE_TYPE_TEXT,
        text=swarm_pb2.TextPayload(content="Reactor Critical"),
        source_agent_id="test-agent"
    )
    
    # Run handler
    await output_manager.handle_alert(envelope)
    
    # Check stdout
    captured = capsys.readouterr()
    assert "[DISPLAY] >>> ALERT: Reactor Critical" in captured.out
