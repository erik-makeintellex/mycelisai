from fastapi import APIRouter, HTTPException, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from shared.schemas import Service
from shared.db import get_db, ServiceDB

router = APIRouter(prefix="/services", tags=["services"])

@router.post("")
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

@router.get("")
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

@router.delete("/{service_id}")
async def delete_service(service_id: str, db: AsyncSession = Depends(get_db)):
    service = await db.get(ServiceDB, service_id)
    if not service:
        raise HTTPException(status_code=404, detail="Service not found")
    await db.delete(service)
    await db.commit()
    return {"status": "deleted", "id": service_id}

@router.post("/{service_id}/check")
async def check_service(service_id: str, db: AsyncSession = Depends(get_db)):
    """Check connectivity to a configured service."""
    service = await db.get(ServiceDB, service_id)
    if not service:
        raise HTTPException(status_code=404, detail="Service not found")

    import httpx
    import time
    
    start_time = time.time()
    status = "offline"
    details = ""

    try:
        if service.type == "mcp_server" or service.type == "api":
            url = service.config.get("url") or service.config.get("health_url")
            if url:
                async with httpx.AsyncClient() as client:
                    # For MCP SSE, we might just check if the endpoint exists
                    # For API, we expect 200
                    try:
                        resp = await client.get(url, timeout=3.0)
                        if resp.status_code < 500: # Any non-server-error is "reachable"
                            status = "online"
                        else:
                            details = f"HTTP {resp.status_code}"
                    except httpx.ConnectError:
                        details = "Connection refused"
                    except httpx.TimeoutException:
                        details = "Timeout"
            else:
                details = "No URL configured"
        
        elif service.type == "database":
            # Placeholder for DB check
            status = "unknown"
            details = "DB check not implemented"
            
        else:
            status = "unknown"
            details = f"Check not implemented for type {service.type}"

    except Exception as e:
        details = str(e)

    latency = int((time.time() - start_time) * 1000)
    
    # Update status in DB
    service.status = status
    await db.commit()

    return {
        "id": service.id,
        "status": status,
        "latency_ms": latency,
        "details": details
    }
