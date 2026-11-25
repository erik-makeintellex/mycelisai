from __future__ import annotations

import subprocess
from invoke import Collection, task

from . import cache as cache_tasks, compose as compose_tasks
from .config import API_HOST, API_PORT, INTERFACE_HOST, INTERFACE_PORT, NAMESPACE
from .k8s_rancher import prune_old_mycelis_images


VALID_TARGETS = {"compose", "k8s"}
VALID_TIERS = {"data", "core", "capabilities", "full"}
COMPOSE_TIER_SERVICES: dict[str, tuple[str, ...]] = {
    "data": ("postgres", "nats"),
    "capabilities": ("searxng",),
    "core": ("core", "interface"),
    "full": ("postgres", "nats", "searxng", "core", "interface"),
}
K8S_TIER_WORKLOADS: dict[str, tuple[str, ...]] = {
    "data": ("statefulset/mycelis-core-postgresql", "statefulset/mycelis-core-nats"),
    "core": ("deployment/mycelis-core",),
    "capabilities": ("deployment/mycelis-core",),
    "full": (
        "statefulset/mycelis-core-postgresql",
        "statefulset/mycelis-core-nats",
        "deployment/mycelis-core",
    ),
}


def _normalize_target(target: str) -> str:
    cleaned = (target or "compose").strip().lower()
    if cleaned not in VALID_TARGETS:
        raise SystemExit(
            f"Unsupported deployment target '{cleaned}'. Choose one of: {', '.join(sorted(VALID_TARGETS))}."
        )
    return cleaned


def _normalize_tier(tier: str) -> str:
    cleaned = (tier or "full").strip().lower()
    if cleaned not in VALID_TIERS:
        raise SystemExit(
            f"Unsupported service tier '{cleaned}'. Choose one of: {', '.join(sorted(VALID_TIERS))}."
        )
    return cleaned


def _compose_start_services(services: tuple[str, ...], *, build: bool = False) -> None:
    compose_tasks._require_compose_env_file()
    env_values = compose_tasks._compose_effective_env()
    compose_tasks._validate_compose_env(env_values)
    if "core" in services:
        env_values = compose_tasks._prepare_wsl_ollama_host(env_values)
    cmd = compose_tasks._compose_command("up", "-d")
    if build:
        cmd.append("--build")
    cmd.extend(services)
    compose_tasks._run_compose(cmd, env=compose_tasks._compose_runtime_env(env_values))


def _compose_stop_services(services: tuple[str, ...]) -> None:
    compose_tasks._require_compose_env_file()
    cmd = compose_tasks._compose_command("stop", *services)
    compose_tasks._run_compose(cmd, env=compose_tasks._compose_runtime_env())


def _k8s_scale_workloads(c, workloads: tuple[str, ...], replicas: int, wait_timeout: int = 180) -> None:
    for workload in workloads:
        c.run(f"kubectl scale {workload} -n {NAMESPACE} --replicas={int(replicas)}", warn=True)
        if replicas > 0:
            c.run(
                f"kubectl rollout status {workload} -n {NAMESPACE} --timeout={int(wait_timeout)}s",
                warn=True,
            )


@task(help={"target": "Deployment target to describe: compose (default) or k8s."})
def plan(_c, target="compose"):
    """Print the tiered deployment and storage strategy for Initial Total Service vs Scale-Up."""
    selected = _normalize_target(target)
    print("=== Mycelis Tiered Deployment & Service Management Plan ===")
    print(f"Active Target Mode: {selected}")
    print()
    print("Strategy Decision Matrix:")
    print("  1. Docker Compose (Primary for Single-Host / WSL Target Server Initial Total Service):")
    print("     - Zero Kubernetes control-plane overhead on the AI/target host.")
    print("     - Direct host-gateway + WSL Ollama relay to GPU inference host (:11434).")
    print("     - Persistent storage volumes: postgres-data (pgvector), nats-data (JetStream),")
    print("       mcp-cache (/data/mcp-cache for stdio npx MCP servers), and /data output block.")
    print("  2. Kubernetes / Helm (Target for Multi-Node / Enterprise Scale-Up):")
    print("     - Enforces NetworkPolicy, PodSecurityContext, ServiceAccount, and cluster PVCs.")
    print("     - Managed via k3d (local cluster validation) or enterprise Helm presets.")
    print()
    print("On-Demand Service Tiers:")
    print("  - data         : PostgreSQL (pgvector) + NATS JetStream stateful foundation")
    print("  - capabilities : On-demand search (SearXNG) and AI relay helpers")
    print("  - core         : Mycelis Core API (:8081) + Next.js Interface (:3000)")
    print("  - full         : Complete stack with migrations, storage-health, and deep health checks")


@task(
    help={
        "tier": "Service tier to bring up: data, capabilities, core, or full (default: full).",
        "target": "Deployment target: compose (default) or k8s.",
        "build": "Build images when starting core/full on compose (default: False).",
        "wait_timeout": "Readiness timeout in seconds (default: 180).",
    }
)
def up(c, tier="full", target="compose", build=False, wait_timeout=180):
    """Bring up a specific service tier on demand for Docker Compose or Kubernetes."""
    selected_tier = _normalize_tier(tier)
    selected_target = _normalize_target(target)
    timeout = int(wait_timeout)

    print(f"=== Mycelis Service Manager: UP (target={selected_target}, tier={selected_tier}) ===")
    if selected_target == "k8s":
        workloads = K8S_TIER_WORKLOADS[selected_tier]
        _k8s_scale_workloads(c, workloads, replicas=1, wait_timeout=timeout)
        return

    if selected_tier == "data":
        compose_tasks.infra_up.body(c, wait_timeout=timeout, migrate=True)
        return

    if selected_tier == "capabilities":
        _compose_start_services(COMPOSE_TIER_SERVICES["capabilities"], build=False)
        env_values = compose_tasks._compose_effective_env()
        searx_port = compose_tasks._compose_host_port(env_values, "MYCELIS_COMPOSE_SEARXNG_PORT", "8088")
        compose_tasks._wait_for_port(searx_port, "SearXNG", timeout_seconds=timeout)
        print(f"Capabilities tier online (SearXNG port {searx_port} ready).")
        return

    if selected_tier == "core":
        compose_tasks.infra_up.body(c, wait_timeout=timeout, migrate=True)
        _compose_start_services(COMPOSE_TIER_SERVICES["core"], build=build)
        compose_tasks._wait_for_http_ok(
            f"http://{API_HOST}:{API_PORT}/healthz",
            "Core health",
            timeout_seconds=timeout,
        )
        compose_tasks._wait_for_http_ok(
            f"http://{INTERFACE_HOST}:{INTERFACE_PORT}/",
            "Frontend",
            timeout_seconds=timeout,
        )
        print("Core tier online.")
        return

    compose_tasks.up.body(c, build=build, wait_timeout=timeout)


@task(
    help={
        "service": "Specific service to wake on demand (postgres, nats, searxng, core, interface).",
        "target": "Deployment target: compose (default) or k8s.",
        "build": "Rebuild container image if applicable (default: False).",
    }
)
def wake(c, service="searxng", target="compose", build=False):
    """Wake a single service on demand without restarting healthy peer services."""
    selected_target = _normalize_target(target)
    svc = (service or "").strip().lower()
    allowed = {"postgres", "nats", "searxng", "core", "interface"}
    if svc not in allowed:
        raise SystemExit(f"Unsupported service '{svc}'. Choose one of: {', '.join(sorted(allowed))}.")

    print(f"=== Mycelis Service Manager: WAKE ({svc} on {selected_target}) ===")
    if selected_target == "k8s":
        mapping = {
            "postgres": ("statefulset/mycelis-core-postgresql",),
            "nats": ("statefulset/mycelis-core-nats",),
            "core": ("deployment/mycelis-core",),
            "interface": ("deployment/mycelis-core",),
            "searxng": ("deployment/mycelis-core",),
        }
        _k8s_scale_workloads(c, mapping[svc], replicas=1)
        return

    _compose_start_services((svc,), build=build)
    print(f"Service '{svc}' is awake.")


@task(
    help={
        "tier": "Tier to hibernate while preserving stateful data volumes: capabilities or core (default: capabilities).",
        "target": "Deployment target: compose (default) or k8s.",
    }
)
def hibernate(c, tier="capabilities", target="compose"):
    """Hibernate non-data tiers (e.g. capabilities or core/interface) while keeping stateful data safe."""
    selected_target = _normalize_target(target)
    selected_tier = _normalize_tier(tier)
    if selected_tier == "data":
        raise SystemExit(
            "Refusing to hibernate the stateful 'data' tier via hibernate; use 'compose.down' if stopping everything."
        )

    services = COMPOSE_TIER_SERVICES["capabilities"] if selected_tier == "capabilities" else ("interface", "core")
    print(f"=== Mycelis Service Manager: HIBERNATE (tier={selected_tier}, target={selected_target}) ===")
    if selected_target == "k8s":
        _k8s_scale_workloads(c, ("deployment/mycelis-core",), replicas=0)
        return

    _compose_stop_services(services)
    print(f"Hibernated {selected_tier} services ({', '.join(services)}); data plane remains intact.")


@task(help={"target": "Deployment target: compose (default) or k8s."})
def status(c, target="compose"):
    """Show tiered service status across Data, Capabilities, and Core tiers."""
    selected_target = _normalize_target(target)
    if selected_target == "k8s":
        c.run(f"kubectl get pods,svc,pvc -n {NAMESPACE}", warn=True)
        return

    compose_tasks.status.body(c)
    env_values = compose_tasks._compose_effective_env()
    searx_port = compose_tasks._compose_host_port(env_values, "MYCELIS_COMPOSE_SEARXNG_PORT", "8088")
    searx_state = "UP" if compose_tasks._port_open(searx_port) else "DOWN (on-demand)"
    print(f"  SearXNG Search  : {searx_state} (127.0.0.1:{searx_port})")


@task(
    name="prune-cache",
    help={
        "user": "Also clean user-level external library caches (default: True).",
        "docker": "Also prune Docker BuildKit build cache and dangling images (default: True).",
    },
)
def prune_cache(c, user=True, docker=True):
    """Keep external library, build artifact, and Docker BuildKit caches clean without touching stateful volumes."""
    print("=== Mycelis Service Manager: PRUNE CACHE ===")
    cache_tasks.clean.body(c, project=True, user=bool(user))
    if docker:
        if type(c).__name__ == "Context":
            prune_old_mycelis_images(max_versions=2)
        compose_tasks._docker_run(["builder", "prune", "-af"], check=False)
        compose_tasks._docker_run(["image", "prune", "-f"], check=False)
        print("Docker BuildKit cache and dangling builder images pruned; active containers and volumes preserved.")


ns = Collection("service")
ns.add_task(plan)
ns.add_task(up)
ns.add_task(wake)
ns.add_task(hibernate)
ns.add_task(status)
ns.add_task(prune_cache)
