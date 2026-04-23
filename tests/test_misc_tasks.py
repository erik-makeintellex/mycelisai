from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

from ops import misc


class _RecvSocket:
    def __init__(self, payload: bytes):
        self.payload = payload
        self.offset = 0

    def recv(self, size: int) -> bytes:
        if self.offset >= len(self.payload):
            return b""
        chunk = self.payload[self.offset:self.offset + size]
        self.offset += len(chunk)
        return chunk


class _FakeSocket:
    def __init__(self):
        self.sent: list[bytes] = []
        self.closed = False

    def sendall(self, data: bytes):
        self.sent.append(data)

    def settimeout(self, _value: float):
        return None

    def close(self):
        self.closed = True


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


def test_read_nats_line_reads_until_crlf():
    sock = _RecvSocket(b"INFO {}\r\nPING\r\n")

    assert misc._read_nats_line(sock) == "INFO {}"
    assert misc._read_nats_line(sock) == "PING"


def test_architecture_sync_requests_all_teams_and_reports_replies(monkeypatch, capsys):
    fake_sock = _FakeSocket()
    handshake = iter(["INFO {}", "PONG", "PONG"])
    replies_sent = False

    class _FakeClock:
        def __init__(self):
            self.value = 1000.0

        def time(self) -> float:
            self.value += 0.25
            return self.value

    clock = _FakeClock()

    def fake_read_nats_line(sock) -> str:
        assert sock is fake_sock
        return next(handshake)

    def fake_drain(sock, _timeout_seconds: float) -> list[tuple[str, str]]:
        nonlocal replies_sent
        assert sock is fake_sock
        if replies_sent:
            return []
        replies_sent = True
        reply_subjects: list[str] = []
        for item in fake_sock.sent:
            if item.startswith(b"SUB "):
                subject = item.decode("utf-8").split()[1]
                if subject.endswith(".signal.status"):
                    reply_subjects.append(subject)
        return [
            (
                subject,
                (
                    '{"meta":{"source_kind":"system","source_channel":"swarm.team.test.internal.response",'
                    '"payload_kind":"status","team_id":"test"},'
                    f'"text":"ack:{idx}"}}'
                ),
            )
            for idx, subject in enumerate(reply_subjects, start=1)
        ]

    monkeypatch.setattr("socket.create_connection", lambda *args, **kwargs: fake_sock)
    monkeypatch.setattr(misc, "_read_nats_line", fake_read_nats_line)
    monkeypatch.setattr(misc, "_drain_nats_messages", fake_drain)
    monkeypatch.setattr("time.time", clock.time)
    monkeypatch.setattr("time.sleep", lambda _n: None)

    misc.architecture_sync.body(Context(), timeout=2)

    output = capsys.readouterr().out
    assert "Dispatched architecture directive to prime-architect" in output
    assert "Dispatched architecture directive to prime-development" in output
    assert "Dispatched architecture directive to agui-design-architect" in output
    assert "All teams replied inside the sync window." in output

    published = b"".join(fake_sock.sent)
    assert b"swarm.team.prime-architect.internal.command" in published
    assert b"swarm.team.prime-development.internal.command" in published
    assert b"swarm.team.agui-design-architect.internal.command" in published
    assert b"swarm.team.prime-architect.signal.status" in published
    assert b"swarm.team.prime-development.signal.status" in published
    assert b"swarm.team.agui-design-architect.signal.status" in published
    assert published.endswith(b"QUIT\r\n")
    assert fake_sock.closed is True


def test_format_sync_reply_prefers_process_result_text():
    message = '{"text":"brief body","tools_used":["publish_signal"]}'

    assert misc._format_sync_reply(message) == "brief body"


def test_format_sync_reply_prefers_wrapped_signal_text():
    message = '{"meta":{"payload_kind":"status"},"text":"wrapped brief"}'

    assert misc._format_sync_reply(message) == "wrapped brief"


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
        "$env:PYTHONPATH='.'; uv run pytest tests/test_core_tasks.py tests/test_ci_tasks.py tests/test_interface_tasks.py tests/test_interface_e2e_tasks.py tests/test_interface_command_tasks.py tests/test_k8s_tasks.py tests/test_lifecycle_tasks.py tests/test_misc_tasks.py -q"
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
    assert "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md" in output
    assert "docs/architecture-library/V8_RUNTIME_CONTRACTS.md" in output
    assert "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md" in output
    assert "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md" in output
    assert "docs/LOCAL_DEV_WORKFLOW.md" in output
    assert "docs/architecture/OPERATIONS.md" in output
    assert "ops/README.md" in output


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
