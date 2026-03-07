from invoke import task, Collection
from .config import ROOT_DIR

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
