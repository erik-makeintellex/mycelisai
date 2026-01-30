from invoke import task, Collection
import os
import sys
import platform
from pathlib import Path

# -- Config --
CLUSTER_NAME = "mycelis-cluster"
NAMESPACE = "mycelis"
ROOT_DIR = Path(__file__).parent.resolve()
CORE_DIR = ROOT_DIR / "core"
SDK_DIR = ROOT_DIR / "sdk/python"

def is_windows():
    return platform.system() == "Windows"

# -- PROTO Namespace --
@task
def proto_generate(c):
    """Generate Go and Python Protobuf stubs."""
    print("Generating Protobufs...")
    
    # 1. Go Generation (Dockerized)
    # Using a standard image approach similar to previous steps
    # We construct the command to run inside docker
    docker_pkg = "golang:1.25" 
    script_content = (
        "apt-get update && apt-get install -y protobuf-compiler && "
        "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && "
        "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && "
        "protoc --go_out=core --go_opt=module=github.com/mycelis/core "
        "--go-grpc_out=core --go-grpc_opt=module=github.com/mycelis/core "
        "proto/swarm/v1/swarm.proto"
    )
    
    # Write to temp file in workspace
    script_path = ROOT_DIR / "scripts" / "gen_proto_go.sh"
    if not script_path.parent.exists():
        script_path.parent.mkdir(parents=True)
        
    script_path.write_text(script_content, encoding="utf-8")
    
    # Run helper script inside container
    cmd_go = f"docker run --rm -v {ROOT_DIR}:/workspace -w /workspace {docker_pkg} sh scripts/gen_proto_go.sh"
    print(f"Running Go Gen: {cmd_go}")
    try:
        c.run(cmd_go)
    finally:
        # Cleanup
        if script_path.exists():
            script_path.unlink()

    # 2. Python Generation (Local via uv)
    out_dir = SDK_DIR / "src/relay/proto"
    if not out_dir.exists():
        out_dir.mkdir(parents=True, exist_ok=True)
        
    cmd_py = (
        f"uv run --with grpcio-tools --with protobuf -m grpc_tools.protoc "
        f"-Iproto "
        f"--python_out={out_dir} "
        f"--grpc_python_out={out_dir} "
        f"proto/swarm/v1/swarm.proto"
    )
    print(f"Running Python Gen: {cmd_py}")
    c.run(cmd_py)

    # Filesystem operations for Python
    init_file = out_dir / "__init__.py"
    if not init_file.exists():
        init_file.touch()
        
    grpc_file = out_dir / "swarm" / "v1" / "swarm_pb2_grpc.py"
    if grpc_file.exists():
        text = grpc_file.read_text(encoding="utf-8")
        new_text = text.replace("import swarm.v1.swarm_pb2 as", "from . import swarm_pb2 as")
        grpc_file.write_text(new_text, encoding="utf-8")

ns_proto = Collection("proto")
ns_proto.add_task(proto_generate, "generate")


# -- K8S Namespace - Split Lifecycle --
@task
def k8s_init(c):
    """
    Initialize Infrastructure Layer.
    Creates Kind cluster and applies persistent infra (NATS, DBs).
    Does NOT apply Core logic.
    """
    print(f"Initializing Cluster: {CLUSTER_NAME}...")
    # 1. Create Cluster (Idempotent)
    try:
        c.run(f"kind get clusters | findstr {CLUSTER_NAME}" if is_windows() else f"kind get clusters | grep {CLUSTER_NAME}", hide=True)
        print("Cluster exists.")
    except:
        c.run(f"kind create cluster --name {CLUSTER_NAME} --config kind-config.yaml")
    
    # 2. Apply Infra Layer
    # Assuming k8s/01-nats.yaml exists. If not, we might fail or imply it.
    # We'll try to apply the directory or specific files if we knew them.
    # User specified k8s/01-nats.yaml ideally. based on current filesystem we have k8s/.
    # We will just apply k8s/01-nats.yaml explicitly as requested.
    print("Applying Infrastructure Layer...")
    # We use -f k8s/ if files are not strictly named yet, or specific if they are.
    # To be safe based on previous 'k8s/' content, we assume k8s/ contains the yaml.
    # But user asked for split. We'll attempt specific file, fallback to dir if needed?
    # No, strict instructions: "Applies k8s/01-nats.yaml".
    c.run("kubectl apply -f k8s/01-nats.yaml")

@task
def k8s_deploy(c):
    """
    Deploy Application Layer.
    Builds Core, Loads to Kind, Applies Core manifests, Restarts.
    """
    print("Deploying Application Layer...")
    
    # 1. Build
    c.run("docker build -t mycelis/core:latest -f core/Dockerfile .")
    
    # 2. Load
    c.run(f"kind load docker-image mycelis/core:latest --name {CLUSTER_NAME}")
    
    # 3. Apply Core
    c.run("kubectl apply -f k8s/02-core.yaml")
    
    # 4. Restart
    # Ignore error if not exists yet
    try:
        c.run(f"kubectl rollout restart deployment/neural-core -n {NAMESPACE}")
    except:
        pass

@task
def k8s_bridge(c):
    """
    Open Development Bridge.
    Forwards cluster ports (NATS:4222, HTTP:8080) to localhost.
    WARNING: This blocks the terminal.
    """
    print("Opening Bridge to Cluster...")
    print("NATS: 127.0.0.1:4222 -> svc/nats")
    print("CORE: 127.0.0.1:8080 -> svc/neural-core")
    
    # We can try to use background tasks or just run sequentially?
    # User said "Use invoke.run(..., asynchronous=True) or warn user".
    # Since we have multiple port forwards, async is better or we just block on one.
    # Windows/Kind port forwarding can be tricky. 
    # Usually `kubectl port-forward` blocks.
    
    # We'll launch the first one async and block on the second?
    # Or just block on NATS if that's the primary need for some tests.
    
    if is_windows():
        print("!! WINDOWS DETECTED !!")
        print("Please run these in separate terminals manually if this hangs:")
        print(f"kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222")
        print(f"kubectl port-forward -n {NAMESPACE} svc/neural-core 8080:8080")
    
    # Attempting synchronous block (easiest for single tunnel, harder for dual)
    # Using 'start' command on windows to spawn new windows?
    if is_windows():
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222")
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/neural-core 8080:8080")
    else:
        p1 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222", asynchronous=True)
        p2 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/neural-core 8080:8080", asynchronous=True)
        print("Bridge active. Press Ctrl+C to stop (if blocking) or kill processes manually.")
        p1.join() 
        p2.join()

ns_k8s = Collection("k8s")
ns_k8s.add_task(k8s_init, "init")
ns_k8s.add_task(k8s_deploy, "deploy")
ns_k8s.add_task(k8s_bridge, "bridge")


# -- CORE Namespace --
@task
def core_test(c):
    """Run Go Core Unit Tests."""
    with c.cd(str(CORE_DIR)):
        c.run("go test ./...")

@task
def core_clean(c):
    """Clean Go Build Artifacts."""
    print("Cleaning Core...")
    with c.cd(str(CORE_DIR)):
        c.run("go clean")
        # Optional: remove binary if we usually output to bin/
        if (CORE_DIR / "bin").exists():
           import shutil
           shutil.rmtree(str(CORE_DIR / "bin"))

@task
def core_build(c):
    """Build Go Core Binary (Local)."""
    print("Building Core (Local)...")
    with c.cd(str(CORE_DIR)):
        # Builds to /bin/server locally
        c.run("go build -v -o bin/server.exe ./cmd/server" if is_windows() else "go build -v -o bin/server ./cmd/server")

ns_core = Collection("core")
ns_core.add_task(core_test, "test")
ns_core.add_task(core_clean, "clean")
ns_core.add_task(core_build, "build")


# -- RELAY Namespace --
@task
def relay_test(c):
    """Run Python Relay SDK Tests."""
    # Assuming location of tests
    c.run("uv run pytest sdk/python")

@task
def relay_demo(c):
    """Run the Reference Worker Agent."""
    print("Starting Reference Worker...")
    # Ensure PYTHONPATH includes the SDK so we can import 'relay'
    # Windows/Linux separator difference handled by env
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/reference_worker.py", env=env)

ns_relay = Collection("relay")
ns_relay.add_task(relay_test, "test")
ns_relay.add_task(relay_demo, "demo")


# -- CLEAN Namespace --
@task
def clean_legacy(c):
    """Remove legacy build files (Makefile, docker-compose)."""
    print("Cleaning legacy artifacts...")
    legacy_files = [
        ROOT_DIR / "Makefile", 
        ROOT_DIR / "Makefile.legacy",
        ROOT_DIR / "docker-compose.yml",
        ROOT_DIR / "deploy/docker/compose.yml" # Checking potential locations
    ]
    
    for p in legacy_files:
        if p.exists():
            print(f"Removing {p}")
            p.unlink()
        else:
            print(f"Skipped {p} (Not found)")

ns_clean = Collection("clean")
ns_clean.add_task(clean_legacy, "legacy")


# -- TEAM Namespace --
@task
def team_sensors(c):
    """Run Sensor Manager (Telemetry Generator)."""
    print("Starting Sensor Manager...")
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/sensors/manager.py", env=env)

@task
def team_output(c):
    """Run Output Manager (Display)."""
    print("Starting Output Manager...")
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/output/manager.py", env=env)

@task
def team_test(c):
    """Run Team Agent Unit Tests."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run pytest agents/tests", env=env)

ns_team = Collection("team")
ns_team.add_task(team_sensors, "sensors")
ns_team.add_task(team_output, "output")
ns_team.add_task(team_test, "test")


# -- ROOT --
ns = Collection()
ns.add_collection(ns_proto)
ns.add_collection(ns_k8s)
ns.add_collection(ns_core)
ns.add_collection(ns_relay)
ns.add_collection(ns_clean)
ns.add_collection(ns_team)
