
import asyncio
import nats
import json
import uuid
import requests
import time
from rich.console import Console

console = Console()

CORE_URL = "http://localhost:8080"

async def main():
    console.print("[bold yellow]🛡️  Governance Smoke Test[/bold yellow]")

    # 1. Connect NATS
    nc = await nats.connect("nats://localhost:4222")
    console.print("✅ Connected to NATS")

    # 2. Test Block Action (DENY)
    # Intent: system.shutdown (Defined in policy.yaml as DENY)
    console.print("\n[bold cyan]🧪 Test 1: Immediate Block (DENY)[/bold cyan]")
    
    # We subscribe to catch if it somehow bounces back (it shouldn't)
    sub = await nc.subscribe("swarm.>")
    
    payload = {
        "id": str(uuid.uuid4()),
        "source_agent_id": "test-script",
        "team_id": "devops",
        "timestamp": {},
        "event": {
            "event_type": "system.shutdown",
            "data": {}
        }
    }
    
    await nc.publish("swarm.test.input", json.dumps(payload).encode())
    console.print(f"   Sent 'system.shutdown' (Expect DENY logging in Core)")
    
    # Wait briefly
    await asyncio.sleep(1)
    
    # 3. Test Require Approval (PARK)
    console.print("\n[bold cyan]🧪 Test 2: Park Request (REQUIRE_APPROVAL)[/bold cyan]")
    # Intent: payment.create amount=100 (Policy: amount > 50 -> Approval)
    
    payload_park = {
        "id": str(uuid.uuid4()),
        "source_agent_id": "finance-bot",
        "team_id": "finance",  # target: team:finance
        "timestamp": {},
        "event": {
            "event_type": "payment.create", 
            "data": {
                "fields": {
                    "amount": {"kind": {"number_value": 100.0}} # Protobuf struct style for Context extract?
                    # Wait, Gatekeeper extracts from Map or Proto?
                    # Gatekeeper.go:
                    # if msg.GetEvent() != nil { for k,v := range Fields ... }
                    # It expects 'structpb.Value'.
                    # Python sending raw JSON to NATS -> Core Proto Unmarshal?
                    # Core uses 'proto.Unmarshal', so we must send PROTOBUF BYTES!
                    # Ah, we can't just send JSON unless Core handles JSON.
                    # Core Router: proto.Unmarshal(msg.Data, &envelope)
                }
            }
        }
    }
    
    # PROBLEM: We need to send valid Protobuf bytes from Python.
    # Alternative: We can use the 'myc inject' CLI if it exists?
    # Or simplified JSON if Core supports it? No, Core is strict Proto.
    
    console.print("[bold red]⚠️  Cannot test NATS injection without Protobuf bindings. Skipping injection.[/bold red]")
    console.print("   (Use `myc inject` when available)")

    # 4. Check Admin API
    console.print("\n[bold cyan]🧪 Test 3: Check Admin Inbox[/bold cyan]")
    try:
        res = requests.get(f"{CORE_URL}/admin/approvals")
        if res.status_code == 200:
            approvals = res.json()
            console.print(f"✅ API Reachable. Pending Approvals: {len(approvals)}")
            for a in approvals:
                console.print(f"   - {a['request_id']}: {a['reason']}")
        else:
            console.print(f"❌ API Error: {res.status_code}")
    except Exception as e:
        console.print(f"❌ Conn Refused: {e}")

    await nc.close()

if __name__ == "__main__":
    asyncio.run(main())
