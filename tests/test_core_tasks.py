from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

from ops import core


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


def test_compile_builds_repo_local_binary(monkeypatch):
    ctx = FakeContext()
    monkeypatch.setattr(core, "is_windows", lambda: False)

    core.compile.body(ctx)

    assert ctx.cd_paths == [str(core.CORE_DIR)]
    assert ctx.commands == ["go build -v -o bin/server ./cmd/server"]


def test_build_uses_compile_and_never_tags_latest(monkeypatch):
    ctx = FakeContext()
    compile_calls: list[str] = []

    monkeypatch.setattr(core.compile, "body", lambda _c: compile_calls.append("compile"))
    monkeypatch.setattr(core, "get_version", lambda _c: "v0.1.0-deadbee")

    core.build.body(ctx)

    assert compile_calls == ["compile"]
    assert ctx.commands == ["docker build -t mycelis/core:v0.1.0-deadbee -f core/Dockerfile ."]
    assert not any("latest" in command for command in ctx.commands)
