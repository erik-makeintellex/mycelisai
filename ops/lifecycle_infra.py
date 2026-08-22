"""Development data-plane policy for local lifecycle tasks."""

import os
from pathlib import Path


def _dotenv_values(path: Path) -> dict[str, str | None]:
    """Read dotenv values when runtime dependencies are installed."""
    try:
        from dotenv import dotenv_values
    except ModuleNotFoundError:
        return {}
    return dict(dotenv_values(path))


def dev_infra_mode(root_dir: Path) -> str:
    """Resolve the explicit mode, with Docker dependencies as the default."""
    configured_mode = os.environ.get("MYCELIS_DEV_INFRA_MODE")
    if configured_mode is None:
        configured_mode = _dotenv_values(root_dir / ".env").get("MYCELIS_DEV_INFRA_MODE")
    mode = str(configured_mode or "compose").strip().lower()
    if mode not in {"", "compose", "k8s"}:
        raise SystemExit("Invalid MYCELIS_DEV_INFRA_MODE. Use compose or k8s; native host services are unsupported.")
    return mode or "compose"


def database_endpoint(root_dir: Path) -> tuple[str, int]:
    """Return the host endpoint used by source-mode Core."""
    values = _dotenv_values(root_dir / ".env")
    host = os.environ.get("DB_HOST") or values.get("DB_HOST") or "127.0.0.1"
    port = os.environ.get("DB_PORT") or values.get("DB_PORT") or "5432"
    return str(host), int(port)


def managed_process_keys(infra_mode: str) -> tuple[str, ...]:
    """Return listeners lifecycle.down owns for the selected mode."""
    if infra_mode == "k8s":
        return "postgres", "nats", "core", "frontend"
    return "core", "frontend"


def ensure_compose_data_plane(infra_mode: str) -> bool:
    """Start only Docker PostgreSQL/NATS when Compose mode is selected."""
    if infra_mode != "compose":
        return False
    from . import compose

    compose.infra_up.body(None, wait_timeout=180, migrate=False)
    return True


def print_development_status(infra_mode: str) -> None:
    print(f"  Dev infra mode  : {infra_mode}")
    if infra_mode == "compose":
        print("  Development     : Docker PostgreSQL/NATS + local Core/Interface")
        print("  Full containers : explicit proof via compose.up; Kubernetes via k8s.*")
    else:
        print("  Development     : Kubernetes bridges + local Core/Interface")


def print_retained_data_plane(infra_mode: str) -> None:
    if infra_mode == "compose":
        print("[4/4] Docker data plane: left running")
        print("  PostgreSQL and NATS remain reusable; inspect them with compose.infra-health.")
        return
    print("[4/4] Kubernetes data plane: left running")
    print("  PostgreSQL and NATS bridges remain reusable; inspect them with k8s.status.")


def stop_compose_data_plane(context) -> None:
    """Stop Compose dependencies without deleting retained volumes."""
    from . import compose

    print("[4/4] Stopping Docker data plane...")
    compose.down.body(context, volumes=False)


def print_shutdown_summary(infra_mode: str, included_data_plane: bool) -> None:
    """Report the exact shutdown boundary instead of implying host-wide control."""
    if infra_mode == "compose" and included_data_plane:
        print("\nLocal app services and Compose data plane stopped. Retained volumes were preserved.")
    elif infra_mode == "compose":
        print("\nLocal app services stopped. Compose PostgreSQL and NATS remain running for reuse.")
    else:
        print("\nLocal app services and Kubernetes port-forwards stopped. Cluster services remain running.")
