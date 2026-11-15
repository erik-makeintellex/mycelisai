import os
import shutil

from invoke import Collection, task

from .config import CLUSTER_NAME, NAMESPACE, ROOT_DIR
from .core import build as core_build
from .k8s_helpers import (
    _check_cluster_status,
    _cluster_list_has_name,
    _deployment_posture,
    _helm_deploy_release,
    _port_open,
    _resolve_k8s_values_file,
    _run_bridge_forwards,
    _start_port_forward_detached,
    _wait_for_local_port,
    _write_kind_config,
)
from .k8s_standards import (
    _chart_dir,
    _chart_version,
    _resolve_package_output_dir,
    _shell_quote,
    _verify_and_package_chart,
    standards,
)
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
            generated_config_path = _write_kind_config(ROOT_DIR)
            c.run(f"kind create cluster --name {CLUSTER_NAME} --config {generated_config_path}")

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
    resolved_values_file = _resolve_k8s_values_file(values_file, root_dir=ROOT_DIR)
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
            root_dir=ROOT_DIR,
            shell_quote=_shell_quote,
        )

    tag = core_build.body(c)
    backend = _k8s_backend()
    _helm_deploy_release(
        c,
        tag=tag,
        backend=backend,
        values_file=resolved_values_file,
        root_dir=ROOT_DIR,
        shell_quote=_shell_quote,
    )


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
    _run_bridge_forwards(c)


@task
def status(c):
    """Check the health of the entire stack."""
    _check_cluster_status(c, _k8s_backend)


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

    _restart_rollout("deployment/mycelis-core")
    if _resource_exists(c, "statefulset/mycelis-core-nats"):
        _restart_rollout("statefulset/mycelis-core-nats")
    elif _resource_exists(c, "deployment/mycelis-core-nats"):
        _restart_rollout("deployment/mycelis-core-nats")
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
    backend = _k8s_backend()
    print("Stopping Local Kubernetes Cluster...")
    if backend == "k3d":
        c.run(f"k3d cluster delete {CLUSTER_NAME}")
    else:
        c.run(f"kind delete cluster --name {CLUSTER_NAME}")

    up(c)

    print("Infrastructure Reset Complete.")
    print("Run 'uv run inv k8s.status' to verify.")


ns = Collection("k8s")
for _task in (init, deploy, standards, wait, up, bridge, status, recover, reset):
    ns.add_task(_task)
