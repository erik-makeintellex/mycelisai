from ops import compose


def test_compose_up_probes_configured_host_ports(monkeypatch):
    ports: list[tuple[str, int]] = []
    urls: list[str] = []
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {
        "MYCELIS_COMPOSE_POSTGRES_PORT": "15432",
        "MYCELIS_COMPOSE_NATS_PORT": "14222",
        "MYCELIS_COMPOSE_CORE_PORT": "18081",
        "MYCELIS_COMPOSE_INTERFACE_PORT": "13000",
    })
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda values: values)
    monkeypatch.setattr(compose, "_run_compose", lambda *args, **kwargs: None)
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: None)
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, **kwargs: ports.append((label, port)) or True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda **kwargs: True)
    monkeypatch.setattr(compose, "_wait_for_http_ok", lambda url, *args, **kwargs: urls.append(url) or True)
    monkeypatch.setattr(compose.status, "body", lambda _c=None: None)

    compose.up.body(None)

    assert ports == [
        ("PostgreSQL", 15432),
        ("NATS", 14222),
        ("Core API", 18081),
        ("Frontend", 13000),
    ]
    assert urls == [f"http://{compose.API_HOST}:18081/healthz", f"http://{compose.INTERFACE_HOST}:13000/"]
