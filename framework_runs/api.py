from __future__ import annotations

import json
from hashlib import sha256
from typing import AsyncIterator
from uuid import uuid4

from fastapi import FastAPI, HTTPException, Path, Request, status
from fastapi.responses import StreamingResponse

from .domain import (
    ApprovalDecisionRequest,
    EXTERNAL_ID_PATTERN,
    RunCreateRequest,
    RunRecord,
    utc_now,
)
from .drivers import ConformanceDriver, Driver
from .lifecycle import append_event, apply_outcome, driver_exception_outcome
from .store import (
    InMemoryRunStore,
    RunStore,
    StoreCapacityError,
    StoreConflictError,
)


TERMINAL_STATUSES = {"completed", "failed", "cancelled"}


def create_app(
    *,
    driver: Driver | None = None,
    store: RunStore | None = None,
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
        active_store: RunStore = request.app.state.store
        return {
            "healthy": True,
            "message": "framework runs facade ready",
            "driver": active_driver.name,
            "framework": active_driver.framework,
            "production_ready": bool(
                active_driver.production_ready and active_store.production_ready
            ),
            "storage": active_store.storage_kind,
        }

    @app.get("/v1/capabilities")
    async def capabilities(request: Request) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        active_store: RunStore = request.app.state.store
        supports_cancellation = bool(
            getattr(active_driver, "supports_cancellation", False)
        )
        cancellation_mode = str(
            getattr(active_driver, "cancellation_mode", "unsupported")
        )
        return {
            "healthy": True,
            "supported_protocols": ["runs_api"],
            "supports_events": True,
            "supports_cancellation": supports_cancellation,
            "supports_approvals": True,
            "supports_usage": False,
            "features": [
                "normalized_runs",
                "completion_candidates",
                "central_approval_authority",
                "externally_supplied_run_identity",
                "structured_correlation",
                "normalized_incremental_events",
                f"driver:{active_driver.name}",
                f"framework:{active_driver.framework}",
                f"storage:{active_store.storage_kind}",
            ],
            "driver": active_driver.name,
            "framework": active_driver.framework,
            "production_ready": bool(
                active_driver.production_ready and active_store.production_ready
            ),
            "correlation_contract": {
                "production_required_fields": [
                    "run_id",
                    "intent_proof_id",
                    "execution_contract_id",
                    "work_item_id",
                    "idempotency_key",
                    "source_kind",
                    "source_channel",
                    "payload_kind",
                    "graph_revision",
                ],
                "legacy_omission": "synthesized_run_only_non_production",
            },
            "cancellation_contract": {
                "mode": cancellation_mode,
                "synchronous_in_flight_preemption": False,
                "safe_point_only": True,
            },
        }

    @app.post("/v1/runs", status_code=status.HTTP_201_CREATED)
    async def create_run(
        payload: RunCreateRequest, request: Request
    ) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        active_store: RunStore = request.app.state.store
        run_id = payload.run_id or str(uuid4())
        correlation_id = payload.correlation_id or run_id
        correlation = payload.correlation.model_dump(mode="json") if payload.correlation else {
            "run_id": run_id,
        }
        correlation_complete = bool(payload.correlation and payload.correlation.complete)
        fingerprint = _request_fingerprint(
            payload, active_driver.name, correlation_id, correlation
        )
        existing = active_store.get(run_id)
        if existing is not None:
            if existing.request_fingerprint == fingerprint:
                return existing.wire()
            raise HTTPException(status_code=409, detail="run id already exists")
        now = utc_now()
        record = RunRecord(
            run_id=run_id,
            correlation_id=correlation_id,
            correlation=correlation,
            correlation_complete=correlation_complete,
            request_fingerprint=fingerprint,
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
            storage_kind=active_store.storage_kind,
            status="accepted",
            created_at=now,
            updated_at=now,
        )
        append_event(record, "accepted", "Run accepted by framework facade.")
        try:
            active_store.put(record)
        except StoreCapacityError as exc:
            raise HTTPException(status_code=503, detail=str(exc)) from exc
        except StoreConflictError as exc:
            concurrent = active_store.get(run_id)
            if concurrent and concurrent.request_fingerprint == fingerprint:
                return concurrent.wire()
            raise HTTPException(status_code=409, detail="run id already exists") from exc

        try:
            outcome = active_driver.start(record.run_id, _run_request(record))
        except Exception as exc:  # driver failures are normalized at the facade
            outcome = driver_exception_outcome(active_driver.name, exc)
        apply_outcome(record, outcome)
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
            for event in record.events:
                yield f"id: {event['sequence']}\ndata: {json.dumps(event, separators=(',', ':'))}\n\n"

        return StreamingResponse(
            stream(),
            media_type="text/event-stream",
            headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
        )

    @app.post("/v1/runs/{run_id}/stop")
    async def stop_run(
        request: Request, run_id: str = Path(min_length=1, max_length=128)
    ) -> dict[str, object]:
        active_store: RunStore = request.app.state.store
        record = _require_run(active_store, run_id)
        if record.status == "cancelled":
            return record.wire()
        if record.status in TERMINAL_STATUSES:
            raise HTTPException(
                status_code=409,
                detail=f"cannot cancel terminal run with status {record.status}",
            )
        active_driver = request.app.state.driver
        cancel_hook = getattr(active_driver, "cancel", None)
        if not bool(getattr(active_driver, "supports_cancellation", False)) or not callable(cancel_hook):
            raise HTTPException(
                status_code=409,
                detail="framework driver does not support safe cancellation",
            )
        try:
            cancel_hook(record.run_id, _run_request(record))
        except Exception as exc:
            raise HTTPException(
                status_code=502,
                detail="framework cancellation failed",
            ) from exc
        record.status = "cancelled"
        record.approval = None
        record.updated_at = utc_now()
        append_event(record, "cancelled", "Run cancelled by central authority.")
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
        active_store: RunStore = request.app.state.store
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
            append_event(record, "failed", record.error["message"], error=record.error)
            active_store.update(record)
            return record.wire()

        run_request = _run_request(record, approval=payload)
        try:
            outcome = active_driver.resume_after_approval(record.run_id, run_request)
        except Exception as exc:
            outcome = driver_exception_outcome(active_driver.name, exc)
        apply_outcome(record, outcome)
        active_store.update(record)
        return record.wire()

    return app


def _require_run(store: RunStore, run_id: str) -> RunRecord:
    if EXTERNAL_ID_PATTERN.fullmatch(run_id) is None:
        raise HTTPException(status_code=404, detail="run not found")
    record = store.get(run_id)
    if record is None:
        raise HTTPException(status_code=404, detail="run not found")
    return record


def _request_fingerprint(
    payload: RunCreateRequest,
    driver_name: str,
    correlation_id: str,
    correlation: dict[str, str],
) -> str:
    normalized = payload.model_dump(mode="json")
    normalized["correlation_id"] = correlation_id
    normalized["correlation"] = correlation
    normalized["driver"] = driver_name
    return sha256(json.dumps(normalized, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def _run_request(
    record: RunRecord, approval: ApprovalDecisionRequest | None = None
) -> RunCreateRequest:
    metadata = dict(record.request_metadata)
    if approval is not None:
        metadata["_framework_resume"] = approval.model_dump(mode="json")
    values: dict[str, object] = {
        "run_id": record.run_id,
        "correlation_id": record.correlation_id,
        "org_id": record.org_id,
        "project_id": record.project_id,
        "user_id": record.user_id,
        "requested_by": record.requested_by,
        "intent": record.intent,
        "instructions": record.instructions,
        "input": record.input,
        "required_protocols": record.required_protocols,
        "required_features": record.required_features,
        "metadata": metadata,
    }
    if record.correlation_complete:
        values["correlation"] = record.correlation
    return RunCreateRequest(**values)


app = create_app()
