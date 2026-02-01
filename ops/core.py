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

from .version import get_version

@task
def build(c):
    """
    Build Go Core Binary + Docker Image.
    Returns the calculated TAG.
    """
    tag = get_version(c)
    print(f"📦 Building Artifact: mycelis/core:{tag}")
    
    # 1. Build Go Binary
    print("   Compiling Go Binary...")
    with c.cd(str(CORE_DIR)):
        c.run("go build -v -o bin/server.exe ./cmd/server" if is_windows() else "go build -v -o bin/server ./cmd/server")
    
    # 2. Build Docker Image
    print(f"   Building Container...")
    # Assumes Dockerfile is in core/Dockerfile and context is root
    c.run(f"docker build -t mycelis/core:{tag} -f core/Dockerfile .")
    
    # 3. Dev Tagging (Warning)
    c.run(f"docker tag mycelis/core:{tag} mycelis/core:latest")
    print("⚠️  Local 'latest' tag created for debugging only. Do NOT push to production.")
    
    return tag

ns = Collection("core")
ns.add_task(test)
ns.add_task(clean)
ns.add_task(build)
