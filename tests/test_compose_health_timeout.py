import pytest

from ops import compose


def test_compose_health_uses_extended_timeout_for_cognitive_status(monkeypatch):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"MYCELIS_API_KEY": "test-key"})
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)
    calls: list[tuple[str, float]] = []
    responses = {
        f"http://{compose.API_HOST}:{compose.API_PORT}/healthz": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/templates": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/brains": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/telemetry/compute": (200, "ok"),
        f"http://{compose.INTERFACE_HOST}:{compose.INTERFACE_PORT}/": (200, "ok"),
        "http://127.0.0.1:8222/varz": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/cognitive/status": (200, '{"text":{"status":"offline"}}'),
    }
    monkeypatch.setattr(
        compose,
        "_http_get",
        lambda url, timeout=3.0, headers=None: calls.append((url, timeout)) or responses[url],
    )

    with pytest.raises(SystemExit):
        compose.health.body(None)

    cognitive_url = f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/cognitive/status"
    assert (cognitive_url, compose.compose_probe.COGNITIVE_STATUS_TIMEOUT_SECONDS) in calls
