import os
import shlex
import socket
import subprocess
import shutil
import time
from pathlib import Path

from invoke import task, Collection
from .config import CLUSTER_NAME, NAMESPACE, is_windows, ROOT_DIR
from .core import build as core_build
from .packaging import relative_to_root, resolve_repo_path, slugify_label, write_checksum_file, write_json
from .version import get_version


def _k8s_backend() -> str:
    requested = os.environ.get("MYCELIS_K8S_BACKEND", "auto").strip().lower()
    if requested not in {"", "auto", "k3d", "kind"}:
        raise SystemExit("Invalid MYCELIS_K8S_BACKEND. Use one of: auto, k3d, kind.")

    if requested in {"", "auto"}:
        if shutil.which("k3d"):
            return "k3d"
        if shutil.which("kind"):
            return "kind"
        raise SystemExit("No supported local Kubernetes backend found. Install k3d or kind.")

    if requested == "k3d" and not shutil.which("k3d"):
        raise SystemExit("MYCELIS_K8S_BACKEND=k3d requires k3d on PATH.")
    if requested == "kind" and not shutil.which("kind"):
        raise SystemExit("MYCELIS_K8S_BACKEND=kind requires kind on PATH.")
    return requested


def _cluster_list_has_name(output: str, backend: str) -> bool:
    for line in output.splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        if backend == "k3d":
            if stripped.lower().startswith("name "):
                continue
            if stripped.split()[0] == CLUSTER_NAME:
                return True
        elif stripped == CLUSTER_NAME:
            return True
    return False


def _cluster_exists(c) -> bool:
    backend = _k8s_backend()
    command = "k3d cluster list" if backend == "k3d" else "kind get clusters"
    result = c.run(command, hide=True, warn=True)
    if not result or not result.ok:
        return False
    return _cluster_list_has_name(result.stdout, backend)


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


def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def _wait_for_local_port(port: int, label: str, timeout_seconds: int = 30, interval_seconds: float = 1.0) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        if _port_open(port):
            return True
        time.sleep(interval_seconds)
    print(f"  ERROR: {label} did not bind localhost:{port} within {timeout_seconds}s")
    return False


def _chart_dir() -> Path:
    return ROOT_DIR / "charts" / "mycelis-core"


def _chart_version() -> str:
    chart_yaml = _chart_dir() / "Chart.yaml"
    for raw_line in chart_yaml.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if line.startswith("version:"):
            return line.split(":", 1)[1].strip().strip('"').strip("'")
    raise SystemExit(f"Unable to determine Helm chart version from {chart_yaml}")


def _resolve_k8s_values_file(explicit_path: str = "") -> Path | None:
    raw_value = explicit_path.strip() or os.environ.get("MYCELIS_K8S_VALUES_FILE", "").strip()
    if not raw_value:
        return None

    resolved = resolve_repo_path(raw_value, root_dir=ROOT_DIR)
    if not resolved.exists():
        raise SystemExit(f"MYCELIS_K8S_VALUES_FILE does not exist: {resolved}")
    return resolved


def _resolve_package_output_dir(raw_path: str = "") -> Path:
    destination = resolve_repo_path(raw_path, root_dir=ROOT_DIR) if raw_path.strip() else (ROOT_DIR / "dist" / "helm")
    destination.mkdir(parents=True, exist_ok=True)
    return destination


def _verify_and_package_chart(
    c,
    *,
    values_file: Path,
    release_label: str,
    package_output_dir: str,
) -> Path:
    chart_dir = _chart_dir()
    output_dir = _resolve_package_output_dir(package_output_dir)
    preset_name = values_file.stem.removeprefix("values-") or "default"
    release_slug = slugify_label(release_label)
    rendered_path = output_dir / f"mycelis-core-{preset_name}-{release_slug}.rendered.yaml"
    manifest_path = output_dir / f"mycelis-core-{preset_name}-{release_slug}.manifest.json"
    chart_version = _chart_version()
    chart_package_path = output_dir / f"mycelis-core-{chart_version}.tgz"

    quoted_chart_dir = shlex.quote(str(chart_dir))
    quoted_values_file = shlex.quote(str(values_file))
    quoted_output_dir = shlex.quote(str(output_dir))
    quoted_rendered_path = shlex.quote(str(rendered_path))
    quoted_release_label = shlex.quote(release_label)

    print(f"Verifying enterprise Helm package for preset '{preset_name}' as {release_label}...")
    c.run(f"helm dependency build {quoted_chart_dir}")
    c.run(f"helm lint {quoted_chart_dir} --values {quoted_values_file}")
    c.run(
        f"helm template mycelis-core {quoted_chart_dir} --namespace {NAMESPACE} --values {quoted_values_file} > {quoted_rendered_path}"
    )
    c.run(
        f"helm package {quoted_chart_dir} --destination {quoted_output_dir} --app-version {quoted_release_label}"
    )

    if not rendered_path.exists():
        raise SystemExit(f"Helm template verification did not produce the rendered bundle: {rendered_path}")
    if not chart_package_path.exists():
        raise SystemExit(f"Helm package verification did not produce the chart archive: {chart_package_path}")

    rendered_checksum_path = write_checksum_file(rendered_path)
    chart_checksum_path = write_checksum_file(chart_package_path)
    write_json(
        manifest_path,
        {
            "artifact_kind": "enterprise_helm_package",
            "chart_archive_checksum_path": relative_to_root(chart_checksum_path, root_dir=ROOT_DIR),
            "chart_archive_path": relative_to_root(chart_package_path, root_dir=ROOT_DIR),
            "chart_version": chart_version,
            "notes": [
                "This slice verifies Helm dependency hydration, lint, template rendering, and chart packaging for a concrete enterprise preset.",
                "Registry publication, signed provenance, and a full installer bundle remain follow-up work.",
            ],
            "preset": preset_name,
            "release_label": release_label,
            "rendered_bundle_checksum_path": relative_to_root(rendered_checksum_path, root_dir=ROOT_DIR),
            "rendered_bundle_path": relative_to_root(rendered_path, root_dir=ROOT_DIR),
            "status": "scaffold",
            "values_file": relative_to_root(values_file, root_dir=ROOT_DIR),
        },
    )
    print(f"Packaged enterprise Helm verification bundle: {manifest_path}")
    return manifest_path


def _deployment_posture(backend: str, values_file: Path | None) -> str:
    if values_file:
        values_name = values_file.name.lower()
        if "windows-ai" in values_name:
            return "enterprise self-hosted with Windows-hosted AI"
        if "enterprise" in values_name:
            return "enterprise self-hosted"
    if backend == "k3d":
        return "k3d validation"
    return f"{backend} validation"


def _start_port_forward_detached(service: str, forward: str):
    if is_windows():
        subprocess.Popen(
            ["kubectl", "port-forward", "-n", NAMESPACE, service, forward],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
        )
    else:
        subprocess.Popen(
            ["kubectl", "port-forward", "-n", NAMESPACE, service, forward],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=True,
        )

@task
def init(c):
    """
    Initialize Infrastructure Layer.
    Creates the preferred local Kubernetes cluster and applies persistent infra (NATS, DBs).
    """
    backend = _k8s_backend()
    print(f"Initializing Local Kubernetes Cluster: {CLUSTER_NAME} ({backend})...")
    if _cluster_exists(c):
        print("Cluster exists.")
    else:
        if backend == "k3d":
            c.run(f"k3d cluster create {CLUSTER_NAME}")
        else:
            # Check if we need to hydrate absolute paths for Windows Kind
            kind_config_path = ROOT_DIR / "kind-config.yaml"
            generated_config_path = ROOT_DIR / "kind-config.gen.yaml"
            with kind_config_path.open("r", encoding="utf-8") as f:
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
            with generated_config_path.open("w", encoding="utf-8") as f:
                f.write(config)

            print(f"Generated absolute config at {generated_config_path}")
            c.run(f"kind create cluster --name {CLUSTER_NAME} --config {generated_config_path}")
    
    
    # Legacy raw manifests removed. Helm chart handles infra.
    print("Infrastructure Ready for Helm.")


@task(
    help={
        "values_file": "Optional Helm values file path. Overrides MYCELIS_K8S_VALUES_FILE when set.",
        "verify_package": "Run enterprise Helm lint/template/package verification without cluster deployment.",
        "release_label": "Artifact label for Helm packaging metadata. Defaults to the computed repo version.",
        "package_output_dir": "Directory for Helm verification artifacts. Defaults to dist/helm.",
    }
)
def deploy(c, values_file="", verify_package=False, release_label="", package_output_dir=""):
    """
    Deploys the core using Helm (Hardened Security).
    Uses Immutable Identity Tagging.
    """
    resolved_values_file = _resolve_k8s_values_file(values_file)
    if verify_package:
        if not resolved_values_file:
            raise SystemExit(
                "A Helm values file is required for --verify-package. Pass --values-file or set MYCELIS_K8S_VALUES_FILE."
            )
        return _verify_and_package_chart(
            c,
            values_file=resolved_values_file,
            release_label=release_label or get_version(c),
            package_output_dir=package_output_dir,
        )

    # 1. Build Artifact (Delegated to Core)
    tag = core_build.body(c)
    
    print(f"Deploying Release: [{tag}]")
    
    print("Building Helm Dependencies...")
    c.run("helm dependency update ./charts/mycelis-core")
    
    backend = _k8s_backend()

    # 2. Load into the local Kubernetes backend
    print(f"   Loading Image into Cluster...")
    if backend == "k3d":
        c.run(f"k3d image import mycelis/core:{tag} -c {CLUSTER_NAME}")
    else:
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
    k8s_text_endpoint = os.getenv("MYCELIS_K8S_TEXT_ENDPOINT", "").strip()
    k8s_media_endpoint = os.getenv("MYCELIS_K8S_MEDIA_ENDPOINT", "").strip()
    values_file = resolved_values_file
    if not api_key:
        raise SystemExit("MYCELIS_API_KEY must be set in .env or shell before deploying the cluster.")
    if values_file and "windows-ai" in values_file.name.lower() and not k8s_text_endpoint:
        raise SystemExit(
            "MYCELIS_K8S_TEXT_ENDPOINT must be set when deploying the enterprise Windows AI preset. "
            "Point it at the Windows GPU host, for example http://192.168.50.156:11434/v1."
        )

    print(f"   Injecting Secrets for DB User: {pg_user}")
    print(f"   Deployment posture: {_deployment_posture(backend, values_file)}")
    if k8s_text_endpoint:
        print(f"   Text AI endpoint: {k8s_text_endpoint}")
    if k8s_media_endpoint:
        print(f"   Media AI endpoint: {k8s_media_endpoint}")
    if values_file:
        print(f"   Helm values file: {values_file}")

    cmd = (
        "helm upgrade --install mycelis-core ./charts/mycelis-core "
        f"--namespace {NAMESPACE} --create-namespace "
        f"--set image.tag={tag} "
        f"--set postgresql.auth.username={pg_user} "
        f"--set postgresql.auth.password={pg_pass} "
        f"--set postgresql.auth.database={pg_db} "
        f"--set coreAuth.apiKey={shlex.quote(api_key)} "
    )
    if values_file:
        cmd += f"--values {shlex.quote(str(values_file))} "
    if k8s_text_endpoint:
        cmd += f"--set-string ai.textEndpoint={shlex.quote(k8s_text_endpoint)} "
    if k8s_media_endpoint:
        cmd += f"--set-string ai.mediaEndpoint={shlex.quote(k8s_media_endpoint)} "
    cmd += "--wait"
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
    print("[1/3] Ensure local Kubernetes cluster + namespace...")
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
    cluster_ready = c.run("kubectl cluster-info", hide=True, warn=True)
    if not cluster_ready.ok:
        raise SystemExit(
            "K8S BRIDGE FAILED: kubectl cannot reach the cluster. "
            "Start Docker and the local Kubernetes backend first with 'uv run inv k8s.status' or 'uv run inv k8s.up'."
        )

    forwards = [
        ("svc/mycelis-core-nats", "4222:4222", 4222, "NATS"),
        ("svc/mycelis-core", "8080:8080", 8080, "Core API bridge"),
        ("svc/mycelis-core-postgresql", "5432:5432", 5432, "PostgreSQL"),
    ]

    failures: list[str] = []
    for service, forward, port, label in forwards:
        if _port_open(port):
            print(f"  {label} already active on localhost:{port}")
            continue
        _start_port_forward_detached(service, forward)
        if _wait_for_local_port(port, label, timeout_seconds=30):
            print(f"  {label} active on localhost:{port}")
        else:
            failures.append(f"{label} localhost:{port}")

    if failures:
        joined = ", ".join(failures)
        raise SystemExit(
            f"K8S BRIDGE FAILED: {joined} did not become reachable. "
            "Inspect cluster readiness with 'uv run inv k8s.status'."
        )

    print("Bridge active.")

@task
def status(c):
    """Check the health of the entire stack."""
    print(f"Checking Infrastructure Status for {CLUSTER_NAME}...")
    try:
        c.run("docker info", hide=True)
        print("Docker: Running")
    except:
        print("Docker: NOT Running.")
        print("Local Kubernetes Cluster: SKIPPED (Docker down)")
        print("Pod Status: SKIPPED")
        print("Persistence (PVC) Status: SKIPPED")
        return

    try:
        backend = _k8s_backend()
        cluster_command = "k3d cluster list" if backend == "k3d" else "kind get clusters"
        cluster_info = c.run(cluster_command, hide=True, warn=True)
        if cluster_info and cluster_info.ok and _cluster_list_has_name(cluster_info.stdout, backend):
            print(f"Local Kubernetes Cluster ({backend} / {CLUSTER_NAME}): Active")
        else:
            print("Local Kubernetes Cluster: Not Found")
            return
    except:
         print(f"Local Kubernetes Cluster: Error checking")
         return

    print("\nPod Status:")
    c.run(f"kubectl get pods -n {NAMESPACE}")
    
    print("\nPersistence (PVC) Status:")
    c.run(f"kubectl get pvc -n {NAMESPACE}")

@task(help={"timeout": "Seconds to wait for recovered rollouts to become ready (default: 180)."})
def recover(c, timeout=180):
    """Attempt to heal the cluster."""
    print("Attempting Recovery...")
    cluster_ready = c.run("kubectl cluster-info", hide=True, warn=True)
    if not cluster_ready.ok:
        raise SystemExit(
            "K8S RECOVER FAILED: kubectl cannot reach the cluster. "
            "Start Docker / the local Kubernetes backend first, then retry."
        )
    restart_failures: list[str] = []

    def _restart_rollout(resource: str) -> None:
        result = c.run(f"kubectl rollout restart {resource} -n {NAMESPACE}", warn=True, hide=True)
        if not result.ok:
            restart_failures.append(resource)

    # Core API
    _restart_rollout("deployment/mycelis-core")
    # NATS can be statefulset or deployment depending on chart version
    if _resource_exists(c, "statefulset/mycelis-core-nats"):
        _restart_rollout("statefulset/mycelis-core-nats")
    elif _resource_exists(c, "deployment/mycelis-core-nats"):
        _restart_rollout("deployment/mycelis-core-nats")
    # PostgreSQL (statefulset)
    if _resource_exists(c, "statefulset/mycelis-core-postgresql"):
        _restart_rollout("statefulset/mycelis-core-postgresql")
    if restart_failures:
        raise SystemExit(
            "K8S RECOVER FAILED: unable to send restart signal(s) for "
            + ", ".join(restart_failures)
            + ". Inspect kubectl access and cluster resources before retrying."
        )
    print("Restart signals sent. Waiting for readiness...")
    wait(c, timeout=timeout)
    print("Recovery complete.")

@task
def reset(c):
    """
    Reset Infrastructure (Teardown + Canonical Bring-Up).
    Standardizes the development environment.
    """
    print("Resetting Infrastructure...")
    
    # 1. Teardown
    backend = _k8s_backend()
    print("Stopping Local Kubernetes Cluster...")
    if backend == "k3d":
        c.run(f"k3d cluster delete {CLUSTER_NAME}")
    else:
        c.run(f"kind delete cluster --name {CLUSTER_NAME}")
    
    # 2. Canonical bring-up
    up(c)
    
    print("Infrastructure Reset Complete.")
    print("Run 'uv run inv k8s.status' to verify.")

ns = Collection("k8s")
ns.add_task(init)
ns.add_task(deploy)
ns.add_task(wait)
ns.add_task(up)
ns.add_task(bridge)
ns.add_task(status)
ns.add_task(recover)
ns.add_task(reset)
