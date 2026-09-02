from __future__ import annotations

import asyncio
import json
from pathlib import Path

import httpx
import pytest

from framework_runs.api import create_app
from framework_runs.domain import ApprovalDecisionRequest, RunCreateRequest, StopRequest
from framework_runs.drivers import ConformanceDriver
from framework_runs.store import InMemoryRunStore


CONTRACT_ROOT = Path(__file__).parents[1] / "contracts" / "framework-runs" / "v1"


def fixture(name: str) -> dict[str, object]:
    return json.loads((CONTRACT_ROOT / name).read_text(encoding="utf-8"))


def payload(run_id: str, *, mode: str = "complete") -> dict[str, object]:
    body = fixture("create_request.json")
    correlation = dict(body["correlation"])  # type: ignore[arg-type]
    correlation.update({
        "run_id": run_id,
        "intent_proof_id": f"proof:{run_id}",
        "execution_contract_id": f"contract:{run_id}",
        "work_item_id": f"work:{run_id}",
        "idempotency_key": f"dispatch:{run_id}",
    })
    body.update({
        "run_id": run_id,
        "correlation": correlation,
        "input": {"conformance_mode": mode, "output": {"answer": 42}},
    })
    return body


@pytest.fixture
def client() -> "ASGIClient":
    return ASGIClient(create_app(
        driver=ConformanceDriver(),
        store=InMemoryRunStore(max_runs=12, max_events_per_run=16),
    ))


def test_golden_documents_are_exact_and_model_valid() -> None:
    create = fixture("create_request.json")
    assert RunCreateRequest.model_validate(create).model_dump(mode="json") == create
    assert StopRequest.model_validate(fixture("stop_request.json"))
    assert ApprovalDecisionRequest.model_validate(fixture("approval_request.json"))

    run = fixture("run_snapshot.json")
    assert set(run) == {
        "run_id", "correlation", "status", "version", "created_at",
        "updated_at", "result", "metadata",
    }
    event = fixture("event.json")
    assert set(event) == {
        "event_id", "sequence", "version", "run_id", "correlation", "kind", "status",
        "message", "timestamp", "result", "metadata",
    }
    receipt = fixture("control_receipt.json")
    assert set(receipt) == {
        "command_id", "run_id", "kind", "state", "version", "created_at",
        "updated_at",
    }
    assert set(fixture("error.json")) == {"error"}


def test_run_version_event_sequence_correlation_and_candidate_manifest(
    client: "ASGIClient",
) -> None:
    response = client.post("/v1/runs", json=payload("core-contract-run"))
    assert response.status_code == 201
    run = response.json()
    assert run["version"] == 2
    output = run["result"]["outputs"][0]
    assert set(output) == {
        "id", "kind", "name", "uri", "content_type", "size_bytes",
        "sha256", "metadata",
    }
    assert output["uri"] == f"candidate://{run['run_id']}/{output['id']}"
    assert output["size_bytes"] > 0
    assert len(output["sha256"]) == 64
    assert set(output["sha256"]) <= set("0123456789abcdef")
    assert output["metadata"]["completion_authority"] == "candidate"
    assert output["metadata"]["requires_core_validation"] is True
    assert output["metadata"]["verified"] is False

    events = sse_events(client.get(f"/v1/runs/{run['run_id']}/events").text)
    assert [
        (event["sequence"], event["version"], event["status"], event["kind"])
        for event in events
    ] == [
        (1, 1, "accepted", "accepted"),
        (2, 2, "completed", "completed"),
    ]
    assert all(event["run_id"] == run["run_id"] for event in events)
    assert all(event["correlation"] == run["correlation"] for event in events)


def test_create_is_identity_stable_and_conflicting_reuse_fails(
    client: "ASGIClient",
) -> None:
    body = payload("core-idempotent-run")
    first = client.post("/v1/runs", json=body)
    replay = client.post("/v1/runs", json=body)
    assert first.status_code == replay.status_code == 201
    assert replay.json() == first.json()
    changed = dict(body)
    changed["intent"] = "Different work under reused identity"
    assert_error(client.post("/v1/runs", json=changed), 409, "run_conflict")


def test_last_event_id_replays_strict_decimal_sequence(client: "ASGIClient") -> None:
    run = client.post("/v1/runs", json=payload("core-cursor-run")).json()
    replay = client.get(
        f"/v1/runs/{run['run_id']}/events", headers={"Last-Event-ID": "1"}
    )
    assert replay.status_code == 200
    assert [event["sequence"] for event in sse_events(replay.text)] == [2]
    assert replay.text.startswith("id: 2\n")

    malformed = client.get(
        f"/v1/runs/{run['run_id']}/events", headers={"Last-Event-ID": "1.0"}
    )
    assert_error(malformed, 422, "invalid_cursor")
    ahead = client.get(
        f"/v1/runs/{run['run_id']}/events", headers={"Last-Event-ID": "3"}
    )
    assert_error(ahead, 409, "cursor_gap")


def test_pruned_cursor_fails_gap_instead_of_silent_partial_replay() -> None:
    client = ASGIClient(create_app(
        driver=ConformanceDriver(),
        store=InMemoryRunStore(max_runs=2, max_events_per_run=1),
    ))
    run = client.post("/v1/runs", json=payload("core-pruned-run")).json()
    response = client.get(
        f"/v1/runs/{run['run_id']}/events", headers={"Last-Event-ID": "0"}
    )
    assert_error(response, 409, "cursor_gap")


def test_stop_command_is_cas_guarded_and_exactly_replayable(
    client: "ASGIClient",
) -> None:
    run = client.post("/v1/runs", json=payload("core-stop-run", mode="hold")).json()
    command = fixture("stop_request.json")
    command.update({"command_id": "stop-core-run", "expected_version": run["version"]})
    first = client.post(f"/v1/runs/{run['run_id']}/stop", json=command)
    assert first.status_code == 202
    assert first.json()["state"] == "applied"
    assert first.json()["version"] == run["version"] + 1
    replay = client.post(f"/v1/runs/{run['run_id']}/stop", json=command)
    assert replay.status_code == 200 and replay.json() == first.json()

    changed = {**command, "reason": "different"}
    assert_error(
        client.post(f"/v1/runs/{run['run_id']}/stop", json=changed),
        409,
        "command_conflict",
    )


def test_control_version_and_approval_identity_fail_closed(client: "ASGIClient") -> None:
    held = client.post("/v1/runs", json=payload("core-stale-run", mode="hold")).json()
    stale = fixture("stop_request.json")
    stale.update({"command_id": "stop-stale", "expected_version": 1})
    response = client.post(f"/v1/runs/{held['run_id']}/stop", json=stale)
    assert response.json() == fixture("error.json")
    assert response.status_code == 409

    waiting = client.post(
        "/v1/runs", json=payload("core-approval-run", mode="approval")
    ).json()
    approval = fixture("approval_request.json")
    approval.update({
        "approval_id": waiting["approval"]["id"],
        "command_id": "approve-core-run",
        "expected_version": waiting["version"],
    })
    wrong = {**approval, "approval_id": "other-approval"}
    assert_error(
        client.post(
            f"/v1/runs/{waiting['run_id']}/approvals/{waiting['approval']['id']}",
            json=wrong,
        ),
        409,
        "approval_mismatch",
    )
    accepted = client.post(
        f"/v1/runs/{waiting['run_id']}/approvals/{waiting['approval']['id']}",
        json=approval,
    )
    assert accepted.status_code == 202 and accepted.json()["kind"] == "approve"
    replay = client.post(
        f"/v1/runs/{waiting['run_id']}/approvals/{waiting['approval']['id']}",
        json=approval,
    )
    assert replay.status_code == 200 and replay.json() == accepted.json()


@pytest.mark.parametrize("mutation", [
    {"legacy_id": "run-1"},
    {"correlation_id": "legacy"},
    {"unknown": True},
])
def test_unknown_and_legacy_create_fields_use_exact_error_envelope(
    client: "ASGIClient", mutation: dict[str, object]
) -> None:
    body = payload("core-invalid-run")
    body.update(mutation)
    assert_error(client.post("/v1/runs", json=body), 422, "invalid_request")


@pytest.mark.parametrize(("field", "value"), [
    ("run_id", " core-whitespace-run"),
    ("intent", "Produce candidate "),
])
def test_create_rejects_noncanonical_surrounding_whitespace(
    client: "ASGIClient", field: str, value: str,
) -> None:
    body = payload("core-whitespace-run")
    body[field] = value
    assert_error(client.post("/v1/runs", json=body), 422, "invalid_request")


def test_controls_reject_blank_or_noncanonical_actor() -> None:
    with pytest.raises(ValueError):
        StopRequest.model_validate({**fixture("stop_request.json"), "actor_id": "   "})
    with pytest.raises(ValueError):
        ApprovalDecisionRequest.model_validate({
            **fixture("approval_request.json"), "actor_id": " operator-golden",
        })


def assert_error(response: httpx.Response, status: int, code: str) -> None:
    assert response.status_code == status
    assert set(response.json()) == {"error"}
    assert set(response.json()["error"]) == {"code", "message", "recoverable"}
    assert response.json()["error"]["code"] == code


def sse_events(body: str) -> list[dict[str, object]]:
    return [
        json.loads(line.removeprefix("data: "))
        for line in body.splitlines()
        if line.startswith("data: ")
    ]


class ASGIClient:
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
