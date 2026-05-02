from ops import interface_runtime as interface
from tests.interface_task_support import FakeContext, FakeResult

def test_test_uses_direct_shell_command(monkeypatch):
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.test.body(ctx)

    assert shell_calls == [["npm", "run", "test"]]


def test_typecheck_uses_direct_shell_command(monkeypatch):
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.typecheck.body(ctx)

    assert shell_calls == [["npx", "tsc", "--noEmit"]]


def test_test_coverage_uses_direct_shell_command(monkeypatch):
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.test_coverage.body(ctx)

    assert shell_calls == [["npx", "vitest", "run", "--coverage"]]


def test_interface_ready_urls_prioritize_requested_host_then_loopback_fallbacks():
    urls = interface._interface_ready_urls("127.0.0.1", 3000)

    assert urls == [
        "http://127.0.0.1:3000",
        "http://localhost:3000",
        "http://[::1]:3000",
    ]


def test_pick_interface_port_falls_back_when_preferred_port_is_busy(monkeypatch):
    occupied_ports = {3000}

    class FakeSocket:
        def __init__(self, family, *args, **kwargs):
            self.family = family
            self.bind_calls: list[tuple[str, int]] = []
            self.closed = False

        def setsockopt(self, *args, **kwargs):
            return None

        def settimeout(self, *args, **kwargs):
            return None

        def connect_ex(self, address):
            return 0 if address[1] in occupied_ports else 111

        def bind(self, address):
            self.bind_calls.append(address)
            if address[1] in occupied_ports:
                raise OSError("address in use")
            self.port = 43210 if address[1] == 3000 else 43211

        def getsockname(self):
            return ("127.0.0.1", getattr(self, "port", 43210))

        def close(self):
            self.closed = True

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            self.close()

    monkeypatch.setattr(interface.socket, "socket", lambda family, *args, **kwargs: FakeSocket(family))

    assert interface._pick_interface_port(3000) == 3100


def test_pick_interface_port_uses_ipv6_bind_host_without_dual_binding(monkeypatch):
    occupied_ports: set[int] = set()
    bind_calls: list[tuple[int, tuple[str, int]]] = []
    sockopts: list[tuple[int, int, int]] = []

    class FakeSocket:
        def __init__(self, family, *args, **kwargs):
            self.family = family
            self.port = 43210

        def setsockopt(self, level, option, value):
            sockopts.append((level, option, value))

        def settimeout(self, *args, **kwargs):
            return None

        def connect_ex(self, address):
            return 0 if address[1] in occupied_ports else 111

        def bind(self, address):
            bind_calls.append((self.family, address))
            if address[1] in occupied_ports:
                raise OSError("address in use")
            self.port = address[1] or 43210

        def getsockname(self):
            if self.family == interface.socket.AF_INET6:
                return ("::", self.port)
            return ("127.0.0.1", self.port)

        def close(self):
            return None

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            self.close()

    monkeypatch.setattr(interface, "INTERFACE_BIND_HOST", "::")
    monkeypatch.setattr(interface, "is_windows", lambda: False)
    monkeypatch.setattr(interface.socket, "socket", lambda family, *args, **kwargs: FakeSocket(family))

    assert interface._pick_interface_port(3000) == 3100
    assert bind_calls == [(interface.socket.AF_INET6, ("::", 3100))]
    assert sockopts == [(interface.socket.IPPROTO_IPV6, interface.socket.IPV6_V6ONLY, 0)]


def test_wait_for_interface_ready_fails_when_managed_server_exits(monkeypatch):
    class FakeServer:
        @staticmethod
        def poll():
            return 1

    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    try:
        interface._wait_for_interface_ready("127.0.0.1", 4310, timeout_seconds=1, process=FakeServer())
    except RuntimeError as exc:
        assert "Managed Interface server exited before it became ready" in str(exc)
        assert "4310" in str(exc)
    else:
        raise AssertionError("expected managed server startup failure")


def test_wait_for_interface_ready_prefers_reachable_port_over_exited_parent(monkeypatch):
    class FakeServer:
        @staticmethod
        def poll():
            return 1

    class FakeHTTPResponse:
        status = 200

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            return False

    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)
    monkeypatch.setattr(interface.urllib.request, "urlopen", lambda url, timeout=5: FakeHTTPResponse())

    assert interface._wait_for_interface_ready("127.0.0.1", 4310, timeout_seconds=1, process=FakeServer()) == "127.0.0.1"


def test_check_does_not_treat_plain_html_words_as_hydration_failure(monkeypatch, capsys):
    class FakeHTTPResponse:
        def __init__(self, body: str):
            self.status = 200
            self._body = body.encode("utf-8")

        def read(self):
            return self._body

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            return False

    monkeypatch.setattr(
        interface.urllib.request,
        "urlopen",
        lambda req, timeout=10: FakeHTTPResponse("<html><body>hydration and error words in static docs text</body></html>"),
    )

    interface.check.body(FakeContext(), port=3000)

    out = capsys.readouterr().out
    assert "ALL PAGES HEALTHY." in out
