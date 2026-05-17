from __future__ import annotations

import pytest

from ops import mycelis_api
from ops.mycelis_api import APIClient, _team_work_payload, run_delivery_proof


class FakeClient(APIClient):
    def __init__(self):
        super().__init__("http://test-core")
        self.calls: list[tuple[str, str, dict | None]] = []
        self.work_item_id = "11111111-1111-4111-8111-111111111111"

    def request(self, method: str, path: str, payload: dict | None = None) -> dict:
        self.calls.append((method, path, payload))
        if path == "/api/v1/capabilities/refresh":
            return {"ok": True, "data": {"count": 2, "manifests": [{}, {}]}}
        if path == "/api/v1/capabilities":
            return {"ok": True, "data": {"count": 1, "manifests": [{}]}}
        if path == "/api/v1/system/deployments/trust":
            return {"ok": True, "data": {"runtime_health": {"status": "online"}}}
        if method == "POST" and path.endswith("/work"):
            return {"ok": True, "data": {"work_item_id": self.work_item_id}}
        if method == "POST" and path.endswith("/interactions"):
            return {"ok": True, "data": {"status": "created"}}
        if method == "GET" and path.endswith("/work?limit=10"):
            return {"ok": True, "data": [{"work_item_id": self.work_item_id}]}
        raise AssertionError(f"unexpected API call: {method} {path}")


def test_delivery_proof_uses_core_api_for_capabilities_deployment_and_team_work():
    client = FakeClient()

    result = run_delivery_proof(client, "delivery-team", "Build target delivery proof")

    assert result == {
        "capability_count": 2,
        "deployment_trust": True,
        "team_id": "delivery-team",
        "work_item_id": client.work_item_id,
    }
    assert client.calls[0][:2] == ("POST", "/api/v1/capabilities/refresh")
    assert client.calls[1][:2] == ("GET", "/api/v1/system/deployments/trust")
    assert client.calls[2][:2] == ("POST", "/api/v1/teams/delivery-team/work")
    assert client.calls[3][:2] == (
        "POST",
        f"/api/v1/teams/delivery-team/work/{client.work_item_id}/interactions",
    )
    assert client.calls[4][:2] == ("GET", "/api/v1/teams/delivery-team/work?limit=10")


def test_delivery_proof_read_only_does_not_create_team_work():
    client = FakeClient()

    result = run_delivery_proof(client, "delivery-team", "Observe only", mutate=False)

    assert result["capability_count"] == 1
    assert result["work_item_id"] == ""
    assert [call[:2] for call in client.calls] == [
        ("GET", "/api/v1/capabilities"),
        ("GET", "/api/v1/system/deployments/trust"),
    ]


def test_delivery_proof_fails_when_team_work_readback_is_missing():
    client = FakeClient()
    client.work_item_id = "22222222-2222-4222-8222-222222222222"

    def missing_readback(method: str, path: str, payload: dict | None = None) -> dict:
        if method == "GET" and path.endswith("/work?limit=10"):
            return {"ok": True, "data": []}
        return FakeClient.request(client, method, path, payload)

    client.request = missing_readback  # type: ignore[method-assign]

    with pytest.raises(SystemExit, match="readback"):
        run_delivery_proof(client, "delivery-team", "Broken readback")


def test_team_work_payload_is_delegated_target_delivery_work():
    payload = _team_work_payload("delivery-team", "Use Mycelis API")

    assert payload["team_id"] == "delivery-team"
    assert payload["execution_shape"] == "delegated_work"
    assert payload["state"] == "queued"
    assert "mycelis_api_self_use" in payload["scope"]


def test_api_client_defaults_read_repo_local_env(tmp_path, monkeypatch):
    (tmp_path / ".env").write_text(
        "MYCELIS_API_KEY=local-key\nMYCELIS_API_HOST=127.0.0.1\nMYCELIS_API_PORT=8081\n",
        encoding="utf-8",
    )
    monkeypatch.setattr(mycelis_api, "ROOT_DIR", tmp_path)
    monkeypatch.delenv("MYCELIS_API_KEY", raising=False)
    monkeypatch.delenv("MYCELIS_API_BASE_URL", raising=False)
    monkeypatch.delenv("MYCELIS_API_HOST", raising=False)
    monkeypatch.delenv("MYCELIS_API_PORT", raising=False)
    monkeypatch.delenv("PORT", raising=False)

    assert mycelis_api.default_base_url() == "http://127.0.0.1:8081"
    assert mycelis_api.api_key() == "local-key"
