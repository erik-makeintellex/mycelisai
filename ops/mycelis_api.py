from __future__ import annotations

import json
import os
from dataclasses import dataclass
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

from invoke import Collection, task

from .compose_env import clean_env_value, read_env_file
from .config import API_HOST, API_PORT, ROOT_DIR


@dataclass
class APIClient:
    base_url: str
    api_key: str = ""
    timeout: float = 10

    def request(self, method: str, path: str, payload: dict[str, Any] | None = None) -> dict[str, Any]:
        body = None
        headers = {"Accept": "application/json"}
        if payload is not None:
            body = json.dumps(payload).encode("utf-8")
            headers["Content-Type"] = "application/json"
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"

        req = Request(
            self.base_url.rstrip("/") + path,
            data=body,
            headers=headers,
            method=method.upper(),
        )
        try:
            with urlopen(req, timeout=self.timeout) as resp:
                text = resp.read().decode("utf-8")
        except HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            raise SystemExit(f"Mycelis API {method} {path} failed: HTTP {exc.code} {detail}") from exc
        except URLError as exc:
            raise SystemExit(f"Mycelis API unavailable at {self.base_url}: {exc.reason}") from exc

        try:
            data = json.loads(text) if text else {}
        except json.JSONDecodeError as exc:
            raise SystemExit(f"Mycelis API {method} {path} returned non-JSON response") from exc
        if isinstance(data, dict) and data.get("ok") is False:
            raise SystemExit(f"Mycelis API {method} {path} returned error: {data.get('error')}")
        return data


def default_base_url() -> str:
    env_file = read_env_file(ROOT_DIR / ".env")
    base_url = os.environ.get("MYCELIS_API_BASE_URL") or env_file.get("MYCELIS_API_BASE_URL")
    if base_url:
        return clean_env_value(base_url)
    host = os.environ.get("MYCELIS_API_HOST") or env_file.get("MYCELIS_API_HOST") or API_HOST
    port = (
        os.environ.get("MYCELIS_API_PORT")
        or os.environ.get("PORT")
        or env_file.get("MYCELIS_API_PORT")
        or env_file.get("PORT")
        or str(API_PORT)
    )
    return f"http://{clean_env_value(host)}:{clean_env_value(str(port))}"


def api_key() -> str:
    return clean_env_value(
        os.environ.get("MYCELIS_API_KEY") or read_env_file(ROOT_DIR / ".env").get("MYCELIS_API_KEY", "")
    )


def _data(resp: dict[str, Any]) -> Any:
    return resp.get("data") if isinstance(resp, dict) and "data" in resp else resp


def _count_manifests(resp: dict[str, Any]) -> int:
    data = _data(resp)
    if isinstance(data, dict):
        manifests = data.get("manifests")
        if isinstance(manifests, list):
            return len(manifests)
        count = data.get("count")
        if isinstance(count, int):
            return count
    return 0


def _team_work_payload(team_id: str, objective: str) -> dict[str, Any]:
    return {
        "team_id": team_id,
        "objective": objective,
        "scope": ["target_delivery", "mycelis_api_self_use"],
        "owner": "Soma delivery proof",
        "execution_shape": "delegated_work",
        "state": "queued",
        "expected_outputs": ["Durable TeamWorkItem visible through Mycelis API"],
        "expected_proof": ["Capability refresh", "Deployment trust read", "Team work readback"],
        "capability_requirements": ["capability_manifest", "team_work_api", "deployment_trust_api"],
        "governance_posture": "optional",
        "recovery_options": ["Run api.delivery-proof again after restoring Core, PostgreSQL, and migrations."],
        "version": "v1",
    }


def run_delivery_proof(client: APIClient, team_id: str, objective: str, mutate: bool = True) -> dict[str, Any]:
    capabilities = client.request(
        "POST" if mutate else "GET",
        "/api/v1/capabilities/refresh" if mutate else "/api/v1/capabilities",
    )
    deployments = client.request("GET", "/api/v1/system/deployments/trust")
    result = {
        "capability_count": _count_manifests(capabilities),
        "deployment_trust": bool(_data(deployments)),
        "team_id": team_id,
        "work_item_id": "",
    }
    if not mutate:
        return result

    created = client.request("POST", f"/api/v1/teams/{team_id}/work", _team_work_payload(team_id, objective))
    work_item = _data(created)
    if not isinstance(work_item, dict) or not work_item.get("work_item_id"):
        raise SystemExit("Mycelis API did not return a durable work_item_id")
    work_item_id = str(work_item["work_item_id"])
    result["work_item_id"] = work_item_id

    client.request(
        "POST",
        f"/api/v1/teams/{team_id}/work/{work_item_id}/interactions",
        {
            "team_id": team_id,
            "work_item_id": work_item_id,
            "source_kind": "internal_tool",
            "source_channel": "uv.inv.api.delivery-proof",
            "actor_ref": "Codex",
            "verb": "record_delivery_proof",
            "summary": "Mycelis used its own API to create and read a delivery work item.",
            "payload_kind": "proof",
            "payload": {"capability_count": result["capability_count"]},
            "version": "v1",
        },
    )
    readback = client.request("GET", f"/api/v1/teams/{team_id}/work?limit=10")
    items = _data(readback)
    if not isinstance(items, list) or not any(
        item.get("work_item_id") == work_item_id for item in items if isinstance(item, dict)
    ):
        raise SystemExit("Mycelis API team-work readback did not include the created work item")
    return result


@task(
    name="delivery-proof",
    help={
        "base_url": "Core API base URL. Defaults to MYCELIS_API_BASE_URL or MYCELIS_API_HOST/MYCELIS_API_PORT.",
        "team_id": "Team id to use for the bounded API delivery proof.",
        "objective": "Objective stored in the created TeamWorkItem.",
        "read_only": "Only read capabilities/deployment trust; do not create team work.",
    },
)
def delivery_proof(c, base_url="", team_id="target-delivery-api-proof", objective="", read_only=False):
    """Use the Mycelis API itself as a delivery proof lane."""
    del c
    objective = objective or "Use the Mycelis API as a target-delivery work lane."
    client = APIClient(base_url=base_url or default_base_url(), api_key=api_key())
    result = run_delivery_proof(client, team_id=team_id, objective=objective, mutate=not read_only)
    print("=== MYCELIS API DELIVERY PROOF ===")
    print(f"Core API: {client.base_url}")
    print(f"Capabilities visible: {result['capability_count']}")
    print(f"Deployment trust readable: {result['deployment_trust']}")
    if result["work_item_id"]:
        print(f"Team work item: {result['team_id']} / {result['work_item_id']}")
    else:
        print("Read-only mode: no TeamWorkItem created.")


ns = Collection("api")
ns.add_task(delivery_proof)
