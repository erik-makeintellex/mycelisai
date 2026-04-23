from __future__ import annotations

from invoke import Context

from ops import wsl_runtime


def test_configured_checkout_rejects_mnt_paths():
    try:
        wsl_runtime._configured_checkout("/mnt/d/mycelis")
    except SystemExit as exc:
        assert "must live on the Linux filesystem" in str(exc)
    else:
        raise AssertionError("expected /mnt path rejection")


def test_refresh_refuses_when_windows_branch_is_not_pushed(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_require_windows_dev_host", lambda: None)
    monkeypatch.setattr(wsl_runtime, "_configured_distro", lambda distro="": "mother-brain")
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")
    monkeypatch.setattr(wsl_runtime, "_configured_remote", lambda remote="": "origin")
    monkeypatch.setattr(wsl_runtime, "_wsl_repo_guard", lambda **_kwargs: None)
    monkeypatch.setattr(
        wsl_runtime,
        "_collect_local_state",
        lambda: {
            "branch": "feature/test",
            "commit": "abc123",
            "upstream": "origin/feature/test",
            "ahead": 1,
            "behind": 0,
            "dirty": 0,
        },
    )
    monkeypatch.setattr(
        wsl_runtime,
        "_assert_current_branch_published",
        lambda remote, branch: (_ for _ in ()).throw(
            SystemExit(f"Windows dev branch '{branch}' is ahead of {remote}/{branch} by 1 commit(s). Push first, then refresh WSL.")
        ),
    )
    monkeypatch.setattr(wsl_runtime, "_current_local_branch", lambda require_attached=True: "feature/test")

    try:
        wsl_runtime.refresh.body(Context())
    except SystemExit as exc:
        assert "Push first" in str(exc)
    else:
        raise AssertionError("expected refresh refusal when branch is unpublished")


def test_refresh_refuses_when_windows_checkout_is_dirty(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_require_windows_dev_host", lambda: None)
    monkeypatch.setattr(wsl_runtime, "_configured_distro", lambda distro="": "mother-brain")
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")
    monkeypatch.setattr(wsl_runtime, "_configured_remote", lambda remote="": "origin")
    monkeypatch.setattr(wsl_runtime, "_wsl_repo_guard", lambda **_kwargs: None)
    monkeypatch.setattr(
        wsl_runtime,
        "_collect_local_state",
        lambda: {
            "branch": "main",
            "commit": "abc123",
            "upstream": "origin/main",
            "ahead": 0,
            "behind": 0,
            "dirty": 2,
        },
    )

    try:
        wsl_runtime.refresh.body(Context())
    except SystemExit as exc:
        assert "dirty path(s)" in str(exc)
    else:
        raise AssertionError("expected refresh refusal for dirty Windows checkout")


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


def test_ensure_wsl_compose_env_bootstraps_from_example(monkeypatch):
    monkeypatch.setattr(wsl_runtime, "_configured_checkout", lambda checkout="": "/home/erik/Projects/mycelisai/scratch")

    run_calls: list[list[str]] = []
    shell_calls: list[str] = []

    def fake_run(command, **_kwargs):
        run_calls.append(command)
        if "test -f .env.compose" in command:
            return wsl_runtime.CommandResult(command=command, returncode=1, stdout="", stderr="")
        if "test -f .env.compose.example" in command:
            return wsl_runtime.CommandResult(command=command, returncode=0, stdout="", stderr="")
        raise AssertionError(f"unexpected command: {command}")

    monkeypatch.setattr(wsl_runtime, "_run_command", fake_run)
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_shell",
        lambda command, **_kwargs: shell_calls.append(command),
    )

    wsl_runtime._ensure_wsl_compose_env(distro="mother-brain", checkout="/home/erik/Projects/mycelisai/scratch")

    assert shell_calls == ["cp .env.compose.example .env.compose"]


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
