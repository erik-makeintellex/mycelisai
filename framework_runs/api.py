from __future__ import annotations

import json
from dataclasses import asdict
from typing import AsyncIterator
from uuid import UUID, uuid4

from fastapi import FastAPI, HTTPException, Path, Request, status
from fastapi.responses import StreamingResponse

from .domain import (
    ApprovalDecisionRequest,
    DriverOutcome,
    RunCreateRequest,
    RunRecord,
    utc_now,
    wire_time,
)
from .drivers import ConformanceDriver, Driver
from .store import InMemoryRunStore, StoreCapacityError


TERMINAL_STATUSES = {"completed", "failed", "cancelled"}
COMPLETION_CANDIDATE_METADATA = {
    "completion_authority": "candidate",
    "requires_core_validation": True,
    "verified": False,
}


def create_app(
    *,
    driver: Driver | None = None,
    store: InMemoryRunStore | None = None,
) -> FastAPI:
    selected_driver = driver or ConformanceDriver()
    selected_store = store or InMemoryRunStore()

    app = FastAPI(
        title="Mycelis Framework Runs Facade",
        version="0.1.0",
    )
    app.state.driver = selected_driver
    app.state.store = selected_store

    @app.get("/health")
    async def health(request: Request) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        active_store: InMemoryRunStore = request.app.state.store
        return {
            "healthy": True,
            "message": "framework runs facade ready",
            "driver": active_driver.name,
            "framework": active_driver.framework,
            "production_ready": bool(
                active_driver.production_ready and active_store.production_ready
            ),
            "storage": "bounded_memory_non_production",
        }

    @app.get("/v1/capabilities")
    async def capabilities(request: Request) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        return {
            "healthy": True,
            "supported_protocols": ["runs_api"],
            "supports_events": True,
            "supports_cancellation": True,
            "supports_approvals": True,
            "supports_usage": False,
            "features": [
                "normalized_runs",
                "completion_candidates",
                "central_approval_authority",
                f"driver:{active_driver.name}",
                f"framework:{active_driver.framework}",
            ],
            "driver": active_driver.name,
            "framework": active_driver.framework,
            "production_ready": active_driver.production_ready,
        }

    @app.post("/v1/runs", status_code=status.HTTP_201_CREATED)
    async def create_run(
        payload: RunCreateRequest, request: Request
    ) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        active_store: InMemoryRunStore = request.app.state.store
        now = utc_now()
        record = RunRecord(
            run_id=str(uuid4()),
            org_id=payload.org_id,
            project_id=payload.project_id,
            user_id=payload.user_id,
            requested_by=payload.requested_by,
            intent=payload.intent,
            instructions=payload.instructions,
            input=payload.input,
            required_protocols=payload.required_protocols,
            required_features=payload.required_features,
            request_metadata=payload.metadata,
            driver_name=active_driver.name,
            status="accepted",
            created_at=now,
            updated_at=now,
        )
        _append_event(record, "accepted", "Run accepted by framework facade.")
        try:
            active_store.put(record)
        except StoreCapacityError as exc:
            raise HTTPException(status_code=503, detail=str(exc)) from exc

        try:
            outcome = active_driver.start(record.run_id, payload)
        except Exception as exc:  # driver failures are normalized at the facade
            outcome = _driver_exception_outcome(active_driver.name, exc)
        _apply_outcome(record, outcome)
        active_store.update(record)
        return record.wire()

    @app.get("/v1/runs/{run_id}")
    async def get_run(
        request: Request, run_id: str = Path(min_length=1, max_length=128)
    ) -> dict[str, object]:
        return _require_run(request.app.state.store, run_id).wire()

    @app.get("/v1/runs/{run_id}/events")
    async def events(
        request: Request, run_id: str = Path(min_length=1, max_length=128)
    ) -> StreamingResponse:
        record = _require_run(request.app.state.store, run_id)

        async def stream() -> AsyncIterator[str]:
            for sequence, event in enumerate(record.events, start=1):
                yield f"id: {sequence}\ndata: {json.dumps(event, separators=(',', ':'))}\n\n"

        return StreamingResponse(
            stream(),
            media_type="text/event-stream",
            headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
        )

    @app.post("/v1/runs/{run_id}/stop")
    async def stop_run(
        request: Request, run_id: str = Path(min_length=1, max_length=128)
    ) -> dict[str, object]:
        active_store: InMemoryRunStore = request.app.state.store
        record = _require_run(active_store, run_id)
        if record.status == "cancelled":
            return record.wire()
        if record.status in TERMINAL_STATUSES:
            raise HTTPException(
                status_code=409,
                detail=f"cannot cancel terminal run with status {record.status}",
            )
        record.status = "cancelled"
        record.approval = None
        record.updated_at = utc_now()
        _append_event(record, "cancelled", "Run cancelled by central authority.")
        active_store.update(record)
        return record.wire()

    @app.post("/v1/runs/{run_id}/approvals/{approval_id}")
    async def submit_approval(
        payload: ApprovalDecisionRequest,
        request: Request,
        run_id: str = Path(min_length=1, max_length=128),
        approval_id: str = Path(min_length=1, max_length=128),
    ) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        active_store: InMemoryRunStore = request.app.state.store
        record = _require_run(active_store, run_id)
        if payload.approval_id != approval_id:
            raise HTTPException(status_code=400, detail="approval id does not match route")
        if record.status != "approval_needed" or record.approval is None:
            raise HTTPException(status_code=409, detail="run has no pending approval")
        if record.approval["id"] != approval_id:
            raise HTTPException(status_code=404, detail="approval not found")

        record.approval = None
        record.updated_at = utc_now()
        if payload.decision == "deny":
            record.status = "failed"
            record.error = {
                "code": "approval_denied",
                "message": "Central operator denied framework execution.",
                "recoverable": True,
                "metadata": {
                    "actor_id": payload.actor_id,
                    "reason": payload.reason,
                },
            }
            _append_event(record, "failed", record.error["message"], error=record.error)
            active_store.update(record)
            return record.wire()

        run_request = RunCreateRequest(
            org_id=record.org_id,
            project_id=record.project_id,
            user_id=record.user_id,
            requested_by=record.requested_by,
            intent=record.intent,
            instructions=record.instructions,
            input=record.input,
            required_protocols=record.required_protocols,
            required_features=record.required_features,
            metadata=record.request_metadata,
        )
        try:
            outcome = active_driver.resume_after_approval(record.run_id, run_request)
        except Exception as exc:
            outcome = _driver_exception_outcome(active_driver.name, exc)
        _apply_outcome(record, outcome)
        active_store.update(record)
        return record.wire()

    return app


def _require_run(store: InMemoryRunStore, run_id: str) -> RunRecord:
    try:
        UUID(run_id)
    except ValueError as exc:
        raise HTTPException(status_code=404, detail="run not found") from exc
    record = store.get(run_id)
    if record is None:
        raise HTTPException(status_code=404, detail="run not found")
    return record


def _append_event(
    record: RunRecord,
    kind: str,
    message: str,
    **payload: object,
) -> None:
    event = {
        "event_id": str(uuid4()),
        "run_id": record.run_id,
        "kind": kind,
        "status": record.status,
        "message": message,
        "timestamp": wire_time(utc_now()),
        "metadata": {
            "driver": record.driver_name,
            "execution_authority": "mycelis_core",
        },
        **payload,
    }
    record.events.append(event)


def _apply_outcome(record: RunRecord, outcome: DriverOutcome) -> None:
    record.status = outcome.status
    record.updated_at = utc_now()
    if outcome.status == "approval_needed":
        if outcome.approval is None:
            record.status = "failed"
            record.error = {
                "code": "invalid_driver_approval",
                "message": "Driver requested approval without approval details.",
                "recoverable": False,
            }
            _append_event(record, "failed", record.error["message"], error=record.error)
            return
        record.approval = {"id": str(uuid4()), **asdict(outcome.approval)}
        _append_event(
            record,
            "approval_needed",
            outcome.message,
            approval=record.approval,
        )
        return
    if outcome.status == "completed":
        record.result = {
            "summary": outcome.message,
            "outputs": [
                {"id": str(uuid4()), **asdict(output)} for output in outcome.outputs
            ],
            "metadata": {**outcome.metadata, **COMPLETION_CANDIDATE_METADATA},
            "finished_at": wire_time(record.updated_at),
        }
        _append_event(
            record,
            "completed",
            outcome.message,
            result=record.result,
            metadata={
                "driver": record.driver_name,
                "execution_authority": "mycelis_core",
                **COMPLETION_CANDIDATE_METADATA,
            },
        )
        return
    if outcome.status == "failed":
        error = outcome.error
        record.error = (
            asdict(error)
            if error is not None
            else {
                "code": "driver_failed",
                "message": outcome.message,
                "recoverable": True,
                "metadata": {},
            }
        )
        _append_event(record, "failed", outcome.message, error=record.error)
        return
    _append_event(record, "progress", outcome.message, metadata={
        "driver": record.driver_name,
        "execution_authority": "mycelis_core",
        **outcome.metadata,
    })


def _driver_exception_outcome(driver_name: str, exc: Exception) -> DriverOutcome:
    from .domain import DriverError

    return DriverOutcome(
        status="failed",
        message="Framework driver failed.",
        error=DriverError(
            code="framework_driver_error",
            message="Framework driver failed.",
            recoverable=True,
            metadata={"driver": driver_name, "exception_type": exc.__class__.__name__},
        ),
    )


app = create_app()
