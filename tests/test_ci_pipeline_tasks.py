import pytest

from ops import ci
from tests.ci_task_support import FakeContext, FakeResult


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
