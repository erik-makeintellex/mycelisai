from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
import nats
import asyncio
from nats.js import JetStreamContext
from contextlib import asynccontextmanager
import os
import json
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy import text
from datetime import datetime

from shared.schemas import AgentConfig, EventMessage, MessageType, AIModel, Team, TeamUpdate, Service, ServiceType, TextMessage
from shared.db import init_db, get_db, AgentDB, TeamDB, AIModelDB, CredentialDB, ServiceDB, UserDB, RoleDB, GroupDB

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await init_db()
    app.state.nc = await nats.connect(os.getenv("NATS_URL", "nats://localhost:4222"))
    app.state.js = app.state.nc.jetstream()
    yield
    # Shutdown
    await app.state.nc.close()

app = FastAPI(title="Mycelis Service Network API", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Allow all origins for development
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.post("/agents/register")
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

@app.delete("/agents/{agent_name}")
async def delete_agent(agent_name: str, db: AsyncSession = Depends(get_db)):
    agent = await db.get(AgentDB, agent_name)
    if not agent:
        raise HTTPException(status_code=404, detail="Agent not found")
    await db.delete(agent)
    await db.commit()
    return {"status": "deleted", "agent": agent_name}

@app.get("/health")
async def health_check(db: AsyncSession = Depends(get_db)):
    """Check system health (DB, NATS) and return environment info."""
    health_status = {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "components": {
            "database": "unknown",
            "nats": "unknown",
            "api": "healthy"
        },
        "environment": {
            "nats_url": os.getenv("NATS_URL", "nats://localhost:4222"),
            "database_url": "configured" if os.getenv("DATABASE_URL") else "missing",
            "backend_host": os.getenv("HOST", "localhost")
        }
    }

    # Check Database
    try:
        await db.execute(text("SELECT 1"))
        health_status["components"]["database"] = "connected"
    except Exception as e:
        health_status["components"]["database"] = f"error: {str(e)}"
        health_status["status"] = "degraded"

    # Check NATS
    nc = app.state.nc
    if nc and nc.is_connected:
        health_status["components"]["nats"] = "connected"
    else:
        health_status["components"]["nats"] = "disconnected"
        health_status["status"] = "degraded"

    return health_status

@app.get("/agents")
async def list_agents(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(AgentDB))
    agents = result.scalars().all()
    return [AgentConfig(**agent.config) for agent in agents]

@app.get("/models")
async def list_models(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(AIModelDB))
    models = result.scalars().all()
    return [AIModel(
        id=m.id, name=m.name, provider=m.provider, 
        context_window=m.context_window, input_price=m.input_price, 
        output_price=m.output_price, description=m.description
    ) for m in models]

@app.post("/models")
async def add_model(model: AIModel, db: AsyncSession = Depends(get_db)):
    db_model = AIModelDB(**model.model_dump())
    db.add(db_model)
    await db.commit()
    return {"status": "added", "model": model}

@app.get("/teams")
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

@app.post("/teams")
async def create_team(team: Team, db: AsyncSession = Depends(get_db)):
    # Auto-provision channels
    team.inter_comm_channel = f"team.{team.id}.chat"
    if "admin" not in team.channels:
        team.channels.append("admin")
    
    # Provision NATS streams
    js: JetStreamContext = app.state.js
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

@app.post("/teams/{team_id}/agents/{agent_name}")
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

@app.delete("/teams/{team_id}/agents/{agent_name}")
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

@app.put("/teams/{team_id}")
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

@app.post("/services")
async def register_service(service: Service, db: AsyncSession = Depends(get_db)):
    db_service = ServiceDB(
        id=service.id,
        name=service.name,
        type=service.type,
        config=service.config,
        status=service.status,
        description=service.description
    )
    db.add(db_service)
    await db.commit()
    await db.refresh(db_service)
    return service

@app.get("/services")
async def list_services(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(ServiceDB))
    services = result.scalars().all()
    return [
        Service(
            id=s.id, name=s.name, type=s.type, config=s.config,
            status=s.status, description=s.description, created_at=s.created_at
        )
        for s in services
    ]

@app.delete("/services/{service_id}")
async def delete_service(service_id: str, db: AsyncSession = Depends(get_db)):
    service = await db.get(ServiceDB, service_id)
    if not service:
        raise HTTPException(status_code=404, detail="Service not found")
    await db.delete(service)
    await db.commit()
    return {"status": "deleted", "id": service_id}

@app.post("/ingest/{channel}")
async def ingest_event(channel: str, event: EventMessage):
    js: JetStreamContext = app.state.js
    try:
        # Ensure the stream exists (simple auto-provisioning for demo)
        stream_name = channel.replace(".", "-")
        try:
            await js.add_stream(name=stream_name, subjects=[f"{channel}.>"])
        except Exception:
            pass # Stream might already exist

        await js.publish(f"{channel}.event", event.model_dump_json().encode())
        return {"status": "published", "id": event.id}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/stream/{channel}")
async def stream_events(channel: str):
    async def event_generator():
        queue = asyncio.Queue()
        
        async def cb(msg):
            await queue.put(msg)

        # Subscribe to NATS (wildcard for all events in channel)
        # We use the shared NATS connection from app.state
        sub = await app.state.js.subscribe(f"{channel}.>", cb=cb)
        
        try:
            while True:
                msg = await queue.get()
                yield f"data: {msg.data.decode()}\n\n"
                await msg.ack()
        except asyncio.CancelledError:
            await sub.unsubscribe()
        finally:
            # Ensure unsubscribe happens if generator exits
            try:
                await sub.unsubscribe()
            except Exception:
                pass

    return StreamingResponse(event_generator(), media_type="text/event-stream")

@app.post("/agents/{agent_name}/chat")
async def chat_with_agent(agent_name: str, message: TextMessage):
    """Send a direct message to an agent."""
    js: JetStreamContext = app.state.js
    
    # Construct AgentMessage
    # We use "user" as the sender for now, or could use a session ID
    sender_id = "user" 
    
    agent_msg = AgentMessage(
        id=f"msg-{datetime.utcnow().timestamp()}",
        source_agent_id="api",
        sender=sender_id,
        recipient=agent_name,
        content=message.content,
        intent="chat",
        type=MessageType.TEXT
    )
    
    # Ensure stream exists
    try:
        await js.add_stream(name="chat", subjects=["chat.>"])
    except Exception:
        pass

    # Publish to chat.agent.{name}
    await js.publish(f"chat.agent.{agent_name}", agent_msg.model_dump_json().encode())
    
    return {"status": "sent", "id": agent_msg.id}

@app.get("/")
async def root():
    return {"message": "Mycelis API is running", "docs": "/docs"}
