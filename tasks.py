from invoke import Collection, task

from ops import auth
from ops import cache
from ops import ci
from ops import cognitive
from ops import compose
from ops import core
from ops import db
from ops import device
from ops import interface
from ops import k8s
from ops import lifecycle
from ops import logging as logging_tasks
from ops import misc
from ops import proto_relay
from ops import quality
from ops import test
from ops import wsl_runtime

ns = Collection()


@task(
    help={
        "optional_engines": "Also install optional local cognitive engine dependencies from cognitive/ (default: False).",
    }
)
def install(c, optional_engines=False):
    """
    Install the default local development dependencies for the supported Core + Interface stack.
    Assumes the shell/session is already running with admin rights.
    """
    print("Installing workspace Python dependencies...")
    from ops.config import ensure_managed_cache_dirs, managed_cache_env
    from ops.cache import ensure_disk_headroom

    ensure_managed_cache_dirs()
    ensure_disk_headroom(min_free_gb=12, reason="install")
    env = managed_cache_env()
    c.run("uv sync --all-packages --dev", env=env)

    print("Installing Go module dependencies...")
    with c.cd("core"):
        c.run("go mod download", env=env)

    print("Installing Interface dependencies...")
    c.run("npm install --prefix interface", env=env)

    if optional_engines:
        print("Installing optional cognitive engine dependencies...")
        with c.cd("cognitive"):
            c.run("uv sync", env=env)
    else:
        print("Skipping optional cognitive engine dependencies.")
        print("  Use `uv run inv install --optional-engines` or `uv run inv cognitive.install` when you need local vLLM/Diffusers helpers.")

    print("Install complete.")


ns.add_task(install)
ns.add_collection(cache.ns)
ns.add_collection(compose.ns)
ns.add_collection(core.ns)
ns.add_collection(k8s.ns)
ns.add_collection(proto_relay.ns_proto)
ns.add_collection(proto_relay.ns_relay)
ns.add_collection(misc.ns_clean)
ns.add_collection(misc.ns_team)
ns.add_collection(interface.ns)
ns.add_collection(auth.ns)
ns.add_collection(logging_tasks.ns)
ns.add_collection(quality.ns)
ns.add_collection(device.ns)
ns.add_collection(test.ns)
ns.add_collection(db.ns)
ns.add_collection(ci.ns)
ns.add_collection(cognitive.ns)
ns.add_collection(lifecycle.ns)
ns.add_collection(wsl_runtime.ns)

