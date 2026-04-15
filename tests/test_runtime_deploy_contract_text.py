from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
TESTING = ROOT / "docs" / "TESTING.md"
OPERATIONS = ROOT / "docs" / "architecture" / "OPERATIONS.md"


def test_compose_testing_contract_points_to_explicit_non_loopback_ai_hosts():
    snippets = [
        (
            README,
            [
                "use a reachable host/IP like `http://192.168.x.x:11434/v1`, not `localhost`",
                "point it at a host-reachable endpoint such as `http://host.docker.internal:11434`",
                "auto-start a WSL-host relay for the AI endpoint when needed",
            ],
        ),
        (
            TESTING,
            [
                "use `MYCELIS_K8S_TEXT_ENDPOINT` and optional `MYCELIS_K8S_MEDIA_ENDPOINT`",
                "explicit reachable AI host instead of a chart-baked or localhost default",
                "keep it container-reachable instead of `localhost`, `127.0.0.1`, or `0.0.0.0`",
                "may relay `MYCELIS_COMPOSE_OLLAMA_HOST` through the WSL host",
            ],
        ),
        (
            OPERATIONS,
            [
                "use explicit reachable AI endpoints for deployed text or media engines instead of localhost assumptions",
                "the Helm chart applies `MYCELIS_K8S_TEXT_ENDPOINT` through provider-specific env overrides",
                "use a reachable Windows IP or hostname such as `http://192.168.x.x:11434/v1`, not `localhost`",
                "can auto-start a WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST`",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Compose/testing contract is missing explicit non-loopback AI endpoint guidance:\n" + "\n".join(missing)
