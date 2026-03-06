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

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return self.command_results.get(command, FakeResult())

    @contextmanager
    def cd(self, _path: str):
        yield


def test_baseline_runs_expected_commands_without_e2e(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)

    ctx = FakeContext(
        {
            "go test ./... -count=1": FakeResult(),
            "cd interface && npm run build": FakeResult(),
            "cd interface && npx tsc --noEmit": FakeResult(),
            "cd interface && npx vitest run --reporter=dot": FakeResult(),
        }
    )

    ci.baseline.body(ctx, e2e=False)

    assert "go test ./... -count=1" in ctx.commands
    assert "cd interface && npm run build" in ctx.commands
    assert "cd interface && npx tsc --noEmit" in ctx.commands
    assert "cd interface && npx vitest run --reporter=dot" in ctx.commands
    assert "cd interface && npx playwright test --reporter=dot" not in ctx.commands


def test_baseline_runs_playwright_when_e2e_enabled(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)

    ctx = FakeContext(
        {
            "go test ./... -count=1": FakeResult(),
            "cd interface && npm run build": FakeResult(),
            "cd interface && npx tsc --noEmit": FakeResult(),
            "cd interface && npx vitest run --reporter=dot": FakeResult(),
            "cd interface && npx playwright test --reporter=dot": FakeResult(),
        }
    )

    ci.baseline.body(ctx, e2e=True)

    assert "cd interface && npx playwright test --reporter=dot" in ctx.commands


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


def test_release_preflight_runs_toolchain_and_baseline_when_clean(monkeypatch):
    monkeypatch.setattr(ci.logging_tasks.check_schema, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.logging_tasks.check_topics, "body", lambda _ctx, **_kwargs: None)
    monkeypatch.setattr(ci.quality.max_lines, "body", lambda _ctx, **_kwargs: None)

    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
            "go version": FakeResult(stdout="go version go1.26.0 windows/amd64\n"),
            "node -v": FakeResult(stdout="v25.2.1\n"),
            "npm -v": FakeResult(stdout="11.6.2\n"),
            "go test ./... -count=1": FakeResult(),
            "cd interface && npm run build": FakeResult(),
            "cd interface && npx tsc --noEmit": FakeResult(),
            "cd interface && npx vitest run --reporter=dot": FakeResult(),
        }
    )

    ci.release_preflight.body(ctx, e2e=False, strict_toolchain=True)

    assert "git status --porcelain" in ctx.commands
    assert "go version" in ctx.commands
    assert "go test ./... -count=1" in ctx.commands
    assert "cd interface && npm run build" in ctx.commands
    assert "cd interface && npx tsc --noEmit" in ctx.commands
