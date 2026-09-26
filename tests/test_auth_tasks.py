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
    compose_example_file = tmp_path / ".env.compose.example"
    env_file.write_text("DB_HOST=127.0.0.1\n", encoding="utf-8")
    example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")
    compose_example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")

    monkeypatch.setattr(auth, "ENV_PATH", env_file)
    monkeypatch.setattr(auth, "ENV_EXAMPLE_PATH", example_file)
    monkeypatch.setattr(auth, "ENV_COMPOSE_EXAMPLE_PATH", compose_example_file)

    auth.dev_key.body(Context(), rotate=False, show=False, value="")

    env_text = env_file.read_text(encoding="utf-8")
    example_text = example_file.read_text(encoding="utf-8")
    assert "MYCELIS_API_KEY=mycelis-dev-" in env_text
    assert f"MYCELIS_API_KEY={auth.SAMPLE_VALUE}" in example_text
    assert f"MYCELIS_BREAK_GLASS_API_KEY={auth.BREAK_GLASS_SAMPLE_VALUE}" in example_text
    assert f"MYCELIS_LOCAL_ADMIN_USERNAME={auth.LOCAL_ADMIN_SAMPLE_USERNAME}" in example_text
    assert f"MYCELIS_BREAK_GLASS_USERNAME={auth.BREAK_GLASS_SAMPLE_USERNAME}" in example_text
    assert "MYCELIS_API_KEY=placeholder" in compose_example_file.read_text(encoding="utf-8")


def test_break_glass_key_generates_and_syncs_example(tmp_path: Path, monkeypatch):
    env_file = tmp_path / ".env"
    example_file = tmp_path / ".env.example"
    compose_example_file = tmp_path / ".env.compose.example"
    env_file.write_text("MYCELIS_API_KEY=primary\n", encoding="utf-8")
    example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")
    compose_example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")

    monkeypatch.setattr(auth, "ENV_PATH", env_file)
    monkeypatch.setattr(auth, "ENV_EXAMPLE_PATH", example_file)
    monkeypatch.setattr(auth, "ENV_COMPOSE_EXAMPLE_PATH", compose_example_file)

    auth.break_glass_key.body(Context(), rotate=False, show=False, value="")

    env_text = env_file.read_text(encoding="utf-8")
    example_text = example_file.read_text(encoding="utf-8")
    assert "MYCELIS_BREAK_GLASS_API_KEY=mycelis-break-glass-" in env_text
    assert f"MYCELIS_BREAK_GLASS_API_KEY={auth.BREAK_GLASS_SAMPLE_VALUE}" in example_text
    assert "MYCELIS_API_KEY=placeholder" in compose_example_file.read_text(encoding="utf-8")


def test_break_glass_key_rotates_with_break_glass_prefix(tmp_path: Path, monkeypatch):
    env_file = tmp_path / ".env"
    example_file = tmp_path / ".env.example"
    compose_example_file = tmp_path / ".env.compose.example"
    env_file.write_text("MYCELIS_BREAK_GLASS_API_KEY=old\n", encoding="utf-8")
    example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")
    compose_example_file.write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")

    monkeypatch.setattr(auth, "ENV_PATH", env_file)
    monkeypatch.setattr(auth, "ENV_EXAMPLE_PATH", example_file)
    monkeypatch.setattr(auth, "ENV_COMPOSE_EXAMPLE_PATH", compose_example_file)

    auth.break_glass_key.body(Context(), rotate=True, show=False, value="")

    env_text = env_file.read_text(encoding="utf-8")
    assert "MYCELIS_BREAK_GLASS_API_KEY=mycelis-break-glass-" in env_text


def test_posture_reports_break_glass_warnings(tmp_path: Path, monkeypatch, capsys):
    env_file = tmp_path / ".env"
    env_file.write_text(
        "\n".join(
            [
                "MYCELIS_API_KEY=primary",
                "MYCELIS_BREAK_GLASS_API_KEY=break",
                "MYCELIS_BREAK_GLASS_USERNAME=recovery-admin",
            ]
        )
        + "\n",
        encoding="utf-8",
    )

    monkeypatch.setattr(auth, "ENV_PATH", env_file)

    auth.posture.body(Context(), compose=False)

    captured = capsys.readouterr().out
    assert "primary_key" in captured
    assert "break_glass_key" in captured
    assert "warning: partial break-glass config" in captured


def _patch_auth_paths(tmp_path: Path, monkeypatch, env_text: str) -> Path:
    env_file = tmp_path / ".env"
    env_file.write_text(env_text, encoding="utf-8")
    (tmp_path / ".env.example").write_text("MYCELIS_API_KEY=placeholder\n", encoding="utf-8")
    monkeypatch.setattr(auth, "ENV_PATH", env_file)
    monkeypatch.setattr(auth, "ENV_EXAMPLE_PATH", tmp_path / ".env.example")
    monkeypatch.setattr(auth, "ENV_COMPOSE_EXAMPLE_PATH", tmp_path / ".env.compose.example")
    return env_file


def test_dev_key_ensures_distinct_web_secrets_without_printing_them(tmp_path: Path, monkeypatch, capsys):
    env_file = _patch_auth_paths(tmp_path, monkeypatch, "DB_HOST=127.0.0.1\n")

    auth.dev_key.body(Context(), rotate=False, show=False, value="")

    values = {name: auth._read_env_value(env_file, name) for name in auth.WEB_SECRET_NAMES + (auth.KEY_NAME,)}
    assert all(len(value) >= 32 for value in values.values())
    assert len(set(values.values())) == len(values)
    out = capsys.readouterr().out
    for name in auth.WEB_SECRET_NAMES:
        assert name in out
        assert values[name] not in out
        assert values[name][:8] + "..." not in out
    example = (tmp_path / ".env.example").read_text(encoding="utf-8")
    for name in auth.WEB_SECRET_NAMES:
        assert f"{name}={auth.WEB_SECRET_SAMPLES[name]}" in example


def test_dev_key_replaces_web_secret_that_reuses_the_api_key(tmp_path: Path, monkeypatch):
    api_key = "mycelis-dev-" + "a" * 40
    kept = "k" * 48
    env_file = _patch_auth_paths(
        tmp_path,
        monkeypatch,
        f"MYCELIS_API_KEY={api_key}\nMYCELIS_WEB_SESSION_SECRET={api_key}\nMYCELIS_WEB_IDENTITY_FORWARD_SECRET={kept}\n",
    )

    auth.dev_key.body(Context(), rotate=False, show=False, value="")

    assert auth._read_env_value(env_file, "MYCELIS_WEB_SESSION_SECRET") not in {api_key, kept}
    assert auth._read_env_value(env_file, "MYCELIS_WEB_IDENTITY_FORWARD_SECRET") == kept
    assert auth._read_env_value(env_file, "MYCELIS_API_KEY") == api_key


def test_posture_warns_about_missing_short_or_reused_web_secrets(tmp_path: Path, monkeypatch, capsys):
    _patch_auth_paths(
        tmp_path,
        monkeypatch,
        "MYCELIS_API_KEY=primary-key-0123456789abcdef0123456789\n"
        "MYCELIS_WEB_SESSION_SECRET=primary-key-0123456789abcdef0123456789\n"
        "MYCELIS_WEB_IDENTITY_FORWARD_SECRET=short\n",
    )

    auth.posture.body(Context(), compose=False)

    out = capsys.readouterr().out
    assert "warning: MYCELIS_WEB_SESSION_SECRET must differ" in out
    assert "warning: MYCELIS_WEB_IDENTITY_FORWARD_SECRET must be at least 32 bytes" in out
    assert "warning: set MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256" in out
    assert "primary-key-0123456789abcdef0123456789" not in out
