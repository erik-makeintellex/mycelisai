from __future__ import annotations

import re
from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR

EVENTS_FILE = ROOT_DIR / "core" / "pkg" / "protocol" / "events.go"
LOGGING_DOC = ROOT_DIR / "docs" / "logging.md"
REQUIRED_LOGGING_DOC_TERMS = {
    "OperationalLogContext",
    "memory.stream",
    "central_review",
    "review_channels",
    "schema_version",
}

DEFAULT_SCHEMA_PATHS = "core/internal"
DEFAULT_TOPIC_PATHS = "core/internal/swarm,core/internal/server,interface/store"

EVENT_CONST_RE = re.compile(r'Event[A-Za-z0-9_]+\s+EventType\s*=\s*"([^"]+)"')
QUOTED_RE = re.compile(r'"([^"]+)"')
TOPIC_LITERAL_RE = re.compile(r'"(swarm\.[^"]+)"')

ALLOWED_NON_MISSION_EVENTS = {"agent.heartbeat"}
CODE_EXTENSIONS = {".go", ".py", ".ts", ".tsx"}


def _expand_paths(paths: str) -> list[Path]:
    items: list[Path] = []
    for raw in paths.split(","):
        value = raw.strip()
        if not value:
            continue
        p = ROOT_DIR / value
        if p.exists():
            items.append(p)
    return items


def _iter_code_files(roots: list[Path]) -> list[Path]:
    files: list[Path] = []
    for root in roots:
        if root.is_file():
            if root.suffix in CODE_EXTENSIONS:
                files.append(root)
            continue

        for path in root.rglob("*"):
            if not path.is_file():
                continue
            if path.suffix not in CODE_EXTENSIONS:
                continue
            if path.name.endswith("_test.go"):
                continue
            if path.name.endswith(".test.ts") or path.name.endswith(".test.tsx"):
                continue
            if "__tests__" in path.parts:
                continue
            files.append(path)
    return files


def _read_event_types(path: Path = EVENTS_FILE) -> set[str]:
    if not path.exists():
        return set()
    text = path.read_text(encoding="utf-8")
    return set(EVENT_CONST_RE.findall(text))


def _check_doc_event_coverage(event_types: set[str], doc_path: Path = LOGGING_DOC) -> list[str]:
    if not doc_path.exists():
        return sorted(event_types)
    body = doc_path.read_text(encoding="utf-8")
    missing = [event for event in sorted(event_types) if event not in body]
    return missing


def _check_doc_contract_terms(doc_path: Path = LOGGING_DOC) -> list[str]:
    if not doc_path.exists():
        return sorted(REQUIRED_LOGGING_DOC_TERMS)
    body = doc_path.read_text(encoding="utf-8")
    return sorted(term for term in REQUIRED_LOGGING_DOC_TERMS if term not in body)


def _collect_schema_violations(files: list[Path], allowed_events: set[str]) -> list[str]:
    violations: list[str] = []
    for path in files:
        try:
            rel = path.relative_to(ROOT_DIR).as_posix()
        except ValueError:
            rel = path.as_posix()
        for lineno, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
            line = raw.strip()
            if not line or line.startswith("//"):
                continue

            relevant = ("EventType:" in raw) or (".Emit(" in raw)
            if not relevant:
                continue

            for token in QUOTED_RE.findall(raw):
                if "." not in token:
                    continue
                if token.startswith("swarm."):
                    continue
                if token in ALLOWED_NON_MISSION_EVENTS:
                    continue
                if token in allowed_events:
                    continue
                # Skip struct tags and JSON keys.
                if token in {"json", "db", "yaml", "event_type"}:
                    continue
                if token.endswith(",omitempty"):
                    continue
                violations.append(f"{rel}:{lineno} -> {token}")
    return violations


def _collect_topic_literal_violations(files: list[Path], allowed_files: set[str]) -> list[str]:
    violations: list[str] = []
    for path in files:
        try:
            rel = path.relative_to(ROOT_DIR).as_posix()
        except ValueError:
            rel = path.as_posix()
        if rel in allowed_files:
            continue

        for lineno, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
            line = raw.strip()
            if not line or line.startswith("//"):
                continue
            matches = TOPIC_LITERAL_RE.findall(raw)
            for topic in matches:
                violations.append(f"{rel}:{lineno} -> {topic}")
    return violations


@task(help={"paths": "Comma-separated roots to scan for schema checks."})
def check_schema(_c, paths=DEFAULT_SCHEMA_PATHS):
    """
    Validate logging schema usage:
    - event literals in runtime code must map to declared EventType constants
    - docs/logging.md must mention each EventType constant
    """
    event_types = _read_event_types()
    if not event_types:
        raise SystemExit(f"LOGGING CHECK FAILED: missing event definitions at {EVENTS_FILE}")

    roots = _expand_paths(paths)
    files = _iter_code_files(roots)
    code_violations = _collect_schema_violations(files, event_types)
    doc_missing = _check_doc_event_coverage(event_types)
    contract_missing = _check_doc_contract_terms()

    if code_violations:
        print("Unknown event literals:")
        for item in code_violations[:80]:
            print(f"  {item}")
        if len(code_violations) > 80:
            print(f"  ... and {len(code_violations) - 80} more")

    if doc_missing:
        print("Event types missing from docs/logging.md:")
        for item in doc_missing:
            print(f"  {item}")

    if contract_missing:
        print("Logging contract terms missing from docs/logging.md:")
        for item in contract_missing:
            print(f"  {item}")

    if code_violations or doc_missing or contract_missing:
        raise SystemExit("LOGGING CHECK FAILED: schema validation errors found.")

    print(f"LOGGING CHECK PASSED: validated {len(files)} files and {len(event_types)} event types.")


@task(
    help={
        "paths": "Comma-separated roots to scan for hardcoded swarm topics.",
        "allow_files": "Comma-separated file paths allowed to contain topic literals.",
    }
)
def check_topics(_c, paths=DEFAULT_TOPIC_PATHS, allow_files="core/pkg/protocol/topics.go"):
    """
    Fail when production code hardcodes swarm.* subjects outside allowed files.
    """
    roots = _expand_paths(paths)
    files = _iter_code_files(roots)
    allowed = {p.strip().replace("\\", "/") for p in allow_files.split(",") if p.strip()}
    violations = _collect_topic_literal_violations(files, allowed)

    if violations:
        print("Hardcoded swarm topics found:")
        for item in violations[:120]:
            print(f"  {item}")
        if len(violations) > 120:
            print(f"  ... and {len(violations) - 120} more")
        raise SystemExit("LOGGING CHECK FAILED: hardcoded topic literals detected.")

    print(f"LOGGING TOPIC CHECK PASSED: scanned {len(files)} files.")


ns = Collection("logging")
ns.add_task(check_schema, name="check-schema")
ns.add_task(check_topics, name="check-topics")
