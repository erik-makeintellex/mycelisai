from __future__ import annotations

import io
import json
import urllib.error

import pytest
from invoke.exceptions import Exit

from ops import cognitive_litellm


CLIENT_KEY = "scoped-client-key-must-not-leak"


class FakeResponse:
    def __init__(self, body: bytes = b""):
        self.body = body

    def __enter__(self):
        return self

    def __exit__(self, *_args):
        return False

    def read(self, limit: int = -1) -> bytes:
        return self.body if limit < 0 else self.body[:limit]


def _run_success(monkeypatch, capsys):
    requests = []

    def fake_urlopen(request, timeout):
        requests.append((request, timeout))
        if request.full_url.endswith("/v1/models"):
            return FakeResponse(json.dumps({"data": [{"id": "mycelis-default"}]}).encode())
        return FakeResponse()

    monkeypatch.setenv(cognitive_litellm.CLIENT_KEY_ENV, CLIENT_KEY)
    monkeypatch.setattr(cognitive_litellm, "_open_without_redirects", fake_urlopen)

    cognitive_litellm.preflight(
        "https://gateway.example.test/proxy/v1",
        cognitive_litellm.CLIENT_KEY_ENV,
        "mycelis-default",
        4,
    )
    return requests, capsys.readouterr().out


def test_preflight_uses_safe_get_probes_and_scoped_auth(monkeypatch, capsys):
    requests, output = _run_success(monkeypatch, capsys)

    assert [request.full_url for request, _timeout in requests] == [
        "https://gateway.example.test/proxy/health/liveliness",
        "https://gateway.example.test/proxy/health/readiness",
        "https://gateway.example.test/proxy/v1/models",
    ]
    assert all(request.get_method() == "GET" for request, _timeout in requests)
    assert [timeout for _request, timeout in requests] == [4.0, 4.0, 4.0]
    assert requests[0][0].get_header("Authorization") is None
    assert requests[1][0].get_header("Authorization") is None
    assert requests[2][0].get_header("Authorization") == f"Bearer {CLIENT_KEY}"
    assert CLIENT_KEY not in output
    assert "Completion  : SKIPPED" in output
    assert "CORRELATION-CAPABLE transport posture" in output
    assert "Production enablement remains open" in output
    assert "non-swarm inference is not scope-correlated" in output


@pytest.mark.parametrize(
    ("endpoint", "message"),
    [
        ("", "--litellm-endpoint is required"),
        ("ftp://gateway.example.test/v1", r"absolute HTTP\(S\) URL"),
        ("https://user:password@gateway.example.test/v1", "must not contain credentials"),
        ("https://gateway.example.test/v1?key=secret", "must not contain query"),
        ("https://gateway.example.test", "must end with /v1"),
        ("http://gateway.example.test/v1", "public LiteLLM endpoints must use HTTPS"),
    ],
)
def test_preflight_rejects_unsafe_or_ambiguous_endpoints(endpoint, message):
    with pytest.raises(Exit, match=message):
        cognitive_litellm._normalize_endpoint(endpoint)


def test_preflight_allows_private_http_endpoint():
    assert cognitive_litellm._normalize_endpoint("http://litellm:4000/v1") == (
        "http://litellm:4000/v1",
        "http://litellm:4000",
        "http://litellm:4000",
    )


def test_preflight_redirect_handler_keeps_scoped_key_on_approved_origin():
    handler = cognitive_litellm._RejectRedirects()
    request = cognitive_litellm.urllib.request.Request(
        "https://gateway.example.test/v1/models",
        headers={"Authorization": f"Bearer {CLIENT_KEY}"},
    )

    assert (
        handler.redirect_request(
            request,
            None,
            302,
            "Found",
            {},
            "https://redirect.example.test/v1/models",
        )
        is None
    )


def test_preflight_rejects_proxy_administration_key_reference(monkeypatch):
    monkeypatch.delenv(cognitive_litellm.CLIENT_KEY_ENV, raising=False)

    with pytest.raises(Exit, match="proxy administration keys are not accepted"):
        cognitive_litellm.preflight(
            "https://gateway.example.test/v1",
            "LITELLM_MASTER_KEY",
            "mycelis-default",
            5,
        )


def test_preflight_requires_scoped_client_key_without_exposing_dotenv_value(
    monkeypatch,
    tmp_path,
):
    monkeypatch.delenv(cognitive_litellm.CLIENT_KEY_ENV, raising=False)
    monkeypatch.setattr(cognitive_litellm, "ROOT_DIR", tmp_path)
    (tmp_path / ".env").write_text(
        f"{cognitive_litellm.CLIENT_KEY_ENV}=\n",
        encoding="utf-8",
    )

    with pytest.raises(Exit) as excinfo:
        cognitive_litellm.preflight(
            "https://gateway.example.test/v1",
            cognitive_litellm.CLIENT_KEY_ENV,
            "mycelis-default",
            5,
        )

    assert str(excinfo.value) == (
        f"{cognitive_litellm.CLIENT_KEY_ENV} must be set in the shell or repo-local .env"
    )


@pytest.mark.parametrize("timeout", [0, 31, "not-a-number"])
def test_preflight_bounds_probe_timeout(timeout):
    with pytest.raises(Exit, match="from 1 through 30 seconds"):
        cognitive_litellm.preflight(
            "https://gateway.example.test/v1",
            cognitive_litellm.CLIENT_KEY_ENV,
            "mycelis-default",
            timeout,
        )


def test_preflight_redacts_transport_failure(monkeypatch, capsys):
    monkeypatch.setenv(cognitive_litellm.CLIENT_KEY_ENV, CLIENT_KEY)
    monkeypatch.setattr(
        cognitive_litellm,
        "_open_without_redirects",
        lambda *_args, **_kwargs: (_ for _ in ()).throw(
            urllib.error.URLError(f"upstream failed with {CLIENT_KEY}")
        ),
    )

    with pytest.raises(Exit) as excinfo:
        cognitive_litellm.preflight(
            "https://gateway.example.test/v1",
            cognitive_litellm.CLIENT_KEY_ENV,
            "mycelis-default",
            5,
        )

    combined = capsys.readouterr().out + str(excinfo.value)
    assert CLIENT_KEY not in combined
    assert "could not reach the configured proxy" in combined


def test_preflight_redacts_auth_response_body(monkeypatch, capsys):
    calls = 0

    def fake_urlopen(request, timeout):
        nonlocal calls
        calls += 1
        if calls < 3:
            return FakeResponse()
        raise urllib.error.HTTPError(
            request.full_url,
            401,
            f"invalid key {CLIENT_KEY}",
            hdrs=None,
            fp=io.BytesIO(f'{{"error":"{CLIENT_KEY}"}}'.encode()),
        )

    monkeypatch.setenv(cognitive_litellm.CLIENT_KEY_ENV, CLIENT_KEY)
    monkeypatch.setattr(cognitive_litellm, "_open_without_redirects", fake_urlopen)

    with pytest.raises(Exit) as excinfo:
        cognitive_litellm.preflight(
            "https://gateway.example.test/v1",
            cognitive_litellm.CLIENT_KEY_ENV,
            "mycelis-default",
            5,
        )

    combined = capsys.readouterr().out + str(excinfo.value)
    assert CLIENT_KEY not in combined
    assert "rejected authentication (HTTP 401)" in combined


def test_preflight_rejects_unlisted_alias_without_printing_model_inventory(
    monkeypatch,
    capsys,
):
    hidden_model = "sensitive-upstream-model-name"

    def fake_urlopen(request, timeout):
        if request.full_url.endswith("/v1/models"):
            return FakeResponse(json.dumps({"data": [{"id": hidden_model}]}).encode())
        return FakeResponse()

    monkeypatch.setenv(cognitive_litellm.CLIENT_KEY_ENV, CLIENT_KEY)
    monkeypatch.setattr(cognitive_litellm, "_open_without_redirects", fake_urlopen)

    with pytest.raises(Exit) as excinfo:
        cognitive_litellm.preflight(
            "https://gateway.example.test/v1",
            cognitive_litellm.CLIENT_KEY_ENV,
            "mycelis-default",
            5,
        )

    combined = capsys.readouterr().out + str(excinfo.value)
    assert hidden_model not in combined
    assert "1 model(s) visible" in combined
