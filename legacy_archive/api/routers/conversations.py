
from fastapi import APIRouter, Request, HTTPException
from shared.db import get_db, ConversationDB, MessageDB
from sqlalchemy import select
from sqlalchemy.orm import joinedload
from typing import List, Optional
from pydantic import BaseModel
import uuid
import uuid
from datetime import datetime

router = APIRouter(tags=["conversations"])

class ConversationCreate(BaseModel):
    title: Optional[str] = None
    user_id: Optional[str] = "user"
    metadata: Optional[dict] = {}

class Conversation(BaseModel):
    id: str
    title: Optional[str]
    created_at: datetime

class Message(BaseModel):
    id: str
    role: str
    sender: str
    content: str
    timestamp: datetime

@router.post("/conversations", response_model=Conversation)
async def create_conversation(conversation: ConversationCreate):
    """Start a new conversation thread."""
    async for session in get_db():
        new_conv = ConversationDB(
            id=str(uuid.uuid4()),
            title=conversation.title,
            user_id=conversation.user_id,
            metadata_=conversation.metadata
        )
        session.add(new_conv)
        await session.commit()
        return Conversation(
            id=new_conv.id,
            title=new_conv.title,
            created_at=new_conv.created_at
        )

@router.get("/conversations", response_model=List[Conversation])
async def list_conversations(user_id: Optional[str] = "user"):
    """List recent conversations."""
    async for session in get_db():
        stmt = select(ConversationDB).where(ConversationDB.user_id == user_id).order_by(ConversationDB.created_at.desc()).limit(50)
        result = await session.execute(stmt)
        rows = result.scalars().all()
        return [Conversation(id=r.id, title=r.title, created_at=r.created_at) for r in rows]

@router.get("/conversations/{conversation_id}/history", response_model=List[Message])
async def get_conversation_history(conversation_id: str):
    """Get message history for a conversation."""
    async for session in get_db():
        stmt = select(MessageDB).where(MessageDB.conversation_id == conversation_id).order_by(MessageDB.timestamp.asc())
        result = await session.execute(stmt)
        rows = result.scalars().all()
        return [Message(
            id=r.id,
            role=r.role,
            sender=r.sender,
            content=r.content,
            timestamp=r.timestamp
        ) for r in rows]
