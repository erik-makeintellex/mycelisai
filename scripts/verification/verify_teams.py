import asyncio
import aiohttp
import logging
import json

# Configuration
API_URL = "http://localhost:8000"

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

async def verify_teams_workflow():
    logger.info("Starting Team Verification Workflow...")
    
    async with aiohttp.ClientSession() as session:
        # 1. Create an Agent (to assign later)
        agent_name = "test-agent-01"
        agent_config = {
            "name": agent_name,
            "languages": ["python"],
            "prompt_config": {"system_prompt": "You are a test agent."},
            "backend": "mock"
        }
        logger.info(f"Creating Agent {agent_name}...")
        async with session.post(f"{API_URL}/agents/register", json=agent_config) as resp:
            if resp.status != 200:
                logger.error(f"Failed to register agent: {await resp.text()}")
                return False
            logger.info("✅ Agent registered")

        # 2. Create a Team
        team_id = "test-team-01"
        team_config = {
            "id": team_id,
            "name": "Test Team",
            "description": "A team for testing assignment",
            "agents": [] # Start empty
        }
        logger.info(f"Creating Team {team_id}...")
        async with session.post(f"{API_URL}/teams", json=team_config) as resp:
            if resp.status != 200:
                logger.error(f"Failed to create team: {await resp.text()}")
                return False
            logger.info("✅ Team created")

        # 3. Assign Agent to Team
        logger.info(f"Assigning {agent_name} to {team_id}...")
        async with session.post(f"{API_URL}/teams/{team_id}/agents/{agent_name}") as resp:
            if resp.status != 200:
                logger.error(f"Failed to assign agent: {await resp.text()}")
                return False
            data = await resp.json()
            if data.get("agent") == agent_name:
                logger.info("✅ Agent assigned successfully")
            else:
                logger.error(f"Unexpected response: {data}")
                return False

        # 4. Verify Assignment via Get Team
        logger.info("Verifying assignment via GET /teams...")
        async with session.get(f"{API_URL}/teams") as resp:
            teams = await resp.json()
            target_team = next((t for t in teams if t["id"] == team_id), None)
            if target_team and agent_name in target_team["agents"]:
                logger.info("✅ Assignment verified in team list")
            else:
                logger.error(f"Agent not found in team list: {target_team}")
                return False

        # 5. Remove Agent from Team
        logger.info(f"Removing {agent_name} from {team_id}...")
        async with session.delete(f"{API_URL}/teams/{team_id}/agents/{agent_name}") as resp:
            if resp.status != 200:
                logger.error(f"Failed to remove agent: {await resp.text()}")
                return False
            logger.info("✅ Agent removed successfully")

        # 6. Cleanup (Optional, but good for repeatability)
        # We might leave them for manual inspection or delete them.
        # Let's delete to keep it clean.
        await session.delete(f"{API_URL}/agents/{agent_name}")
        # API doesn't have DELETE /teams yet? Let's check schemas/main.py... 
        # I didn't see DELETE /teams/{id} in main.py earlier.
        
    logger.info("🎉 Team Assignment Workflow Verified!")
    return True

if __name__ == "__main__":
    asyncio.run(verify_teams_workflow())
