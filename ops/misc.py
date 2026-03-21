from invoke import task, Collection
from .config import ROOT_DIR

WORKTREE_REVIEW_TARGETS = (
    "README.md",
    "V8_DEV_STATE.md",
    "docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md",
    "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
    "docs/TESTING.md",
)

WORKTREE_BASELINE_INSTALLS = (
    "uv run inv install",
)

WORKTREE_BASELINE_COMMANDS = (
    "uv run inv ci.entrypoint-check",
    "uv run inv ci.baseline",
)

WORKTREE_AREA_RULES = (
    {
        "name": "Core runtime",
        "prefixes": ("core/", "proto/"),
        "installs": ("cd core && go mod download",),
        "commands": ("uv run inv core.test",),
    },
    {
        "name": "Interface",
        "prefixes": ("interface/",),
        "installs": ("uv run inv interface.install",),
        "commands": (
            "uv run inv interface.test",
            "cd interface && npx tsc --noEmit",
            "uv run inv interface.build",
        ),
    },
    {
        "name": "Cognitive",
        "prefixes": ("cognitive/",),
        "installs": ("uv run inv cognitive.install",),
        "commands": (),
    },
    {
        "name": "Python automation",
        "prefixes": ("agents/", "cli/", "ops/", "scripts/", "sdk/python/", "tests/"),
        "exact_paths": ("pyproject.toml", "tasks.py"),
        "installs": ("uv sync --all-packages --dev",),
        "commands": (
            "$env:PYTHONPATH='.'; uv run pytest tests/test_ci_tasks.py tests/test_misc_tasks.py -q",
        ),
    },
    {
        "name": "Docs and state",
        "prefixes": ("docs/",),
        "exact_paths": ("README.md", "V7_DEV_STATE.md", "V8_DEV_STATE.md", "mycelis-architecture-v7.md"),
        "installs": (),
        "commands": (
            "$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q",
        ),
    },
    {
        "name": "Infra and deploy",
        "prefixes": ("charts/", "deploy/", "k8s/"),
        "installs": (),
        "commands": (),
    },
)

# -- CLEAN --
@task
def legacy(c):
    """Remove legacy build files."""
    legacy_files = [
        ROOT_DIR / "Makefile", 
        ROOT_DIR / "Makefile.legacy",
        ROOT_DIR / "docker-compose.yml"
    ]
    for p in legacy_files:
        if p.exists():
            p.unlink()
            print(f"Removed {p}")

ns_clean = Collection("clean")
ns_clean.add_task(legacy)

# -- TEAM --
@task
def sensors(c):
    """Run Sensor Manager."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/sensors/manager.py", env=env)

@task
def output(c):
    """Run Output Manager."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run python agents/output/manager.py", env=env)

@task
def test(c):
    """Run Team Agent Unit Tests."""
    env = {"PYTHONPATH": "sdk/python/src"}
    c.run("uv run pytest agents/tests", env=env)


def _unique_strings(items):
    ordered = []
    seen = set()
    for item in items:
        if item in seen:
            continue
        seen.add(item)
        ordered.append(item)
    return ordered


def _normalize_git_path(path: str) -> str:
    normalized = path.strip()
    if " -> " in normalized:
        normalized = normalized.split(" -> ", 1)[1]
    return normalized.replace("\\", "/")


def _parse_porcelain_entries(status_output: str):
    entries = []
    for raw_line in status_output.splitlines():
        if not raw_line.strip():
            continue
        if len(raw_line) < 4:
            continue
        entries.append(
            {
                "status": raw_line[:2].strip() or "??",
                "path": _normalize_git_path(raw_line[3:]),
            }
        )
    return entries


def _match_worktree_rule(path: str):
    for rule in WORKTREE_AREA_RULES:
        if path in rule.get("exact_paths", ()):
            return rule
        if any(path.startswith(prefix) for prefix in rule.get("prefixes", ())):
            return rule
    return {
        "name": "Unclassified",
        "installs": (),
        "commands": (),
    }


def _build_worktree_triage(status_output: str):
    entries = _parse_porcelain_entries(status_output)
    area_counts = {}
    priority_installs = []
    recommended_commands = list(WORKTREE_BASELINE_COMMANDS)

    for entry in entries:
        rule = _match_worktree_rule(entry["path"])
        entry["area"] = rule["name"]
        area_counts[rule["name"]] = area_counts.get(rule["name"], 0) + 1
        priority_installs.extend(rule.get("installs", ()))
        recommended_commands.extend(rule.get("commands", ()))

    areas = [
        {"name": name, "count": count}
        for name, count in sorted(area_counts.items())
    ]

    return {
        "entries": entries,
        "areas": areas,
        "priority_installs": _unique_strings(priority_installs),
        "recommended_commands": _unique_strings(recommended_commands),
    }


@task(name="worktree-triage")
def worktree_triage(c):
    """Summarize dirty-worktree scope, install checks, and evidence commands.

    This is a local maintenance helper under ops/. It must not register,
    persist, or imply runtime teams inside core/config/teams or runtime
    registries.
    """
    print("=== WORKTREE TRIAGE ===")

    status = c.run("git status --porcelain", hide=True, warn=True)
    if status.exited != 0:
        raise SystemExit("WORKTREE TRIAGE FAILED: unable to read git status.")

    triage = _build_worktree_triage(status.stdout or "")
    entries = triage["entries"]

    if entries:
        print(f"Working tree: dirty ({len(entries)} changed path(s))")
    else:
        print("Working tree: clean")

    print("\nExpected review targets:")
    for target in WORKTREE_REVIEW_TARGETS:
        print(f"  - {target}")

    print("\nDependency reset:")
    for command in WORKTREE_BASELINE_INSTALLS:
        print(f"  - {command}")

    if triage["priority_installs"]:
        print("\nPriority install checks:")
        for command in triage["priority_installs"]:
            print(f"  - {command}")
    else:
        print("\nPriority install checks:")
        print("  - none triggered by current paths")

    print("\nRecommended evidence commands:")
    for command in triage["recommended_commands"]:
        print(f"  - {command}")

    if entries:
        print("\nArea breakdown:")
        for area in triage["areas"]:
            print(f"  - {area['name']}: {area['count']} path(s)")

        print("\nChanged paths:")
        preview = entries[:20]
        for entry in preview:
            print(f"  - [{entry['status']}] {entry['path']} ({entry['area']})")
        if len(entries) > len(preview):
            print(f"  - ... and {len(entries) - len(preview)} more")
    else:
        print("\nChanged paths:")
        print("  - none")


def _read_nats_line(sock) -> str:
    data = bytearray()
    while True:
        chunk = sock.recv(1)
        if not chunk:
            raise ConnectionError("NATS connection closed")
        data.extend(chunk)
        if data.endswith(b"\r\n"):
            return data[:-2].decode("utf-8", errors="replace")


def _drain_nats_messages(sock, timeout_seconds: float) -> list[tuple[str, str]]:
    import socket
    import time

    messages: list[tuple[str, str]] = []
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            sock.settimeout(max(0.1, deadline - time.time()))
            line = _read_nats_line(sock)
        except socket.timeout:
            break
        if not line:
            continue
        if line == "PING":
            sock.sendall(b"PONG\r\n")
            continue
        if not line.startswith("MSG "):
            continue

        parts = line.split()
        subject = parts[1]
        payload_len = int(parts[-1])
        payload = bytearray()
        while len(payload) < payload_len:
            chunk = sock.recv(payload_len - len(payload))
            if not chunk:
                raise ConnectionError("NATS connection closed during payload read")
            payload.extend(chunk)
        trailer = sock.recv(2)
        if trailer != b"\r\n":
            raise ConnectionError("Invalid NATS message trailer")
        messages.append((subject, payload.decode("utf-8", errors="replace")))

    return messages


def _format_sync_reply(message: str) -> str:
    import json

    text = message.strip()
    if not text:
        return text

    try:
        payload = json.loads(text)
    except json.JSONDecodeError:
        return text

    if isinstance(payload, dict):
        reply_text = payload.get("text")
        if isinstance(reply_text, str) and reply_text.strip():
            return reply_text.strip()
        nested = payload.get("payload")
        if isinstance(nested, str) and nested.strip():
            return nested.strip()
        if isinstance(nested, dict):
            nested_text = nested.get("text")
            if isinstance(nested_text, str) and nested_text.strip():
                return nested_text.strip()
    return text


@task(name="architecture-sync")
def architecture_sync(c, timeout=12):
    """Synchronize architect, development, and AGUI teams over the NATS bus."""
    import socket
    import time

    directives = {
        "prime-architect": {
            "command_subject": "swarm.team.prime-architect.internal.command",
            "reply_subjects": [
                "swarm.team.prime-architect.signal.status",
                "swarm.team.prime-architect.signal.result",
            ],
            "message": (
                "Central architecture directive: keep the next-target workflow aligned to strict gate order. "
                "P0 remains the active phase. Require concrete test evidence before any phase advancement, "
                "coordinate development and AGUI work to the target goals, and reply with a concise execution brief. "
                "Do not use tools for this sync. Respond in plain text with at most 6 short lines."
            ),
        },
        "prime-development": {
            "command_subject": "swarm.team.prime-development.internal.command",
            "reply_subjects": [
                "swarm.team.prime-development.signal.status",
                "swarm.team.prime-development.signal.result",
            ],
            "message": (
                "Development directive: focus on P0 closure, memory-restart reliability, logging standardization, "
                "error-handling normalization, and no-regression verification. Go remains the primary implementation "
                "language for backend/runtime work. Python is limited to tasks, management scripting, and tests. "
                "Reply with the top implementation/testing priorities. Do not use tools for this sync. "
                "Respond in plain text with at most 5 short lines."
            ),
        },
        "agui-design-architect": {
            "command_subject": "swarm.team.agui-design-architect.internal.command",
            "reply_subjects": [
                "swarm.team.agui-design-architect.signal.status",
                "swarm.team.agui-design-architect.signal.result",
            ],
            "message": (
                "AGUI directive: align base UI updates to architecture truth. Prioritize workflow-composer onboarding, "
                "gate-state visibility, system status, team roster visibility, and operator-safe error presentation. "
                "Do not invent client-side workflow semantics that diverge from backend gates. "
                "Reply with the top UI architecture priorities. Do not use tools for this sync. "
                "Respond in plain text with at most 5 short lines."
            ),
        },
    }

    print("=== Team Architecture Sync ===")
    print("Transport: NATS")
    print("Role: central architect")
    print("Target: keep architect, development, and AGUI teams aligned to current goals and testing gates.\n")

    sock = socket.create_connection(("127.0.0.1", 4222), timeout=5)
    try:
        info_line = _read_nats_line(sock)
        if not info_line.startswith("INFO "):
            raise RuntimeError(f"Unexpected NATS handshake: {info_line}")

        sock.sendall(b"CONNECT {\"verbose\":false,\"pedantic\":false}\r\n")

        sid = 1
        for config in directives.values():
            for subject in config["reply_subjects"]:
                sock.sendall(f"SUB {subject} {sid}\r\n".encode("utf-8"))
                sid += 1

        sock.sendall(b"PING\r\n")
        if _read_nats_line(sock) != "PONG":
            raise RuntimeError("NATS did not acknowledge the subscription flush")

        for team_id, config in directives.items():
            payload = config["message"].encode("utf-8")
            sock.sendall(
                f"PUB {config['command_subject']} {len(payload)}\r\n".encode("utf-8")
            )
            sock.sendall(payload + b"\r\n")
            print(f"Dispatched architecture directive to {team_id}")

        sock.sendall(b"PING\r\n")
        if _read_nats_line(sock) != "PONG":
            raise RuntimeError("NATS did not acknowledge the publish flush")

        acknowledgements: list[tuple[str, str]] = []
        responded_teams: set[str] = set()
        deadline = time.time() + float(timeout)
        while time.time() < deadline and len(responded_teams) < len(directives):
            new_messages = _drain_nats_messages(sock, 0.8)
            acknowledgements.extend(new_messages)
            for subject, _ in new_messages:
                for team_id, config in directives.items():
                    if subject in config["reply_subjects"]:
                        responded_teams.add(team_id)
            time.sleep(0.2)

        for subject, message in acknowledgements:
            print(f"[reply] {subject}: {_format_sync_reply(message)}")

        missing = [
            team_id
            for team_id in directives
            if team_id not in responded_teams
        ]
        if missing:
            print("\nMissing team replies:")
            for team_id in missing:
                print(f"  - {team_id}")
        else:
            print("\nAll teams replied inside the sync window.")
    finally:
        try:
            sock.sendall(b"QUIT\r\n")
        except OSError:
            pass
        sock.close()

ns_team = Collection("team")
ns_team.add_task(sensors)
ns_team.add_task(output)
ns_team.add_task(test)
ns_team.add_task(architecture_sync)
ns_team.add_task(worktree_triage)
