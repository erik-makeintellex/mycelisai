from invoke import task, Collection
from .config import ROOT_DIR, SDK_DIR

# -- PROTO --
@task
def generate(c):
    """Generate Go and Python Protobuf stubs."""
    print("Generating Protobufs...")
    
    # 1. Go (Dockerized)
    docker_pkg = "golang:1.25" 
    script_content = (
        "apt-get update && apt-get install -y protobuf-compiler && "
        "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && "
        "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && "
        "protoc --go_out=core --go_opt=module=github.com/mycelis/core "
        "--go-grpc_out=core --go-grpc_opt=module=github.com/mycelis/core "
        "proto/swarm/v1/swarm.proto"
    )
    
    script_path = ROOT_DIR / "scripts" / "gen_proto_go.sh"
    if not script_path.parent.exists():
        script_path.parent.mkdir(parents=True)
    script_path.write_text(script_content, encoding="utf-8")
    
    cmd_go = f"docker run --rm -v {ROOT_DIR}:/workspace -w /workspace {docker_pkg} sh scripts/gen_proto_go.sh"
    try:
        c.run(cmd_go)
    finally:
        if script_path.exists():
            script_path.unlink()

    # 2. Python (Local)
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
    c.run(cmd_py)

    # Post-process Python
    grpc_file = out_dir / "swarm" / "v1" / "swarm_pb2_grpc.py"
    if grpc_file.exists():
        text = grpc_file.read_text(encoding="utf-8")
        new_text = text.replace("import swarm.v1.swarm_pb2 as", "from . import swarm_pb2 as")
        grpc_file.write_text(new_text, encoding="utf-8")

ns_proto = Collection("proto")
ns_proto.add_task(generate)

# -- RELAY --
@task
def test(c):
    """Run Python Relay SDK Tests."""
    c.run("uv run pytest sdk/python")

@task
def demo(c):
    """Run the Reference Worker Agent."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/reference_worker.py", env=env)

ns_relay = Collection("relay")
ns_relay.add_task(test)
ns_relay.add_task(demo)
