from invoke import task, Collection
from .config import CLUSTER_NAME, NAMESPACE, is_windows, ROOT_DIR

@task
def init(c):
    """
    Initialize Infrastructure Layer.
    Creates Kind cluster and applies persistent infra (NATS, DBs).
    """
    print(f"Initializing Cluster: {CLUSTER_NAME}...")
    try:
        c.run(f"kind get clusters | findstr {CLUSTER_NAME}" if is_windows() else f"kind get clusters | grep {CLUSTER_NAME}", hide=True)
        print("Cluster exists.")
    except:
        c.run(f"kind create cluster --name {CLUSTER_NAME} --config kind-config.yaml")
    
    print("Applying Infrastructure Layer...")
    c.run("kubectl apply -f legacy_archive/k8s_raw/00-namespace.yaml")
    c.run("kubectl apply -f legacy_archive/k8s_raw/01-nats.yaml")

from .version import get_version
from .core import build as core_build

@task
def deploy(c):
    """
    Deploys the core using Helm (Hardened Security).
    Uses Immutable Identity Tagging.
    """
    # 1. Build Artifact (Delegated to Core)
    tag = core_build(c)
    
    print(f"🚀 Deploying Release: [{tag}]")
    
    print("📦 Building Helm Dependencies...")
    c.run("helm dependency update ./charts/mycelis-core")
    
    # 2. Load into Kind
    print(f"   Loading Image into Cluster...")
    c.run(f"kind load docker-image mycelis/core:{tag} --name {CLUSTER_NAME}")
    
    # 3. Helm Upgrade (Atomic with Tag Override)
    print("   Applying Helm Chart...")
    cmd = (
        "helm upgrade --install mycelis-core ./charts/mycelis-core "
        f"--namespace {NAMESPACE} --create-namespace "
        f"--set image.tag={tag} "
        "--wait"
    )
    c.run(cmd)
    
    # 4. Restart to ensure fresh config
    c.run(f"kubectl rollout restart deployment/mycelis-core -n {NAMESPACE}")
    print(f"✅ Deployment Complete ({tag}).")

@task
def bridge(c):
    """
    Open Development Bridge.
    Forwards cluster ports (NATS:4222, HTTP:8080) to localhost.
    """
    print("Opening Bridge to Cluster...")
    if is_windows():
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222")
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/mycelis-core 8080:8080")
    else:
        p1 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222", asynchronous=True)
        p2 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/mycelis-core 8080:8080", asynchronous=True)
        print("Bridge active. Press Ctrl+C to stop.")
        p1.join() 
        p2.join()

@task
def status(c):
    """Check the health of the entire stack."""
    print(f"📊 Checking Infrastructure Status for {CLUSTER_NAME}...")
    try:
        c.run("docker info", hide=True)
        print("✅ Docker: Running")
    except:
        print("❌ Docker: NOT Running.")
        return

    try:
        cluster_info = c.run(f"kind get clusters | findstr {CLUSTER_NAME}" if is_windows() else f"kind get clusters | grep {CLUSTER_NAME}", hide=True)
        if CLUSTER_NAME in cluster_info.stdout:
            print(f"✅ Kind Cluster ({CLUSTER_NAME}): Active")
        else:
            print(f"❌ Kind Cluster: Not Found")
            return
    except:
         print(f"❌ Kind Cluster: Error checking")
         return

    print("\n📦 Pod Status:")
    c.run(f"kubectl get pods -n {NAMESPACE}")

@task
def recover(c):
    """Attempt to heal the cluster."""
    print("🚑 Attempting Recovery...")
    c.run(f"kubectl rollout restart deployment/nats -n {NAMESPACE}")
    c.run(f"kubectl rollout restart deployment/neural-core -n {NAMESPACE}")
    print("✅ Restart signals sent.")

ns = Collection("k8s")
ns.add_task(init)
ns.add_task(deploy)
ns.add_task(bridge)
ns.add_task(status)
ns.add_task(recover)
