"""auth.dev-key --admin-password: keep the local admin plaintext and its SHA-256 consistent
in .env, so e2e (and any local tooling) signs in from .env with no prompt or export.
Every value here is a fixture; nothing reads a real .env."""
from __future__ import annotations

import hashlib
from pathlib import Path

import pytest
from invoke import Context

from ops import auth

WEB_SECRETS = (
    "MYCELIS_WEB_SESSION_SECRET=fixture-session-secret-0000000000000000000000\n"
    "MYCELIS_WEB_IDENTITY_FORWARD_SECRET=fixture-forward-secret-000000000000000000000\n"
    "MYCELIS_API_KEY=fixture-api-key-value-0000000000\n"
)


def _sha(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


@pytest.fixture
def env_file(tmp_path: Path, monkeypatch) -> Path:
    path = tmp_path / ".env"
    monkeypatch.setattr(auth, "ENV_PATH", path)
    monkeypatch.setattr(auth, "ENV_EXAMPLE_PATH", tmp_path / ".env.example")
    return path


def _run(admin_password: str, show: bool = False):
    auth.dev_key.body(Context(), rotate=False, show=show, value="", admin_password=admin_password)


def _admin(env_file: Path) -> tuple[str, str]:
    return (auth._read_env_value(env_file, auth.LOCAL_PASSWORD_NAME), auth._read_env_value(env_file, auth.LOCAL_PASSWORD_SHA256_NAME))


def test_sync_on_hash_only_env_leaves_the_local_admin_unchanged(env_file, capsys):
    env_file.write_text(WEB_SECRETS + f"MYCELIS_LOCAL_ADMIN_PASSWORD=\nMYCELIS_LOCAL_ADMIN_PASSWORD_SHA256={_sha('old-fixture')}\n", encoding="utf-8")

    _run("sync")

    assert _admin(env_file) == ("", _sha("old-fixture"))
    out = capsys.readouterr().out
    assert "--admin-password=generate" in out and "unchanged" in out


def test_generate_writes_a_matching_plaintext_and_hash_without_printing_it(env_file, capsys):
    env_file.write_text(WEB_SECRETS + f"MYCELIS_LOCAL_ADMIN_PASSWORD=\nMYCELIS_LOCAL_ADMIN_PASSWORD_SHA256={_sha('old-fixture')}\n", encoding="utf-8")

    _run("generate")

    password, digest = _admin(env_file)
    assert len(password) >= 20 and digest == _sha(password)
    assert password not in capsys.readouterr().out


def test_explicit_password_sets_plaintext_and_matching_hash(env_file):
    env_file.write_text(WEB_SECRETS, encoding="utf-8")

    _run("fixture-owner-chosen-password")

    assert _admin(env_file) == ("fixture-owner-chosen-password", _sha("fixture-owner-chosen-password"))


def test_sync_rederives_a_mismatched_hash_from_the_plaintext(env_file, capsys):
    env_file.write_text(WEB_SECRETS + f"MYCELIS_LOCAL_ADMIN_PASSWORD=fixture-plain\nMYCELIS_LOCAL_ADMIN_PASSWORD_SHA256={_sha('something-else')}\n", encoding="utf-8")

    _run("sync")

    assert _admin(env_file) == ("fixture-plain", _sha("fixture-plain"))
    assert "fixture-plain" not in capsys.readouterr().out


def test_password_reusing_an_api_key_is_refused(env_file):
    env_file.write_text(WEB_SECRETS, encoding="utf-8")
    with pytest.raises(SystemExit):
        _run("fixture-api-key-value-0000000000")


def test_no_admin_flag_never_touches_the_local_admin(env_file):
    env_file.write_text(WEB_SECRETS + f"MYCELIS_LOCAL_ADMIN_PASSWORD=fixture-plain\nMYCELIS_LOCAL_ADMIN_PASSWORD_SHA256={_sha('something-else')}\n", encoding="utf-8")

    _run("")

    assert _admin(env_file) == ("fixture-plain", _sha("something-else"))


def test_posture_warns_when_plaintext_and_hash_disagree(env_file, capsys):
    env_file.write_text(WEB_SECRETS + f"MYCELIS_LOCAL_ADMIN_PASSWORD=fixture-plain\nMYCELIS_LOCAL_ADMIN_PASSWORD_SHA256={_sha('something-else')}\n", encoding="utf-8")

    auth.posture.body(Context(), compose=False)

    assert "does not match" in capsys.readouterr().out
