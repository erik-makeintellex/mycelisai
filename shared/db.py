from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker, declarative_base
from sqlalchemy import Column, String, Integer, Float, JSON, DateTime, ForeignKey, Boolean
from datetime import datetime
import os
import uuid

DATABASE_URL = os.getenv("DATABASE_URL", "postgresql+asyncpg://mycelis:password@localhost/mycelis")

engine = create_async_engine(DATABASE_URL, echo=True)
AsyncSessionLocal = sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)
Base = declarative_base()

class AgentDB(Base):
    __tablename__ = "agents"
    name = Column(String, primary_key=True)
    config = Column(JSON)
    created_at = Column(DateTime, default=datetime.utcnow)

class TeamDB(Base):
    __tablename__ = "teams"
    id = Column(String, primary_key=True)
    name = Column(String)
    description = Column(String, nullable=True)
    agents = Column(JSON) # List of agent names
    channels = Column(JSON) # List of channels
    inter_comm_channel = Column(String, nullable=True)
    resource_access = Column(JSON)
    shared_context = Column(JSON)
    created_at = Column(DateTime, default=datetime.utcnow)

class AIModelDB(Base):
    __tablename__ = "models"
    id = Column(String, primary_key=True)
    name = Column(String)
    provider = Column(String)
    context_window = Column(Integer)
    input_price = Column(Float, default=0.0)
    output_price = Column(Float, default=0.0)
    description = Column(String, nullable=True)

class CredentialDB(Base):
    __tablename__ = "credentials"
    id = Column(String, primary_key=True) # e.g., "openai-key"
    value = Column(String) # Encrypted value (mocked as plain for now)
    owner_id = Column(String) # Agent or Team ID
    type = Column(String) # "api_key", "secret", etc.

class ServiceDB(Base):
    __tablename__ = "services"
    id = Column(String, primary_key=True)
    name = Column(String, nullable=False)
    type = Column(String, nullable=False)
    config = Column(JSON, nullable=False)
    status = Column(String, default="active")
    description = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

class MCPServerDB(Base):
    __tablename__ = "mcp_servers"
    name = Column(String, primary_key=True)
    description = Column(String, nullable=True)
    config = Column(JSON, nullable=False)
    tools = Column(JSON, default=[])
    status = Column(String, default="stopped")

class MCPPermissionDB(Base):
    __tablename__ = "mcp_permissions"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    agent_id = Column(String, nullable=True)
    team_id = Column(String, nullable=True)
    server_name = Column(String, nullable=False)
    tool_name = Column(String, default="*")
    allowed = Column(Boolean, default=True)



class UserDB(Base):
    __tablename__ = "users"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    username = Column(String, unique=True, nullable=False)
    email = Column(String, nullable=True)
    password_hash = Column(String, nullable=False)
    roles = Column(JSON, default=[])
    group_ids = Column(JSON, default=[])
    created_at = Column(DateTime, default=datetime.utcnow)

class RoleDB(Base):
    __tablename__ = "roles"
    name = Column(String, primary_key=True)
    permissions = Column(JSON, default=[])

class GroupDB(Base):
    __tablename__ = "groups"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    name = Column(String, nullable=False)
    parent_group_id = Column(String, nullable=True)
    resource_access_policy = Column(JSON, default={})

    resource_access_policy = Column(JSON, default={})

class ConversationDB(Base):
    __tablename__ = "conversations"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    title = Column(String, nullable=True)
    user_id = Column(String, nullable=True) # Optional link to user
    created_at = Column(DateTime, default=datetime.utcnow)
    metadata_ = Column("metadata", JSON, default={}) # naming collision with sqlachemy metadata

class MessageDB(Base):
    __tablename__ = "messages"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    conversation_id = Column(String, ForeignKey("conversations.id", ondelete="CASCADE"), nullable=False)
    timestamp = Column(DateTime, default=datetime.utcnow)
    role = Column(String, nullable=False) # user, agent, system
    sender = Column(String, nullable=False) # agent_name or "user"
    content = Column(String, nullable=False)
    type = Column(String, default="text") # text, event, tool_call

class MemoryDB(Base):
    __tablename__ = "memories"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    agent_id = Column(String, ForeignKey("agents.name", ondelete="CASCADE"), nullable=False)
    content = Column(String, nullable=False)
    tags = Column(JSON, default=[])
    created_at = Column(DateTime, default=datetime.utcnow)

async def init_db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

async def get_db():
    async with AsyncSessionLocal() as session:
        yield session
