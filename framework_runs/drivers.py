from __future__ import annotations

from copy import deepcopy
from typing import Any, Protocol, runtime_checkable

from .domain import (
    DriverApproval,
    DriverError,
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
    production_ready: bool

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
    production_ready = False

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
    """Small real adapter around an injected compiled LangGraph graph.

    LangGraph remains optional. Constructing this adapter fails closed unless
    the dependency is installed and the injected graph exposes ``invoke``.
    The graph receives ``input.graph_input`` (or the complete request input)
    and a stable ``thread_id`` equal to the normalized run id.
    """

    name = "langgraph"
    framework = "langgraph"
    production_ready = False

    def __init__(self, graph: Any) -> None:
        try:
            import langgraph  # noqa: F401
        except ImportError as exc:
            raise DriverDependencyError(
                "LangGraphDriver requires the optional 'langgraph' dependency"
            ) from exc
        if graph is None or not callable(getattr(graph, "invoke", None)):
            raise TypeError("LangGraphDriver requires a compiled graph with invoke()")
        self._graph = graph

    def start(self, run_id: str, request: RunCreateRequest) -> DriverOutcome:
        graph_input = deepcopy(request.input.get("graph_input", request.input))
        result = self._graph.invoke(
            graph_input,
            config={"configurable": {"thread_id": run_id}},
        )
        return DriverOutcome(
            status="completed",
            message="LangGraph produced a completion candidate.",
            outputs=(
                DriverOutput(
                    kind="framework_state",
                    name="LangGraph state",
                    content_type="application/json",
                    metadata={"state": result},
                ),
            ),
            metadata={"framework": "langgraph"},
        )

    def resume_after_approval(
        self, run_id: str, request: RunCreateRequest
    ) -> DriverOutcome:
        raise RuntimeError(
            "Generic LangGraph approval resume is not configured; inject a governed "
            "driver that maps framework interrupts to Mycelis approvals"
        )
