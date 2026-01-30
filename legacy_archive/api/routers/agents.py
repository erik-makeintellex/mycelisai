from fastapi import APIRouter, HTTPException, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from shared.schemas import AgentConfig
from shared.db import get_db, AgentDB

router = APIRouter(prefix="/agents", tags=["agents"])

@router.post("/register")
async def register_agent(config: AgentConfig, db: AsyncSession = Depends(get_db)):
    agent = AgentDB(name=config.name, config=config.model_dump())
    db.add(agent)
    try:
        await db.commit()
    except Exception:
        await db.rollback()
        # If exists, update
        existing = await db.get(AgentDB, config.name)
        if existing:
            existing.config = config.model_dump()
            await db.commit()
            
    return {"status": "registered", "agent": config}

@router.delete("/{agent_name}")
async def delete_agent(agent_name: str, db: AsyncSession = Depends(get_db)):
    """Delete an agent and all associated data (Memories, History, Config)."""
    from shared.db import MessageDB, MemoryDB, ConversationDB, AgentDB
    from sqlalchemy import delete
    
    agent = await db.get(AgentDB, agent_name)
    if not agent:
        raise HTTPException(status_code=404, detail="Agent not found")
    
    # 1. Delete Memories
    await db.execute(delete(MemoryDB).where(MemoryDB.agent_id == agent_name))
    
    # 2. Delete Conversation History
    target_session_id = f"session-{agent_name}"
    
    # Delete messages first
    await db.execute(delete(MessageDB).where(MessageDB.conversation_id == target_session_id))
    
    # Also delete any messages where this agent was the sender in ANY conversation
    # This prevents ghost messages from appearing in other contexts if shared
    await db.execute(delete(MessageDB).where(MessageDB.sender == agent_name))

    # Delete the conversation itself
    await db.execute(delete(ConversationDB).where(ConversationDB.id == target_session_id))
    
    # 3. Delete Agent Config
    await db.delete(agent)
    
    await db.commit()
    return {"status": "deleted", "agent": agent_name}

@router.delete("/{agent_name}/memory")
async def clear_agent_memory(agent_name: str, db: AsyncSession = Depends(get_db)):
    """Clear all conversation history and memories for an agent."""
    from shared.db import MessageDB, MemoryDB, ConversationDB
    from sqlalchemy import delete
    
    # 1. Delete Memories
    await db.execute(delete(MemoryDB).where(MemoryDB.agent_id == agent_name))
    
    # 2. Clear Session History
    target_session_id = f"session-{agent_name}"
    await db.execute(delete(MessageDB).where(MessageDB.conversation_id == target_session_id))
    
    # 3. Clear Broadcast/Group Messages sent by agent
    await db.execute(delete(MessageDB).where(MessageDB.sender == agent_name))
    
    await db.commit()
    return {"status": "memory_cleared", "agent": agent_name}

@router.get("")
async def list_agents(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(AgentDB))
    agents = result.scalars().all()
    return [AgentConfig(**agent.config) for agent in agents]

from shared.schemas import AgentTemplate

TEMPLATES = [
    AgentTemplate(
        id="researcher",
        name="Researcher",
        description="Autonomous web researcher that gathers and summarizes information.",
        role="researcher",
        capabilities=["web-search", "summarization", "fact-checking"],
        system_prompt_template="You are a meticulous researcher. Your goal is to gather accurate information from provided sources and summarize it clearly. Cite your sources.",
        default_inputs=["research.request"],
        default_outputs=["research.summary"]
    ),
    AgentTemplate(
        id="coder",
        name="Coder",
        description="Software engineer agent capable of writing and refactoring code.",
        role="engineer",
        capabilities=["code-generation", "refactoring", "debugging"],
        system_prompt_template="You are an expert software engineer. Write clean, efficient, and well-documented code. Follow best practices and design patterns.",
        default_inputs=["code.request"],
        default_outputs=["code.result"]
    ),
    AgentTemplate(
        id="reviewer",
        name="Reviewer",
        description="Code reviewer that provides constructive feedback and security analysis.",
        role="reviewer",
        capabilities=["code-review", "security-audit", "performance-analysis"],
        system_prompt_template="You are a senior code reviewer. Analyze the provided code for logic errors, security vulnerabilities, and style violations. Provide constructive feedback.",
        default_inputs=["review.request"],
        default_outputs=["review.feedback"]
    ),
    AgentTemplate(
        id="coordinator",
        name="Coordinator",
        description="Task manager that breaks down complex goals and delegates to other agents.",
        role="manager",
        capabilities=["task-decomposition", "delegation", "planning"],
        system_prompt_template="You are a project coordinator. Break down high-level goals into actionable tasks. Assign these tasks to the most suitable agents based on their capabilities.",
        default_inputs=["task.new"],
        default_outputs=["task.assignment"]
    )
]

@router.get("/templates")
async def list_templates():
    return TEMPLATES
