"""Development data-plane policy for local lifecycle tasks."""

import os
from pathlib import Path


def dev_infra_mode(root_dir: Path) -> str:
    """Resolve the explicit mode, with Docker dependencies as the default."""
    configured_mode = os.environ.get("MYCELIS_DEV_INFRA_MODE")
    if configured_mode is None:
        try:
            from dotenv import dotenv_values

            configured_mode = dotenv_values(root_dir / ".env").get("MYCELIS_DEV_INFRA_MODE")
        except ModuleNotFoundError:
            configured_mode = None
    mode = str(configured_mode or "compose").strip().lower()
    if mode not in {"", "compose", "native", "k8s"}:
        raise SystemExit("Invalid MYCELIS_DEV_INFRA_MODE. Use compose, native, or k8s.")
    return mode or "compose"


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
    elif infra_mode == "native":
        print("  Development     : host PostgreSQL/NATS + local Core/Interface")
        print("  Docker/K8s      : explicit proof lanes via compose.* or k8s.*")
    else:
        print("  Development     : Kubernetes bridges + local Core/Interface")


def print_retained_data_plane(infra_mode: str) -> None:
    if infra_mode == "compose":
        print("[4/4] Docker data plane: left running")
        print("  PostgreSQL and NATS remain reusable; inspect them with compose.infra-health.")
        return
    print("[4/4] Native infrastructure: left running")
    print("  PostgreSQL and NATS are development dependencies; inspect them with native-infra.status.")
