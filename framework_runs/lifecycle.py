from __future__ import annotations

from dataclasses import asdict
from uuid import uuid4

from .domain import DriverError, DriverOutcome, RunRecord, utc_now, wire_time


COMPLETION_CANDIDATE_METADATA = {
    "completion_authority": "candidate",
    "requires_core_validation": True,
    "verified": False,
}


def append_event(
    record: RunRecord,
    kind: str,
    message: str,
    *,
    metadata: dict[str, object] | None = None,
    **payload: object,
) -> None:
    record.event_sequence += 1
    event_metadata = {
        "driver": record.driver_name,
        "execution_authority": "mycelis_core",
        "correlation_id": record.correlation_id,
        "correlation": record.correlation,
        "correlation_complete": record.correlation_complete,
        **(metadata or {}),
    }
    record.events.append(
        {
            "event_id": str(uuid4()),
            "sequence": record.event_sequence,
            "run_id": record.run_id,
            "correlation_id": record.correlation_id,
            "correlation": record.correlation,
            "kind": kind,
            "status": record.status,
            "message": message,
            "timestamp": wire_time(utc_now()),
            "metadata": event_metadata,
            **payload,
        }
    )


def apply_outcome(record: RunRecord, outcome: DriverOutcome) -> None:
    record.updated_at = utc_now()
    if outcome.events:
        record.status = "running"
    for event in outcome.events:
        append_event(record, "progress", event.message, metadata=event.metadata)
    record.status = outcome.status

    if outcome.status == "approval_needed":
        if outcome.approval is None:
            record.status = "failed"
            record.error = {
                "code": "invalid_driver_approval",
                "message": "Driver requested approval without approval details.",
                "recoverable": False,
            }
            append_event(record, "failed", record.error["message"], error=record.error)
            return
        record.approval = {"id": str(uuid4()), **asdict(outcome.approval)}
        append_event(
            record,
            "approval_needed",
            outcome.message,
            approval=record.approval,
        )
        return
    if outcome.status == "completed":
        outputs = []
        for output in outcome.outputs:
            wire_output = asdict(output)
            wire_output["metadata"] = {
                **output.metadata,
                **COMPLETION_CANDIDATE_METADATA,
            }
            outputs.append({"id": str(uuid4()), **wire_output})
        record.result = {
            "summary": outcome.message,
            "outputs": outputs,
            "metadata": {**outcome.metadata, **COMPLETION_CANDIDATE_METADATA},
            "finished_at": wire_time(record.updated_at),
        }
        append_event(
            record,
            "completed",
            outcome.message,
            result=record.result,
            metadata=COMPLETION_CANDIDATE_METADATA,
        )
        return
    if outcome.status == "failed":
        record.error = asdict(outcome.error) if outcome.error else {
            "code": "driver_failed",
            "message": outcome.message,
            "recoverable": True,
            "metadata": {},
        }
        append_event(record, "failed", outcome.message, error=record.error)
        return
    append_event(record, "progress", outcome.message, metadata=outcome.metadata)


def driver_exception_outcome(driver_name: str, exc: Exception) -> DriverOutcome:
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
