import typer
import asyncio
import httpx
from rich.console import Console
from rich.table import Table
import nats
import os
import sys
import time
from uuid import uuid4

# Add SDK path for Protobufs if needed, or use HTTP API where possible
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '../sdk/python/src')))

try:
    from scip.proto import envelope_pb2
except ImportError as e:
    # Just warn, don't crash, so other commands work
    print(f"Warning: Failed to import SCIP protos: {e}")
    envelope_pb2 = None


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
def think(prompt: str, profile: str = typer.Option("chat", help="Cognitive Profile (coder, chat, logic)")):
    """Infer a response using the Cognitive Matrix."""
    payload = {"profile": profile, "prompt": prompt}
    
    try:
        r = httpx.post(f"{API_URL}/api/v1/cognitive/infer", json=payload, timeout=30.0)
        if r.status_code == 200:
            data = r.json()
            # Print meta to stderr (info)
            console.print(f"[dim italic]Used Model: {data.get('model_used', 'unknown')}[/dim italic]", style="grey50")
            # Print raw text to stdout
            print(data.get("text", ""))
        else:
            console.print(f"❌ Error {r.status_code}: {r.text}", style="red")
    except Exception as e:
        console.print(f"❌ Connection Failed: {e}", style="red")

@app.command()
def inject(intent: str, payload: str, type: str = "text"):
    """Inject a stimulus into the swarm (SCIP)."""
    if not envelope_pb2:
        console.print("❌ SCIP Protos not found. Run 'inv proto.generate'")
        return

    console.print(f"💉 Injecting [bold]{intent}[/bold]...")

    async def run_inject():
        try:
            nc = await nats.connect(NATS_URL)
            
            # Construct Envelope
            env = envelope_pb2.SignalEnvelope()
            env.trace_id = str(uuid4())
            env.timestamp = int(time.time() * 1e9)
            env.sender_id = "user:cli"
            env.target_id = "broadcast"
            env.intent = intent
            env.data_type = envelope_pb2.TEXT_UTF8 if type == "text" else envelope_pb2.BINARY
            env.payload = payload.encode('utf-8')
            
            # Serialize
            data = env.SerializeToString()
            
            # Publish to broad subject
            subject = f"scip.{intent}"
            await nc.publish(subject, data)
            console.print(f"✅ Sent {len(data)} bytes to {subject}")
            
            await nc.close()
        except Exception as e:
            console.print(f"❌ Error: {e}")

    asyncio.run(run_inject())

@app.command()
def snoop():
    """Spy on the Neural Bus (SCIP)."""
    if not envelope_pb2:
        console.print("❌ SCIP Protos not found. Run 'inv proto.generate'")
        return

    console.print("🕵️  Snooping on [bold]scip.>[/bold]...")
    
    async def run_snoop():
        try:
            nc = await nats.connect(NATS_URL)
            
            async def cb(msg):
                try:
                    env = envelope_pb2.SignalEnvelope()
                    env.ParseFromString(msg.data)
                    
                    console.print(f"\n[bold cyan]📨 Envelope ({msg.subject}):[/bold cyan]")
                    console.print(f"  Trace: {env.trace_id}")
                    console.print(f"  From:  {env.sender_id} -> {env.target_id}")
                    console.print(f"  Intent: [bold green]{env.intent}[/bold green]")
                    console.print(f"  Type:   {envelope_pb2.DataType.Name(env.data_type)}")
                    console.print(f"  Size:   {len(env.payload)} bytes")
                    if env.data_type == envelope_pb2.TEXT_UTF8:
                        console.print(f"  Data:   [italic]{env.payload.decode()}[/italic]")
                except Exception as e:
                    console.print(f"❌ Malformed Packet: {e}")

            await nc.subscribe("scip.>", cb=cb)
            
            # Keep running
            while True:
                await asyncio.sleep(1)
        except Exception as e:
            console.print(f"❌ NATS Error: {e}")

    try:
        asyncio.run(run_snoop())
    except KeyboardInterrupt:
        console.print("\nStopped.")

@app.command()
def admin_create(username: str, role: str = "admin"):
    """Bootstrap the first Admin user (and Identity Schema)."""
    console.print(f"👑 Bootstrapping Admin: [bold]{username}[/bold]...")
    
    import psycopg2
    from psycopg2.extras import RealDictCursor
    
    # 1. Connect (Try localhost:5432 - assuming Port Forward)
    # Default creds for local dev (often postgres/postgres or mycelis/password)
    # We'll try common defaults
    DSN = os.getenv("PG_DSN", "postgresql://postgres:postgres@localhost:5432/postgres")
    
    try:
        conn = psycopg2.connect(DSN)
        conn.autocommit = True
        cur = conn.cursor()
        console.print("✅ Connected to Database")
        
        # 2. Apply Schema
        schema_path = os.path.join(os.path.dirname(__file__), '../core/internal/identity/schema.sql')
        if os.path.exists(schema_path):
            console.print("📜 Applying Identity Schema...")
            with open(schema_path, 'r') as f:
                schema_sql = f.read()
                cur.execute(schema_sql)
            console.print("✅ Schema Applied")
        else:
            console.print(f"⚠️ Schema file not found at {schema_path}")
            
        # 3. Create User
        user_id = str(uuid4())
        try:
            cur.execute(
                "INSERT INTO users (id, username, role) VALUES (%s, %s, %s) ON CONFLICT (username) DO NOTHING",
                (user_id, username, role)
            )
            console.print(f"✅ User '{username}' inserted/verified (ID: {user_id})")
        except Exception as e:
            console.print(f"⚠️ Failed to insert user: {e}")

        # 4. Generate Token (Mock for now, normally JWT)
        # We assume the UI just checks for presence in localStorage, but ideally we return a signed token.
        # For Minimum Viable Bootstrapper, we return the UUID.
        console.print(f"\n🔑 [bold green]Session Token:[/bold green] {user_id}")
        console.print("👉 Copy this to your UI Settings or Local Storage.")

        conn.close()
        
    except Exception as e:
        console.print(f"❌ Database Error: {e}")
        console.print("[dim]Ensure 'uv run inv k8s.bridge' is running to forward port 5432.[/dim]")

if __name__ == "__main__":
    app()
