from __future__ import annotations

import pytest

from ops import compose, service_tiers


class _DummyContext:
    def __init__(self) -> None:
        self.commands: list[str] = []

    def run(self, cmd: str, warn: bool = False):
        self.commands.append(cmd)
        return type("Result", (), {"ok": True, "return_code": 0})()


def test_service_plan_prints_strategy_matrix(capsys):
    service_tiers.plan.body(None, target="compose")

    out = capsys.readouterr().out
    assert "Strategy Decision Matrix" in out
    assert "Docker Compose (Primary for Single-Host" in out
    assert "Kubernetes / Helm (Target for Multi-Node" in out
    assert "mcp-cache" in out


def test_service_up_data_tier_compose_delegates_to_infra_up(monkeypatch):
    calls: list[tuple[int, bool]] = []
    monkeypatch.setattr(
        compose.infra_up,
        "body",
        lambda _c, wait_timeout=180, migrate=False: calls.append((wait_timeout, migrate)),
    )

    service_tiers.up.body(None, tier="data", target="compose", wait_timeout=90)

    assert calls == [(90, True)]


def test_service_up_capabilities_tier_compose_starts_searxng(monkeypatch, capsys):
    started: list[tuple[tuple[str, ...], bool]] = []
    waited: list[tuple[int, str, int]] = []

    monkeypatch.setattr(
        service_tiers,
        "_compose_start_services",
        lambda services, build=False: started.append((services, build)),
    )
    monkeypatch.setattr(
        compose,
        "_compose_effective_env",
        lambda: {"MYCELIS_COMPOSE_SEARXNG_PORT": "18088"},
    )
    monkeypatch.setattr(
        compose,
        "_wait_for_port",
        lambda port, label, timeout_seconds=60: waited.append((port, label, timeout_seconds)) or True,
    )

    service_tiers.up.body(None, tier="capabilities", target="compose", wait_timeout=75)

    assert started == [(("searxng",), False)]
    assert waited == [(18088, "SearXNG", 75)]
    assert "Capabilities tier online" in capsys.readouterr().out


def test_service_up_k8s_scales_expected_workloads():
    ctx = _DummyContext()

    service_tiers.up.body(ctx, tier="data", target="k8s", wait_timeout=60)

    assert any("kubectl scale statefulset/mycelis-core-postgresql" in cmd for cmd in ctx.commands)
    assert any("kubectl scale statefulset/mycelis-core-nats" in cmd for cmd in ctx.commands)
    assert any("kubectl rollout status statefulset/mycelis-core-postgresql" in cmd for cmd in ctx.commands)


def test_service_wake_starts_single_compose_service(monkeypatch):
    started: list[tuple[tuple[str, ...], bool]] = []
    monkeypatch.setattr(
        service_tiers,
        "_compose_start_services",
        lambda services, build=False: started.append((services, build)),
    )

    service_tiers.wake.body(None, service="searxng", target="compose", build=False)

    assert started == [(("searxng",), False)]


def test_service_wake_rejects_unknown_service():
    with pytest.raises(SystemExit) as excinfo:
        service_tiers.wake.body(None, service="unknown", target="compose")

    assert "Unsupported service" in str(excinfo.value)


def test_service_hibernate_stops_capabilities_while_protecting_data(monkeypatch):
    stopped: list[tuple[str, ...]] = []
    monkeypatch.setattr(
        service_tiers,
        "_compose_stop_services",
        lambda services: stopped.append(services),
    )

    service_tiers.hibernate.body(None, tier="capabilities", target="compose")

    assert stopped == [("searxng",)]

    with pytest.raises(SystemExit) as excinfo:
        service_tiers.hibernate.body(None, tier="data", target="compose")

    assert "Refusing to hibernate the stateful 'data' tier" in str(excinfo.value)


def test_service_status_compose_reports_searxng_state(monkeypatch, capsys):
    monkeypatch.setattr(compose.status, "body", lambda _c: print("base-status"))
    monkeypatch.setattr(compose, "_compose_effective_env", lambda: {"MYCELIS_COMPOSE_SEARXNG_PORT": "8088"})
    monkeypatch.setattr(compose, "_port_open", lambda port: False)

    service_tiers.status.body(None, target="compose")

    out = capsys.readouterr().out
    assert "base-status" in out
    assert "SearXNG Search  : DOWN (on-demand)" in out


def test_wsl_exec_command_runs_natively_when_already_inside_wsl(monkeypatch):
    monkeypatch.setattr(compose, "running_in_wsl", lambda: True)
    assert compose._wsl_exec_command("sh", "-lc", "echo ok") == ["sh", "-lc", "echo ok"]


def test_service_prune_cache_cleans_project_user_and_docker_builder_caches(monkeypatch, capsys):
    cleaned: list[tuple[bool, bool]] = []
    docker_cmds: list[list[str]] = []

    monkeypatch.setattr(
        service_tiers.cache_tasks.clean,
        "body",
        lambda _c, project=True, user=False: cleaned.append((project, user)),
    )
    monkeypatch.setattr(
        compose,
        "_docker_run",
        lambda args, check=True: docker_cmds.append(args),
    )

    service_tiers.prune_cache.body(None, user=True, docker=True)

    assert cleaned == [(True, True)]
    assert docker_cmds == [["builder", "prune", "-af"], ["image", "prune", "-f"]]
    assert "Docker BuildKit cache and dangling builder images pruned" in capsys.readouterr().out



def test_compose_effective_env_honors_searxng_and_mcp_cache_overrides(monkeypatch):
    monkeypatch.setenv("MYCELIS_COMPOSE_SEARXNG_PORT", "18088")
    monkeypatch.setenv("MYCELIS_MCP_NPM_CACHE", "/custom/mcp-cache")

    effective = compose._compose_effective_env({"MYCELIS_COMPOSE_SEARXNG_PORT": "8088"})

    assert effective["MYCELIS_COMPOSE_SEARXNG_PORT"] == "18088"
    assert effective["MYCELIS_MCP_NPM_CACHE"] == "/custom/mcp-cache"


