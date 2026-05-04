import pytest

from ops import ci
from tests.ci_task_support import FakeContext, FakeResult


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
