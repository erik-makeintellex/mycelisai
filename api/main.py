from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import nats
import os
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends
from datetime import datetime

from shared.db import init_db, get_db
from api.routers import agents, teams, services, channels, models, chat

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
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include Routers
app.include_router(agents.router)
app.include_router(teams.router)
app.include_router(services.router)
app.include_router(channels.router)
app.include_router(models.router)
app.include_router(chat.router)

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

    # Check Ollama
    import httpx
    ollama_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
    try:
        async with httpx.AsyncClient() as client:
            resp = await client.get(f"{ollama_url}/api/tags", timeout=2.0)
            if resp.status_code == 200:
                health_status["components"]["ollama"] = "connected"
            else:
                health_status["components"]["ollama"] = f"error: {resp.status_code}"
                health_status["status"] = "degraded"
    except Exception as e:
        health_status["components"]["ollama"] = f"disconnected: {str(e)}"
        health_status["status"] = "degraded"

    return health_status

@app.get("/")
async def root():
    return {"message": "Mycelis API is running", "docs": "/docs"}
