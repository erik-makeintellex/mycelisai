from __future__ import annotations

import re
import subprocess
import tomllib
from pathlib import Path

from ops import proto_relay


ROOT = Path(__file__).resolve().parents[1]

RETIRED_REPOSITORY_PATHS = (
    "cli/README.md",
    "cli/__init__.py",
    "cli/main.py",
    "cli/pyproject.toml",
    "cli/tests/__init__.py",
    "cli/tests/test_main.py",
    "core/internal/identity/schema.sql",
    "deploy/charts/mycelis-core/Chart.yaml",
    "deploy/charts/mycelis-core/values.yaml",
    "deploy/docker/Dockerfile.core",
    "interface/e2e/specs/v8-ui-testing-agentry.spec.ts",
    "k8s/network-policy-provisioner.yaml",
    "k8s/network-policy-registry.yaml",
    "k8s/registry-deployment.yaml",
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
    "sdk/python/src/relay/proto/envelope_pb2.py",
    "sdk/python/src/relay/proto/envelope_pb2_grpc.py",
    "sdk/python/src/relay/proto/swarm/v1/swarm_pb2_grpc.py",
    "sdk/python/src/scip/proto/envelope_pb2_grpc.py",
    "tests/test_v8_3_outcome_docs.py",
    "tests/test_v8_3_ui_ux_brief.py",
    "tests/test_framework_runs_sidecar.py",
    "tests/ui/browser_qa_plan_generic_complex_app_output.md",
    "tests/ui/browser_qa_plan_workspace_chat.md",
    "tests/ui/browser_qa_workflow_variants_reboot.md",
)


def test_retired_repository_paths_are_absent():
    present = [path for path in RETIRED_REPOSITORY_PATHS if (ROOT / path).exists()]

    assert not present, f"retired tooling paths restored: {present}"


def test_personal_monokle_config_is_ignored_and_untracked():
    assert ".monokle" in (ROOT / ".gitignore").read_text(encoding="utf-8").splitlines()
    tracked = subprocess.run(
        ["git", "ls-files", "--", ".monokle"],
        cwd=ROOT,
        text=True,
        capture_output=True,
        check=True,
    ).stdout
    assert tracked == ""


def test_uv_workspace_omits_retired_cli_and_console_alias():
    project = tomllib.loads((ROOT / "pyproject.toml").read_text(encoding="utf-8"))
    workspace_members = project["tool"]["uv"]["workspace"]["members"]

    assert "cli" not in workspace_members
    assert "scripts" not in project["project"]


def test_pytest_discovery_is_bounded_to_owned_test_roots():
    project = tomllib.loads((ROOT / "pyproject.toml").read_text(encoding="utf-8"))

    assert project["tool"]["pytest"]["ini_options"]["testpaths"] == [
        "tests", "agents/tests", "sdk/python/tests",
    ]


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


def test_proto_generation_pins_all_generator_versions(monkeypatch, tmp_path):
    sdk_dir = tmp_path / "sdk" / "python"
    monkeypatch.setattr(proto_relay, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(proto_relay, "SDK_DIR", sdk_dir)
    context = _RecordingContext()

    proto_relay.generate.body(context)

    go_script = proto_relay._go_generation_script()
    assert proto_relay.GO_TOOLCHAIN_IMAGE in context.commands[0]
    assert "@sha256:" in proto_relay.GO_TOOLCHAIN_IMAGE
    assert f"protobuf-compiler={proto_relay.PROTOBUF_COMPILER_PACKAGE.split('=', 1)[1]}" in go_script
    assert f"protoc-gen-go@{proto_relay.PROTOC_GEN_GO_VERSION}" in go_script
    assert "protoc-gen-go-grpc" not in go_script
    assert "--go-grpc_out" not in go_script
    assert "@latest" not in go_script

    assert len(context.commands) == 3
    for command in context.commands[1:]:
        assert f"grpcio-tools=={proto_relay.GRPCIO_TOOLS_VERSION}" in command
        assert f"protobuf=={proto_relay.PYTHON_PROTOBUF_VERSION}" in command
        assert "--grpc_python_out" not in command
