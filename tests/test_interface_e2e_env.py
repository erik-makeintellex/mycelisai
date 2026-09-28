"""Regression coverage for E2EENV: Playwright must receive the resolved
`.env` (including a linked worktree's fallback to the main checkout's copy),
notably `MYCELIS_LOCAL_ADMIN_PASSWORD`. See docs/TESTING.md's lease rules for
why e2e cleanup must never stop a foreign (e.g. Docker-published) listener.

Never reads or prints a real secret: every `.env` here is a fixture written
to a pytest tmp_path.
"""
from __future__ import annotations

import os

import pytest

from ops import interface_env
from ops import interface_runtime as runtime


FAKE_PASSWORD = "fixture-not-a-real-secret"  # noqa: S105 - test fixture value only


def _write_env_file(path, **values: str) -> None:
    path.write_text(
        "\n".join(f"{key}={value}" for key, value in values.items()) + "\n",
        encoding="utf-8",
    )


@pytest.fixture
def isolated_env(monkeypatch):
    """Clear the one variable under test so nothing leaks in from the real
    parent shell or from an earlier test's `_load_env()` mutation of the
    process-wide `os.environ` in this pytest session."""
    monkeypatch.delenv("MYCELIS_LOCAL_ADMIN_PASSWORD", raising=False)
    monkeypatch.delenv("PORT", raising=False)
    yield


def test_interface_subprocess_env_includes_password_loaded_from_dotenv(isolated_env, monkeypatch, tmp_path):
    """`interface.test` / `interface.build` run commands through
    `_interface_subprocess_env`; it must hand the child process the
    `.env`-resolved `MYCELIS_LOCAL_ADMIN_PASSWORD`, not just cache paths."""
    main_checkout_env = tmp_path / "main-checkout" / ".env"
    main_checkout_env.parent.mkdir(parents=True)
    _write_env_file(main_checkout_env, MYCELIS_LOCAL_ADMIN_PASSWORD=FAKE_PASSWORD)

    monkeypatch.setattr(
        interface_env,
        "resolve_env_file",
        lambda name, checkout_root=None: main_checkout_env if name == ".env" else None,
    )

    process_env = interface_env._interface_subprocess_env()

    assert process_env.get("MYCELIS_LOCAL_ADMIN_PASSWORD") == FAKE_PASSWORD
    assert "PORT" not in process_env


def test_empty_dotenv_value_never_blanks_a_value_set_in_the_shell(isolated_env, monkeypatch, tmp_path):
    """Owner repro (2026-09-28): the local admin is configured by
    MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256, so `.env` carries an empty
    `MYCELIS_LOCAL_ADMIN_PASSWORD=` line. `load_dotenv(override=True)` then
    replaced the password exported for the e2e run with "", and Playwright
    global-setup refused to start. An empty `.env` value must not blank a
    non-empty shell value; a non-empty `.env` value still wins."""
    main_checkout_env = tmp_path / "main-checkout" / ".env"
    main_checkout_env.parent.mkdir(parents=True)
    _write_env_file(main_checkout_env, MYCELIS_LOCAL_ADMIN_PASSWORD="", MYCELIS_E2EENV_FIXTURE_OVERRIDE="from-dotenv")
    monkeypatch.setenv("MYCELIS_LOCAL_ADMIN_PASSWORD", FAKE_PASSWORD)
    monkeypatch.setenv("MYCELIS_E2EENV_FIXTURE_OVERRIDE", "from-shell")
    monkeypatch.setattr(
        interface_env,
        "resolve_env_file",
        lambda name, checkout_root=None: main_checkout_env if name == ".env" else None,
    )

    env = interface_env._task_env(None)

    assert env.get("MYCELIS_LOCAL_ADMIN_PASSWORD") == FAKE_PASSWORD
    assert env.get("MYCELIS_E2EENV_FIXTURE_OVERRIDE") == "from-dotenv"


def test_playwright_subprocess_env_includes_password_from_linked_worktree_fallback(isolated_env, monkeypatch, tmp_path):
    """The exact E2EENV repro: a linked worktree has no local `.env`, so the
    e2e task must fall back to the main checkout's `.env` and forward the
    resolved password to the `npx playwright test` child process env — the
    process global-setup.ts reads to sign in as local admin."""
    main_checkout_env = tmp_path / "main-checkout" / ".env"
    main_checkout_env.parent.mkdir(parents=True)
    _write_env_file(main_checkout_env, MYCELIS_LOCAL_ADMIN_PASSWORD=FAKE_PASSWORD)

    # Simulate the worktree fallback: no local .env/.env.compose, only the
    # main checkout's .env resolves (ops/config.py's resolve_env_file already
    # does this walk via git-common-dir; here we stand in for that lookup so
    # the test does not depend on real git worktree plumbing).
    monkeypatch.setattr(
        interface_env,
        "resolve_env_file",
        lambda name, checkout_root=None: main_checkout_env if name == ".env" else None,
    )

    captured_popen_env: dict[str, dict[str, str]] = {}

    class FakePopen:
        def __init__(self, *args, **kwargs):
            captured_popen_env["env"] = kwargs["env"]

        def poll(self):
            return 0

    monkeypatch.setattr(runtime.subprocess, "Popen", FakePopen)
    monkeypatch.setattr(runtime, "_playwright_last_run_path", lambda: tmp_path / "does-not-exist.json")
    monkeypatch.setattr(runtime.shutil, "which", lambda _name: None)

    # This is the real production call chain for `interface.e2e`: build the
    # Playwright env exactly like the `e2e` task does, then hand it to the
    # streaming runner that actually launches `npx playwright test`.
    playwright_env = interface_env._build_playwright_env(live_backend=False, port=4321)
    runtime._run_playwright_command_streaming(["npx", "playwright", "test"], extra_env=playwright_env)

    assert captured_popen_env["env"].get("MYCELIS_LOCAL_ADMIN_PASSWORD") == FAKE_PASSWORD


def test_playwright_subprocess_env_includes_password_when_only_task_env_runs_first(isolated_env, monkeypatch, tmp_path):
    """Isolate the exact defect the packet flagged: if the streaming runner's
    own `os.environ.copy()` happens to be the first env-loading call in the
    process (e.g. `--server-mode=external` short-circuits earlier work, or a
    future refactor reorders the `e2e` task), the resolved `.env` password
    must still reach the child process. This must not depend on some earlier,
    unrelated call having already mutated the real `os.environ` first."""
    main_checkout_env = tmp_path / "main-checkout" / ".env"
    main_checkout_env.parent.mkdir(parents=True)
    _write_env_file(main_checkout_env, MYCELIS_LOCAL_ADMIN_PASSWORD=FAKE_PASSWORD)

    monkeypatch.setattr(
        interface_env,
        "resolve_env_file",
        lambda name, checkout_root=None: main_checkout_env if name == ".env" else None,
    )

    captured_popen_env: dict[str, dict[str, str]] = {}

    class FakePopen:
        def __init__(self, *args, **kwargs):
            captured_popen_env["env"] = kwargs["env"]

        def poll(self):
            return 0

    monkeypatch.setattr(runtime.subprocess, "Popen", FakePopen)
    monkeypatch.setattr(runtime, "_playwright_last_run_path", lambda: tmp_path / "does-not-exist.json")
    monkeypatch.setattr(runtime.shutil, "which", lambda _name: None)

    # No prior call to _build_playwright_env / _task_env in this test: the
    # streaming runner's own os.environ.copy() + _task_env() is the first and
    # only env-loading call.
    runtime._run_playwright_command_streaming(
        ["npx", "playwright", "test"],
        extra_env={"PLAYWRIGHT_SKIP_WEBSERVER": "1"},
    )

    assert captured_popen_env["env"].get("MYCELIS_LOCAL_ADMIN_PASSWORD") == FAKE_PASSWORD
