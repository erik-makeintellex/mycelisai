from enum import Enum
from typing import Dict, List, Optional, Any, Union
from pydantic import BaseModel, Field
from datetime import datetime

class MessageType(str, Enum):
    EVENT = "event"
    BINARY = "binary"
    TEXT = "text"

class BaseMessage(BaseModel):
    id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    source_agent_id: str
    type: MessageType

class EventMessage(BaseMessage):
    type: MessageType = MessageType.EVENT
    payload: Dict[str, Any]
    stream_id: Optional[str] = None

class BinaryMessage(BaseMessage):
    type: MessageType = MessageType.BINARY
    payload: bytes
    content_type: str = "application/octet-stream"

class TextMessage(BaseMessage):
    type: MessageType = MessageType.TEXT
    content: str

class AgentMessage(BaseMessage):
    type: MessageType = MessageType.TEXT  # Reuse TEXT type for now, or define specific AGENT type
    sender: str
    recipient: str
    content: str
    intent: Optional[str] = None  # e.g., "ask", "reply", "inform"

    context_id: Optional[str] = None  # To track conversation threads

class ToolCallMessage(BaseMessage):
    type: MessageType = MessageType.TEXT # Using TEXT for JSON payload compatibility for now
    tool_name: str
    arguments: Dict[str, Any]
    call_id: str

class ToolResultMessage(BaseMessage):
    type: MessageType = MessageType.TEXT
    call_id: str
    result: Any
    is_error: bool = False

class MCPServer(BaseModel):
    name: str
    description: Optional[str] = None
    config: Dict[str, Any]  # e.g., {"command": "python", "args": ["script.py"]}
    tools: List[Dict[str, Any]] = []
    status: str = "stopped"

class MCPPermission(BaseModel):
    id: Optional[str] = None
    agent_id: Optional[str] = None
    team_id: Optional[str] = None
    server_name: str
    tool_name: str = "*"  # Wildcard allowed
    allowed: bool = True


class MessagingConfig(BaseModel):
    inputs: List[str] = []
    outputs: List[str] = []

class DeploymentConfig(BaseModel):
    replicas: int = 1
    constraints: Dict[str, str] = {}

class AgentConfig(BaseModel):
    name: str
    languages: List[str]
    prompt_config: Dict[str, Any]
    capabilities: List[str] = []
    messaging: MessagingConfig = Field(default_factory=MessagingConfig)
    deployment: DeploymentConfig = Field(default_factory=DeploymentConfig)
    backend: str = "openai"  # Default backend
    host: str = "192.168.50.156" # Default to host IP for WSL/Docker access
    port: int = 8000
    inter_comm_channel: Optional[str] = None  # Channel for inter-agent communication

class AIModel(BaseModel):
    id: str
    name: str
    provider: str
    context_window: int
    input_price: float = 0.0
    output_price: float = 0.0
    description: Optional[str] = None

class Team(BaseModel):
    id: str
    name: str
    description: Optional[str] = None
    agents: List[str] = []  # List of Agent IDs
    channels: List[str] = []  # List of associated channels
    inter_comm_channel: Optional[str] = None  # Dedicated channel for team chat
    resource_access: Dict[str, str] = {}  # Resource ID -> Access Level (e.g., "read", "write")
    shared_context: Dict[str, Any] = {}
    created_at: datetime = Field(default_factory=datetime.utcnow)

class TeamUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    channels: Optional[List[str]] = None
    resource_access: Optional[Dict[str, str]] = None
    shared_context: Optional[Dict[str, Any]] = None

class ServiceType(str, Enum):
    IOT_DEVICE = "iot_device"
    API = "api"
    DATABASE = "database"
    OTHER = "other"

class Service(BaseModel):
    id: str
    name: str
    type: ServiceType
    config: Dict[str, Any]  # e.g., MQTT topic, API URL
    status: str = "active"
    description: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)

class Permission(str, Enum):
    AGENT_CREATE = "agent.create"
    AGENT_VIEW = "agent.view"
    AGENT_INTERACT = "agent.interact"
    TEAM_CREATE = "team.create"
    TEAM_MANAGE = "team.manage"
    CHANNEL_SUBSCRIBE = "channel.subscribe"
    CHANNEL_PUBLISH = "channel.publish"
    ADMIN = "*"

class Role(BaseModel):
    name: str
    permissions: List[str] = []

class Group(BaseModel):
    id: str
    name: str
    parent_group_id: Optional[str] = None
    resource_access_policy: Dict[str, str] = {}

class User(BaseModel):
    id: str
    username: str
    email: Optional[str] = None
    roles: List[str] = []
    group_ids: List[str] = []
    created_at: datetime = Field(default_factory=datetime.utcnow)

class UserCreate(BaseModel):
    username: str
    password: str
    email: Optional[str] = None

class Token(BaseModel):
    access_token: str
    token_type: str
