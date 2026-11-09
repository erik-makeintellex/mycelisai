from __future__ import annotations

from invoke import Context

from ops import wsl_runtime


def test_validate_runs_expected_wsl_commands_and_windows_probe(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_require_windows_dev_host", lambda: None)
    monkeypatch.setattr(wsl_runtime, "_configured_distro", lambda distro="": "mother-brain")
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")
    monkeypatch.setattr(
        wsl_runtime,
        "_collect_wsl_state",
        lambda **_kwargs: {"branch": "main", "commit": "abc123", "dirty": 0},
    )
    monkeypatch.setattr(wsl_runtime, "_print_repo_state", lambda *args, **kwargs: None)
    monkeypatch.setattr(wsl_runtime, "_ensure_wsl_compose_env", lambda **_kwargs: None)
    monkeypatch.setattr(wsl_runtime, "_ensure_wsl_output_block_path", lambda **_kwargs: None)

    shell_calls: list[str] = []
    probe_calls: list[str] = []
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_shell",
        lambda command, **_kwargs: shell_calls.append(command),
    )
    monkeypatch.setattr(wsl_runtime, "_probe_windows_gui", lambda url: probe_calls.append(url))

    wsl_runtime.validate.body(Context(), lane="runtime")

    assert shell_calls == [
        "uv run inv install",
        "uv run inv ci.release-preflight --lane=runtime --no-e2e",
        "uv run inv auth.posture --compose",
        "uv run inv compose.up --build --wait-timeout=240",
        "uv run inv compose.health",
        "uv run inv compose.storage-health",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts --live-backend --workers=1 --server-mode=start",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/team-creation.spec.ts --live-backend --workers=1 --server-mode=start",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts --live-backend --workers=1 --server-mode=start",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/workspace-live-backend.spec.ts --live-backend --workers=1 --server-mode=start",
    ]
    assert probe_calls == ["http://localhost:3000"]


def test_validate_release_lane_uses_runtime_preflight_before_compose_owned_gates(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_require_windows_dev_host", lambda: None)
    monkeypatch.setattr(wsl_runtime, "_configured_distro", lambda distro="": "mother-brain")
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")
    monkeypatch.setattr(
        wsl_runtime,
        "_collect_wsl_state",
        lambda **_kwargs: {"branch": "main", "commit": "abc123", "dirty": 0},
    )
    monkeypatch.setattr(wsl_runtime, "_print_repo_state", lambda *args, **kwargs: None)
    monkeypatch.setattr(wsl_runtime, "_ensure_wsl_compose_env", lambda **_kwargs: None)
    monkeypatch.setattr(wsl_runtime, "_ensure_wsl_output_block_path", lambda **_kwargs: None)
    monkeypatch.setattr(wsl_runtime, "_probe_windows_gui", lambda _url: None)

    shell_calls: list[str] = []
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_shell",
        lambda command, **_kwargs: shell_calls.append(command),
    )

    wsl_runtime.validate.body(Context(), lane="release")

    assert "uv run inv ci.release-preflight --lane=runtime --no-e2e" in shell_calls
    assert "uv run inv ci.release-preflight --lane=release --no-e2e" not in shell_calls
    assert "uv run inv compose.health" in shell_calls
    assert "uv run inv compose.storage-health" in shell_calls


def test_ensure_wsl_compose_env_bootstraps_from_example(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")
    run_calls: list[list[str]] = []
    shell_calls: list[str] = []

    def fake_run(command, **_kwargs):
        run_calls.append(command)
        command_text = " ".join(command)
        if "test -f .env" in command_text and ".example" not in command_text:
            return wsl_runtime.CommandResult(command=command, returncode=1, stdout="", stderr="")
        if "test -f .env.example" in command_text or "test -f .env.compose.example" in command_text:
            return wsl_runtime.CommandResult(command=command, returncode=0, stdout="", stderr="")
        raise AssertionError(f"unexpected command: {command}")

    monkeypatch.setattr(wsl_runtime, "_run_command", fake_run)
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_shell",
        lambda command, **_kwargs: shell_calls.append(command),
    )

    wsl_runtime._ensure_wsl_compose_env(distro="mother-brain", checkout="/home/erik/Projects/mycelisai/scratch")

    assert shell_calls == ["cp .env.example .env", "cp .env.compose.example .env.compose"]


def test_ensure_wsl_output_block_path_creates_configured_directory(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")

    shell_calls: list[str] = []
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_shell",
        lambda command, **_kwargs: shell_calls.append(command),
    )

    wsl_runtime._ensure_wsl_output_block_path(distro="mother-brain", checkout="/home/erik/Projects/mycelisai/scratch")

    assert len(shell_calls) == 1
    assert "MYCELIS_OUTPUT_HOST_PATH" in shell_calls[0]
    assert "workspace/docker-compose/data" in shell_calls[0]


def test_cycle_runs_refresh_then_validate(monkeypatch):
    calls: list[tuple[str, dict[str, str]]] = []

    monkeypatch.setattr(
        wsl_runtime.refresh,
        "body",
        lambda _ctx, **kwargs: calls.append(("refresh", kwargs)),
    )
    monkeypatch.setattr(
        wsl_runtime.validate,
        "body",
        lambda _ctx, **kwargs: calls.append(("validate", kwargs)),
    )

    wsl_runtime.cycle.body(Context(), branch="main", lane="runtime")

    assert calls == [
        (
            "refresh",
            {
                "branch": "main",
                "ref": "",
                "distro": "",
                "checkout": "",
                "remote": "",
            },
        ),
        (
            "validate",
            {
                "lane": "runtime",
                "distro": "",
                "checkout": "",
                "gui_url": "",
                "compose_wait_timeout": "",
            },
        ),
    ]
