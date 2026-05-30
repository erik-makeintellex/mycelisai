from invoke import Context

from ops import lifecycle


def test_health_uses_extended_timeout_for_cognitive_status(monkeypatch):
    calls: list[tuple[str, float]] = []
    monkeypatch.setattr(lifecycle, "_load_env", lambda: None)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: True)
    monkeypatch.setattr(
        lifecycle,
        "_http_get",
        lambda url, timeout=3.0, headers=None: calls.append((url, timeout)) or (200, "ok"),
    )

    lifecycle.health.body(Context())

    timeouts = {url.rsplit("/", 1)[-1]: timeout for url, timeout in calls}
    assert timeouts["status"] == lifecycle.COGNITIVE_STATUS_TIMEOUT_SECONDS
    assert timeouts["templates"] == 5.0
