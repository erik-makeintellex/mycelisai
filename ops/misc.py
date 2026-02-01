from invoke import task, Collection
from .config import ROOT_DIR

# -- CLEAN --
@task
def legacy(c):
    """Remove legacy build files."""
    legacy_files = [
        ROOT_DIR / "Makefile", 
        ROOT_DIR / "Makefile.legacy",
        ROOT_DIR / "docker-compose.yml"
    ]
    for p in legacy_files:
        if p.exists():
            p.unlink()
            print(f"Removed {p}")

ns_clean = Collection("clean")
ns_clean.add_task(legacy)

# -- TEAM --
@task
def sensors(c):
    """Run Sensor Manager."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/sensors/manager.py", env=env)

@task
def output(c):
    """Run Output Manager."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/output/manager.py", env=env)

@task
def test(c):
    """Run Team Agent Unit Tests."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run pytest agents/tests", env=env)

ns_team = Collection("team")
ns_team.add_task(sensors)
ns_team.add_task(output)
ns_team.add_task(test)
