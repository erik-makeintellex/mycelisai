from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from shared.schemas import AIModel
from shared.db import get_db, AIModelDB

router = APIRouter(prefix="/models", tags=["models"])

@router.get("")
async def list_models(db: AsyncSession = Depends(get_db)):
    # Sync with Ollama
    import httpx
    import os
    ollama_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
    try:
        async with httpx.AsyncClient() as client:
            resp = await client.get(f"{ollama_url}/api/tags", timeout=1.0)
            if resp.status_code == 200:
                data = resp.json()
                for model in data.get("models", []):
                    name = model["name"]
                    # Check if exists
                    existing = await db.get(AIModelDB, name)
                    if not existing:
                        new_model = AIModelDB(
                            id=name,
                            name=name,
                            provider="ollama",
                            context_window=4096, # Default
                            input_price=0,
                            output_price=0,
                            description=f"Ollama model: {name}"
                        )
                        db.add(new_model)
                await db.commit()
    except Exception as e:
        print(f"Ollama sync failed: {e}")

    result = await db.execute(select(AIModelDB))
    models = result.scalars().all()
    return [AIModel(
        id=m.id, name=m.name, provider=m.provider, 
        context_window=m.context_window, input_price=m.input_price, 
        output_price=m.output_price, description=m.description
    ) for m in models]

@router.post("")
async def add_model(model: AIModel, db: AsyncSession = Depends(get_db)):
    db_model = AIModelDB(**model.model_dump())
    db.add(db_model)
    await db.commit()
    return {"status": "added", "model": model}
