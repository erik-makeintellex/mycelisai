from __future__ import annotations

import asyncio
import json
import sys
from types import ModuleType, SimpleNamespace

import httpx
import pytest

from framework_runs.api import create_app
from framework_runs.domain import DriverEvent, DriverOutcome
from framework_runs.drivers import DriverDependencyError, LangGraphDriver
from framework_runs.store import InMemoryRunStore


class FakeCommand:
    def __init__(self, *, resume: object) -> None:
        self.resume = resume


class FakeGraph:
    def __init__(self) -> None:
        self.calls: list[tuple[object, dict[str, object]]] = []

    def invoke(self, graph_input: object, *, config: object) -> object:
        raise AssertionError("stream should be preferred")

    def stream(self, graph_input: object, **kwargs: object):
        self.calls.append((graph_input, kwargs))
        if isinstance(graph_input, FakeCommand):
            yield {"review": {"approved": True}}
            return
        yield {"draft": {"step": 1}}
        yield {
            "__interrupt__": (
                SimpleNamespace(value="Approve publish", id="int-1"),
            )
        }

    def get_state(self, config: object) -> object:
        return SimpleNamespace(values={"candidate": "ready"})


def _install_fake_langgraph(monkeypatch: pytest.MonkeyPatch) -> None:
    fake_package = ModuleType("langgraph")
    fake_types = ModuleType("langgraph.types")
    fake_types.Command = FakeCommand  # type: ignore[attr-defined]
    monkeypatch.setitem(sys.modules, "langgraph", fake_package)
    monkeypatch.setitem(sys.modules, "langgraph.types", fake_types)


def _payload(run_id: str) -> dict[str, object]:
    return {
        "run_id": run_id,
        "correlation": {
            "run_id": run_id,
            "intent_proof_id": "proof-langgraph-1",
            "execution_contract_id": "contract-langgraph-1",
            "team_id": "team-langgraph-1",
            "outcome_id": "outcome-langgraph-1",
            "work_item_id": "work-langgraph-1",
            "idempotency_key": f"dispatch:{run_id}",
            "source_kind": "web_api",
            "source_channel": "api.intent.confirm-action",
            "payload_kind": "command",
            "graph_revision": "worker-graph-v1",
        },
        "intent": "Run injected graph",
        "input": {"graph_input": {"topic": "bounded"}},
    }


def test_langgraph_stream_interrupt_resume_and_cancel_hooks(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    _install_fake_langgraph(monkeypatch)
    graph = FakeGraph()
    cancellations: list[tuple[str, dict[str, object]]] = []
    driver = LangGraphDriver(
        graph=graph,
        cancel_hook=lambda run_id, config: cancellations.append((run_id, config)),
    )
    client = ASGIClient(
        create_app(driver=driver, store=InMemoryRunStore(max_runs=4))
    )
    payload = _payload("core-langgraph-1")

    created = client.post("/v1/runs", json=payload).json()
    assert created["status"] == "approval_needed"
    assert created["approval"]["summary"] == "Approve publish"

    approval_id = created["approval"]["id"]
    resumed = client.post(
        f"/v1/runs/core-langgraph-1/approvals/{approval_id}",
        json={
            "approval_id": approval_id,
            "decision": "approve",
            "command_id": "approve-langgraph-1",
            "expected_version": created["version"],
            "actor_id": "operator-1",
        },
    )
    assert resumed.status_code == 202
    resumed = client.get("/v1/runs/core-langgraph-1").json()
    assert resumed["status"] == "completed"
    assert resumed["result"]["outputs"][0]["metadata"]["state"] == {
        "candidate": "ready"
    }
    events = _sse_events(client.get("/v1/runs/core-langgraph-1/events").text)
    assert [event["kind"] for event in events] == [
        "accepted",
        "progress",
        "approval_needed",
        "progress",
        "completed",
    ]
    assert graph.calls[0][1]["config"] == {
        "configurable": {
            "thread_id": "core-langgraph-1",
            "mycelis_run_id": "core-langgraph-1",
            **payload["correlation"],
        }
    }
    assert isinstance(graph.calls[1][0], FakeCommand)

    held = client.post("/v1/runs", json=_payload("core-langgraph-2")).json()
    stopped = client.post(f"/v1/runs/{held['run_id']}/stop", json={
        "command_id": "stop-langgraph-2", "expected_version": held["version"],
        "actor_id": "operator-1",
    })
    assert stopped.status_code == 202
    assert client.get(f"/v1/runs/{held['run_id']}").json()["status"] == "cancelled"
    assert cancellations[0][0] == "core-langgraph-2"


def test_langgraph_without_cancel_hook_fails_closed(monkeypatch: pytest.MonkeyPatch) -> None:
    _install_fake_langgraph(monkeypatch)
    client = ASGIClient(create_app(
        driver=LangGraphDriver(graph=FakeGraph()),
        store=InMemoryRunStore(max_runs=2),
    ))
    capabilities = client.get("/v1/capabilities").json()
    assert capabilities["supports_cancellation"] is False
    assert capabilities["cancellation_contract"]["mode"] == "unsupported"
    assert capabilities["cancellation_contract"]["synchronous_in_flight_preemption"] is False

    created = client.post("/v1/runs", json=_payload("core-no-cancel")).json()
    events_before = client.get("/v1/runs/core-no-cancel/events").text
    stopped = client.post("/v1/runs/core-no-cancel/stop", json={
        "command_id": "stop-no-cancel", "expected_version": created["version"],
        "actor_id": "operator-1",
    })
    assert stopped.status_code == 409
    assert stopped.json()["error"]["code"] == "unsupported_control"
    assert client.get("/v1/runs/core-no-cancel").json() == created
    assert client.get("/v1/runs/core-no-cancel/events").text == events_before


def test_langgraph_failed_cancel_hook_preserves_state_and_redacts_error(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    _install_fake_langgraph(monkeypatch)

    def fail_cancel(run_id: str, config: dict[str, object]) -> None:
        raise RuntimeError("cancel-hook-secret")

    client = ASGIClient(create_app(
        driver=LangGraphDriver(graph=FakeGraph(), cancel_hook=fail_cancel),
        store=InMemoryRunStore(max_runs=2),
    ))
    created = client.post("/v1/runs", json=_payload("core-failed-cancel")).json()
    events_before = client.get("/v1/runs/core-failed-cancel/events").text
    stopped = client.post("/v1/runs/core-failed-cancel/stop", json={
        "command_id": "stop-failed-cancel", "expected_version": created["version"],
        "actor_id": "operator-1",
    })
    assert stopped.status_code == 502
    assert stopped.json()["error"]["code"] == "control_failed"
    assert "cancel-hook-secret" not in stopped.text
    assert client.get("/v1/runs/core-failed-cancel").json() == created
    assert client.get("/v1/runs/core-failed-cancel/events").text == events_before


def test_langgraph_adapter_is_dependency_gated_or_requires_real_graph() -> None:
    try:
        import langgraph  # noqa: F401
    except ImportError:
        with pytest.raises(
            DriverDependencyError, match="optional 'langgraph' dependency"
        ):
            LangGraphDriver(graph=object())
    else:
        with pytest.raises(TypeError, match="compiled graph with invoke"):
            LangGraphDriver(graph=object())


def test_driver_event_metadata_cannot_replace_facade_authority() -> None:
    class HostileMetadataDriver:
        name = "hostile-metadata"
        framework = "test"

        def start(self, _run_id: str, _request: object) -> DriverOutcome:
            event = DriverEvent("Untrusted event.", {"driver": "forged", "execution_authority": "adapter"})
            return DriverOutcome(status="running", message="Still running.", events=(event,))

    client = ASGIClient(create_app(driver=HostileMetadataDriver(), store=InMemoryRunStore(max_runs=1)))
    assert client.post("/v1/runs", json=_payload("core-hostile")).status_code == 201
    retained = _sse_events(client.get("/v1/runs/core-hostile/events").text)[1]
    assert retained["metadata"]["driver"] == "hostile-metadata"
    assert retained["metadata"]["execution_authority"] == "mycelis_core"


def _sse_events(body: str) -> list[dict[str, object]]:
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
