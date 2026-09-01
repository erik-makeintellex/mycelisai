from __future__ import annotations

from contextlib import contextmanager
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from threading import Thread

import pytest
from invoke.exceptions import Exit

from ops import cognitive_litellm


CLIENT_KEY = "loopback-scoped-key-must-not-leak"


class LiteLLMTestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.server.requests.append(
            {
                "method": self.command,
                "path": self.path,
                "authorization": self.headers.get("Authorization"),
            }
        )

        if self.path in {"/health/liveliness", "/health/readiness"}:
            self._send_json(200, {"status": "ok"})
            return
        if self.path == "/v1/models":
            if self.server.redirect_models:
                self.send_response(302)
                self.send_header("Location", "/redirect-target")
                self.end_headers()
                return
            if self.headers.get("Authorization") != f"Bearer {CLIENT_KEY}":
                self._send_json(401, {"error": CLIENT_KEY})
                return
            self._send_json(200, {"data": [{"id": "mycelis-default"}]})
            return
        if self.path == "/redirect-target":
            self._send_json(200, {"data": [{"id": "mycelis-default"}]})
            return
        self._send_json(404, {"error": "not found"})

    def _send_json(self, status, payload):
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, _format, *_args):
        return


@contextmanager
def loopback_litellm(*, redirect_models=False):
    server = ThreadingHTTPServer(("127.0.0.1", 0), LiteLLMTestHandler)
    server.requests = []
    server.redirect_models = redirect_models
    thread = Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        host, port = server.server_address
        yield f"http://{host}:{port}", server.requests
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=5)


def test_preflight_real_loopback_transport_uses_only_safe_get_paths(monkeypatch, capsys):
    monkeypatch.setenv(cognitive_litellm.CLIENT_KEY_ENV, CLIENT_KEY)

    with loopback_litellm() as (origin, requests):
        cognitive_litellm.preflight(
            f"{origin}/v1",
            cognitive_litellm.CLIENT_KEY_ENV,
            "mycelis-default",
            5,
        )

    assert requests == [
        {
            "method": "GET",
            "path": "/health/liveliness",
            "authorization": None,
        },
        {
            "method": "GET",
            "path": "/health/readiness",
            "authorization": None,
        },
        {
            "method": "GET",
            "path": "/v1/models",
            "authorization": f"Bearer {CLIENT_KEY}",
        },
    ]
    assert not any("completion" in request["path"] for request in requests)
    output = capsys.readouterr().out
    assert CLIENT_KEY not in output
    assert "[OK] Authenticated model alias: mycelis-default" in output
    assert "Completion  : SKIPPED" in output


def test_preflight_real_loopback_rejects_redirect_without_forwarding_key(
    monkeypatch,
    capsys,
):
    monkeypatch.setenv(cognitive_litellm.CLIENT_KEY_ENV, CLIENT_KEY)

    with loopback_litellm(redirect_models=True) as (origin, requests):
        with pytest.raises(Exit) as excinfo:
            cognitive_litellm.preflight(
                f"{origin}/v1",
                cognitive_litellm.CLIENT_KEY_ENV,
                "mycelis-default",
                5,
            )

    assert [request["path"] for request in requests] == [
        "/health/liveliness",
        "/health/readiness",
        "/v1/models",
    ]
    assert requests[-1]["authorization"] == f"Bearer {CLIENT_KEY}"
    assert not any(request["path"] == "/redirect-target" for request in requests)
    combined = capsys.readouterr().out + str(excinfo.value)
    assert CLIENT_KEY not in combined
    assert str(excinfo.value) == "LiteLLM authenticated model discovery failed (HTTP 302)"
