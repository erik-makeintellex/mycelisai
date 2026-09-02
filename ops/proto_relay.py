import os
import tempfile
from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR, SDK_DIR

GO_TOOLCHAIN_IMAGE = "golang:1.26-bookworm@sha256:9fdc884aacc3bec89b20ffc69f4bb369c78210e3e4f600387b5128b12c199f81"
PROTOBUF_COMPILER_PACKAGE = "protobuf-compiler=3.21.12-3"
PROTOC_GEN_GO_VERSION = "v1.36.11"
GRPCIO_TOOLS_VERSION = "1.76.0"
PYTHON_PROTOBUF_VERSION = "6.33.5"


def _go_generation_script() -> str:
    return (
        "apt-get update && "
        f"apt-get install -y {PROTOBUF_COMPILER_PACKAGE} && "
        "go install google.golang.org/protobuf/cmd/protoc-gen-go@"
        f"{PROTOC_GEN_GO_VERSION} && "
        "protoc --go_out=core --go_opt=module=github.com/mycelis/core "
        "proto/swarm/v1/swarm.proto proto/envelope.proto"
    )


def _python_generator_prefix() -> str:
    return (
        f"uv run --with grpcio-tools=={GRPCIO_TOOLS_VERSION} "
        f"--with protobuf=={PYTHON_PROTOBUF_VERSION} -m grpc_tools.protoc"
    )


# -- PROTO --
@task
def generate(c):
    """Generate Go and Python Protobuf stubs."""
    print("Generating Protobufs...")

    script_content = _go_generation_script()
    temp_dir = ROOT_DIR / "workspace" / "tool-cache" / "proto"
    temp_dir.mkdir(parents=True, exist_ok=True)
    script_fd, script_name = tempfile.mkstemp(prefix="gen-proto-go-", suffix=".sh", dir=temp_dir)
    os.close(script_fd)
    script_path = Path(script_name)
    script_path.write_text(script_content, encoding="utf-8")

    script_rel = script_path.relative_to(ROOT_DIR).as_posix()
    cmd_go = f"docker run --rm -v {ROOT_DIR}:/workspace -w /workspace {GO_TOOLCHAIN_IMAGE} sh {script_rel}"
    try:
        c.run(cmd_go)
    finally:
        if script_path.exists():
            script_path.unlink()

    out_dir = SDK_DIR / "src/relay/proto"
    if not out_dir.exists():
        out_dir.mkdir(parents=True, exist_ok=True)
        
    cmd_py = (
        f"{_python_generator_prefix()} "
        f"-Iproto "
        f"--python_out={out_dir} "
        f"proto/swarm/v1/swarm.proto"
    )
    c.run(cmd_py)

    scip_out_dir = SDK_DIR / "src/scip/proto"
    if not scip_out_dir.exists():
        scip_out_dir.mkdir(parents=True, exist_ok=True)

    cmd_py_scip = (
        f"{_python_generator_prefix()} "
        f"-Iproto "
        f"--python_out={scip_out_dir} "
        f"proto/envelope.proto"
    )
    c.run(cmd_py_scip)

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
