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
        inboxes: list[str] = []
        for item in fake_sock.sent:
            if item.startswith(b"SUB "):
                inboxes.append(item.decode("utf-8").split()[1])
        return [
            (subject, f'{{"text":"ack:{idx}","tools_used":[]}}')
            for idx, subject in enumerate(inboxes, start=1)
        ]

    monkeypatch.setattr("socket.create_connection", lambda *args, **kwargs: fake_sock)
    monkeypatch.setattr(misc, "_read_nats_line", fake_read_nats_line)
    monkeypatch.setattr(misc, "_drain_nats_messages", fake_drain)
    monkeypatch.setattr("time.time", clock.time)
    monkeypatch.setattr("time.sleep", lambda _n: None)

    misc.architecture_sync.body(Context(), timeout=2)

    output = capsys.readouterr().out
    assert "Requested execution brief from prime-architect" in output
    assert "Requested execution brief from prime-development" in output
    assert "Requested execution brief from agui-design-architect" in output
    assert "All teams replied inside the sync window." in output

    published = b"".join(fake_sock.sent)
    assert b"swarm.council.prime-architect-agent.request" in published
    assert b"swarm.council.prime-development-agent.request" in published
    assert b"swarm.council.agui-design-architect-agent.request" in published
    assert published.endswith(b"QUIT\r\n")
    assert fake_sock.closed is True


def test_format_sync_reply_prefers_process_result_text():
    message = '{"text":"brief body","tools_used":["publish_signal"]}'

    assert misc._format_sync_reply(message) == "brief body"
