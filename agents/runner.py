import asyncio
import argparse
import logging
import os
import sys
import aiohttp

# Add project root to path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

from agents.base import BaseAgent
from shared.schemas import AgentConfig

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def fetch_agent_config(api_url: str, agent_name: str) -> AgentConfig:
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{api_url}/agents") as response:
            if response.status != 200:
                raise RuntimeError(f"Failed to fetch agents: {response.status}")
            
            agents_data = await response.json()
            for agent_data in agents_data:
                if agent_data["name"] == agent_name:
                    return AgentConfig(**agent_data)
            
            raise RuntimeError(f"Agent {agent_name} not found in registry")

async def main():
    parser = argparse.ArgumentParser(description="Run a Mycelis Agent")
    parser.add_argument("name", help="Name of the agent to run")
    parser.add_argument("--api", default="http://localhost:8000", help="API URL")
    parser.add_argument("--nats", default="nats://localhost:4222", help="NATS URL")
    
    args = parser.parse_args()
    
    try:
        logger.info(f"Fetching config for agent: {args.name}")
        config = await fetch_agent_config(args.api, args.name)
        
        logger.info(f"Starting agent {args.name} with backend {config.backend}")
        agent = BaseAgent(config, nats_url=args.nats)
        await agent.run()
        
    except Exception as e:
        logger.error(f"Failed to run agent: {e}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())
