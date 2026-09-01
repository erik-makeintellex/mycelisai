from __future__ import annotations

import pytest
from invoke import Context
from invoke.exceptions import Exit

from ops import cognitive


@pytest.mark.parametrize(
    "task_func",
    [
        cognitive.install.body,
        cognitive.llm.body,
        cognitive.media.body,
        cognitive.up.body,
        cognitive.status.body,
    ],
)
def test_optional_local_cognitive_tasks_fail_cleanly_on_windows(monkeypatch, task_func):
    monkeypatch.setattr(cognitive, "is_windows", lambda: True)

    with pytest.raises(Exit) as excinfo:
        task_func(Context())

    assert "not supported on Windows hosts" in str(excinfo.value)


def test_engine_secret_resolves_shell_reference(monkeypatch):
    monkeypatch.setenv("MYCELIS_TEXT_ENGINE_API_KEY", "local-test-secret")

    assert cognitive._resolve_engine_secret(
        {"api_key_secret_ref": "env:MYCELIS_TEXT_ENGINE_API_KEY"},
        "api_key_secret_ref",
    ) == "local-test-secret"


def test_engine_secret_falls_back_to_repo_dotenv(monkeypatch, tmp_path):
    monkeypatch.delenv("MYCELIS_TEXT_ENGINE_API_KEY", raising=False)
    monkeypatch.setattr(cognitive, "ROOT_DIR", tmp_path)
    (tmp_path / ".env").write_text(
        "MYCELIS_TEXT_ENGINE_API_KEY=dotenv-test-secret\n",
        encoding="utf-8",
    )

    assert cognitive._resolve_engine_secret(
        {"api_key_secret_ref": "env:MYCELIS_TEXT_ENGINE_API_KEY"},
        "api_key_secret_ref",
    ) == "dotenv-test-secret"


@pytest.mark.parametrize("reference", ["", "raw-secret", "env:bad-name"])
def test_engine_secret_rejects_unmanaged_reference(reference):
    with pytest.raises(Exit, match="must be an env:NAME secret reference"):
        cognitive._resolve_engine_secret(
            {"api_key_secret_ref": reference},
            "api_key_secret_ref",
        )


def test_status_routes_litellm_mode_without_local_engine_host_gate(monkeypatch):
    calls = []
    monkeypatch.setattr(cognitive, "is_windows", lambda: True)
    monkeypatch.setattr(
        cognitive.cognitive_litellm,
        "preflight",
        lambda *args: calls.append(args),
    )

    cognitive.status.body(
        Context(),
        litellm=True,
        litellm_endpoint="https://gateway.example.test/v1",
        litellm_api_key_env="LITELLM_PROXY_API_KEY",
        litellm_model="mycelis-default",
        timeout=7,
    )

    assert calls == [
        (
            "https://gateway.example.test/v1",
            "LITELLM_PROXY_API_KEY",
            "mycelis-default",
            7,
        )
    ]


def test_status_requires_explicit_litellm_mode_for_gateway_options():
    with pytest.raises(Exit, match="pass --litellm"):
        cognitive.status.body(
            Context(),
            litellm_endpoint="https://gateway.example.test/v1",
        )
