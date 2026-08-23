from __future__ import annotations

import re
from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path

from invoke import Context

import tasks


ROOT = Path(__file__).resolve().parents[1]


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
        "api",
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
        "wsl",
    ]


def test_root_collection_exports_expected_task_surface():
    assert sorted(tasks.ns.task_names.keys()) == [
        "api.delivery-proof",
        "auth.break-glass-key",
        "auth.dev-key",
        "auth.posture",
        "cache.apply-user-policy",
        "cache.clean",
        "cache.guard",
        "cache.status",
        "ci.baseline",
        "ci.build",
        "ci.check",
        "ci.entrypoint-check",
        "ci.lint",
        "ci.release-preflight",
        "ci.service-check",
        "ci.test",
        "ci.toolchain-check",
        "clean.disk-status",
        "clean.generated",
        "clean.reports",
        "clean.windows-dev-residue",
        "clean.wsl-handoff",
        "cognitive.install",
        "cognitive.llm",
        "cognitive.media",
        "cognitive.media-gateway",
        "cognitive.status",
        "cognitive.stop",
        "cognitive.up",
        "compose.down",
        "compose.health",
        "compose.infra-health",
        "compose.infra-up",
        "compose.logs",
        "compose.migrate",
        "compose.status",
        "compose.storage-health",
        "compose.up",
        "compose.warm-cognitive",
        "core.build",
        "core.clean",
        "core.compile",
        "core.package",
        "core.restart",
        "core.run",
        "core.smoke",
        "core.stop",
        "core.test",
        "db.clear-runtime-context",
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
        "k8s.standards",
        "k8s.status",
        "k8s.up",
        "k8s.wait",
        "lifecycle.down",
        "lifecycle.first-boot-proof",
        "lifecycle.health",
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
        "test.coverage",
        "wsl.cycle",
        "wsl.refresh",
        "wsl.status",
        "wsl.validate",
    ]


def test_root_task_surface_stays_within_operator_budget():
    assert len(tasks.ns.task_names) <= 95


def test_documented_invoke_commands_are_registered():
    documentation = {
        ROOT / "README.md",
        ROOT / "AGENTS.md",
        ROOT / "ops" / "README.md",
        *sorted((ROOT / "docs").rglob("*.md")),
        *sorted((ROOT / "architecture").rglob("*.md")),
    }
    command_pattern = re.compile(r"uv run inv ([a-z][a-z0-9]*(?:[.-][a-z0-9]+)*)(?![a-z0-9.*-])")
    documented_tasks = {
        match
        for path in documentation
        for match in command_pattern.findall(path.read_text(encoding="utf-8"))
    }
    missing = sorted(documented_tasks - set(tasks.ns.task_names))

    assert not missing, f"Documentation references unregistered Invoke tasks: {missing}"


def test_install_skips_optional_engines_by_default(capsys):
    ctx = FakeContext()

    tasks.install.body(ctx, optional_engines=False)

    assert ctx.commands == [
        "uv sync --all-packages --dev",
        'uv run python -c "import RNS; print(RNS.__version__)"',
        "uvx --from rns rnstatus --help",
        "go mod download",
        "npm ci --prefix interface",
        "npx --prefix interface playwright install chromium",
    ]
    assert ctx.cd_paths == ["core"]
    output = capsys.readouterr().out
    assert "Skipping optional cognitive engine dependencies." in output


def test_install_can_include_optional_engines():
    ctx = FakeContext()

    tasks.install.body(ctx, optional_engines=True)

    assert ctx.commands == [
        "uv sync --all-packages --dev",
        'uv run python -c "import RNS; print(RNS.__version__)"',
        "uvx --from rns rnstatus --help",
        "go mod download",
        "npm ci --prefix interface",
        "npx --prefix interface playwright install chromium",
        "uv sync",
    ]
    assert ctx.cd_paths == ["core", "cognitive"]
