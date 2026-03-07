from __future__ import annotations

from invoke import Context

from ops import misc


class _RecvSocket:
    def __init__(self, payload: bytes):
        self.payload = payload
        self.offset = 0

    def recv(self, size: int) -> bytes:
        if self.offset >= len(self.payload):
            return b""
        chunk = self.payload[self.offset:self.offset + size]
        self.offset += len(chunk)
        return chunk


class _FakeSocket:
    def __init__(self):
        self.sent: list[bytes] = []
        self.closed = False

    def sendall(self, data: bytes):
        self.sent.append(data)

    def settimeout(self, _value: float):
        return None

    def close(self):
        self.closed = True


def test_read_nats_line_reads_until_crlf():
    sock = _RecvSocket(b"INFO {}\r\nPING\r\n")

    assert misc._read_nats_line(sock) == "INFO {}"
    assert misc._read_nats_line(sock) == "PING"


def test_architecture_sync_requests_all_teams_and_reports_replies(monkeypatch, capsys):
    fake_sock = _FakeSocket()
    handshake = iter(["INFO {}", "PONG", "PONG"])
    replies_sent = False

    class _FakeClock:
        def __init__(self):
            self.value = 1000.0

        def time(self) -> float:
            self.value += 0.25
            return self.value

    clock = _FakeClock()

    def fake_read_nats_line(sock) -> str:
        assert sock is fake_sock
        return next(handshake)

    def fake_drain(sock, _timeout_seconds: float) -> list[tuple[str, str]]:
        nonlocal replies_sent
        assert sock is fake_sock
        if replies_sent:
            return []
        replies_sent = True
        reply_subjects: list[str] = []
        for item in fake_sock.sent:
            if item.startswith(b"SUB "):
                subject = item.decode("utf-8").split()[1]
                if subject.endswith(".signal.status"):
                    reply_subjects.append(subject)
        return [
            (
                subject,
                (
                    '{"meta":{"source_kind":"system","source_channel":"swarm.team.test.internal.response",'
                    '"payload_kind":"status","team_id":"test"},'
                    f'"text":"ack:{idx}"}}'
                ),
            )
            for idx, subject in enumerate(reply_subjects, start=1)
        ]

    monkeypatch.setattr("socket.create_connection", lambda *args, **kwargs: fake_sock)
    monkeypatch.setattr(misc, "_read_nats_line", fake_read_nats_line)
    monkeypatch.setattr(misc, "_drain_nats_messages", fake_drain)
    monkeypatch.setattr("time.time", clock.time)
    monkeypatch.setattr("time.sleep", lambda _n: None)

    misc.architecture_sync.body(Context(), timeout=2)

    output = capsys.readouterr().out
    assert "Dispatched architecture directive to prime-architect" in output
    assert "Dispatched architecture directive to prime-development" in output
    assert "Dispatched architecture directive to agui-design-architect" in output
    assert "All teams replied inside the sync window." in output

    published = b"".join(fake_sock.sent)
    assert b"swarm.team.prime-architect.internal.command" in published
    assert b"swarm.team.prime-development.internal.command" in published
    assert b"swarm.team.agui-design-architect.internal.command" in published
    assert b"swarm.team.prime-architect.signal.status" in published
    assert b"swarm.team.prime-development.signal.status" in published
    assert b"swarm.team.agui-design-architect.signal.status" in published
    assert published.endswith(b"QUIT\r\n")
    assert fake_sock.closed is True


def test_format_sync_reply_prefers_process_result_text():
    message = '{"text":"brief body","tools_used":["publish_signal"]}'

    assert misc._format_sync_reply(message) == "brief body"


def test_format_sync_reply_prefers_wrapped_signal_text():
    message = '{"meta":{"payload_kind":"status"},"text":"wrapped brief"}'

    assert misc._format_sync_reply(message) == "wrapped brief"
