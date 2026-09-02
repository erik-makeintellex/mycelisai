from __future__ import annotations

import re
import tomllib
from pathlib import Path

from ops import proto_relay


ROOT = Path(__file__).resolve().parents[1]

RETIRED_TOOLING_PATHS = (
    "cli/README.md",
    "cli/__init__.py",
    "cli/main.py",
    "cli/pyproject.toml",
    "cli/tests/__init__.py",
    "cli/tests/test_main.py",
    "core/internal/identity/schema.sql",
    "scripts/apply_migrations.py",
    "scripts/deploy_schema.py",
    "scripts/dev.py",
    "scripts/dev/forward_ports.sh",
    "scripts/dev/stop_ports.sh",
    "scripts/drills/disconnect.py",
    "scripts/fix_streams.py",
    "scripts/qa/audit_archivist.go",
    "scripts/qa/go.mod",
    "scripts/qa/go.sum",
    "scripts/qa/inject_artifact.go",
    "scripts/recover_agents.py",
    "scripts/reproduce_issue.py",
    "scripts/run_researcher.py",
    "scripts/seed_registry.py",
    "scripts/send_message.py",
    "scripts/verification/list_streams.py",
    "scripts/verification/verify_agent_flow.py",
    "scripts/verification/verify_governance.py",
    "scripts/verification/verify_mcp.py",
    "scripts/verification/verify_sse.py",
    "scripts/verification/verify_stack.py",
    "scripts/verification/verify_teams.py",
    "scripts/verify_governance.py",
    "scripts/verify_hierarchy.py",
    "scripts/verify_memory.py",
    "scripts/verify_team_topology.py",
)


def test_obsolete_cli_scripts_and_duplicate_identity_schema_are_absent():
    present = [path for path in RETIRED_TOOLING_PATHS if (ROOT / path).exists()]

    assert not present, f"retired tooling paths restored: {present}"


def test_uv_workspace_omits_retired_cli_and_console_alias():
    project = tomllib.loads((ROOT / "pyproject.toml").read_text(encoding="utf-8"))
    workspace_members = project["tool"]["uv"]["workspace"]["members"]

    assert "cli" not in workspace_members
    assert "scripts" not in project["project"]


def test_uv_lock_omits_retired_cli_package():
    lock_text = (ROOT / "uv.lock").read_text(encoding="utf-8")

    assert not re.search(r'^\s*"cli",$', lock_text, flags=re.MULTILINE)
    assert not re.search(r'^name = "cli"$', lock_text, flags=re.MULTILINE)


class _RecordingContext:
    def __init__(self):
        self.commands: list[str] = []

    def run(self, command: str, **_kwargs):
        self.commands.append(command)


def test_proto_generation_uses_governed_temporary_path(monkeypatch, tmp_path):
    sdk_dir = tmp_path / "sdk" / "python"
    monkeypatch.setattr(proto_relay, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(proto_relay, "SDK_DIR", sdk_dir)
    context = _RecordingContext()

    proto_relay.generate.body(context)

    assert "scripts/" not in context.commands[0]
    assert "workspace/tool-cache/proto/gen-proto-go-" in context.commands[0]
    assert not list((tmp_path / "workspace" / "tool-cache" / "proto").glob("gen-proto-go-*.sh"))
