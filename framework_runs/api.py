from __future__ import annotations

import json
from typing import AsyncIterator

from fastapi import FastAPI, HTTPException, Path, Request, status
from fastapi.responses import JSONResponse, StreamingResponse

from .domain import (
    ApprovalDecisionRequest,
    EXTERNAL_ID_PATTERN,
    RunCreateRequest,
    RunRecord,
    StopRequest,
    utc_now,
)
from .drivers import ConformanceDriver, Driver
from .lifecycle import append_event, apply_outcome, driver_exception_outcome
from .http_contract import fail, install_error_handlers, parse_last_event_id, request_fingerprint
from .store import (
    InMemoryRunStore,
    RunStore,
    StoreCapacityError,
    StoreConflictError,
    StoreVersionError,
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
    install_error_handlers(app)

    @app.get("/health")
    async def health(request: Request) -> dict[str, object]:
        active_driver: Driver = request.app.state.driver
        return {
            "healthy": True,
            "message": "framework runs facade ready",
            "driver": active_driver.name,
            "framework": active_driver.framework,
            "production_ready": False,
            "storage": request.app.state.store.storage_kind,
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
            "production_ready": False,
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
        run_id = payload.run_id
        correlation = payload.correlation.model_dump(mode="json")
        fingerprint = request_fingerprint(payload, discriminator=active_driver.name)
        existing = active_store.get(run_id)
        if existing is not None:
            if existing.request_fingerprint == fingerprint:
                return existing.wire()
            fail(409, "run_conflict", "Run id already exists with different content.")
        now = utc_now()
        record = RunRecord(
            run_id=run_id,
            correlation=correlation,
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
            fail(503, "capacity_exhausted", "All bounded run slots are active.", True)
        except StoreConflictError as exc:
            concurrent = active_store.get(run_id)
            if concurrent and concurrent.request_fingerprint == fingerprint:
                return concurrent.wire()
            fail(409, "run_conflict", "Run id already exists with different content.")

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
        cursor = parse_last_event_id(request.headers.get("Last-Event-ID"))
        if record.events:
            first_sequence = int(record.events[0]["sequence"])
            last_sequence = int(record.events[-1]["sequence"])
            if cursor > last_sequence or cursor < first_sequence - 1:
                fail(409, "cursor_gap", "Event cursor cannot be replayed.", True)

        async def stream() -> AsyncIterator[str]:
            for event in record.events:
                if int(event["sequence"]) <= cursor:
                    continue
                data = json.dumps(event, separators=(",", ":"))
                yield f"id: {event['sequence']}\ndata: {data}\n\n"

        return StreamingResponse(
            stream(),
            media_type="text/event-stream",
            headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
        )

    @app.post("/v1/runs/{run_id}/stop")
    async def stop_run(
        payload: StopRequest,
        request: Request,
        run_id: str = Path(min_length=1, max_length=128),
    ) -> JSONResponse:
        active_store: RunStore = request.app.state.store
        fingerprint = request_fingerprint(payload, discriminator=f"stop:{run_id}")
        replay = _command_replay(active_store, payload.command_id, fingerprint)
        if replay is not None:
            return JSONResponse(status_code=200, content=replay)
        record = _require_run(active_store, run_id)
        if record.status in TERMINAL_STATUSES:
            fail(409, "invalid_run_state", "Terminal run cannot be stopped.")
        replay = _begin_command(
            active_store, payload.command_id, fingerprint, record,
            payload.expected_version, "stop",
        )
        if replay is not None:
            return JSONResponse(status_code=200, content=replay)
        active_driver = request.app.state.driver
        cancel_hook = getattr(active_driver, "cancel", None)
        supports_cancel = bool(getattr(active_driver, "supports_cancellation", False))
        if not supports_cancel or not callable(cancel_hook):
            error = _wire_error(
                "unsupported_control", "Driver does not support safe cancellation."
            )
            active_store.fail_command(payload.command_id, error)
            fail(409, "unsupported_control", "Driver does not support safe cancellation.")
        try:
            cancel_hook(record.run_id, _run_request(record))
        except Exception as exc:
            error = _wire_error(
                "control_failed", "Framework cancellation failed.", True
            )
            active_store.fail_command(payload.command_id, error)
            fail(502, "control_failed", "Framework cancellation failed.", True)
        record.status = "cancelled"
        record.approval = None
        record.updated_at = utc_now()
        record.version += 1
        append_event(record, "cancelled", "Run cancelled by central authority.")
        receipt = active_store.complete_command(payload.command_id, record)
        return JSONResponse(status_code=202, content=receipt.wire())

    @app.post("/v1/runs/{run_id}/approvals/{approval_id}")
    async def submit_approval(
        payload: ApprovalDecisionRequest,
        request: Request,
        run_id: str = Path(min_length=1, max_length=128),
        approval_id: str = Path(min_length=1, max_length=128),
    ) -> JSONResponse:
        active_driver: Driver = request.app.state.driver
        active_store: RunStore = request.app.state.store
        fingerprint = request_fingerprint(payload, discriminator=f"approval:{run_id}:{approval_id}")
        replay = _command_replay(active_store, payload.command_id, fingerprint)
        if replay is not None:
            return JSONResponse(status_code=200, content=replay)
        record = _require_run(active_store, run_id)
        if payload.approval_id != approval_id:
            fail(409, "approval_mismatch", "Approval id does not match route.")
        if record.status != "approval_needed" or record.approval is None:
            fail(409, "invalid_run_state", "Run has no pending approval.")
        if record.approval["id"] != approval_id:
            fail(404, "approval_not_found", "Approval not found.")
        kind = "approve" if payload.decision == "approve" else "deny"
        replay = _begin_command(
            active_store, payload.command_id, fingerprint, record,
            payload.expected_version, kind,
        )
        if replay is not None:
            return JSONResponse(status_code=200, content=replay)

        record.approval = None
        record.updated_at = utc_now()
        if payload.decision == "deny":
            record.status = "failed"
            record.version += 1
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
            receipt = active_store.complete_command(payload.command_id, record)
            return JSONResponse(status_code=202, content=receipt.wire())

        run_request = _run_request(record, approval=payload)
        try:
            outcome = active_driver.resume_after_approval(record.run_id, run_request)
        except Exception as exc:
            outcome = driver_exception_outcome(active_driver.name, exc)
        apply_outcome(record, outcome)
        receipt = active_store.complete_command(payload.command_id, record)
        return JSONResponse(status_code=202, content=receipt.wire())

    return app


def _require_run(store: RunStore, run_id: str) -> RunRecord:
    if EXTERNAL_ID_PATTERN.fullmatch(run_id) is None:
        fail(404, "run_not_found", "Run not found.")
    record = store.get(run_id)
    if record is None:
        fail(404, "run_not_found", "Run not found.")
    return record


def _command_replay(store: RunStore, command_id: str, fingerprint: str) -> dict[str, object] | None:
    try:
        receipt = store.replay_command(command_id, fingerprint)
    except StoreConflictError:
        fail(409, "command_conflict", "Command id already has different content.")
    return receipt.wire() if receipt is not None else None


def _begin_command(
    store: RunStore, command_id: str, fingerprint: str, record: RunRecord,
    expected_version: int, kind: str,
) -> dict[str, object] | None:
    try:
        receipt, created = store.begin_command(
            command_id, fingerprint, record.run_id, expected_version, kind
        )
    except StoreVersionError:
        fail(409, "version_conflict", "Run version does not match expected_version.", True)
    except StoreConflictError:
        fail(409, "command_conflict", "Run already has a conflicting command.")
    return None if created else receipt.wire()


def _wire_error(code: str, message: str, recoverable: bool = False) -> dict[str, object]:
    return {"code": code, "message": message, "recoverable": recoverable}


def _run_request(
    record: RunRecord, approval: ApprovalDecisionRequest | None = None
) -> RunCreateRequest:
    metadata = dict(record.request_metadata)
    values: dict[str, object] = {
        "run_id": record.run_id,
        "correlation": record.correlation,
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
    request = RunCreateRequest(**values)
    if approval is None:
        return request
    internal_metadata = dict(request.metadata)
    internal_metadata["_framework_resume"] = approval.model_dump(mode="json")
    return request.model_copy(update={"metadata": internal_metadata})


app = create_app()
