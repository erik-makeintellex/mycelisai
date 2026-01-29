
from fastmcp import FastMCP
from pydantic import BaseModel, Field
from typing import List, Optional
import os
import sys

# Ensure shared modules are accessible
sys.path.append(os.path.join(os.path.dirname(__file__), "../../"))

from shared.db import get_db, MemoryDB
from shared.logger import get_logger
from sqlalchemy import select, delete
from sqlalchemy.ext.asyncio import AsyncSession
import uuid

# Initialize FastMCP Server
mcp = FastMCP("memory-mcp", dependencies=["asyncpg", "sqlalchemy", "structlog"])
log = get_logger("services.memory_mcp")

class RememberInput(BaseModel):
    agent_id: str = Field(..., description="The ID (name) of the agent storing the memory")
    content: str = Field(..., description="The content to remember")
    tags: List[str] = Field(default=[], description="Tags for categorization")

class RecallInput(BaseModel):
    agent_id: str = Field(..., description="The ID (name) of the agent recalling memory")
    query: str = Field(..., description="Search query or Resonance keywords")
    limit: int = Field(5, description="Max memories to return")

@mcp.tool()
async def remember(agent_id: str, content: str, tags: List[str] = []) -> str:
    """Store a persistent memory snippet for an agent."""
    log.info("remember_call", agent=agent_id, content_preview=content[:20])
    async for session in get_db():
        try:
            memory = MemoryDB(
                id=str(uuid.uuid4()),
                agent_id=agent_id,
                content=content,
                tags=tags
            )
            session.add(memory)
            await session.commit()
            return f"Memory stored: {memory.id}"
        except Exception as e:
            await session.rollback()
            log.error("remember_failed", error=str(e))
            return f"Error storing memory: {str(e)}"

@mcp.tool()
async def recall(agent_id: str, query: str, limit: int = 5) -> List[dict]:
    """Retrieve memories relevant to a query (Simple contains search for v1)."""
    log.info("recall_call", agent=agent_id, query=query)
    async for session in get_db():
        try:
            # Simple text search for now. V2 will use pgvector.
            stmt = select(MemoryDB).where(
                MemoryDB.agent_id == agent_id
            ).where(
                MemoryDB.content.ilike(f"%{query}%")
            ).limit(limit)
            
            result = await session.execute(stmt)
            memories = result.scalars().all()
            
            return [{
                "id": m.id,
                "content": m.content,
                "tags": m.tags,
                "created_at": str(m.created_at)
            } for m in memories]
        except Exception as e:
            log.error("recall_failed", error=str(e))
            return []

if __name__ == "__main__":
    mcp.run()
