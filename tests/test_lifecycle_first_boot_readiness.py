from __future__ import annotations

import subprocess

import pytest

from ops import lifecycle_first_boot_readiness as readiness


def result(returncode: int, stdout: str = "", stderr: str = "") -> subprocess.CompletedProcess[str]:
    return subprocess.CompletedProcess([], returncode, stdout, stderr)


def test_is_transient_error_matches_known_restart_signatures():
    assert readiness.is_transient_error("FATAL: the database system is starting up")
    assert readiness.is_transient_error("psql: error: connection to server ... Connection refused")
    assert not readiness.is_transient_error("ERROR:  syntax error at or near \"SELCT\"")
    assert not readiness.is_transient_error("")


def test_redact_replaces_only_known_secret_values():
    secrets = {"DB_PASSWORD": "s3cr3t-value", "MYCELIS_API_KEY": "key-abc"}
    text = "connection failed for user with password s3cr3t-value and key key-abc"

    rendered = readiness.redact(text, secrets)

    assert "s3cr3t-value" not in rendered
    assert "key-abc" not in rendered
    assert "connection failed for user with password ***" in rendered


def test_secret_values_only_pulls_recognised_keys():
    environment = {"DB_PASSWORD": "pw", "DB_HOST": "postgres", "MYCELIS_API_KEY": "", "OTHER": "x"}

    picked = readiness.secret_values(environment)

    assert picked == {"DB_PASSWORD": "pw"}


def test_retry_transient_retries_transient_failures_then_succeeds():
    attempts = iter([
        result(1, stderr="FATAL: the database system is starting up"),
        result(1, stderr="psql: error: connection refused"),
        result(0, stdout="ok"),
    ])
    sleeps: list[float] = []

    outcome = readiness.retry_transient(
        "probe",
        lambda: next(attempts),
        policy=readiness.RetryPolicy(attempts=5, base_delay=0.1, max_delay=1.0),
        sleep=sleeps.append,
    )

    assert outcome.stdout == "ok"
    assert sleeps == [0.1, 0.2]  # exponential backoff between the two transient retries


def test_retry_transient_fails_fast_on_non_transient_error_without_exhausting_attempts():
    calls = []

    def run():
        calls.append(1)
        return result(1, stderr='ERROR:  syntax error at or near "SELCT"')

    with pytest.raises(SystemExit, match="syntax error"):
        readiness.retry_transient(
            "schema install",
            run,
            policy=readiness.RetryPolicy(attempts=10, base_delay=0.1),
            sleep=lambda _delay: pytest.fail("should not sleep on a non-transient failure"),
        )

    assert len(calls) == 1


def test_retry_transient_gives_up_after_bounded_attempts_and_surfaces_real_stderr():
    def run():
        return result(1, stderr="server closed the connection unexpectedly")

    with pytest.raises(SystemExit, match="failed after 3 attempt\\(s\\): server closed the connection unexpectedly"):
        readiness.retry_transient(
            "probe",
            run,
            policy=readiness.RetryPolicy(attempts=3, base_delay=0.01, max_delay=0.01),
            sleep=lambda _delay: None,
        )


def test_retry_transient_redacts_secrets_in_raised_message():
    def run():
        return result(1, stderr="auth failed for password s3cr3t")

    with pytest.raises(SystemExit) as excinfo:
        readiness.retry_transient(
            "probe",
            run,
            policy=readiness.RetryPolicy(attempts=1),
            secrets={"DB_PASSWORD": "s3cr3t"},
            sleep=lambda _delay: None,
        )

    assert "s3cr3t" not in str(excinfo.value)
    assert "***" in str(excinfo.value)


def test_wait_for_isolated_postgres_ready_requires_select_one_not_just_pg_isready():
    # Regression for the real race: pg_isready can observe Postgres's own
    # temporary init-phase server and report ready moments before it restarts.
    pg_isready_calls = []
    select_one_calls = []

    def pg_isready():
        pg_isready_calls.append(1)
        return result(0)  # the temp server always answers pg_isready

    select_one_results = iter([
        result(1, stderr="FATAL: the database system is shutting down"),
        result(0, stdout="1"),
    ])

    def select_one():
        select_one_calls.append(1)
        return next(select_one_results)

    readiness.wait_for_isolated_postgres_ready(
        pg_isready=pg_isready,
        select_one=select_one,
        policy=readiness.RetryPolicy(attempts=5, base_delay=0.01, max_delay=0.01),
        sleep=lambda _delay: None,
    )

    assert len(pg_isready_calls) == 2
    assert len(select_one_calls) == 2


def test_wait_for_isolated_postgres_ready_gives_up_and_surfaces_real_output():
    def pg_isready():
        return result(1, stdout="/var/run/postgresql:5432 - no response")

    with pytest.raises(SystemExit, match="never became ready.*no response"):
        readiness.wait_for_isolated_postgres_ready(
            pg_isready=pg_isready,
            select_one=lambda: pytest.fail("select_one should not run while pg_isready keeps failing"),
            policy=readiness.RetryPolicy(attempts=2, base_delay=0.01, max_delay=0.01),
            sleep=lambda _delay: None,
        )
