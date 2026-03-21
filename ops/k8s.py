from invoke import task, Collection
from .config import CLUSTER_NAME, NAMESPACE, is_windows, ROOT_DIR
from .core import build as core_build


def _cluster_exists(c) -> bool:
    cmd = (
        f"kind get clusters | findstr {CLUSTER_NAME}"
        if is_windows()
        else f"kind get clusters | grep {CLUSTER_NAME}"
    )
    result = c.run(cmd, hide=True, warn=True)
    return bool(result and result.ok and CLUSTER_NAME in result.stdout)


def _resource_exists(c, resource: str) -> bool:
    result = c.run(f"kubectl get {resource} -n {NAMESPACE}", hide=True, warn=True)
    return bool(result and result.ok)


def _wait_rollout(c, resource: str, timeout_seconds: int = 180, required: bool = True) -> bool:
    if not _resource_exists(c, resource):
        if required:
            print(f"  ERROR: Missing required resource: {resource}")
        else:
            print(f"  Skipping optional resource (not found): {resource}")
        return False

    print(f"  Waiting for {resource} rollout...")
    timeout = f"{int(timeout_seconds)}s"
    result = c.run(
        f"kubectl rollout status {resource} -n {NAMESPACE} --timeout={timeout}",
        warn=True,
    )
    if result.ok:
        print(f"  {resource} ready")
        return True
    print(f"  ERROR: {resource} failed readiness check within {timeout}")
    return False

@task
def init(c):
    """
    Initialize Infrastructure Layer.
    Creates Kind cluster and applies persistent infra (NATS, DBs).
    """
    print(f"Initializing Cluster: {CLUSTER_NAME}...")
    if _cluster_exists(c):
        print("Cluster exists.")
    else:
        # Check if we need to hydrate absolute paths for Windows Kind
        with open("kind-config.yaml", "r") as f:
            config = f.read()
        
        # Replace relative paths with absolute
        # We assume relative paths start with ./ops/
        abs_ops = str(ROOT_DIR / "ops").replace("\\", "/")
        # Basic substitution - robust enough for this specific file
        config = config.replace("./ops", abs_ops)
        
        # Fix Logs path too
        abs_logs = str(ROOT_DIR / "logs").replace("\\", "/")
        config = config.replace("./logs", abs_logs)

        # Write temp
        with open("kind-config.gen.yaml", "w") as f:
            f.write(config)

        print(f"Generated absolute config at kind-config.gen.yaml")
        c.run(f"kind create cluster --name {CLUSTER_NAME} --config kind-config.gen.yaml")
    
    
    # Legacy raw manifests removed. Helm chart handles infra.
    print("Infrastructure Ready for Helm.")


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
    import shlex
    load_dotenv(os.path.join(ROOT_DIR, ".env")) # Explicit path
    
    pg_user = os.getenv("POSTGRES_USER", "mycelis")
    pg_pass = os.getenv("POSTGRES_PASSWORD", "password")
    pg_db = os.getenv("POSTGRES_DB", "cortex")
    api_key = os.getenv("MYCELIS_API_KEY", "")
    if not api_key:
        raise SystemExit("MYCELIS_API_KEY must be set in .env or shell before deploying the cluster.")

    print(f"   Injecting Secrets for DB User: {pg_user}")

    cmd = (
        "helm upgrade --install mycelis-core ./charts/mycelis-core "
        f"--namespace {NAMESPACE} --create-namespace "
        f"--set image.tag={tag} "
        f"--set postgresql.auth.username={pg_user} "
        f"--set postgresql.auth.password={pg_pass} "
        f"--set postgresql.auth.database={pg_db} "
        f"--set coreAuth.apiKey={shlex.quote(api_key)} "
        "--wait"
    )
    c.run(cmd)
    
    # 4. Restart to ensure fresh config
    c.run(f"kubectl rollout restart deployment/mycelis-core -n {NAMESPACE}")
    print(f"Deployment Complete ({tag}).")


@task(help={"timeout": "Seconds to wait per rollout readiness check (default: 180)."})
def wait(c, timeout=180):
    """
    Wait for in-cluster dependency order to be healthy.
    Order: PostgreSQL -> NATS -> Core API deployment.
    """
    timeout = int(timeout)
    print(f"Waiting for cluster services in namespace '{NAMESPACE}'...")

    ns_result = c.run(f"kubectl get namespace {NAMESPACE}", hide=True, warn=True)
    if not ns_result.ok:
        print(f"ERROR: Namespace '{NAMESPACE}' not found. Run 'uv run inv k8s.init' first.")
        raise SystemExit(1)

    pg_ready = _wait_rollout(c, "statefulset/mycelis-core-postgresql", timeout_seconds=timeout)

    # NATS chart shape can vary by chart version (statefulset or deployment).
    nats_ready = _wait_rollout(
        c, "statefulset/mycelis-core-nats", timeout_seconds=timeout, required=False
    )
    if not nats_ready:
        nats_ready = _wait_rollout(
            c, "deployment/mycelis-core-nats", timeout_seconds=timeout, required=False
        )

    core_ready = _wait_rollout(c, "deployment/mycelis-core", timeout_seconds=timeout)

    if not (pg_ready and nats_ready and core_ready):
        print("\nCluster readiness failed. Inspect pods with:")
        print(f"  kubectl get pods -n {NAMESPACE}")
        print(f"  kubectl describe pods -n {NAMESPACE}")
        raise SystemExit(1)

    print("\nCluster ready (PostgreSQL, NATS, Core API).")


@task(help={"timeout": "Seconds to wait per rollout readiness check (default: 180)."})
def up(c, timeout=180):
    """
    Canonical cluster bring-up sequence.
    Order: init -> deploy -> wait.
    """
    print("=== Mycelis Cluster Up ===\n")
    print("[1/3] Ensure Kind cluster + namespace...")
    init(c)
    print("\n[2/3] Deploy Helm release...")
    deploy(c)
    print("\n[3/3] Wait for service readiness...")
    wait(c, timeout=timeout)
    print("\nCluster bring-up complete.")
    print("Next steps:")
    print("  1) uv run inv k8s.bridge")
    print("  2) uv run inv db.migrate")
    print("  3) uv run inv lifecycle.up --frontend")


@task
def bridge(c):
    """
    Open Development Bridge.
    Forwards cluster ports (NATS:4222, HTTP:8080, PG:5432) to localhost.
    """
    print("Starting Port-Forward Proxy (NATS:4222, HTTP:8080, PG:5432)...")
    if is_windows():
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/mycelis-core-nats 4222:4222")
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/mycelis-core 8080:8080")
        c.run(f"start kubectl port-forward -n {NAMESPACE} svc/mycelis-core-postgresql 5432:5432")
    else:
        p1 = c.run(f"kubectl port-forward -n {NAMESPACE} svc/mycelis-core-nats 4222:4222", asynchronous=True)
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
    # Core API
    c.run(f"kubectl rollout restart deployment/mycelis-core -n {NAMESPACE}", warn=True)
    # NATS can be statefulset or deployment depending on chart version
    if _resource_exists(c, "statefulset/mycelis-core-nats"):
        c.run(f"kubectl rollout restart statefulset/mycelis-core-nats -n {NAMESPACE}", warn=True)
    elif _resource_exists(c, "deployment/mycelis-core-nats"):
        c.run(f"kubectl rollout restart deployment/mycelis-core-nats -n {NAMESPACE}", warn=True)
    # PostgreSQL (statefulset)
    if _resource_exists(c, "statefulset/mycelis-core-postgresql"):
        c.run(f"kubectl rollout restart statefulset/mycelis-core-postgresql -n {NAMESPACE}", warn=True)
    print("Restart signals sent.")

@task
def reset(c):
    """
    Reset Infrastructure (Teardown + Canonical Bring-Up).
    Standardizes the development environment.
    """
    print("Resetting Infrastructure...")
    
    # 1. Teardown
    print("Stopping Cluster...")
    c.run(f"kind delete cluster --name {CLUSTER_NAME}")
    
    # 2. Canonical bring-up
    up(c)
    
    print("Infrastructure Reset Complete.")
    print("Run 'inv k8s.status' to verify.")

ns = Collection("k8s")
ns.add_task(init)
ns.add_task(deploy)
ns.add_task(wait)
ns.add_task(up)
ns.add_task(bridge)
ns.add_task(status)
ns.add_task(recover)
ns.add_task(reset)
