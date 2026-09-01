from __future__ import annotations

from collections import OrderedDict
from copy import deepcopy
from threading import RLock
from typing import Protocol, runtime_checkable

from .domain import RunRecord


class StoreCapacityError(RuntimeError):
    pass


class StoreConflictError(RuntimeError):
    pass


@runtime_checkable
class RunStore(Protocol):
    """Persistence boundary for atomic normalized run records.

    Production implementations must durably preserve run identity, event order,
    and create conflicts across process restarts.
    """

    production_ready: bool
    storage_kind: str

    def put(self, record: RunRecord) -> None: ...

    def get(self, run_id: str) -> RunRecord | None: ...

    def update(self, record: RunRecord) -> None: ...


class InMemoryRunStore:
    """Bounded, process-local conformance storage; never production durable."""

    production_ready = False
    storage_kind = "bounded_memory_non_production"

    def __init__(self, *, max_runs: int = 256, max_events_per_run: int = 256) -> None:
        if max_runs < 1 or max_events_per_run < 1:
            raise ValueError("store limits must be positive")
        self.max_runs = max_runs
        self.max_events_per_run = max_events_per_run
        self._runs: OrderedDict[str, RunRecord] = OrderedDict()
        self._lock = RLock()

    def put(self, record: RunRecord) -> None:
        with self._lock:
            if record.run_id in self._runs:
                raise StoreConflictError("run id already exists")
            if record.run_id not in self._runs and len(self._runs) >= self.max_runs:
                self._prune_one_terminal_run()
            if record.run_id not in self._runs and len(self._runs) >= self.max_runs:
                raise StoreCapacityError("all bounded run slots are active")
            self._runs[record.run_id] = deepcopy(record)

    def get(self, run_id: str) -> RunRecord | None:
        with self._lock:
            record = self._runs.get(run_id)
            return deepcopy(record) if record is not None else None

    def update(self, record: RunRecord) -> None:
        with self._lock:
            if record.run_id not in self._runs:
                raise KeyError(record.run_id)
            record.events = record.events[-self.max_events_per_run :]
            self._runs[record.run_id] = deepcopy(record)

    def _prune_one_terminal_run(self) -> None:
        for run_id, record in self._runs.items():
            if record.status in {"completed", "failed", "cancelled"}:
                del self._runs[run_id]
                return
