from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

from ops import misc


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


def test_clean_generated_only_removes_allowlisted_paths(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)

    generated = (
        tmp_path / ".venv",
        tmp_path / "interface" / "node_modules",
        tmp_path / "interface" / ".next",
        tmp_path / "workspace" / "tool-cache",
        tmp_path / "interface" / "test-results",
        tmp_path / "interface" / "playwright-report",
        tmp_path / ".pytest_cache",
        tmp_path / "core" / "bin",
    )
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", generated)
    monkeypatch.setattr(
        misc,
        "REPORT_ARTIFACT_TARGETS",
        (
            tmp_path / "interface" / "test-results",
            tmp_path / "interface" / "playwright-report",
            tmp_path / ".pytest_cache",
        ),
    )
    monkeypatch.setattr(
        misc,
        "WSL_HANDOFF_TARGETS",
        (
            tmp_path / ".venv",
            tmp_path / "interface" / "node_modules",
            tmp_path / "interface" / ".next",
        ),
    )

    for target in generated:
        target.mkdir(parents=True, exist_ok=True)
        (target / "marker.txt").write_text("x", encoding="utf-8")

    runtime_data = tmp_path / "workspace" / "docker-compose" / "data"
    runtime_data.mkdir(parents=True, exist_ok=True)
    (runtime_data / "keep.txt").write_text("keep", encoding="utf-8")

    misc.clean_generated.body(Context())

    output = capsys.readouterr().out
    assert "workspace/docker-compose/data is intentionally untouched." in output
    assert "keep heavy build/test artifacts in the WSL checkout" in output
    for target in generated:
        assert not target.exists()
    assert runtime_data.exists()
    assert (runtime_data / "keep.txt").exists()


def test_clean_windows_dev_residue_requires_windows(monkeypatch):
    monkeypatch.setattr(misc, "is_windows", lambda: False)

    try:
        misc.clean_windows_dev_residue.body(Context())
    except SystemExit as exc:
        assert "Windows-only" in str(exc)
    else:
        raise AssertionError("expected SystemExit for non-Windows host")


def test_clean_windows_dev_residue_reports_source_only_guidance(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(misc, "is_windows", lambda: True)
    targets = (
        tmp_path / ".venv",
        tmp_path / "interface" / "node_modules",
    )
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", targets)

    for target in targets:
        target.mkdir(parents=True, exist_ok=True)
        (target / "marker.txt").write_text("x", encoding="utf-8")

    misc.clean_windows_dev_residue.body(Context())

    output = capsys.readouterr().out
    assert "Windows source-only reminder:" in output
    assert "run install/build/test/compose from the WSL checkout" in output
    for target in targets:
        assert not target.exists()


def test_clean_disk_status_reports_repo_total_and_vhd_guidance(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(misc, "ROOT_DIR", tmp_path)
    targets = (
        tmp_path / ".venv",
        tmp_path / "workspace" / "tool-cache",
    )
    monkeypatch.setattr(misc, "GENERATED_ARTIFACT_TARGETS", targets)

    first = targets[0]
    first.mkdir(parents=True, exist_ok=True)
    (first / "one.bin").write_bytes(b"a" * 1024)

    second = targets[1]
    second.mkdir(parents=True, exist_ok=True)
    (second / "two.bin").write_bytes(b"b" * 2048)

    misc.clean_disk_status.body(Context())

    output = capsys.readouterr().out
    assert ".venv: present" in output
    assert "workspace/tool-cache: present" in output
    assert "Repo-local generated total:" in output
    assert "Windows should stay source-only" in output
    assert "WSL VHD slack space are outside repo cleanup" in output
    assert "wsl --shutdown" in output


def test_worktree_triage_on_windows_mentions_wsl_validation(monkeypatch, capsys):
    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )
    monkeypatch.setattr(misc, "is_windows", lambda: True)

    misc.worktree_triage.body(ctx)

    output = capsys.readouterr().out
    assert "Windows host note:" in output
    assert "run heavy validation from the WSL checkout" in output
