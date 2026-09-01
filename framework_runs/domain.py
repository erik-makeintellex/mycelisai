from __future__ import annotations

import re
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator


RunStatus = Literal[
    "accepted",
    "running",
    "approval_needed",
    "completed",
    "failed",
    "cancelled",
]
EventKind = Literal[
    "accepted",
    "progress",
    "approval_needed",
    "completed",
    "failed",
    "cancelled",
]
EXTERNAL_ID_PATTERN = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$")


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def wire_time(value: datetime) -> str:
    return value.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


class RunCorrelation(BaseModel):
    model_config = ConfigDict(extra="forbid")

    run_id: str = Field(min_length=1, max_length=128)
    intent_proof_id: str = Field(default="", max_length=128)
    execution_contract_id: str = Field(default="", max_length=128)
    team_id: str = Field(default="", max_length=128)
    outcome_id: str = Field(default="", max_length=128)
    work_item_id: str = Field(default="", max_length=128)
    idempotency_key: str = Field(default="", max_length=128)
    source_kind: str = Field(default="", max_length=128)
    source_channel: str = Field(default="", max_length=128)
    payload_kind: str = Field(default="", max_length=128)
    graph_revision: str = Field(default="", max_length=128)

    @field_validator("*")
    @classmethod
    def identifiers_must_be_safe(cls, value: str) -> str:
        value = value.strip()
        if value and EXTERNAL_ID_PATTERN.fullmatch(value) is None:
            raise ValueError("correlation identifier contains unsupported characters")
        return value

    @property
    def complete(self) -> bool:
        return bool(
            self.intent_proof_id
            and self.execution_contract_id
            and self.work_item_id
            and self.idempotency_key
            and self.source_kind
            and self.source_channel
            and self.payload_kind
            and self.graph_revision
        )


class RunCreateRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    run_id: str = Field(default="", max_length=128)
    correlation_id: str = Field(default="", max_length=128)
    correlation: RunCorrelation | None = None
    org_id: str = Field(default="", max_length=256)
    project_id: str = Field(default="", max_length=256)
    user_id: str = Field(default="", max_length=256)
    requested_by: str = Field(default="", max_length=256)
    intent: str = Field(min_length=1, max_length=16_384)
    instructions: str = Field(default="", max_length=65_536)
    input: dict[str, Any] = Field(default_factory=dict)
    required_protocols: list[str] = Field(default_factory=list, max_length=32)
    required_features: list[str] = Field(default_factory=list, max_length=256)
    metadata: dict[str, Any] = Field(default_factory=dict)

    @field_validator("intent")
    @classmethod
    def intent_must_not_be_blank(cls, value: str) -> str:
        value = value.strip()
        if not value:
            raise ValueError("intent must not be blank")
        return value

    @field_validator("run_id", "correlation_id")
    @classmethod
    def external_ids_must_be_safe(cls, value: str) -> str:
        value = value.strip()
        if value and EXTERNAL_ID_PATTERN.fullmatch(value) is None:
            raise ValueError("identifier contains unsupported characters")
        return value

    @model_validator(mode="after")
    def structured_correlation_must_match(self) -> RunCreateRequest:
        if self.correlation is None:
            return self
        if not self.run_id or self.correlation.run_id != self.run_id:
            raise ValueError("correlation.run_id must match top-level run_id")
        if not self.correlation.complete:
            raise ValueError("structured correlation is incomplete")
        return self


class ApprovalDecisionRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    approval_id: str
    decision: Literal["approve", "deny"]
    actor_id: str = ""
    reason: str = ""
    metadata: dict[str, Any] = Field(default_factory=dict)


@dataclass(frozen=True)
class DriverApproval:
    kind: str
    summary: str
    risk_level: str
    requested_action: str
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class DriverOutput:
    kind: str
    name: str = ""
    uri: str = ""
    content_type: str = ""
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class DriverError:
    code: str
    message: str
    recoverable: bool = True
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class DriverEvent:
    message: str
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class DriverOutcome:
    status: Literal["running", "approval_needed", "completed", "failed"]
    message: str
    outputs: tuple[DriverOutput, ...] = ()
    approval: DriverApproval | None = None
    error: DriverError | None = None
    metadata: dict[str, Any] = field(default_factory=dict)
    events: tuple[DriverEvent, ...] = ()


@dataclass
class RunRecord:
    run_id: str
    correlation_id: str
    correlation: dict[str, str]
    correlation_complete: bool
    request_fingerprint: str
    org_id: str
    project_id: str
    user_id: str
    requested_by: str
    intent: str
    instructions: str
    input: dict[str, Any]
    required_protocols: list[str]
    required_features: list[str]
    request_metadata: dict[str, Any]
    driver_name: str
    storage_kind: str
    status: RunStatus
    created_at: datetime
    updated_at: datetime
    approval: dict[str, Any] | None = None
    result: dict[str, Any] | None = None
    error: dict[str, Any] | None = None
    events: list[dict[str, Any]] = field(default_factory=list)
    event_sequence: int = 0

    def wire(self) -> dict[str, Any]:
        body: dict[str, Any] = {
            "run_id": self.run_id,
            "correlation_id": self.correlation_id,
            "correlation": self.correlation,
            "status": self.status,
            "created_at": wire_time(self.created_at),
            "updated_at": wire_time(self.updated_at),
            "metadata": {
                **self.request_metadata,
                "driver": self.driver_name,
                "execution_authority": "mycelis_core",
                "storage": self.storage_kind,
                "correlation": self.correlation,
                "correlation_complete": self.correlation_complete,
                "request_context": {
                    "org_id": self.org_id,
                    "project_id": self.project_id,
                    "user_id": self.user_id,
                    "requested_by": self.requested_by,
                    "correlation_id": self.correlation_id,
                    "required_protocols": self.required_protocols,
                    "required_features": self.required_features,
                },
            },
        }
        if self.approval is not None:
            body["approval"] = self.approval
        if self.result is not None:
            body["result"] = self.result
        if self.error is not None:
            body["error"] = self.error
        return body
