from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

import tasks


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
