import asyncio
import asyncpg
import os
from pathlib import Path

# DB Config
DB_USER = "mycelis"
DB_PASS = "password"
DB_NAME = "cortex"
# Assuming localhost:5432 based on typical setup and k8s.bridge
DB_HOST = "localhost" 
DB_PORT = "5432"

DSN = f"postgresql://{DB_USER}:{DB_PASS}@{DB_HOST}:{DB_PORT}/{DB_NAME}"

MIGRATIONS_DIR = Path("core/migrations")
FILES = [
    "003_mission_hierarchy.up.sql",
    "004_registry.up.sql"
]

async def apply_migration(conn, filename):
    filepath = MIGRATIONS_DIR / filename
    print(f"Applying {filename}...")
    if not filepath.exists():
        print(f"Error: {filepath} not found!")
        return
    
    with open(filepath, "r") as f:
        sql = f.read()
    
    try:
        # Execute the SQL
        # Splitting by statement might be needed if the driver doesn't support multistatement well,
        # but asyncpg usually handles executing a block of SQL.
        # However, for migrations, often better to execute as one block or use explicit transaction.
        async with conn.transaction():
             await conn.execute(sql)
        print(f"✅ Successfully applied {filename}")
    except Exception as e:
        print(f"❌ Failed to apply {filename}: {e}")
        # Dont exit, let user verify.

async def main():
    print(f"Connecting to {DSN}...")
    try:
        conn = await asyncpg.connect(DSN)
    except Exception as e:
        print(f"❌ Connection failed: {e}")
        return

    try:
        for filename in FILES:
            await apply_migration(conn, filename)
            
        # Verification Step
        print("\n--- VERIFICATION ---")
        
        # Check Missions
        res = await conn.fetch("SELECT column_name FROM information_schema.columns WHERE table_name = 'missions'")
        print(f"Missions Table Columns: {[r['column_name'] for r in res]}")
        
        # Check Teams
        res = await conn.fetch("SELECT column_name FROM information_schema.columns WHERE table_name = 'teams' AND column_name IN ('parent_id', 'path')")
        print(f"Teams Recursive Columns: {[r['column_name'] for r in res]}")
        
        # Check Registry
        res = await conn.fetch("SELECT count(*) FROM information_schema.tables WHERE table_name = 'connector_templates'")
        print(f"Connector Templates Table Count: {res[0]['count']}")
        
    finally:
        await conn.close()

if __name__ == "__main__":
    asyncio.run(main())
