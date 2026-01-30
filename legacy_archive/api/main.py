from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import nats
import os
import uuid
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends
from datetime import datetime

from shared.db import init_db, get_db
# Changed imports for structlog wrapper
from shared.logger import setup_logger, get_logger, set_trace_id, reset_trace_id
import structlog

from api.routers import agents, teams, services, channels, models, chat, conversations

# Initialize logger for this module
logger = get_logger("api.main")

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Setup Logger (Console for dev by default, JSON if LOG_JSON=true)
    setup_logger("api", level=os.getenv("LOG_LEVEL", "INFO"))
    
    logger.info("server_startup", message="API Server Starting...")

    # Startup
    await init_db()
    app.state.nc = await nats.connect(os.getenv("NATS_URL", "nats://localhost:4222"))
    app.state.js = app.state.nc.jetstream()
    
    logger.info("server_ready", message="API Server Ready")
    yield
    # Shutdown
    await app.state.nc.close()
    logger.info("server_shutdown", message="API Server Shutdown")

app = FastAPI(title="Mycelis Service Network API", lifespan=lifespan)

# Trace ID Middleware
@app.middleware("http")
async def add_process_context(request: Request, call_next):
    trace_id = request.headers.get("X-Trace-ID", str(uuid.uuid4()))
    token = set_trace_id(trace_id)
    try:
        response = await call_next(request)
        return response
    finally:
        reset_trace_id(token)

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
app.include_router(conversations.router)

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

@app.get("/config")
async def get_config():
    """Get current configuration"""
    return {
        "ollama_base_url": os.getenv("OLLAMA_BASE_URL", "http://localhost:11434"),
        "nats_url": os.getenv("NATS_URL", "nats://localhost:4222"),
        "database_url": "configured" if os.getenv("DATABASE_URL") else "missing"
    }

@app.post("/config")
async def update_config(config: dict):
    """Update configuration (requires API restart to apply)"""
    # For now, return the config - in production this would update a ConfigMap
    return {
        "status": "received",
        "message": "Configuration update received. Restart API to apply changes.",
        "config": config
    }
