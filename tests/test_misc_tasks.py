from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

from ops import misc, misc_support


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


def test_team_namespace_has_no_raw_nats_publisher():
    """`team.architecture-sync` published raw to swarm subjects from Python,
    bypassing Core dispatch, receipts and mission events (CONS-SURFACE)."""
    assert sorted(misc.ns_team.task_names) == ["worktree-triage"]
    for removed in (
        "architecture_sync",
        "_architecture_sync_directives",
        "_read_nats_line",
        "_drain_nats_messages",
        "_format_sync_reply",
        "_format_sync_output",
    ):
        assert not hasattr(misc, removed), removed
    assert not hasattr(misc_support, "architecture_sync_directives")


def test_build_worktree_triage_maps_changed_paths_to_installs_and_commands():
    triage = misc._build_worktree_triage(
        "\n".join(
            [
                " M core/internal/swarm/team.go",
                " M interface/components/dashboard/OperationsBoard.tsx",
                " M ops/misc.py",
                "R  docs/old.md -> docs/new.md",
            ]
        )
    )

    assert [area["name"] for area in triage["areas"]] == [
        "Core runtime",
        "Docs and state",
        "Interface",
        "Python automation",
    ]
    assert "cd core && go mod download" in triage["priority_installs"]
    assert "uv run inv interface.install" in triage["priority_installs"]
    assert "uv sync --all-packages --dev" in triage["priority_installs"]
    assert "uv run inv core.test" in triage["recommended_commands"]
    assert "uv run inv core.compile" in triage["recommended_commands"]
    assert "uv run inv interface.test" in triage["recommended_commands"]
    assert "uv run inv interface.typecheck" in triage["recommended_commands"]
    assert "uv run inv interface.build" in triage["recommended_commands"]
    assert (
        "$env:PYTHONPATH='.'; uv run pytest tests/test_core_tasks.py tests/test_ci_pipeline_tasks.py tests/test_ci_preflight_tasks.py tests/test_ci_runtime_posture_tasks.py tests/test_ci_service_tasks.py tests/test_interface_tasks.py tests/test_interface_e2e_tasks.py tests/test_interface_command_tasks.py tests/test_k8s_tasks.py tests/test_lifecycle_tasks.py tests/test_misc_tasks.py tests/test_cleanup_tasks.py -q"
        in triage["recommended_commands"]
    )
    assert "uv run inv ci.build" in triage["recommended_commands"]
    assert "$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q" in triage["recommended_commands"]


def test_worktree_triage_reports_clean_tree(capsys):
    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    misc.worktree_triage.body(ctx)

    output = capsys.readouterr().out
    assert "Working tree: clean" in output
    assert "uv run inv install" in output
    assert "uv run inv ci.entrypoint-check" in output
    assert "uv run inv ci.baseline" in output
    assert "none triggered by current paths" in output


def test_worktree_triage_expected_targets_cover_task_contract_docs(capsys):
    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    misc.worktree_triage.body(ctx)

    output = capsys.readouterr().out
    assert "docs/architecture-library/MYCELIS_CANONICAL_PRD.md" in output
    assert "docs/LOCAL_DEV_WORKFLOW.md" in output
    assert "docs/architecture/OPERATIONS.md" in output
    assert "ops/README.md" in output


def test_worktree_triage_has_no_windows_host_note(monkeypatch, capsys):
    """WSL `dev` is the single development surface; there is no Windows-side note."""
    ctx = FakeContext(
        {
            "git status --porcelain": FakeResult(stdout=""),
        }
    )

    misc.worktree_triage.body(ctx)

    output = capsys.readouterr().out
    assert "Windows" not in output
