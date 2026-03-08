from invoke import Collection, task

from ops import auth
from ops import ci
from ops import cognitive
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

ns = Collection()


@task
def install(c):
    """
    Install all local development dependencies.
    Assumes the shell/session is already running with admin rights.
    """
    print("Installing workspace Python dependencies...")
    c.run("uv sync --all-packages --dev")

    print("Installing Go module dependencies...")
    with c.cd("core"):
        c.run("go mod download")

    print("Installing Interface dependencies...")
    c.run("npm install --prefix interface")

    print("Installing Cognitive dependencies...")
    with c.cd("cognitive"):
        c.run("uv sync")

    print("Install complete.")


ns.add_task(install)
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

