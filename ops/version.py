from invoke import run
from .config import ROOT_DIR

def get_version(c=None):
    """
    Calculates the immutable identity tag: v{MAJOR}.{MINOR}.{PATCH}-{SHORT_SHA}
    """
    # 1. Read VERSION file
    version_file = ROOT_DIR / "VERSION"
    if not version_file.exists():
        raise FileNotFoundError("VERSION file missing in project root")
    
    semver = version_file.read_text().strip()
    
    # 2. Get Git SHA
    # Use invoke context if provided, otherwise subprocess (but try to use c if possible for consistency)
    if c:
        res = c.run("git rev-parse --short HEAD", hide=True)
        sha = res.stdout.strip()
    else:
        # Fallback if no context (though tasks usually provide one)
        import subprocess
        sha = subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], cwd=str(ROOT_DIR)).decode().strip()
            
    # 3. Form Tag
    tag = f"v{semver}-{sha}"
    return tag
