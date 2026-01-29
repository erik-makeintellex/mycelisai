from datetime import datetime
import uuid
import json
from sqlalchemy import text, select
# Direct Import from Shared, bypassing fastmcp dependence
from shared.db import MemoryDB, get_db

async def remember(content: str, agent_id: str = "default") -> str:
    """
    Store a piece of information in long-term memory.
    
    Args:
        content: The fact or observation to remember.
        agent_id: The identity of the agent storing the memory.
    """
    # Use the async generator pattern from shared.db
    async for session in get_db():
        try:
            memory = MemoryDB(
                id=f"mem-{uuid.uuid4().hex}",
                agent_id=agent_id,
                content=content,
            )
            session.add(memory)
            await session.commit()
            return f"Memory stored: {content}"
        except Exception as e:
            return f"Failed to store memory: {e}"
        # We break after one use since get_db is a generator intended for dependency injection,
        # but here we just need one session.
        break
    return "Failed to acquire database session."

async def recall(query: str, agent_id: str = "default") -> str:
    """
    Search long-term memory for relevant information.
    
    Args:
        query: The search query.
        agent_id: The agent_id to restrict search to.
    """
    async for session in get_db():
        try:
            # V1: Simple semantic search simulation (LIKE query)
            stmt = select(MemoryDB).where(
                MemoryDB.agent_id == agent_id,
                MemoryDB.content.ilike(f"%{query}%")
            ).limit(5)
            
            result = await session.execute(stmt)
            memories = result.scalars().all()
            
            if not memories:
                return "No relevant memories found."
            
            return "\n".join([f"- {m.content}" for m in memories])
        except Exception as e:
            return f"Failed to recall memories: {e}"
        break
    return "Failed to acquire database session."
