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


def _payload(
    run_id: str,
    intent: str,
    *,
    input: dict[str, object] | None = None,
    **updates: object,
) -> dict[str, object]:
    payload: dict[str, object] = {
        "run_id": run_id,
        "correlation": {
            "run_id": run_id,
            "intent_proof_id": f"proof:{run_id}",
            "execution_contract_id": f"contract:{run_id}",
            "team_id": "team-1",
            "outcome_id": "outcome-1",
            "work_item_id": f"work:{run_id}",
            "idempotency_key": f"dispatch:{run_id}",
            "source_kind": "web_api",
            "source_channel": "api.intent.confirm-action",
            "payload_kind": "command",
            "graph_revision": "worker-graph-v1",
        },
        "intent": intent,
        "input": input or {},
    }
    payload.update(updates)
    return payload

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
    assert capabilities["correlation_contract"]["production_required_fields"]

def test_create_get_and_events_normalize_ids_and_completion_candidate(
    client: "ASGIClient",
) -> None:
    response = client.post(
        "/v1/runs",
        json=_payload(
            "core-run-normalize",
            "Produce a bounded result",
            org_id="org-1",
            project_id="project-1",
            user_id="operator-1",
            requested_by="Soma",
            required_protocols=["runs_api"],
            required_features=["candidate_outputs"],
            input={
                "output_kind": "document",
                "output_name": "Result",
                "output_uri": "workspace://candidate/result.md",
                "output": {"body": "candidate"},
            },
            metadata={"purpose": "conformance"},
        ),
    )
    assert response.status_code == 201
    run = response.json()
    assert run["run_id"] == "core-run-normalize"
    assert run["status"] == "completed"
    assert run["metadata"]["execution_authority"] == "mycelis_core"
    assert run["correlation"] == _payload(
        "core-run-normalize", "unused"
    )["correlation"]
    assert run["metadata"]["request_context"] == {
        "org_id": "org-1",
        "project_id": "project-1",
        "user_id": "operator-1",
        "requested_by": "Soma",
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
        json=_payload(
            "core-run-approval",
            "Wait for approval",
            input={
                "conformance_mode": "approval",
                "approval_summary": "Allow bounded operation",
            },
        ),
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
        json=_payload(
            "core-run-deny", "Deny operation", input={"conformance_mode": "approval"}
        ),
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
        json=_payload("core-run-hold", "Hold", input={"conformance_mode": "hold"}),
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

    completed = client.post(
        "/v1/runs", json=_payload("core-run-complete", "Complete")
    ).json()
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
    terminal = test_client.post(
        "/v1/runs", json=_payload("core-run-terminal", "Complete")
    ).json()
    replacement = test_client.post(
        "/v1/runs",
        json=_payload(
            "core-run-replacement", "Hold", input={"conformance_mode": "hold"}
        ),
    )
    assert replacement.status_code == 201
    assert test_client.get(f"/v1/runs/{terminal['run_id']}").status_code == 404
    at_capacity = test_client.post(
        "/v1/runs",
        json=_payload(
            "core-run-second-hold", "Second hold", input={"conformance_mode": "hold"}
        ),
    )
    assert at_capacity.status_code == 503


def test_supplied_identity_is_stable_idempotent_and_conflict_safe() -> None:
    store = InMemoryRunStore(max_runs=4)
    assert isinstance(store, RunStore)
    test_client = ASGIClient(create_app(driver=ConformanceDriver(), store=store))
    payload = {
        "run_id": "core-run-001",
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
    unsafe = _payload("core-run-unsafe", "Rejected identity")
    unsafe["run_id"] = "unsafe/run"
    assert test_client.post("/v1/runs", json=unsafe).status_code == 422


def test_create_and_approval_payloads_forbid_unknown_fields(client: "ASGIClient") -> None:
    create_payload = _payload("core-run-extra", "Reject extra")
    create_payload["correlation_id"] = "legacy-id"
    assert client.post("/v1/runs", json=create_payload).status_code == 422
    assert client.post("/v1/runs", json=_payload("core-run-duplicate", "Reject duplicate", metadata={"run_id": "duplicate"})).status_code == 422
    assert client.post("/v1/runs", json=_payload("core-run-resume", "Reject internal control", metadata={"_framework_resume": {}})).status_code == 422
    assert client.post("/v1/runs", json=_payload("core-run-protocol", "Reject protocol", required_protocols=["responses_api"])).status_code == 422
    created = client.post(
        "/v1/runs",
        json=_payload(
            "core-run-approval-extra",
            "Reject approval extra",
            input={"conformance_mode": "approval"},
        ),
    ).json()
    approval_id = created["approval"]["id"]
    assert client.post(
        f"/v1/runs/{created['run_id']}/approvals/{approval_id}",
        json={
            "approval_id": approval_id,
            "decision": "approve",
            "unknown": True,
        },
    ).status_code == 422


def test_driver_failure_does_not_expose_exception_message() -> None:
    class FailingDriver:
        name = "failing"
        framework = "test"

        def start(self, run_id: str, request: object) -> object:
            raise RuntimeError("provider-secret-must-not-leak")

        def resume_after_approval(self, run_id: str, request: object) -> object:
            raise RuntimeError("provider-secret-must-not-leak")

    test_client = ASGIClient(
        create_app(driver=FailingDriver(), store=InMemoryRunStore(max_runs=1))
    )
    response = test_client.post(
        "/v1/runs", json=_payload("core-run-fail", "Fail safely")
    )

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
