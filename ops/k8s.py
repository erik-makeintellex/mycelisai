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
    
    
    # Legacy raw manifests removed. Helm chart handles infra.
    print("Infrastructure Ready for Helm.")

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
    
    print(f"Deploying Release: [{tag}]")
    
    print("Building Helm Dependencies...")
    c.run("helm dependency update ./charts/mycelis-core")
    
    # 2. Load into Kind
    print(f"   Loading Image into Cluster...")
    c.run(f"kind load docker-image mycelis/core:{tag} --name {CLUSTER_NAME}")
    
    # 3. Helm Upgrade (Atomic with Tag Override)
    print("   Applying Helm Chart...")
    # Load .env secrets
    import os
    from dotenv import load_dotenv
    load_dotenv(os.path.join(ROOT_DIR, ".env")) # Explicit path
    
    pg_user = os.getenv("POSTGRES_USER", "mycelis")
    pg_pass = os.getenv("POSTGRES_PASSWORD", "password")
    pg_db = os.getenv("POSTGRES_DB", "cortex")

    print(f"   Injecting Secrets for DB User: {pg_user}")

    cmd = (
        "helm upgrade --install mycelis-core ./charts/mycelis-core "
        f"--namespace {NAMESPACE} --create-namespace "
        f"--set image.tag={tag} "
        f"--set postgresql.auth.username={pg_user} "
        f"--set postgresql.auth.password={pg_pass} "
        f"--set postgresql.auth.database={pg_db} "
        "--wait"
    )
    c.run(cmd)
    
    # 4. Restart to ensure fresh config
    c.run(f"kubectl rollout restart deployment/mycelis-core -n {NAMESPACE}")
    print(f"Deployment Complete ({tag}).")

@task
def bridge(c):
    """
    Open Development Bridge.
    Forwards cluster ports (NATS:4222, HTTP:8080) to localhost.
    """
    print("Starting Port-Forward Proxy (NATS:4222, HTTP:8080, PG:5432)...")
    if is_windows():
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222")
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/mycelis-core 8080:8080")
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/mycelis-core-postgresql 5432:5432")
    else:
        p1 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/nats 4222:4222", asynchronous=True)
        p2 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/mycelis-core 8080:8080", asynchronous=True)
        p3 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/mycelis-core-postgresql 5432:5432", asynchronous=True)
        print("Bridge active. Press Ctrl+C to stop.")
        p1.join() 
        p2.join()
        p3.join()

@task
def status(c):
    """Check the health of the entire stack."""
    print(f"Checking Infrastructure Status for {CLUSTER_NAME}...")
    try:
        c.run("docker info", hide=True)
        print("Docker: Running")
    except:
        print("Docker: NOT Running.")
        return

    try:
        cluster_info = c.run(f"kind get clusters | findstr {CLUSTER_NAME}" if is_windows() else f"kind get clusters | grep {CLUSTER_NAME}", hide=True)
        if CLUSTER_NAME in cluster_info.stdout:
            print(f"Kind Cluster ({CLUSTER_NAME}): Active")
        else:
            print(f"Kind Cluster: Not Found")
            return
    except:
         print(f"Kind Cluster: Error checking")
         return

    print("\nPod Status:")
    c.run(f"kubectl get pods -n {NAMESPACE}")
    
    print("\nPersistence (PVC) Status:")
    c.run(f"kubectl get pvc -n {NAMESPACE}")

@task
def recover(c):
    """Attempt to heal the cluster."""
    print("Attempting Recovery...")
    c.run(f"kubectl rollout restart deployment/nats -n {NAMESPACE}")
    c.run(f"kubectl rollout restart deployment/neural-core -n {NAMESPACE}")
    print("Restart signals sent.")

@task
def reset(c):
    """
    Reset Infrastructure (Teardown + Init + Deploy).
    Standardizes the development environment.
    """
    print("Resetting Infrastructure...")
    
    # 1. Teardown
    print("Stopping Cluster...")
    c.run(f"kind delete cluster --name {CLUSTER_NAME}")
    
    # 2. Init
    init(c)
    
    # 3. Deploy
    deploy(c)
    
    print("Infrastructure Reset Complete.")
    print("Run 'inv k8s.status' to verify.")

ns = Collection("k8s")
ns.add_task(init)
ns.add_task(deploy)
ns.add_task(bridge)
ns.add_task(status)
ns.add_task(recover)
ns.add_task(reset)
