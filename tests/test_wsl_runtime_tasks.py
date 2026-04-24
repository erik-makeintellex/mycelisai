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


def test_fetch_wsl_remote_repairs_github_https_auth_with_windows_gcm(monkeypatch):
    fetch_results = [
        wsl_runtime.CommandResult(
            command=["git", "fetch"],
            returncode=128,
            stdout="",
            stderr="fatal: could not read Username for 'https://github.com': terminal prompts disabled",
        ),
        wsl_runtime.CommandResult(command=["git", "fetch"], returncode=0, stdout="", stderr=""),
    ]
    config_calls: list[tuple[str, ...]] = []
    fetch_calls: list[tuple[str, ...]] = []

    def fake_run_wsl_git(*args, **_kwargs):
        if args == ("remote", "get-url", "origin"):
            return wsl_runtime.CommandResult(
                command=["git", *args],
                returncode=0,
                stdout="https://github.com/example/private-repo.git\n",
                stderr="",
            )
        if args[:3] == ("config", "--local", "credential.helper"):
            config_calls.append(args)
            return wsl_runtime.CommandResult(command=["git", *args], returncode=0, stdout="", stderr="")
        raise AssertionError(f"unexpected git command: {args}")

    def fake_run_wsl_git_noninteractive(*args, **_kwargs):
        fetch_calls.append(args)
        return fetch_results.pop(0)

    monkeypatch.setattr(wsl_runtime, "_run_wsl_git", fake_run_wsl_git)
    monkeypatch.setattr(wsl_runtime, "_run_wsl_git_noninteractive", fake_run_wsl_git_noninteractive)
    monkeypatch.setattr(
        wsl_runtime,
        "_find_windows_gcm_helper",
        lambda **_kwargs: "/mnt/c/Program Files/Git/mingw64/bin/git-credential-manager.exe",
    )

    wsl_runtime._fetch_wsl_remote_with_auth_repair("origin", distro="mother-brain", checkout="/repo")

    assert fetch_calls == [
        ("fetch", "--prune", "origin"),
        ("fetch", "--prune", "origin"),
    ]
    assert config_calls == [
        (
            "config",
            "--local",
            "credential.helper",
            "/mnt/c/Program\\ Files/Git/mingw64/bin/git-credential-manager.exe",
        )
    ]


def test_fetch_wsl_remote_reports_actionable_ssh_auth_guidance(monkeypatch):
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_git",
        lambda *args, **_kwargs: wsl_runtime.CommandResult(
            command=["git", *args],
            returncode=0,
            stdout="git@github.com:example/private-repo.git\n",
            stderr="",
        )
        if args == ("remote", "get-url", "origin")
        else (_ for _ in ()).throw(AssertionError(f"unexpected git command: {args}")),
    )
    monkeypatch.setattr(
        wsl_runtime,
        "_run_wsl_git_noninteractive",
        lambda *args, **_kwargs: wsl_runtime.CommandResult(
            command=["git", *args],
            returncode=128,
            stdout="",
            stderr="git@github.com: Permission denied (publickey).\nfatal: Could not read from remote repository.",
        ),
    )

    try:
        wsl_runtime._fetch_wsl_remote_with_auth_repair("origin", distro="mother-brain", checkout="/repo")
    except SystemExit as exc:
        message = str(exc)
        assert "WSL git auth is not ready" in message
        assert "ssh -T git@github.com" in message
        assert "do not copy source trees" in message
    else:
        raise AssertionError("expected actionable WSL git auth failure")


def test_clean_wsl_proof_checkout_preserves_runtime_cache_roots(monkeypatch):
    calls: list[tuple[str, ...]] = []
    monkeypatch.setattr(wsl_runtime, "_run_wsl_git", lambda *args, **_kwargs: calls.append(args))

    wsl_runtime._clean_wsl_proof_checkout(distro="mother-brain", checkout="/repo")

    assert calls == [
        (
            "clean",
            "-fdx",
            "-e",
            "workspace/tool-cache/",
            "-e",
            "workspace/logs/",
            "-e",
            "workspace/docker-compose/",
        )
    ]


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
