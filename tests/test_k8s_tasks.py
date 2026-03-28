from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path

from invoke import Context

from ops import k8s


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""

    @property
    def ok(self) -> bool:
        return self.exited == 0


class FakeContext(Context):
    def __init__(self, command_results: dict[str, FakeResult] | None = None):
        super().__init__()
        self.command_results = command_results or {}
        self.commands: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return self.command_results.get(command, FakeResult())

    @contextmanager
    def cd(self, _path: str):
        yield


def test_deploy_uses_core_build_task_body(monkeypatch):
    ctx = FakeContext()
    build_calls: list[str] = []

    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: build_calls.append("build") or "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")

    k8s.deploy.body(ctx)

    assert build_calls == ["build"]
    assert any("kind load docker-image mycelis/core:v0.1.0-deadbee" in command for command in ctx.commands)
    assert any("--set image.tag=v0.1.0-deadbee" in command for command in ctx.commands)


def test_init_reads_and_writes_kind_config_under_root_dir(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    monkeypatch.setattr(k8s, "_cluster_exists", lambda _c: False)
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)

    (tmp_path / "kind-config.yaml").write_text(
        "mounts:\n  - hostPath: ./ops\n  - hostPath: ./logs\n",
        encoding="utf-8",
    )

    k8s.init.body(ctx)

    generated = tmp_path / "kind-config.gen.yaml"
    assert generated.exists()
    generated_text = generated.read_text(encoding="utf-8")
    assert str(tmp_path / "ops").replace("\\", "/") in generated_text
    assert str(tmp_path / "logs").replace("\\", "/") in generated_text
    assert any(str(generated) in command for command in ctx.commands)


def test_reset_prints_managed_status_command(monkeypatch, capsys):
    ctx = FakeContext()
    monkeypatch.setattr(k8s, "up", lambda _ctx: ctx.commands.append("up"))

    k8s.reset.body(ctx)

    output = capsys.readouterr().out
    assert "uv run inv k8s.status" in output
    assert "Run 'inv k8s.status' to verify." not in output
