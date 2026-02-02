
import typer
import asyncio
import httpx
from rich.console import Console
from rich.table import Table
import nats
import os
import sys

# Add SDK path for Protobufs if needed, or use HTTP API where possible
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../sdk/python/src')))

app = typer.Typer(name="myc", help="Synaptic Injector - Mycelis CLI")
console = Console()

API_URL = os.getenv("MYCELIS_API_URL", "http://localhost:8080")
NATS_URL = os.getenv("NATS_URL", "nats://localhost:4222")

@app.command()
def status():
    """Check the health of the Neural Core."""
    console.print(f"Checking Pulse at [bold cyan]{API_URL}[/bold cyan]...")
    try:
        # Check API
        r = httpx.get(f"{API_URL}/api/v1/health", timeout=2.0)
        if r.status_code == 200:
            console.print("✅ [bold green]Core Online[/bold green]")
            console.print(f"   Status: {r.json()}")
        else:
            console.print(f"⚠️ [bold yellow]Core Unstable[/bold yellow]: {r.status_code}")
    except Exception as e:
        console.print(f"❌ [bold red]Core Unreachable[/bold red]: {e}")
        
    # Check NATS (Simple connect)
    async def check_nats():
        try:
            nc = await nats.connect(NATS_URL)
            console.print("✅ [bold green]NATS Connected[/bold green]")
            await nc.close()
        except Exception as e:
            console.print(f"❌ [bold red]NATS Unreachable[/bold red]: {e}")
            
    try:
        asyncio.run(check_nats())
    except Exception:
        pass

@app.command()
def scan():
    """Scan for active agents in the swarm."""
    console.print("Scanning Registry...")
    # TODO: Use NATS Request or API Registry endpoint if exists
    # For now, let's peek at the Registry via API if we added one, otherwise NATS
    
    # We don't have a specific GET /api/v1/agents yet in the plan?
    # Let's fallback to "Not Implemented" or query memory for "agent.heartbeat"
    
    console.print("⚠️  Registry API not exposed. Querying Memory Stream for unique agents...")
    
    try:
        r = httpx.get(f"{API_URL}/api/v1/memory/stream?limit=100", timeout=2.0)
        if r.status_code == 200:
            logs = r.json()
            agents = set()
            for log in logs:
                if log['source']:
                    agents.add(log['source'])
            
            table = Table(title="Detected Agents (Last 100 Events)")
            table.add_column("Agent ID", style="cyan")
            table.add_column("Status", style="green")
            
            for agent in agents:
                table.add_row(agent, "Active")
                
            console.print(table)
        else:
            console.print("❌ Failed to fetch memory.")
    except Exception as e:
        console.print(f"❌ Error: {e}")

@app.command()
def inject(intent: str, payload: str):
    """Inject a stimulus into the swarm."""
    console.print(f"💉 Injecting [bold]{intent}[/bold]...")
    # TODO: Implement NATS publish using RelayClient logic
    console.print("⚠️  Injection requires Protobuf. Using Stub.")

if __name__ == "__main__":
    app()
