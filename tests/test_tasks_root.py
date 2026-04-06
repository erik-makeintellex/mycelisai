from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

import tasks
from ops import test as test_tasks


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""


class FakeContext(Context):
    def __init__(self):
        super().__init__()
        self.commands: list[str] = []
        self.cd_paths: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return FakeResult()

    @contextmanager
    def cd(self, path: str):
        self.cd_paths.append(path)
        yield


def test_root_collection_registers_expected_namespaces():
    assert sorted(tasks.ns.collections.keys()) == [
        "auth",
        "cache",
        "ci",
        "clean",
        "cognitive",
        "compose",
        "core",
        "db",
        "device",
        "interface",
        "k8s",
        "lifecycle",
        "logging",
        "proto",
        "quality",
        "relay",
        "team",
        "test",
    ]


def test_root_collection_exports_expected_task_surface():
    assert sorted(tasks.ns.task_names.keys()) == [
        "auth.dev-key",
        "cache.apply-user-policy",
        "cache.clean",
        "cache.status",
        "ci.baseline",
        "ci.build",
        "ci.check",
        "ci.deploy",
        "ci.entrypoint-check",
        "ci.lint",
        "ci.release-preflight",
        "ci.service-check",
        "ci.test",
        "ci.toolchain-check",
        "clean.legacy",
        "cognitive.install",
        "cognitive.llm",
        "cognitive.media",
        "cognitive.status",
        "cognitive.stop",
        "cognitive.up",
        "compose.down",
        "compose.health",
        "compose.logs",
        "compose.migrate",
        "compose.status",
        "compose.up",
        "core.build",
        "core.clean",
        "core.compile",
        "core.restart",
        "core.run",
        "core.smoke",
        "core.stop",
        "core.test",
        "db.create",
        "db.migrate",
        "db.reset",
        "db.status",
        "device.boot",
        "install",
        "interface.build",
        "interface.check",
        "interface.clean",
        "interface.dev",
        "interface.e2e",
        "interface.install",
        "interface.lint",
        "interface.restart",
        "interface.stop",
        "interface.test",
        "interface.test-coverage",
        "interface.typecheck",
        "k8s.bridge",
        "k8s.deploy",
        "k8s.init",
        "k8s.recover",
        "k8s.reset",
        "k8s.status",
        "k8s.up",
        "k8s.wait",
        "lifecycle.down",
        "lifecycle.health",
        "lifecycle.memory-restart",
        "lifecycle.restart",
        "lifecycle.status",
        "lifecycle.up",
        "logging.check-schema",
        "logging.check-topics",
        "proto.generate",
        "quality.max-lines",
        "relay.demo",
        "relay.test",
        "team.architecture-sync",
        "team.worktree-triage",
        "test.all",
        "test.coverage",
        "test.e2e",
    ]


def test_install_skips_optional_engines_by_default(capsys):
    ctx = FakeContext()

    tasks.install.body(ctx, optional_engines=False)

    assert ctx.commands == [
        "uv sync --all-packages --dev",
        "go mod download",
        "npm install --prefix interface",
    ]
    assert ctx.cd_paths == ["core"]
    output = capsys.readouterr().out
    assert "Skipping optional cognitive engine dependencies." in output


def test_install_can_include_optional_engines():
    ctx = FakeContext()

    tasks.install.body(ctx, optional_engines=True)

    assert ctx.commands == [
        "uv sync --all-packages --dev",
        "go mod download",
        "npm install --prefix interface",
        "uv sync",
    ]
    assert ctx.cd_paths == ["core", "cognitive"]


def test_test_all_normalizes_failures_to_system_exit(monkeypatch, capsys):
    monkeypatch.setattr(test_tasks.core.test, "body", lambda c: None)
    monkeypatch.setattr(test_tasks.interface.test, "body", lambda c: (_ for _ in ()).throw(SystemExit(3)))

    with __import__("pytest").raises(SystemExit) as excinfo:
        test_tasks.all.body(FakeContext())

    assert excinfo.value.code == 1
    assert "Test Failure: see task output above." in capsys.readouterr().out


def test_test_e2e_alias_forwards_workers_and_server_mode(monkeypatch):
    captured: dict[str, object] = {}

    monkeypatch.setattr(
        test_tasks.interface.e2e,
        "body",
        lambda c, headed=False, project="", spec="", live_backend=False, workers="", server_mode="dev": captured.update(
            {
                "headed": headed,
                "project": project,
                "spec": spec,
                "live_backend": live_backend,
                "workers": workers,
                "server_mode": server_mode,
            }
        ),
    )

    test_tasks.e2e.body(
        FakeContext(),
        headed=True,
        project="chromium",
        spec="e2e/specs/navigation.spec.ts",
        live_backend=True,
        workers="1",
        server_mode="start",
    )

    assert captured == {
        "headed": True,
        "project": "chromium",
        "spec": "e2e/specs/navigation.spec.ts",
        "live_backend": True,
        "workers": "1",
        "server_mode": "start",
    }
