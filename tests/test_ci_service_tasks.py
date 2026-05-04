import pytest

from ops import ci
from tests.ci_task_support import FakeContext, FakeResult


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
