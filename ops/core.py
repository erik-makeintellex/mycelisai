from invoke import task, Collection
from .config import CORE_DIR, is_windows

@task
def test(c):
    """Run Go Core Unit Tests."""
    with c.cd(str(CORE_DIR)):
        c.run("go test ./...")

@task
def clean(c):
    """Clean Go Build Artifacts."""
    print("Cleaning Core...")
    with c.cd(str(CORE_DIR)):
        c.run("go clean")
        if (CORE_DIR / "bin").exists():
           import shutil
           shutil.rmtree(str(CORE_DIR / "bin"))

@task
def build(c):
    """Build Go Core Binary (Local)."""
    print("Building Core (Local)...")
    with c.cd(str(CORE_DIR)):
        c.run("go build -v -o bin/server.exe ./cmd/server" if is_windows() else "go build -v -o bin/server ./cmd/server")

ns = Collection("core")
ns.add_task(test)
ns.add_task(clean)
ns.add_task(build)
