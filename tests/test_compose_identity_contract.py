from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
COMPOSE_FILE = ROOT / "docker-compose.yml"


def test_compose_shares_web_identity_secrets_between_interface_and_core():
    text = COMPOSE_FILE.read_text(encoding="utf-8")

    session_secret = "MYCELIS_WEB_SESSION_SECRET: ${MYCELIS_WEB_SESSION_SECRET:-${MYCELIS_API_KEY}}"
    forward_secret = (
        "MYCELIS_WEB_IDENTITY_FORWARD_SECRET: "
        "${MYCELIS_WEB_IDENTITY_FORWARD_SECRET:-${MYCELIS_API_KEY}}"
    )

    assert text.count(session_secret) == 2
    assert text.count(forward_secret) == 2
