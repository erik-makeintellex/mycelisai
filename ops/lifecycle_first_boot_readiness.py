"""Deterministic readiness and transient-retry helpers for isolated first-boot Postgres exec.

The official Postgres image runs a temporary, socket-only server to apply its
own init scripts, then stops it and execs the real, long-running server.
`pg_isready` can observe the temporary server and report success moments
before it stops for that handoff, so a caller that `docker compose exec`s a
follow-on command right after a single "ready" probe can race the brief
window where nothing is listening. This module makes readiness require a
real `SELECT 1` as the target role/database and retries only exec failures
whose stderr matches a known transient shutdown/restart signature, so a real
failure (bad SQL, wrong credentials, missing image) still fails fast with its
real stderr surfaced.
"""

from __future__ import annotations

import time
from dataclasses import dataclass
from typing import Callable, Protocol


TRANSIENT_MARKERS = (
    "connection refused",
    "could not connect to server",
    "the database system is starting up",
    "the database system is shutting down",
    "terminating connection due to administrator command",
    "server closed the connection unexpectedly",
    "no route to host",
    "eof detected",
    "unexpected eof",
    "container is restarting",
)

SECRET_ENV_KEYS = (
    "POSTGRES_PASSWORD",
    "DB_PASSWORD",
    "MYCELIS_API_KEY",
    "MYCELIS_WEB_SESSION_SECRET",
    "MYCELIS_WEB_IDENTITY_FORWARD_SECRET",
)


class ExecResult(Protocol):
    returncode: int
    stdout: str
    stderr: str


def secret_values(environment: dict[str, str]) -> dict[str, str]:
    """Pick only the fixture's generated secrets out of its environment map."""
    return {key: environment[key] for key in SECRET_ENV_KEYS if environment.get(key)}


def redact(text: str, secrets: dict[str, str]) -> str:
    rendered = text or ""
    for value in secrets.values():
        if value:
            rendered = rendered.replace(value, "***")
    return rendered


def is_transient_error(text: str) -> bool:
    lowered = (text or "").lower()
    return any(marker in lowered for marker in TRANSIENT_MARKERS)


@dataclass(frozen=True)
class RetryPolicy:
    attempts: int = 20
    base_delay: float = 0.5
    max_delay: float = 3.0


def _describe(result: ExecResult) -> str:
    stderr = (getattr(result, "stderr", "") or "").strip()
    stdout = (getattr(result, "stdout", "") or "").strip()
    return stderr or stdout or "(no output captured)"


def retry_transient(
    description: str,
    run: Callable[[], ExecResult],
    *,
    policy: RetryPolicy = RetryPolicy(),
    succeeded: Callable[[ExecResult], bool] = lambda result: result.returncode == 0,
    secrets: dict[str, str] | None = None,
    sleep: Callable[[float], None] = time.sleep,
) -> ExecResult:
    """Call `run()` until `succeeded()`, retrying only transient failures with backoff.

    Raises SystemExit with the real, secret-redacted output once attempts are
    exhausted or a non-transient failure is observed.
    """
    if policy.attempts < 1:
        raise ValueError("policy.attempts must be at least 1")
    delay = policy.base_delay
    for attempt in range(1, policy.attempts + 1):
        result = run()
        if succeeded(result):
            return result
        detail = redact(_describe(result), secrets or {})
        if attempt == policy.attempts or not is_transient_error(detail):
            raise SystemExit(f"{description} failed after {attempt} attempt(s): {detail}")
        sleep(delay)
        delay = min(delay * 2, policy.max_delay)
    raise SystemExit(f"{description}: no attempts executed.")


def wait_for_isolated_postgres_ready(
    *,
    pg_isready: Callable[[], ExecResult],
    select_one: Callable[[], ExecResult],
    policy: RetryPolicy = RetryPolicy(),
    secrets: dict[str, str] | None = None,
    sleep: Callable[[float], None] = time.sleep,
) -> None:
    """Wait until Postgres accepts a real `SELECT 1` as the target role/database.

    A readiness wait's own "not ready yet" is expected during startup, so any
    failure of either probe retries (bounded, with backoff) until the final
    attempt, which raises with the real output of whichever probe failed.
    """

    def probe() -> ExecResult:
        ready = pg_isready()
        return ready if ready.returncode != 0 else select_one()

    def succeeded(result: ExecResult) -> bool:
        return result.returncode == 0 and (getattr(result, "stdout", "") or "").strip() == "1"

    delay = policy.base_delay
    for attempt in range(1, policy.attempts + 1):
        result = probe()
        if succeeded(result):
            return
        if attempt == policy.attempts:
            detail = redact(_describe(result), secrets or {})
            raise SystemExit(
                f"Isolated first-boot PostgreSQL never became ready as mycelis/cortex "
                f"after {attempt} attempt(s): {detail}"
            )
        sleep(delay)
        delay = min(delay * 2, policy.max_delay)
