from fastapi import APIRouter, HTTPException, Depends, Request
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from shared.schemas import Team, TeamUpdate
from shared.db import get_db, TeamDB, AgentDB
from nats.js import JetStreamContext

router = APIRouter(prefix="/teams", tags=["teams"])

@router.get("")
async def list_teams(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(TeamDB))
    teams = result.scalars().all()
    return [Team(
        id=t.id, name=t.name, description=t.description,
        agents=t.agents, channels=t.channels,
        inter_comm_channel=t.inter_comm_channel,
        resource_access=t.resource_access, shared_context=t.shared_context,
        created_at=t.created_at
    ) for t in teams]

@router.post("")
async def create_team(team: Team, request: Request, db: AsyncSession = Depends(get_db)):
    # Auto-provision channels
    team.inter_comm_channel = f"team.{team.id}.chat"
    if "admin" not in team.channels:
        team.channels.append("admin")
    
    # Provision NATS streams
    js: JetStreamContext = request.app.state.js
    try:
        await js.add_stream(name=f"team-{team.id}", subjects=[f"team.{team.id}.>"])
    except Exception:
        pass

    db_team = TeamDB(
        id=team.id, name=team.name, description=team.description,
        agents=team.agents, channels=team.channels,
        inter_comm_channel=team.inter_comm_channel,
        resource_access=team.resource_access, shared_context=team.shared_context
    )
    db.add(db_team)
    await db.commit()
    return {"status": "created", "team": team}

@router.post("/{team_id}/agents/{agent_name}")
async def add_agent_to_team(team_id: str, agent_name: str, db: AsyncSession = Depends(get_db)):
    team = await db.get(TeamDB, team_id)
    if not team:
        raise HTTPException(status_code=404, detail="Team not found")
    
    # Verify agent exists
    agent = await db.get(AgentDB, agent_name)
    if not agent:
        raise HTTPException(status_code=404, detail="Agent not found")
    
    agents_list = list(team.agents)
    if agent_name not in agents_list:
        agents_list.append(agent_name)
        team.agents = agents_list
        await db.commit()
    
    return {"status": "added", "team_id": team_id, "agent": agent_name}

@router.delete("/{team_id}/agents/{agent_name}")
async def remove_agent_from_team(team_id: str, agent_name: str, db: AsyncSession = Depends(get_db)):
    team = await db.get(TeamDB, team_id)
    if not team:
        raise HTTPException(status_code=404, detail="Team not found")
    
    agents_list = list(team.agents)
    if agent_name in agents_list:
        agents_list.remove(agent_name)
        team.agents = agents_list
        await db.commit()
    
    return {"status": "removed", "team_id": team_id, "agent": agent_name}

@router.put("/{team_id}")
async def update_team(team_id: str, update: TeamUpdate, db: AsyncSession = Depends(get_db)):
    team = await db.get(TeamDB, team_id)
    if not team:
        raise HTTPException(status_code=404, detail="Team not found")
    
    update_data = update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(team, key, value)
    
    await db.commit()
    await db.refresh(team)
    
    return Team(
        id=team.id, name=team.name, description=team.description,
        agents=team.agents, channels=team.channels,
        inter_comm_channel=team.inter_comm_channel,
        resource_access=team.resource_access, shared_context=team.shared_context,
        created_at=team.created_at
    )

@router.delete("/{team_id}")
async def delete_team(team_id: str, db: AsyncSession = Depends(get_db)):
    team = await db.get(TeamDB, team_id)
    if not team:
        raise HTTPException(status_code=404, detail="Team not found")
    
    await db.delete(team)
    await db.commit()
    return {"status": "deleted", "team_id": team_id}
