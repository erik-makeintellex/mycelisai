"""Isolated first-boot product-state assertions, parameterized over a psql runner.

Kept separate from `lifecycle_first_boot_isolated` so that module's fixture
lifecycle and Compose-exec plumbing stays under the repository's file-size
policy; these functions take a `psql` callable (a fixture-bound `-c`/`-f`
runner returning trimmed stdout) instead of a `Fixture`, so they carry no
Compose or Docker dependency of their own.
"""

from __future__ import annotations

import json
import re
import urllib.request
from typing import Callable

from .db_schema import SCHEMA_COMPATIBILITY_CHECKS

PsqlRunner = Callable[..., str]


def count_tables(psql: PsqlRunner, tables: tuple[str, ...]) -> dict[str, int]:
    names = tuple(dict.fromkeys(tables))
    if not names or any(re.fullmatch(r"[a-z][a-z0-9_]*", name) is None for name in names):
        raise SystemExit("Invalid first-boot table-count contract.")
    sql = " UNION ALL ".join(f"SELECT '{name}', COUNT(*)::bigint FROM {name}" for name in names) + " ORDER BY 1;"
    counts: dict[str, int] = {}
    for line in psql("-c", sql).splitlines():
        name, separator, value = line.partition("|")
        if separator and value.isdecimal():
            counts[name] = int(value)
    if set(counts) != set(names):
        raise SystemExit("Isolated first-boot proof did not receive every table count.")
    return counts


def invocation_tables(psql: PsqlRunner) -> tuple[str, ...]:
    sql = "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_name LIKE '%invocation%' ORDER BY table_name;"
    return tuple(name for name in psql("-c", sql).splitlines() if re.fullmatch(r"[a-z][a-z0-9_]*", name))


def assert_schema_compatible(psql: PsqlRunner) -> None:
    for label, sql in SCHEMA_COMPATIBILITY_CHECKS:
        if psql("-c", sql).strip() != "1":
            raise SystemExit(f"Isolated first-boot schema compatibility failed: {label}.")
    print("  [OK] Canonical schema compatibility")


def assert_empty_user_state(psql: PsqlRunner, tables: tuple[str, ...], label: str) -> None:
    counts = count_tables(psql, tables + invocation_tables(psql))
    dirty = {name: count for name, count in counts.items() if count}
    if dirty:
        raise SystemExit(f"Isolated first-boot {label} created user state: {dirty}.")
    print(f"  [OK] Empty user state after {label}")


def assert_nats_empty(monitor_port: int) -> None:
    url = f"http://127.0.0.1:{monitor_port}/jsz?streams=1"
    with urllib.request.urlopen(url, timeout=5) as response:
        payload = json.load(response)
    if response.status != 200 or int(payload.get("streams", 0)) or int(payload.get("messages", 0)):
        raise SystemExit("Isolated first-boot NATS retained streams or messages.")
    print("  [OK] Empty fixture JetStream")
