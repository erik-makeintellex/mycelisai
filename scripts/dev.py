#!/usr/bin/env -S uv run
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "click",
#     "docker",
#     "watchdog",
#     "grpcio-tools",
#     "protobuf",
#     "pytest",
#     "pytest-asyncio",
#     "nats-py",
#     "aiohttp",
# ]
# ///


import click
import subprocess
import sys
import os
import platform
from pathlib import Path

# Constants
ROOT_DIR = Path(__file__).parent.parent
CORE_DIR = ROOT_DIR / "core"
PROTO_DIR = ROOT_DIR / "proto"

def run_cmd(cmd, cwd=None, env=None):
    """Run a shell command and stream output."""
    click.echo(click.style(f"> {cmd}", fg="green"))
    try:
        subprocess.run(
            cmd, 
            shell=True, 
            check=True, 
            cwd=cwd or ROOT_DIR,
            env={**os.environ, **(env or {})}
        )
    except subprocess.CalledProcessError as e:
        click.echo(click.style(f"Error: Command failed with exit code {e.returncode}", fg="red"))
        sys.exit(e.returncode)

@click.group()
def cli():
    """Mycelis Universal Dev Runner."""
    pass

@cli.command()
def up():
    """Start the infrastructure stack (NATS, Postgres) via Docker Compose."""
    click.echo("Starting infrastructure...")
    # Ensure we use the correct compose file if exists, or check for individual containers
    # For now, we assume a standard compose file at root or deploy/
    compose_file = ROOT_DIR / "deploy" / "docker" / "compose.yml"
    
    if not compose_file.exists():
        click.echo(click.style(f"Warning: {compose_file} not found. Creating temporary dev stack...", fg="yellow"))
        # Fallback to simple NATS start
        run_cmd("docker run -d --name mycelis-nats -p 4222:4222 -p 8222:8222 nats:latest -js")
    else:
        run_cmd(f"docker compose -f {compose_file} up -d")

@cli.command()
def down():
    """Stop the infrastructure stack."""
    run_cmd("docker rm -f mycelis-nats || true")

@cli.command()
def proto():
    """Generate Go code from Protobuf definitions."""
    click.echo("Generating Protobufs...")
    
    # Ensure protoc availability or use a dockerized generator
    # We'll use a dockerized approach for maximum portability
    
    # 1. Mount proto dir to /defs
    # 2. Mount core/pkg/pb to /out
    
    uid = os.getuid() if hasattr(os, 'getuid') else 0
    gid = os.getgid() if hasattr(os, 'getgid') else 0
    
    docker_cmd = (
        f"docker run --rm -v {ROOT_DIR}:/workspace -w /workspace "
        f"ghcr.io/mycelis/proto-builder:latest "
        f"protoc --go_out=. --go_opt=paths=source_relative "
        f"--go-grpc_out=. --go-grpc_opt=paths=source_relative "
        f"proto/swarm/v1/swarm.proto"
    )
    
    # Fallback: check if local protoc exists
    try:
        subprocess.run("protoc --version", shell=True, check=True, stdout=subprocess.DEVNULL)
        has_protoc = True
    except:
        has_protoc = False

    if has_protoc:
        # Use local protoc
        # We need to ensure the go plugins are installed
        # "uvx" can't easily install go binaries, so we rely on user or a go-based tool
        run_cmd(
            "protoc -I. "
            "--go_out=. --go_opt=module=github.com/mycelis/core "
            "proto/swarm/v1/swarm.proto"
        )
    else:
        click.echo("Local protoc not found. Using Docker...")
        # Note: We haven't built the builder image yet, so this might fail in this specific turn.
        # Ideally we use a standard image like namely/protoc-all or buf
        # Use namely/protoc-all which has protoc and go plugins
        # We mount current dir to /defs
        # Command: -f proto/swarm/v1/swarm.proto -l go -o . --go-source-relative
        
        # NOTE: standard protoc-all entrypoint expects -d directory. 
        # We'll use a direct invocation if possible or format args for it.
        # Actually, let's use a simpler image definition if generic:
        # "rvolosatovs/protoc" is small.
        # Or just "golang:1.24-alpine" and install protoc? No, that's slow.
        
        # Let's try executing the command we prepared earlier but with a real image
        # 'ghcr.io/mycelis/proto-builder' doesn't exist.
        # We will use 'jaegertracing/protobuf' or similar which usually has standard tools.
        # Or better, let's use a standard alpine image and install protoc + protoc-gen-go quickly? 
        # No, that's repeated cost.
        
        # We will assume the user has docker.
        # We'll use `namely/protoc-all`
        
        click.echo("Running protoc via Docker (namely/protoc-all)...")
        # namely/protoc-all input is usually a directory.
        # Let's try to run the raw protoc command inside a container that has it.
        # 'bufbuild/buf' is increasingly standard.
        
        # Workaround: For this environment, we'll try 'rvolosatovs/protoc' which is a tiny wrapper.
        # It mirrors local protoc args.
        
        # We need the go plugins (protoc-gen-go) which might not be in base images.
        # 'kyleconroy/sqlc' ? No.
        
        # Robust Strategy: 'golang' image, install tools, gen.
        # Cache it? 
        
        # Let's look at the command string again.
        # We need: protoc --go_out=. --go_opt=module=...
        
        # Robust Strategy: 'golang' image, install tools, gen.
        
        # Windows handling for subprocess with list of args is better than shell=True for complex quotes
        # But we need sh -c inside the container.
        # We will use a script file inside /workspace if needed, or simple one-liner with minimal quotes.
        
        # Using a list for docker run to avoid local shell parsing issues
        docker_args = [
            "docker", "run", "--rm",
            "-v", f"{ROOT_DIR}:/workspace",
            "-w", "/workspace",
            "golang:1.24",
            "sh", "-c",
            "apt-get update && apt-get install -y protobuf-compiler && "
            "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && "
            "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && "
            "protoc --go_out=core --go_opt=module=github.com/mycelis/core "
            "--go-grpc_out=core --go-grpc_opt=module=github.com/mycelis/core "
            "proto/swarm/v1/swarm.proto"
        ]
        
        click.echo("Running dockerized protoc...")
        # We use run_cmd which is shell=True. This causes the list to be joined or is that only if it is a string?
        # run_cmd expects a string if it uses shell=True effectively.
        # Let's bypass run_cmd for this complex command to use subprocess.run with list (shell=False)
        try:
            subprocess.run(docker_args, check=True)
        except subprocess.CalledProcessError as e:
            click.echo(click.style(f"Docker protoc failed: {e}", fg="red"))
            sys.exit(e.returncode)

@cli.command(name="build-go") # Explicit name to match Makefile usage if preferred, or fix Makefile
def build_go():
    """Build the Go Core binary."""
    click.echo("Building Go Core...")
    # Fix: CWD logic in run_cmd might be failing if it relies on string concatenation
    # Re-verify run_cmd logic or just explicit path
    run_cmd("go build -o bin/server ./cmd/server", cwd=str(CORE_DIR))

@cli.command(name="proto-py")
def proto_py():
    """Generate Python stubs from Protobuf definitions."""
    click.echo("Generating Python stubs...")
    
    # 1. Ensure Output Dir
    out_dir = ROOT_DIR / "sdk/python/src/relay/proto"
    out_dir.mkdir(parents=True, exist_ok=True)
    
    # 2. Run protoc via grpc_tools (installed via uv dependency)
    # We run against the 'proto' directory as include root
    # Output to sdk/python/src/relay/proto
    
    proto_file = "proto/swarm/v1/swarm.proto"
    
    cmd = [
        sys.executable, "-m", "grpc_tools.protoc",
        "-Iproto",
        f"--python_out={out_dir}",
        f"--grpc_python_out={out_dir}",
        proto_file
    ]
    
    click.echo(f"Running: {' '.join(str(c) for c in cmd)}")
    try:
        subprocess.run(cmd, check=True, cwd=ROOT_DIR)
    except subprocess.CalledProcessError as e:
        click.echo(click.style(f"Protoc failed: {e}", fg="red"))
        sys.exit(e.returncode)

    # 3. Fix Relative Imports (Python Protoc quirk)
    # The generated file 'swarm_pb2_grpc.py' imports 'swarm.v1.swarm_pb2' 
    # but in our package structure it needs to be relative or absolute to the package.
    # Since we output to sdk/python/src/relay/proto, the package is 'relay.proto.swarm.v1' effectively?
    # Wait, the protoc output usually mirrors the proto package structure if not careful.
    # 'proto/swarm/v1/swarm.proto' -> 'sdk/python/src/relay/proto/swarm/v1/swarm_pb2.py'
    # Our include path is '-Iproto', so the generated file structure inside out_dir will conform to 'swarm/v1/'.
    
    # Let's inspect what we typically get.
    # If we generated into '.../relay/proto', we likely have '.../relay/proto/swarm/v1/swarm_pb2.py'
    # The grpc file will output 'import swarm.v1.swarm_pb2' which fails if 'swarm' is not in python path.
    # harmonizing to 'from . import swarm_pb2' works if they are in same dir.
    
    # We'll use a precise fix for the generated grpc file.
    # It will be at: sdk/python/src/relay/proto/swarm/v1/swarm_pb2_grpc.py
    
    grpc_file = out_dir / "swarm" / "v1" / "swarm_pb2_grpc.py"
    if grpc_file.exists():
        click.echo(f"Fixing imports in {grpc_file}...")
        text = grpc_file.read_text(encoding="utf-8")
        # Replace absolute import with relative/local
        # Pattern match might be needed, but usually it's static string.
        # "import swarm.v1.swarm_pb2 as swarm__pb2" -> "from . import swarm_pb2 as swarm__pb2"
        # Or similar.
        
        # We'll try a generic replacement that covers common cases or just the specific one.
        # Based on prev Makefile attempt: 'import swarm.v1.swarm_pb2 as'
        
        new_text = text.replace("import swarm.v1.swarm_pb2 as", "from . import swarm_pb2 as")
        grpc_file.write_text(new_text, encoding="utf-8")
    
    # Also verify __init__.py existence in subdirs if needed?
    # Python 3 namespace packages might handle it, but good practice to be safe.
    pass

@cli.command(name="test-python")
def test_python():
    """Run Python SDK Unit Tests."""
    click.echo("Running Python SDK Tests...")
    # Using sys.executable to ensure we run in the same uv environment
    cmd = [sys.executable, "-m", "pytest", "sdk/python/tests"]
    try:
        subprocess.run(cmd, check=True, cwd=ROOT_DIR)
    except subprocess.CalledProcessError as e:
        sys.exit(e.returncode)

@cli.command(name="verify-topology")
def verify_topology():
    """Run Topology Verification Script (Relay -> NATS -> Core)."""
    click.echo("Running Topology Verification...")
    cmd = [sys.executable, "scripts/verify_team_topology.py"]
    try:
        subprocess.run(cmd, check=True, cwd=ROOT_DIR)
    except subprocess.CalledProcessError as e:
        sys.exit(e.returncode)

@cli.command()
def watch():
    """Watch for changes and rebuild."""
    # TODO: Implement watchdog logic
    pass

if __name__ == "__main__":
    cli()
