from __future__ import annotations

from pathlib import Path

from invoke import Context

from ops import auth


def test_upsert_env_value_adds_missing_key(tmp_path: Path):
    env_file = tmp_path / ".env"
    env_file.write_text("DB_HOST=127.0.0.1\n", encoding="utf-8")

    auth._upsert_env_value(env_file, "MYCELIS_API_KEY", "abc123")

    text = env_file.read_text(encoding="utf-8")
    assert "DB_HOST=127.0.0.1" in text
    assert "MYCELIS_API_KEY=abc123" in text


def test_upsert_env_value_replaces_existing_key(tmp_path: Path):
    env_file = tmp_path / ".env"
    env_file.write_text("MYCELIS_API_KEY=old\nDB_HOST=127.0.0.1\n", encoding="utf-8")

    auth._upsert_env_value(env_file, "MYCELIS_API_KEY", "new")

    text = env_file.read_text(encoding="utf-8")
    assert "MYCELIS_API_KEY=new" in text
    assert "MYCELIS_API_KEY=old" not in text


def test_dev_key_generates_when_missing_and_syncs_example(tmp_path: Path, monkeypatch):
    env_file = tmp_path / ".env"
    example_file = tmp_path / ".env.example"
    env_file.write_text("DB_HOST=127.0.0.1\n", encoding="utf-8")
    example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")

    monkeypatch.setattr(auth, "ENV_PATH", env_file)
    monkeypatch.setattr(auth, "ENV_EXAMPLE_PATH", example_file)

    auth.dev_key.body(Context(), rotate=False, show=False, value="")

    env_text = env_file.read_text(encoding="utf-8")
    example_text = example_file.read_text(encoding="utf-8")
    assert "MYCELIS_API_KEY=mycelis-dev-" in env_text
    assert f"MYCELIS_API_KEY={auth.SAMPLE_VALUE}" in example_text
