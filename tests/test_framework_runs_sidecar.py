from __future__ import annotations

import json
import asyncio
from uuid import UUID

import httpx
import pytest

from framework_runs.api import create_app
from framework_runs.drivers import ConformanceDriver
from framework_runs.store import InMemoryRunStore, RunStore


@pytest.fixture
def client() -> "ASGIClient":
    return ASGIClient(
        create_app(
            driver=ConformanceDriver(),
            store=InMemoryRunStore(max_runs=8, max_events_per_run=16),
        )
    )


def test_health_and_capabilities_disclose_non_production_boundary(
    client: "ASGIClient",
) -> None:
    health = client.get("/health")
    assert health.status_code == 200
    assert health.json() == {
        "healthy": True,
        "message": "framework runs facade ready",
        "driver": "conformance",
        "framework": "built_in",
        "production_ready": False,
        "storage": "bounded_memory_non_production",
    }

    capabilities = client.get("/v1/capabilities").json()
    assert capabilities["supported_protocols"] == ["runs_api"]
    assert capabilities["supports_events"] is True
    assert capabilities["supports_cancellation"] is True
    assert capabilities["supports_approvals"] is True
    assert capabilities["supports_usage"] is False
    assert capabilities["cancellation_contract"] == {
        "mode": "local_held_run",
        "synchronous_in_flight_preemption": False,
        "safe_point_only": True,
    }
    assert "completion_candidates" in capabilities["features"]
    assert capabilities["correlation_contract"]["legacy_omission"] == "synthesized_run_only_non_production"


def test_create_get_and_events_normalize_ids_and_completion_candidate(
    client: "ASGIClient",
) -> None:
    response = client.post(
        "/v1/runs",
        json={
            "org_id": "org-1",
            "project_id": "project-1",
            "user_id": "operator-1",
            "requested_by": "Soma",
            "intent": "Produce a bounded result",
            "required_protocols": ["runs_api"],
            "required_features": ["candidate_outputs"],
            "input": {
                "output_kind": "document",
                "output_name": "Result",
                "output_uri": "workspace://candidate/result.md",
                "output": {"body": "candidate"},
            },
            "metadata": {"intent_proof_id": "proof-1"},
            "ignored_go_field": "safe to ignore",
        },
    )
    assert response.status_code == 201
    run = response.json()
    UUID(run["run_id"])
    assert run["status"] == "completed"
    assert run["metadata"]["execution_authority"] == "mycelis_core"
    assert run["correlation"] == {"run_id": run["run_id"]}
    assert run["metadata"]["correlation_complete"] is False
    assert run["metadata"]["request_context"] == {
        "org_id": "org-1",
        "project_id": "project-1",
        "user_id": "operator-1",
        "requested_by": "Soma",
        "correlation_id": run["run_id"],
        "required_protocols": ["runs_api"],
        "required_features": ["candidate_outputs"],
    }
    assert run["result"]["metadata"] == {
        "resumed_after_approval": False,
        "completion_authority": "candidate",
        "requires_core_validation": True,
        "verified": False,
    }
    assert run["result"]["outputs"][0]["uri"] == "workspace://candidate/result.md"
    assert run["result"]["outputs"][0]["metadata"]["verified"] is False

    assert client.get(f"/v1/runs/{run['run_id']}").json() == run
    events_response = client.get(f"/v1/runs/{run['run_id']}/events")
    assert events_response.headers["content-type"].startswith("text/event-stream")
    events = _sse_events(events_response.text)
    assert [event["kind"] for event in events] == ["accepted", "completed"]
    assert all(event["run_id"] == run["run_id"] for event in events)
    assert events[-1]["metadata"]["completion_authority"] == "candidate"
    assert events[-1]["metadata"]["verified"] is False


def test_approval_lifecycle_is_central_and_resume_remains_candidate(
    client: "ASGIClient",
) -> None:
    created = client.post(
        "/v1/runs",
        json={
            "intent": "Wait for approval",
            "input": {
                "conformance_mode": "approval",
                "approval_summary": "Allow bounded operation",
            },
        },
    ).json()
    assert created["status"] == "approval_needed"
    approval_id = created["approval"]["id"]
    UUID(approval_id)

    mismatch = client.post(
        f"/v1/runs/{created['run_id']}/approvals/{approval_id}",
        json={"approval_id": str(UUID(int=0)), "decision": "approve"},
    )
    assert mismatch.status_code == 400

    approved = client.post(
        f"/v1/runs/{created['run_id']}/approvals/{approval_id}",
        json={
            "approval_id": approval_id,
            "decision": "approve",
            "actor_id": "operator-1",
        },
    )
    assert approved.status_code == 200
    run = approved.json()
    assert run["status"] == "completed"
    assert "approval" not in run
    assert run["result"]["metadata"]["resumed_after_approval"] is True
    assert run["result"]["metadata"]["completion_authority"] == "candidate"

    events = _sse_events(
        client.get(f"/v1/runs/{created['run_id']}/events").text
    )
    assert [event["kind"] for event in events] == [
        "accepted",
        "approval_needed",
        "completed",
    ]


def test_denied_approval_fails_without_framework_resume(client: "ASGIClient") -> None:
    created = client.post(
        "/v1/runs",
        json={"intent": "Deny operation", "input": {"conformance_mode": "approval"}},
    ).json()
    approval_id = created["approval"]["id"]
    denied = client.post(
        f"/v1/runs/{created['run_id']}/approvals/{approval_id}",
        json={
            "approval_id": approval_id,
            "decision": "deny",
            "actor_id": "operator-1",
            "reason": "outside approved scope",
        },
    )
    assert denied.status_code == 200
    run = denied.json()
    assert run["status"] == "failed"
    assert run["error"]["code"] == "approval_denied"
    assert run["error"]["metadata"]["reason"] == "outside approved scope"
    assert "result" not in run


def test_cancel_is_idempotent_and_cannot_regress_terminal_run(
    client: "ASGIClient",
) -> None:
    held = client.post(
        "/v1/runs",
        json={"intent": "Hold", "input": {"conformance_mode": "hold"}},
    ).json()
    assert held["status"] == "running"

    stopped = client.post(f"/v1/runs/{held['run_id']}/stop")
    assert stopped.status_code == 200
    assert stopped.json()["status"] == "cancelled"
    assert client.post(f"/v1/runs/{held['run_id']}/stop").status_code == 200
    assert [
        event["kind"]
        for event in _sse_events(
            client.get(f"/v1/runs/{held['run_id']}/events").text
        )
    ] == ["accepted", "progress", "cancelled"]

    completed = client.post("/v1/runs", json={"intent": "Complete"}).json()
    terminal_stop = client.post(f"/v1/runs/{completed['run_id']}/stop")
    assert terminal_stop.status_code == 409
    assert client.get(f"/v1/runs/{completed['run_id']}").json()["status"] == "completed"


def test_bounded_store_prunes_terminal_but_never_active_run() -> None:
    test_client = ASGIClient(
        create_app(
            driver=ConformanceDriver(),
            store=InMemoryRunStore(max_runs=1, max_events_per_run=8),
        )
    )
    terminal = test_client.post("/v1/runs", json={"intent": "Complete"}).json()
    replacement = test_client.post(
        "/v1/runs",
        json={"intent": "Hold", "input": {"conformance_mode": "hold"}},
    )
    assert replacement.status_code == 201
    assert test_client.get(f"/v1/runs/{terminal['run_id']}").status_code == 404
    at_capacity = test_client.post(
        "/v1/runs",
        json={"intent": "Second hold", "input": {"conformance_mode": "hold"}},
    )
    assert at_capacity.status_code == 503


def test_supplied_identity_is_stable_idempotent_and_conflict_safe() -> None:
    store = InMemoryRunStore(max_runs=4)
    assert isinstance(store, RunStore)
    test_client = ASGIClient(create_app(driver=ConformanceDriver(), store=store))
    payload = {
        "run_id": "core-run-001",
        "correlation_id": "mission-001",
        "correlation": {
            "run_id": "core-run-001",
            "intent_proof_id": "proof-001",
            "execution_contract_id": "contract-001",
            "team_id": "team-001",
            "outcome_id": "outcome-001",
            "work_item_id": "work-001",
            "idempotency_key": "confirm:proof-001",
            "source_kind": "web_api",
            "source_channel": "api.intent.confirm-action",
            "payload_kind": "command",
            "graph_revision": "worker-graph-v1",
        },
        "intent": "Produce one candidate",
        "input": {"output": {"answer": 42}},
    }
    first = test_client.post("/v1/runs", json=payload)
    repeated = test_client.post("/v1/runs", json=payload)
    assert first.status_code == repeated.status_code == 201
    assert repeated.json() == first.json()
    assert first.json()["run_id"] == "core-run-001"
    assert first.json()["correlation_id"] == "mission-001"
    assert first.json()["correlation"] == payload["correlation"]
    events = _sse_events(test_client.get("/v1/runs/core-run-001/events").text)
    assert all(event["correlation"] == payload["correlation"] for event in events)
    changed = {**payload, "correlation": {**payload["correlation"], "idempotency_key": "confirm:other"}}
    conflict = test_client.post("/v1/runs", json=changed)
    assert conflict.status_code == 409
    mismatch = {**payload, "correlation": {**payload["correlation"], "run_id": "other-run"}}
    assert test_client.post("/v1/runs", json=mismatch).status_code == 422
    incomplete = {**payload, "run_id": "core-run-002", "correlation": {"run_id": "core-run-002"}}
    assert test_client.post("/v1/runs", json=incomplete).status_code == 422
    assert test_client.post("/v1/runs", json={
        "run_id": "unsafe/run",
        "intent": "Rejected identity",
    }).status_code == 422


def test_driver_failure_does_not_expose_exception_message() -> None:
    class FailingDriver:
        name = "failing"
        framework = "test"
        production_ready = False

        def start(self, run_id: str, request: object) -> object:
            raise RuntimeError("provider-secret-must-not-leak")

        def resume_after_approval(self, run_id: str, request: object) -> object:
            raise RuntimeError("provider-secret-must-not-leak")

    test_client = ASGIClient(
        create_app(driver=FailingDriver(), store=InMemoryRunStore(max_runs=1))
    )
    response = test_client.post("/v1/runs", json={"intent": "Fail safely"})

    assert response.status_code == 201
    payload = response.json()
    assert payload["status"] == "failed"
    assert payload["error"]["message"] == "Framework driver failed."
    assert payload["error"]["metadata"]["exception_type"] == "RuntimeError"
    assert "provider-secret-must-not-leak" not in response.text


def _sse_events(body: str) -> list[dict[str, object]]:
    return [
        json.loads(line.removeprefix("data: "))
        for line in body.splitlines()
        if line.startswith("data: ")
    ]


class ASGIClient:
    """Small synchronous facade over HTTPX's non-blocking ASGI transport."""

    def __init__(self, app: object) -> None:
        self.app = app

    def get(self, path: str, **kwargs: object) -> httpx.Response:
        return self.request("GET", path, **kwargs)

    def post(self, path: str, **kwargs: object) -> httpx.Response:
        return self.request("POST", path, **kwargs)

    def request(self, method: str, path: str, **kwargs: object) -> httpx.Response:
        async def send() -> httpx.Response:
            async with httpx.AsyncClient(
                transport=httpx.ASGITransport(app=self.app),
                base_url="http://framework-runs.test",
            ) as client:
                return await client.request(method, path, **kwargs)

        return asyncio.run(send())
