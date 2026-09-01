from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator


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


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def wire_time(value: datetime) -> str:
    return value.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


class RunCreateRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

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
class DriverOutcome:
    status: Literal["running", "approval_needed", "completed", "failed"]
    message: str
    outputs: tuple[DriverOutput, ...] = ()
    approval: DriverApproval | None = None
    error: DriverError | None = None
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass
class RunRecord:
    run_id: str
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
    status: RunStatus
    created_at: datetime
    updated_at: datetime
    approval: dict[str, Any] | None = None
    result: dict[str, Any] | None = None
    error: dict[str, Any] | None = None
    events: list[dict[str, Any]] = field(default_factory=list)

    def wire(self) -> dict[str, Any]:
        body: dict[str, Any] = {
            "run_id": self.run_id,
            "status": self.status,
            "created_at": wire_time(self.created_at),
            "updated_at": wire_time(self.updated_at),
            "metadata": {
                **self.request_metadata,
                "driver": self.driver_name,
                "execution_authority": "mycelis_core",
                "storage": "bounded_memory_non_production",
                "request_context": {
                    "org_id": self.org_id,
                    "project_id": self.project_id,
                    "user_id": self.user_id,
                    "requested_by": self.requested_by,
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
