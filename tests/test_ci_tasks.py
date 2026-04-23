from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

import pytest
from invoke import Context

from ops import ci


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""


class FakeContext(Context):
    def __init__(self, command_results: dict[str, FakeResult]):
        super().__init__()
        self.command_results = command_results
        self.commands: list[str] = []
        self.cd_paths: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return self.command_results.get(command, FakeResult())

    @contextmanager
    def cd(self, path: str):
        self.cd_paths.append(path)
        yield


def test_baseline_runs_expected_commands_without_e2e(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)
    build_calls: list[str] = []
    test_calls: list[str] = []
    typecheck_calls: list[str] = []
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(ci.interface_tasks.stop, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.clean, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.test, "body", lambda _ctx: test_calls.append("test"))
    monkeypatch.setattr(ci.interface_tasks.typecheck, "body", lambda _ctx: typecheck_calls.append("typecheck"))

    ctx = FakeContext(
        {
            "go test ./... -count=1": FakeResult(),
        }
    )

    ci.baseline.body(ctx, e2e=False)

    assert "go test ./... -count=1" in ctx.commands
    assert "npx playwright test --reporter=dot" not in ctx.commands
    assert build_calls == ["build"]
    assert test_calls == ["test"]
    assert typecheck_calls == ["typecheck"]


def test_baseline_runs_playwright_when_e2e_enabled(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)
    build_calls: list[str] = []
    test_calls: list[str] = []
    typecheck_calls: list[str] = []
    e2e_calls: list[dict[str, object]] = []
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(ci.interface_tasks.stop, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.clean, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.test, "body", lambda _ctx: test_calls.append("test"))
    monkeypatch.setattr(ci.interface_tasks.typecheck, "body", lambda _ctx: typecheck_calls.append("typecheck"))
    monkeypatch.setattr(ci.interface_tasks.e2e, "body", lambda _ctx, **kwargs: e2e_calls.append(kwargs))

    ctx = FakeContext(
        {
            "go test ./... -count=1": FakeResult(),
        }
    )

    ci.baseline.body(ctx, e2e=True)

    assert e2e_calls == [{"workers": "1", "server_mode": "start"}]
    assert build_calls == ["build", "build"]
    assert test_calls == ["test"]
    assert typecheck_calls == ["typecheck"]


def test_baseline_skips_playwright_when_prior_steps_failed(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)
    build_calls: list[str] = []
    test_calls: list[str] = []
    typecheck_calls: list[str] = []
    e2e_calls: list[dict[str, object]] = []
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(ci.interface_tasks.stop, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.clean, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.test, "body", lambda _ctx: test_calls.append("test"))
    monkeypatch.setattr(ci.interface_tasks.typecheck, "body", lambda _ctx: typecheck_calls.append("typecheck"))
    monkeypatch.setattr(ci.interface_tasks.e2e, "body", lambda _ctx, **kwargs: e2e_calls.append(kwargs))

    ctx = FakeContext(
        {
            "go test ./... -count=1": FakeResult(exited=1, stderr="core tests failed"),
        }
    )

    with pytest.raises(SystemExit):
        ci.baseline.body(ctx)

    assert e2e_calls == []
    assert build_calls == ["build"]
    assert test_calls == ["test"]
    assert typecheck_calls == ["typecheck"]


def test_baseline_runs_playwright_by_default(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)
    build_calls: list[str] = []
    test_calls: list[str] = []
    typecheck_calls: list[str] = []
    e2e_calls: list[dict[str, object]] = []
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(ci.interface_tasks.stop, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.clean, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.test, "body", lambda _ctx: test_calls.append("test"))
    monkeypatch.setattr(ci.interface_tasks.typecheck, "body", lambda _ctx: typecheck_calls.append("typecheck"))
    monkeypatch.setattr(ci.interface_tasks.e2e, "body", lambda _ctx, **kwargs: e2e_calls.append(kwargs))

    ctx = FakeContext(
        {
            "go test ./... -count=1": FakeResult(),
        }
    )

    ci.baseline.body(ctx)

    assert e2e_calls == [{"workers": "1", "server_mode": "start"}]
    assert build_calls == ["build", "build"]
    assert test_calls == ["test"]
    assert typecheck_calls == ["typecheck"]


def test_build_reuses_core_compile_and_interface_build_tasks(monkeypatch):
    compile_calls: list[str] = []
    build_calls: list[str] = []

    monkeypatch.setattr(ci.core_tasks.compile, "body", lambda _ctx: compile_calls.append("compile"))
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))

    ci.build.body(FakeContext({}))

    assert compile_calls == ["compile"]
    assert build_calls == ["build"]


def test_lint_reuses_interface_lint_task(monkeypatch):
    lint_calls: list[str] = []
    monkeypatch.setattr(ci.interface_tasks.lint, "body", lambda _ctx: lint_calls.append("lint"))

    ctx = FakeContext(
        {
            "go vet ./...": FakeResult(),
        }
    )

    ci.lint.body(ctx)

    assert "go vet ./..." in ctx.commands
    assert lint_calls == ["lint"]


def test_service_check_runs_health_only_by_default(monkeypatch):
    health_calls: list[str] = []
    e2e_calls: list[str] = []

    monkeypatch.setattr(ci.lifecycle.health, "body", lambda _ctx, **_kwargs: health_calls.append("health"))
    monkeypatch.setattr(ci.interface_tasks.e2e, "body", lambda _ctx, **_kwargs: e2e_calls.append("e2e"))

    ci.service_check.body(FakeContext({}), live_backend=False)

    assert health_calls == ["health"]
    assert e2e_calls == []


def test_service_check_runs_live_backend_browser_proof_when_requested(monkeypatch):
    health_calls: list[str] = []
    up_calls: list[tuple[bool, bool]] = []
    migrate_calls: list[str] = []
    build_calls: list[str] = []
    e2e_calls: list[dict[str, object]] = []

    monkeypatch.setattr(ci.lifecycle.up, "body", lambda _ctx, **kwargs: up_calls.append((kwargs["frontend"], kwargs["build"])))
    monkeypatch.setattr(ci.db_tasks, "schema_bootstrapped", lambda: False)
    monkeypatch.setattr(ci.db_tasks.migrate, "body", lambda _ctx: migrate_calls.append("migrate"))
    monkeypatch.setattr(ci.lifecycle.health, "body", lambda _ctx, **_kwargs: health_calls.append("health"))
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(
        ci.interface_tasks.e2e,
        "body",
        lambda _ctx, **kwargs: e2e_calls.append(kwargs),
    )

    ci.service_check.body(FakeContext({}), live_backend=True)

    assert up_calls == [(False, False)]
    assert migrate_calls == ["migrate"]
    assert health_calls == ["health"]
    assert build_calls == ["build"]
    assert e2e_calls == [
        {
            "project": "chromium",
            "spec": "e2e/specs/soma-governance-live.spec.ts",
            "live_backend": True,
            "workers": "1",
            "server_mode": "start",
        }
    ]


def test_service_check_skips_live_backend_browser_proof_when_prereqs_fail(monkeypatch):
    health_calls: list[str] = []
    up_calls: list[tuple[bool, bool]] = []
    migrate_calls: list[str] = []
    build_calls: list[str] = []
    e2e_calls: list[dict[str, object]] = []

    monkeypatch.setattr(ci.lifecycle.up, "body", lambda _ctx, **kwargs: up_calls.append((kwargs["frontend"], kwargs["build"])))
    monkeypatch.setattr(ci.db_tasks, "schema_bootstrapped", lambda: False)
    monkeypatch.setattr(ci.db_tasks.migrate, "body", lambda _ctx: migrate_calls.append("migrate") or (_ for _ in ()).throw(SystemExit(1)))
    monkeypatch.setattr(ci.lifecycle.health, "body", lambda _ctx, **_kwargs: health_calls.append("health"))
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(
        ci.interface_tasks.e2e,
        "body",
        lambda _ctx, **kwargs: e2e_calls.append(kwargs),
    )

    with pytest.raises(SystemExit):
        ci.service_check.body(FakeContext({}), live_backend=True)

    assert up_calls == [(False, False)]
    assert migrate_calls == ["migrate"]
    assert health_calls == ["health"]
    assert build_calls == []
    assert e2e_calls == []


def test_service_check_skips_migrate_when_schema_is_already_initialized(monkeypatch):
    health_calls: list[str] = []
    up_calls: list[tuple[bool, bool]] = []
    migrate_calls: list[str] = []
    build_calls: list[str] = []
    e2e_calls: list[dict[str, object]] = []

    monkeypatch.setattr(ci.lifecycle.up, "body", lambda _ctx, **kwargs: up_calls.append((kwargs["frontend"], kwargs["build"])))
    monkeypatch.setattr(ci.db_tasks, "schema_bootstrapped", lambda: True)
    monkeypatch.setattr(ci.db_tasks.migrate, "body", lambda _ctx: migrate_calls.append("migrate"))
    monkeypatch.setattr(ci.lifecycle.health, "body", lambda _ctx, **_kwargs: health_calls.append("health"))
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: build_calls.append("build"))
    monkeypatch.setattr(
        ci.interface_tasks.e2e,
        "body",
        lambda _ctx, **kwargs: e2e_calls.append(kwargs),
    )

    ci.service_check.body(FakeContext({}), live_backend=True)

    assert up_calls == [(False, False)]
    assert migrate_calls == []
    assert health_calls == ["health"]
    assert build_calls == ["build"]
    assert e2e_calls == [
        {
            "project": "chromium",
            "spec": "e2e/specs/soma-governance-live.spec.ts",
            "live_backend": True,
            "workers": "1",
            "server_mode": "start",
        }
    ]


def test_toolchain_check_warns_when_not_strict():
    ctx = FakeContext(
        {
            "go version": FakeResult(stdout="go version go1.25.6 windows/amd64\n"),
            "node -v": FakeResult(stdout="v25.2.1\n"),
            "npm -v": FakeResult(stdout="11.6.2\n"),
        }
    )

    ci.toolchain_check.body(ctx, strict=False)


def test_entrypoint_check_verifies_runner_matrix():
    ctx = FakeContext(
        {
            "uv run inv -l": FakeResult(stdout="Available tasks:\n"),
            "uvx inv -l": FakeResult(exited=1, stderr="Package `inv` does not provide any executables.\n"),
            "uvx --from invoke inv -l": FakeResult(stdout="Available tasks:\n"),
        }
    )

    ci.entrypoint_check.body(ctx)

    assert ctx.commands == [
        "uv run inv -l",
        "uvx inv -l",
        "uvx --from invoke inv -l",
    ]


def test_entrypoint_check_fails_when_bare_uvx_behavior_changes():
    ctx = FakeContext(
        {
            "uv run inv -l": FakeResult(stdout="Available tasks:\n"),
            "uvx inv -l": FakeResult(stdout="Available tasks:\n"),
            "uvx --from invoke inv -l": FakeResult(stdout="Available tasks:\n"),
        }
    )

    with pytest.raises(SystemExit):
        ci.entrypoint_check.body(ctx)


def test_toolchain_check_fails_when_strict_and_mismatch():
    ctx = FakeContext(
        {
            "go version": FakeResult(stdout="go version go1.25.6 windows/amd64\n"),
            "node -v": FakeResult(stdout="v25.2.1\n"),
            "npm -v": FakeResult(stdout="11.6.2\n"),
        }
    )

    with pytest.raises(SystemExit):
        ci.toolchain_check.body(ctx, strict=True)


def test_release_preflight_fails_on_dirty_tree_before_baseline():
    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=" M README.md\n"),
        }
    )

    with pytest.raises(SystemExit):
        ci.release_preflight.body(ctx, e2e=False, strict_toolchain=False)

    assert ctx.commands == ["git status --porcelain"]


def test_release_preflight_rejects_unknown_lane_before_running_checks():
    ctx = FakeContext({})

    with pytest.raises(SystemExit) as excinfo:
        ci.release_preflight.body(ctx, lane="enterprise")

    assert "unsupported lane 'enterprise'" in str(excinfo.value)
    assert ctx.commands == []


def test_release_preflight_runs_toolchain_and_baseline_when_clean(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.test, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.typecheck, "body", lambda _ctx: None)

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
            "go version": FakeResult(stdout="go version go1.26.0 windows/amd64\n"),
            "node -v": FakeResult(stdout="v25.2.1\n"),
            "npm -v": FakeResult(stdout="11.6.2\n"),
            "go test ./... -count=1": FakeResult(),
        }
    )

    ci.release_preflight.body(ctx, e2e=False, strict_toolchain=True)

    assert "git status --porcelain" in ctx.commands
    assert "go version" in ctx.commands
    assert "go test ./... -count=1" in ctx.commands


def test_release_preflight_release_lane_runs_runtime_and_service_stages(monkeypatch):
    stage_order: list[str] = []
    baseline_calls: list[dict[str, object]] = []
    service_calls: list[dict[str, object]] = []

    monkeypatch.setattr(
        ci.toolchain_check,
        "body",
        lambda _ctx, **kwargs: stage_order.append(f"toolchain:{kwargs['strict']}"),
    )
    monkeypatch.setattr(ci, "_runtime_posture_check", lambda _ctx: stage_order.append("runtime"))
    monkeypatch.setattr(
        ci.baseline,
        "body",
        lambda _ctx, **kwargs: (stage_order.append("baseline"), baseline_calls.append(kwargs)),
    )
    monkeypatch.setattr(
        ci.service_check,
        "body",
        lambda _ctx, **kwargs: (stage_order.append("service"), service_calls.append(kwargs)),
    )

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    ci.release_preflight.body(ctx, lane="release", e2e=False, strict_toolchain=True)

    assert ctx.commands == ["git status --porcelain"]
    assert stage_order == ["toolchain:True", "runtime", "baseline", "service"]
    assert baseline_calls == [{"e2e": False}]
    assert service_calls == [{"live_backend": True}]


def test_release_preflight_release_lane_keeps_baseline_e2e_enabled_by_default(monkeypatch):
    stage_order: list[str] = []
    baseline_calls: list[dict[str, object]] = []
    service_calls: list[dict[str, object]] = []

    monkeypatch.setattr(ci.toolchain_check, "body", lambda _ctx, **_kwargs: stage_order.append("toolchain"))
    monkeypatch.setattr(ci, "_runtime_posture_check", lambda _ctx: stage_order.append("runtime"))
    monkeypatch.setattr(
        ci.baseline,
        "body",
        lambda _ctx, **kwargs: (stage_order.append("baseline"), baseline_calls.append(kwargs)),
    )
    monkeypatch.setattr(
        ci.service_check,
        "body",
        lambda _ctx, **kwargs: (stage_order.append("service"), service_calls.append(kwargs)),
    )

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    ci.release_preflight.body(ctx, lane="release")

    assert ctx.commands == ["git status --porcelain"]
    assert stage_order == ["toolchain", "runtime", "baseline", "service"]
    assert baseline_calls == [{"e2e": True}]
    assert service_calls == [{"live_backend": True}]


def test_release_preflight_runs_service_check_when_requested(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.interface_tasks.build, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.test, "body", lambda _ctx: None)
    monkeypatch.setattr(ci.interface_tasks.typecheck, "body", lambda _ctx: None)

    service_calls: list[dict[str, object]] = []
    monkeypatch.setattr(
        ci.service_check,
        "body",
        lambda _ctx, **kwargs: service_calls.append(kwargs),
    )

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
            "go version": FakeResult(stdout="go version go1.26.0 windows/amd64\n"),
            "node -v": FakeResult(stdout="v25.2.1\n"),
            "npm -v": FakeResult(stdout="11.6.2\n"),
            "go test ./... -count=1": FakeResult(),
        }
    )

    monkeypatch.setattr(ci.interface_tasks.e2e, "body", lambda _ctx, **_kwargs: None)

    ci.release_preflight.body(ctx, e2e=False, strict_toolchain=True, service_health=True, live_backend=True)
    assert service_calls == [{"live_backend": True}]


def test_release_preflight_live_backend_flag_implies_service_check(monkeypatch):
    monkeypatch.setattr(ci.toolchain_check, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.baseline, "body", lambda _ctx, **_kwargs: None)
    runtime_calls: list[str] = []
    service_calls: list[dict[str, object]] = []
    monkeypatch.setattr(ci, "_runtime_posture_check", lambda _ctx: runtime_calls.append("runtime"))
    monkeypatch.setattr(
        ci.service_check,
        "body",
        lambda _ctx, **kwargs: service_calls.append(kwargs),
    )

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    ci.release_preflight.body(ctx, e2e=False, live_backend=True)

    assert runtime_calls == []
    assert service_calls == [{"live_backend": True}]


def test_runtime_posture_check_fails_when_no_explicit_endpoint_is_configured(monkeypatch, tmp_path):
    headroom_calls: list[dict[str, object]] = []
    monkeypatch.setattr(
        ci.cache_tasks,
        "ensure_disk_headroom",
        lambda **kwargs: headroom_calls.append(kwargs),
    )
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    for env_name in (
        "MYCELIS_COMPOSE_OLLAMA_HOST",
        "MYCELIS_K8S_TEXT_ENDPOINT",
        "MYCELIS_K8S_MEDIA_ENDPOINT",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT",
    ):
        monkeypatch.delenv(env_name, raising=False)

    with pytest.raises(SystemExit) as excinfo:
        ci._runtime_posture_check(FakeContext({}))

    assert headroom_calls == [{"min_free_gb": 12, "reason": "release preflight posture"}]
    assert "no explicit AI endpoint configured" in str(excinfo.value)


def test_runtime_posture_check_reads_compose_env_file_when_process_env_is_empty(monkeypatch, tmp_path):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    for env_name in (
        "MYCELIS_COMPOSE_OLLAMA_HOST",
        "MYCELIS_K8S_TEXT_ENDPOINT",
        "MYCELIS_K8S_MEDIA_ENDPOINT",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT",
    ):
        monkeypatch.delenv(env_name, raising=False)

    (tmp_path / ".env.compose").write_text(
        "MYCELIS_COMPOSE_OLLAMA_HOST=http://192.168.50.156:11434\n",
        encoding="utf-8",
    )

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        return 200, "{}"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == ["http://192.168.50.156:11434/api/tags"]


def test_runtime_posture_check_includes_provider_specific_endpoint_from_env_file(monkeypatch, tmp_path):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    for env_name in (
        "MYCELIS_COMPOSE_OLLAMA_HOST",
        "MYCELIS_K8S_TEXT_ENDPOINT",
        "MYCELIS_K8S_MEDIA_ENDPOINT",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT",
        "MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_ENDPOINT",
    ):
        monkeypatch.delenv(env_name, raising=False)

    (tmp_path / ".env").write_text(
        "MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_ENDPOINT=http://192.168.50.157:11434/v1\n",
        encoding="utf-8",
    )

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        return 401, "unauthorized"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == ["http://192.168.50.157:11434/v1/models"]


def test_runtime_posture_check_rejects_loopback_ai_endpoint(monkeypatch):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://127.0.0.1:11434")

    with pytest.raises(SystemExit):
        ci._runtime_posture_check(FakeContext({}))


def test_runtime_posture_check_probes_compose_ai_endpoint_with_fallback(monkeypatch):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://10.0.0.5:11434")

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        if url.endswith("/api/tags"):
            return 404, "not found"
        return 401, "unauthorized"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == [
        "http://10.0.0.5:11434/api/tags",
        "http://10.0.0.5:11434/v1/models",
    ]


def test_runtime_posture_check_probes_k8s_ai_endpoint(monkeypatch, tmp_path):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    monkeypatch.setenv("MYCELIS_K8S_TEXT_ENDPOINT", "http://10.0.0.6:11434/v1")

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        return 401, "unauthorized"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == [
        "http://10.0.0.6:11434/v1/models",
    ]


def test_release_preflight_runs_runtime_posture_when_requested(monkeypatch):
    monkeypatch.setattr(ci.toolchain_check, "body", lambda _ctx, **_kwargs: None)
    baseline_calls: list[dict[str, object]] = []
    runtime_calls: list[str] = []
    monkeypatch.setattr(ci.baseline, "body", lambda _ctx, **kwargs: baseline_calls.append(kwargs))
    monkeypatch.setattr(ci, "_runtime_posture_check", lambda _ctx: runtime_calls.append("runtime"))

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    ci.release_preflight.body(
        ctx,
        e2e=False,
        strict_toolchain=True,
        service_health=False,
        live_backend=False,
        runtime_posture=True,
    )

    assert runtime_calls == ["runtime"]
    assert baseline_calls == [{"e2e": False}]
