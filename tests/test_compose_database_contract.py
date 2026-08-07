from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def test_compose_database_role_uses_the_local_core_password_contract():
    compose_text = (ROOT / "docker-compose.yml").read_text(encoding="utf-8")

    assert "POSTGRES_PASSWORD: ${DB_PASSWORD:-password}" in compose_text
    assert "POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-password}" not in compose_text
