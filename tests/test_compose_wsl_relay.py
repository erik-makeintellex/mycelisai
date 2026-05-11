from __future__ import annotations

from ops import compose_wsl_relay


def test_wsl_ollama_relay_is_restartable():
    calls: list[tuple[list[str], bool]] = []

    def docker_runner(args: list[str], check: bool = True):
        calls.append((args, check))

        class Result:
            stdout = "relay-container\n"
            returncode = 0

        return Result()

    compose_wsl_relay.ensure(
        "127.0.0.1",
        11434,
        11435,
        relay_name="mycelis-home-ollama-relay",
        relay_image="alpine:3.21",
        inspect_relay_labels=lambda: None,
        stop_relay=lambda: None,
        docker_runner=docker_runner,
    )

    run_args = calls[0][0]
    assert "--restart" in run_args
    assert "unless-stopped" in run_args
    assert "mycelis.relay.restart_policy=unless-stopped" in run_args
    assert "--rm" not in run_args


def test_wsl_ollama_relay_missing_restart_label_is_recreated():
    calls: list[tuple[str, int, int]] = []
    values = compose_wsl_relay.prepare_host(
        {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434"},
        docker_host_mode=lambda: "wsl",
        running_in_wsl=lambda: False,
        clean_env_value=lambda value: value.strip(),
        parse_network_endpoint=lambda _url: ("host.docker.internal", 11434),
        relay_port=lambda _values: 11435,
        inspect_relay_labels=lambda: {
            "mycelis.relay.listen_port": "11435",
            "mycelis.relay.target_host": "127.0.0.1",
            "mycelis.relay.target_port": "11434",
        },
        wsl_http_available=lambda url: url == "http://127.0.0.1:11434",
        ensure_relay=lambda host, port, relay_port: calls.append((host, port, relay_port)),
    )

    assert values["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://host.docker.internal:11435"
    assert calls == [("127.0.0.1", 11434, 11435)]
