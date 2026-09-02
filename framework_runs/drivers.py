from __future__ import annotations

import json
from copy import deepcopy
from collections.abc import Callable, Mapping
from typing import Any, Protocol, runtime_checkable

from .domain import (
    DriverApproval,
    DriverError,
    DriverEvent,
    DriverOutcome,
    DriverOutput,
    RunCreateRequest,
)


class DriverDependencyError(RuntimeError):
    """Raised when an optional framework adapter cannot be loaded."""


@runtime_checkable
class Driver(Protocol):
    """Execution boundary behind the normalized Runs API.

    Drivers report execution state and candidate outputs only. They do not
    approve work or create authoritative Mycelis completion proof.
    """

    name: str
    framework: str
    supports_cancellation: bool
    cancellation_mode: str

    def start(self, run_id: str, request: RunCreateRequest) -> DriverOutcome: ...

    def resume_after_approval(
        self, run_id: str, request: RunCreateRequest
    ) -> DriverOutcome: ...


class ConformanceDriver:
    """Deterministic, non-production driver for protocol conformance tests.

    ``input.conformance_mode`` may be ``complete`` (default), ``approval``,
    ``hold``, or ``fail``. This vocabulary is intentionally namespaced so it
    cannot be mistaken for a general-purpose framework executor.
    """

    name = "conformance"
    framework = "built_in"
    supports_cancellation = True
    cancellation_mode = "local_held_run"

    def start(self, run_id: str, request: RunCreateRequest) -> DriverOutcome:
        mode = str(request.input.get("conformance_mode", "complete")).strip().lower()
        if mode == "approval":
            return DriverOutcome(
                status="approval_needed",
                message="Framework execution requires central operator approval.",
                approval=DriverApproval(
                    kind="framework_action",
                    summary=str(
                        request.input.get("approval_summary", "Approve conformance action")
                    ),
                    risk_level=str(request.input.get("risk_level", "medium")),
                    requested_action=str(
                        request.input.get("requested_action", "continue_framework_run")
                    ),
                    metadata={"driver": self.name},
                ),
            )
        if mode == "hold":
            return DriverOutcome(
                status="running",
                message="Conformance run is waiting for cancellation or external progress.",
            )
        if mode == "fail":
            return DriverOutcome(
                status="failed",
                message="Conformance driver reported failure.",
                error=DriverError(
                    code="conformance_failure",
                    message=str(request.input.get("failure_message", "Requested failure")),
                ),
            )
        if mode != "complete":
            return DriverOutcome(
                status="failed",
                message="Unsupported conformance mode.",
                error=DriverError(
                    code="unsupported_conformance_mode",
                    message=f"Unsupported conformance_mode {mode!r}",
                    recoverable=False,
                ),
            )
        return self._completion(request, resumed=False)

    def resume_after_approval(
        self, run_id: str, request: RunCreateRequest
    ) -> DriverOutcome:
        return self._completion(request, resumed=True)

    def cancel(self, run_id: str, request: RunCreateRequest) -> None:
        mode = str(request.input.get("conformance_mode", "complete")).strip().lower()
        if mode != "hold":
            raise RuntimeError("conformance cancellation requires a held run")

    def _completion(
        self, request: RunCreateRequest, *, resumed: bool
    ) -> DriverOutcome:
        output = DriverOutput(
            kind=str(request.input.get("output_kind", "framework_result")),
            name=str(request.input.get("output_name", "Conformance result")),
            uri=str(request.input.get("output_uri", "")),
            content_type=str(request.input.get("content_type", "application/json")),
            metadata={
                "value": deepcopy(request.input.get("output", {"intent": request.intent})),
                "conformance_only": True,
            },
        )
        return DriverOutcome(
            status="completed",
            message=(
                "Conformance driver produced a completion candidate after approval."
                if resumed
                else "Conformance driver produced a completion candidate."
            ),
            outputs=(output,),
            metadata={"resumed_after_approval": resumed},
        )


class LangGraphDriver:
    """Optional facade around an injected, externally checkpointed graph.

    LangGraph remains optional. Constructing this adapter fails closed unless
    the dependency is installed and the injected graph exposes ``invoke``.
    The facade uses the authoritative run id as LangGraph's stable thread id,
    normalizes streamed updates, and maps interrupts to central approvals.
    """

    name = "langgraph"
    framework = "langgraph"

    def __init__(
        self,
        graph: Any,
        *,
        cancel_hook: Callable[[str, dict[str, Any]], None] | None = None,
    ) -> None:
        try:
            from langgraph.types import Command
        except ImportError as exc:
            raise DriverDependencyError(
                "LangGraphDriver requires the optional 'langgraph' dependency"
            ) from exc
        if graph is None or not callable(getattr(graph, "invoke", None)):
            raise TypeError("LangGraphDriver requires a compiled graph with invoke()")
        self._graph = graph
        self._command_type = Command
        self._cancel_hook = cancel_hook
        self.supports_cancellation = cancel_hook is not None
        self.cancellation_mode = "hook" if cancel_hook is not None else "unsupported"

    def start(self, run_id: str, request: RunCreateRequest) -> DriverOutcome:
        graph_input = deepcopy(request.input.get("graph_input", request.input))
        return self._execute(run_id, request, graph_input, resumed=False)

    def resume_after_approval(
        self, run_id: str, request: RunCreateRequest
    ) -> DriverOutcome:
        decision = deepcopy(request.metadata.get("_framework_resume", {}))
        command = self._command_type(resume={
            "decision": decision.get("decision", "approve"),
            "actor_id": decision.get("actor_id", ""),
            "reason": decision.get("reason", ""),
            "metadata": decision.get("metadata", {}),
        })
        return self._execute(run_id, request, command, resumed=True)

    def cancel(self, run_id: str, request: RunCreateRequest) -> None:
        if self._cancel_hook is None:
            raise RuntimeError("LangGraph cancellation hook is not configured")
        self._cancel_hook(run_id, self._config(run_id, request))

    def _execute(
        self,
        run_id: str,
        request: RunCreateRequest,
        graph_input: Any,
        *,
        resumed: bool,
    ) -> DriverOutcome:
        config = self._config(run_id, request)
        stream = getattr(self._graph, "stream", None)
        events: list[DriverEvent] = []
        result: Any = None
        interrupts: tuple[Any, ...] = ()
        if callable(stream):
            for index, chunk in enumerate(stream(
                graph_input,
                config=config,
                stream_mode="updates",
                subgraphs=True,
            ), start=1):
                result = _stream_data(chunk)
                chunk_interrupts = _interrupts(result)
                if chunk_interrupts:
                    interrupts = chunk_interrupts
                    break
                events.append(DriverEvent(
                    message="LangGraph emitted an incremental update.",
                    metadata={
                        "framework": "langgraph",
                        "stream_mode": "updates",
                        "stream_index": index,
                    },
                ))
            if not interrupts:
                result = self._checkpoint_value(config, result)
        else:
            result = self._graph.invoke(graph_input, config=config)
            interrupts = _interrupts(result)
            result = getattr(result, "value", result)

        if interrupts:
            return DriverOutcome(
                status="approval_needed",
                message="LangGraph execution requires central operator approval.",
                approval=_approval_from_interrupts(interrupts),
                metadata={"framework": "langgraph", "resumed": resumed},
                events=tuple(events),
            )
        return DriverOutcome(
            status="completed",
            message="LangGraph produced a completion candidate.",
            outputs=(DriverOutput(
                kind="framework_state",
                name="LangGraph state",
                content_type="application/json",
                metadata=_candidate_state(result),
            ),),
            metadata={"framework": "langgraph", "resumed_after_approval": resumed},
            events=tuple(events),
        )

    def _checkpoint_value(self, config: dict[str, Any], fallback: Any) -> Any:
        get_state = getattr(self._graph, "get_state", None)
        if not callable(get_state):
            return fallback
        try:
            snapshot = get_state(config)
        except Exception:
            return fallback
        return getattr(snapshot, "values", fallback)

    @staticmethod
    def _config(run_id: str, request: RunCreateRequest) -> dict[str, Any]:
        configurable = {
            "thread_id": run_id,
            "mycelis_run_id": run_id,
        }
        configurable.update(request.correlation.model_dump(mode="json"))
        return {"configurable": configurable}


def _stream_data(chunk: Any) -> Any:
    if isinstance(chunk, tuple) and len(chunk) == 2:
        return chunk[1]
    if isinstance(chunk, Mapping) and "data" in chunk and "type" in chunk:
        return chunk["data"]
    return chunk


def _interrupts(value: Any) -> tuple[Any, ...]:
    direct = getattr(value, "interrupts", None)
    if direct:
        return tuple(direct)
    if not isinstance(value, Mapping):
        return ()
    embedded = value.get("__interrupt__") or value.get("interrupts")
    if embedded is None:
        return ()
    return tuple(embedded) if isinstance(embedded, (list, tuple)) else (embedded,)


def _approval_from_interrupts(interrupts: tuple[Any, ...]) -> DriverApproval:
    first = interrupts[0]
    value = getattr(first, "value", first)
    summary = "Approve the interrupted LangGraph action."
    if isinstance(value, str) and value.strip():
        summary = value.strip()[:512]
    elif isinstance(value, Mapping):
        candidate = value.get("summary") or value.get("message")
        if isinstance(candidate, str) and candidate.strip():
            summary = candidate.strip()[:512]
    interrupt_ids = [getattr(item, "id", "") for item in interrupts]
    return DriverApproval(
        kind="langgraph_interrupt",
        summary=summary,
        risk_level="medium",
        requested_action="resume_langgraph_run",
        metadata={
            "framework": "langgraph",
            "interrupt_count": len(interrupts),
            "interrupt_ids": [value for value in interrupt_ids if value],
        },
    )


def _candidate_state(value: Any) -> dict[str, Any]:
    try:
        json.dumps(value)
    except (TypeError, ValueError, OverflowError):
        return {"state_type": type(value).__name__, "state_serializable": False}
    else:
        return {"state": deepcopy(value)}
