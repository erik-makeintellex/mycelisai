from __future__ import annotations

import os
import socket
import subprocess
import time
from pathlib import Path

from dotenv import load_dotenv

from .config import CLUSTER_NAME, NAMESPACE, ROOT_DIR, is_windows
from .k8s_rancher import (
    _ensure_rancher_desktop_path,
    _rancher_desktop_candidate_dirs,
    _rancher_desktop_context_available,
    rancher_target_context,
    rancher_target_url,
)
from .k8s_standards import _shell_quote
from .packaging import resolve_repo_path


def _cluster_list_has_name(output: str, backend: str) -> bool:
    target_ctx = rancher_target_context()
    for line in output.splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        if backend == "rancher-desktop":
            if stripped.lstrip("*").strip().split()[0] in {"rancher-desktop", target_ctx}:
                return True
        elif backend == "k3d":
            if stripped.lower().startswith("name "):
                continue
            if stripped.split()[0] == CLUSTER_NAME:
                return True
        elif stripped == CLUSTER_NAME:
            return True
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


def _resolve_k8s_values_file(explicit_path: str = "", root_dir: Path = ROOT_DIR) -> Path | None:
    raw_value = explicit_path.strip() or os.environ.get("MYCELIS_K8S_VALUES_FILE", "").strip()
    if not raw_value:
        return None

    resolved = resolve_repo_path(raw_value, root_dir=root_dir)
    if not resolved.exists():
        raise SystemExit(f"MYCELIS_K8S_VALUES_FILE does not exist: {resolved}")
    return resolved


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


def _write_kind_config(root_dir: Path = ROOT_DIR) -> Path:
    kind_config_path = root_dir / "kind-config.yaml"
    generated_config_path = root_dir / "kind-config.gen.yaml"
    config = kind_config_path.read_text(encoding="utf-8")
    abs_ops = str(root_dir / "ops").replace("\\", "/")
    config = config.replace("./ops", abs_ops)
    abs_logs = str(root_dir / "logs").replace("\\", "/")
    config = config.replace("./logs", abs_logs)
    generated_config_path.write_text(config, encoding="utf-8")
    print(f"Generated absolute config at {generated_config_path}")
    return generated_config_path


def _helm_deploy_release(
    c,
    *,
    tag: str,
    backend: str,
    values_file: Path | None,
    root_dir: Path = ROOT_DIR,
    shell_quote=_shell_quote,
) -> None:
    print(f"Deploying Release: [{tag}]")
    print("Building Helm Dependencies...")
    c.run("helm dependency update ./charts/mycelis-core", warn=True)

    print("   Loading Image into Cluster...")
    if backend == "rancher-desktop":
        target_ctx = rancher_target_context()
        target_url = rancher_target_url()
        c.run(f"kubectl config use-context {target_ctx}", hide=True, warn=True)
        print(f"   Target Rancher cluster: {target_ctx} ({target_url})")
        if os.getenv("MYCELIS_RANCHER_RUNTIME", "moby").strip().lower() == "containerd":
            c.run(f"docker save mycelis/core:{tag} | nerdctl -n k8s.io load")
        else:
            print(f"   Rancher Desktop Moby runtime active: image mycelis/core:{tag} shared with cluster.")
    elif backend == "k3d":
        c.run(f"k3d image import mycelis/core:{tag} -c {CLUSTER_NAME}")
    else:
        c.run(f"kind load docker-image mycelis/core:{tag} --name {CLUSTER_NAME}")

    print("   Applying Helm Chart...")
    load_dotenv(os.path.join(root_dir, ".env"))
    pg_user = os.getenv("POSTGRES_USER", "mycelis")
    pg_pass = os.getenv("POSTGRES_PASSWORD", "password")
    pg_db = os.getenv("POSTGRES_DB", "cortex")
    api_key = os.getenv("MYCELIS_API_KEY", "")
    k8s_text_endpoint = os.getenv("MYCELIS_K8S_TEXT_ENDPOINT", "").strip()
    k8s_media_endpoint = os.getenv("MYCELIS_K8S_MEDIA_ENDPOINT", "").strip()
    k8s_nats_url = os.getenv("MYCELIS_K8S_NATS_URL", "").strip()
    k8s_db_host = os.getenv("MYCELIS_K8S_DB_HOST", "").strip()
    k8s_with_interface = os.getenv("MYCELIS_K8S_WITH_INTERFACE", "").strip().lower() in {"1", "true", "yes"}
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
        f"--set coreAuth.apiKey={shell_quote(api_key)} "
    )
    if values_file:
        cmd += f"--values {shell_quote(values_file)} "
    elif backend == "rancher-desktop":
        cmd += (
            "--set service.type=LoadBalancer "
            "--set dataVolume.storageClassName=local-path "
            "--set nats.jetstream.fileStore.pvc.storageClassName=local-path "
            "--set postgresql.primary.persistence.storageClass=local-path "
        )
        if not k8s_text_endpoint:
            k8s_text_endpoint = "http://host.rancher-desktop.internal:11434/v1"
    if k8s_text_endpoint:
        cmd += f"--set-string ai.textEndpoint={shell_quote(k8s_text_endpoint)} "
    if k8s_media_endpoint:
        cmd += f"--set-string ai.mediaEndpoint={shell_quote(k8s_media_endpoint)} "
    if k8s_nats_url:
        cmd += f"--set-string serviceAddresses.natsUrl={shell_quote(k8s_nats_url)} "
    if k8s_db_host:
        cmd += f"--set-string serviceAddresses.dbHost={shell_quote(k8s_db_host)} "
    if k8s_with_interface:
        cmd += f"--set interface.enabled=true --set interface.image.tag={tag} "
    cmd += "--wait"
    c.run(cmd)
    c.run(f"kubectl rollout restart deployment/mycelis-core -n {NAMESPACE}")
    print(f"Deployment Complete ({tag}).")


def _run_bridge_forwards(c) -> None:
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


def _check_cluster_status(c, backend_fn) -> None:
    _ensure_rancher_desktop_path()
    print(f"Checking Infrastructure Status for {CLUSTER_NAME}...")
    try:
        c.run("docker info", hide=True)
        print("Docker: Running")
    except Exception:
        print("Docker: NOT Running.")
        print("Local Kubernetes Cluster: SKIPPED (Docker down)")
        print("Pod Status: SKIPPED")
        print("Persistence (PVC) Status: SKIPPED")
        return

    try:
        backend = backend_fn()
        if backend == "rancher-desktop":
            cluster_command = "kubectl config get-contexts -o name"
            label = "rancher-desktop"
        elif backend == "k3d":
            cluster_command = "k3d cluster list"
            label = f"{backend} / {CLUSTER_NAME}"
        else:
            cluster_command = "kind get clusters"
            label = f"{backend} / {CLUSTER_NAME}"
        cluster_info = c.run(cluster_command, hide=True, warn=True)
        if cluster_info and cluster_info.ok and _cluster_list_has_name(cluster_info.stdout, backend):
            print(f"Local Kubernetes Cluster ({label}): Active")
        else:
            print("Local Kubernetes Cluster: Not Found")
            return
    except Exception:
        print("Local Kubernetes Cluster: Error checking")
        return

    print("\nPod Status:")
    c.run(f"kubectl get pods -n {NAMESPACE}")

    print("\nPersistence (PVC) Status:")
    c.run(f"kubectl get pvc -n {NAMESPACE}")
