import pytest
import sys
import os
from fastapi.testclient import TestClient

# Add project root to sys.path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

def test_imports_shared():
    """Test that shared modules can be imported."""
    try:
        from shared import schemas
        from shared import db
    except ImportError as e:
        pytest.fail(f"Failed to import shared modules: {e}")

def test_imports_api():
    """Test that API modules can be imported."""
    try:
        from api import main
    except ImportError as e:
        pytest.fail(f"Failed to import api modules: {e}")

def test_imports_agents():
    """Test that agent modules can be imported."""
    try:
        from agents import base
    except ImportError as e:
        pytest.fail(f"Failed to import agent modules: {e}")

def test_api_instantiation():
    """Test that the FastAPI app can be instantiated."""
    from api.main import app
    with TestClient(app) as client:
        # We don't need to actually make a request, just instantiating the client 
        # triggers startup events (if any) and verifies the app structure.
        # However, since our startup event connects to DB/NATS, we might want to mock that 
        # or just check if the object exists for a basic smoke test.
        assert app is not None
        assert app.title == "Mycelis Service Network API"

def test_schema_validity():
    """Test that Pydantic schemas are valid and can be instantiated."""
    from shared.schemas import AgentConfig, MessageType
    
    # Test basic schema instantiation
    config = AgentConfig(
        name="test-agent",
        languages=["python"],
        prompt_config={"system": "you are a test"},
        capabilities=["test"]
    )
    assert config.name == "test-agent"
    assert config.backend == "openai" # Default value check
