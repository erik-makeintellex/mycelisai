from invoke import task, Collection
from .config import CORE_DIR, ROOT_DIR, ensure_managed_cache_dirs, is_windows, managed_cache_env


def _task_env(extra=None):
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


def _binary_output_path() -> str:
    return "bin/server.exe" if is_windows() else "bin/server"

@task
def test(c):
    """Run Go Core Unit Tests."""
    with c.cd(str(CORE_DIR)):
        c.run("go test ./...", env=_task_env())


@task
def compile(c):
    """Compile the Go Core binary without building a container image."""
    print("Compiling Core binary...")
    with c.cd(str(CORE_DIR)):
        c.run(
            f"go build -v -o {_binary_output_path()} ./cmd/server",
            env=_task_env(),
        )

@task
def clean(c):
    """Clean Go Build Artifacts."""
    print("Cleaning Core...")
    with c.cd(str(CORE_DIR)):
        c.run("go clean", env=_task_env())
        if (CORE_DIR / "bin").exists():
           import shutil
           shutil.rmtree(str(CORE_DIR / "bin"))

from .version import get_version

@task
def build(c):
    """
    Build Go Core Binary + Docker Image.
    Returns the calculated TAG.
    """
    tag = get_version(c)
    print(f"Building Artifact: mycelis/core:{tag}")

    # 1. Build Go Binary
    compile.body(c)

    # 2. Build Docker Image
    print(f"   Building Container...")
    # Assumes Dockerfile is in core/Dockerfile and context is root
    c.run(f"docker build -t mycelis/core:{tag} -f core/Dockerfile .")

    return tag

def _load_env():
    """Load .env into the process environment for local execution.
    Uses override=True so .env values win over system env vars
    (e.g. Windows User OLLAMA_HOST=0.0.0.0 bind address).
    """
    from dotenv import load_dotenv
    load_dotenv(str(ROOT_DIR / ".env"), override=True)

@task
def run(c):
    """
    Run the Core Service locally (Native).
    Stops any existing instance first to avoid port conflicts.
    """
    stop(c)
    _load_env()
    import os, sys
    os.environ["PYTHONIOENCODING"] = "utf-8"
    # Reconfigure stdout/stderr to UTF-8 so Go server emoji logs don't crash invoke
    if is_windows() and hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    print("Starting Mycelis Core (Native)...")
    with c.cd(str(CORE_DIR)):
        bin_name = "server.exe" if is_windows() else "server"
        env = _task_env({"PYTHONIOENCODING": "utf-8"})
        if is_windows():
            c.run(f"bin\\{bin_name}", pty=False, in_stream=False, env=env)
        else:
            c.run(f"./bin/{bin_name}", pty=True, env=env)

@task
def stop(c):
    """
    Stop the Core Service (Kill).
    """
    print("Stopping Core...")
    if is_windows():
        c.run("taskkill /F /IM server.exe", warn=True)
        c.run("taskkill /F /IM core.exe", warn=True) # Handle legacy naming
    else:
        c.run("pkill server", warn=True)
        c.run("pkill core", warn=True)

@task
def restart(c):
    """
    Restart the Core Service (Kill + Run).
    """
    print("Restarting Core...")
    stop(c)
    run(c)

@task
def smoke(c):
    """
    Run Governance Smoke Tests (Go).
    """
    print("Running Smoke Tests...")
    with c.cd(str(CORE_DIR)):
        c.run("go run ./cmd/smoke/main.go", env=_task_env())

ns = Collection("core")
ns.add_task(test)
ns.add_task(compile)
ns.add_task(clean)
ns.add_task(build)
ns.add_task(run)
ns.add_task(stop)
ns.add_task(restart)
ns.add_task(smoke)
