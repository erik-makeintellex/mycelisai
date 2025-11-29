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
    agent = await db.get(AgentDB, agent_name)
    if not agent:
        raise HTTPException(status_code=404, detail="Agent not found")
    await db.delete(agent)
    await db.commit()
    return {"status": "deleted", "agent": agent_name}

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
