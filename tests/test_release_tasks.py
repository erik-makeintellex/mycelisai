from __future__ import annotations

from dataclasses import dataclass
from typing import Callable

from invoke import Context
import pytest

from ops import release, version


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""


class FakeContext(Context):
    def __init__(self, handler: Callable[[str], FakeResult] | None = None):
        super().__init__()
        self.commands: list[str] = []
        self._handler = handler

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        if self._handler is not None:
            return self._handler(command)
        return FakeResult()


def test_validate_release_tag_accepts_release_and_rc_tags():
    assert version.validate_release_tag("v0.8.3") == "v0.8.3"
    assert version.validate_release_tag("  v0.8.3-rc1  ") == "v0.8.3-rc1"
    assert version.validate_release_tag("v1.2.3-rc12") == "v1.2.3-rc12"


@pytest.mark.parametrize("bad_tag", ["0.8.3", "v0.8", "v0.8.3-beta", "v0.8.3-rc", ""])
def test_validate_release_tag_rejects_invalid_tags(bad_tag: str):
    with pytest.raises(ValueError, match="Invalid release tag"):
        version.validate_release_tag(bad_tag)


def test_read_semver_reads_and_validates_version_file(tmp_path):
    (tmp_path / "VERSION").write_text("0.8.3\n", encoding="utf-8")
    assert version.read_semver(tmp_path) == "0.8.3"

    (tmp_path / "VERSION").write_text("invalid-ver\n", encoding="utf-8")
    with pytest.raises(ValueError, match="Invalid semantic version"):
        version.read_semver(tmp_path)


def test_get_version_builds_semver_sha_tag(monkeypatch):
    ctx = FakeContext(lambda _cmd: FakeResult(stdout="abc1234\n"))
    monkeypatch.setattr(version, "read_semver", lambda _root=None: "0.8.3")

    assert version.get_version(ctx) == "v0.8.3-abc1234"
    assert ctx.commands == ["git rev-parse --short HEAD"]


def test_release_status_reports_branch_version_tag_and_prunable_branches(monkeypatch, capsys):
    responses = {
        "git rev-parse --abbrev-ref HEAD": FakeResult(stdout="dev\n"),
        "git status --porcelain": FakeResult(stdout=""),
        "git describe --tags --abbrev=0 main": FakeResult(stdout="v0.8.3\n"),
        "git rev-list --left-right --count main...dev": FakeResult(stdout="0\t2\n"),
        "git branch --merged main": FakeResult(
            stdout="* dev\n  main\n  archive/self-hosted-pivot-stash\n  feature/foo\n"
        ),
    }
    ctx = FakeContext(lambda cmd: responses[cmd])
    monkeypatch.setattr(release, "read_semver", lambda: "0.8.3")

    release.status.body(ctx)

    out = capsys.readouterr().out
    assert "Branch: dev (clean)" in out
    assert "Tracked VERSION: 0.8.3" in out
    assert "Latest tag on main: v0.8.3" in out
    assert "Divergence (main...dev): 0\t2" in out
    assert "archive/self-hosted-pivot-stash, feature/foo" in out


def test_prune_branches_deletes_merged_feature_and_archive_branches_only():
    merged_out = "* dev\n  main\n  archive/self-hosted-pivot-stash\n  feature/foo\n"
    ctx = FakeContext(
        lambda cmd: FakeResult(stdout=merged_out) if cmd == "git branch --merged main" else FakeResult()
    )

    release.prune_branches.body(ctx, remote=True, dry_run=False)

    assert ctx.commands == [
        "git branch --merged main",
        "git branch -d archive/self-hosted-pivot-stash",
        "git branch -d feature/foo",
        "git fetch --prune origin",
    ]


def test_prune_branches_dry_run_never_mutates_refs(capsys):
    merged_out = "* main\n  dev\n  archive/self-hosted-pivot-stash\n"
    ctx = FakeContext(lambda _cmd: FakeResult(stdout=merged_out))

    release.prune_branches.body(ctx, remote=True, dry_run=True)

    assert ctx.commands == ["git branch --merged main"]
    out = capsys.readouterr().out
    assert "[DRY-RUN] Would delete merged branch: archive/self-hosted-pivot-stash" in out


def test_stage_fails_closed_when_working_tree_is_dirty():
    ctx = FakeContext(lambda _cmd: FakeResult(stdout=" M README.md\n"))

    with pytest.raises(SystemExit, match="Working tree must be clean"):
        release.stage.body(ctx, tag="v0.8.3")


def test_stage_fails_closed_when_tag_is_invalid():
    ctx = FakeContext()

    with pytest.raises(ValueError, match="Invalid release tag"):
        release.stage.body(ctx, tag="0.8.3")


def test_stage_dry_run_validates_gates_without_mutating_refs(monkeypatch, capsys):
    max_lines_calls: list[int] = []
    monkeypatch.setattr(release.quality.max_lines, "body", lambda _c, limit=300: max_lines_calls.append(limit))
    ctx = FakeContext(lambda _cmd: FakeResult(stdout=""))

    release.stage.body(ctx, tag="v0.8.3-rc1", dry_run=True)

    assert max_lines_calls == [300]
    assert ctx.commands == ["git status --porcelain", "git diff --check"]
    out = capsys.readouterr().out
    assert "git checkout main" in out
    assert "git merge --ff-only dev" in out
    assert 'git tag -a v0.8.3-rc1 -m "Release v0.8.3-rc1 staged on main from dev"' in out


def test_stage_executes_fast_forward_tag_prune_and_return_to_source_branch(monkeypatch):
    max_lines_calls: list[int] = []
    monkeypatch.setattr(release.quality.max_lines, "body", lambda _c, limit=300: max_lines_calls.append(limit))

    def _handler(cmd: str) -> FakeResult:
        if cmd == "git branch --merged main":
            return FakeResult(stdout="* main\n  dev\n  archive/self-hosted-pivot-stash\n")
        return FakeResult(stdout="")

    ctx = FakeContext(_handler)

    release.stage.body(ctx, tag="v0.8.3", message="Stabilized V8.3 milestone", source_branch="dev", prune=True)

    assert max_lines_calls == [300]
    assert ctx.commands == [
        "git status --porcelain",
        "git diff --check",
        "git checkout main",
        "git merge --ff-only dev",
        'git tag -a v0.8.3 -m "Stabilized V8.3 milestone"',
        "git branch --merged main",
        "git branch -d archive/self-hosted-pivot-stash",
        "git checkout dev",
    ]
