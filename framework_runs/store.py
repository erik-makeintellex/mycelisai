from __future__ import annotations

from collections import OrderedDict
from copy import deepcopy
from threading import RLock
from typing import Protocol, runtime_checkable

from .domain import ControlReceipt, RunRecord, utc_now


class StoreCapacityError(RuntimeError):
    pass


class StoreConflictError(RuntimeError):
    pass


class StoreVersionError(RuntimeError):
    def __init__(self, current_version: int) -> None:
        self.current_version = current_version
        super().__init__("run version does not match expected_version")


@runtime_checkable
class RunStore(Protocol):
    """Conformance storage boundary for normalized run records."""

    storage_kind: str

    def put(self, record: RunRecord) -> None: ...

    def get(self, run_id: str) -> RunRecord | None: ...

    def update(self, record: RunRecord) -> None: ...

    def begin_command(
        self, command_id: str, fingerprint: str, run_id: str,
        expected_version: int, kind: str,
    ) -> tuple[ControlReceipt, bool]: ...

    def replay_command(self, command_id: str, fingerprint: str) -> ControlReceipt | None: ...

    def complete_command(self, command_id: str, record: RunRecord) -> ControlReceipt: ...

    def fail_command(self, command_id: str, error: dict[str, object]) -> None: ...


class InMemoryRunStore:
    """Bounded, process-local conformance storage; never production durable."""

    storage_kind = "bounded_memory_non_production"

    def __init__(self, *, max_runs: int = 256, max_events_per_run: int = 256) -> None:
        if max_runs < 1 or max_events_per_run < 1:
            raise ValueError("store limits must be positive")
        self.max_runs = max_runs
        self.max_events_per_run = max_events_per_run
        self._runs: OrderedDict[str, RunRecord] = OrderedDict()
        self._commands: dict[str, tuple[str, ControlReceipt]] = {}
        self._pending_by_run: dict[str, str] = {}
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

    def begin_command(
        self, command_id: str, fingerprint: str, run_id: str,
        expected_version: int, kind: str,
    ) -> tuple[ControlReceipt, bool]:
        with self._lock:
            existing = self._commands.get(command_id)
            if existing is not None:
                if existing[0] != fingerprint:
                    raise StoreConflictError("command id already has different content")
                return deepcopy(existing[1]), False
            record = self._runs.get(run_id)
            if record is None:
                raise KeyError(run_id)
            if record.version != expected_version:
                raise StoreVersionError(record.version)
            if run_id in self._pending_by_run:
                raise StoreConflictError("run already has a pending command")
            now = utc_now()
            receipt = ControlReceipt(
                command_id=command_id,
                run_id=run_id,
                kind=kind,  # type: ignore[arg-type]
                state="pending",
                version=record.version,
                created_at=now,
                updated_at=now,
            )
            self._commands[command_id] = (fingerprint, receipt)
            self._pending_by_run[run_id] = command_id
            return deepcopy(receipt), True

    def replay_command(self, command_id: str, fingerprint: str) -> ControlReceipt | None:
        with self._lock:
            existing = self._commands.get(command_id)
            if existing is None:
                return None
            if existing[0] != fingerprint:
                raise StoreConflictError("command id already has different content")
            return deepcopy(existing[1])

    def complete_command(self, command_id: str, record: RunRecord) -> ControlReceipt:
        with self._lock:
            fingerprint, receipt = self._commands[command_id]
            current = self._runs.get(record.run_id)
            if current is None:
                raise KeyError(record.run_id)
            if current.version != receipt.version:
                raise StoreVersionError(current.version)
            record.version = current.version + 1
            record.events = record.events[-self.max_events_per_run :]
            self._runs[record.run_id] = deepcopy(record)
            receipt.state = "applied"
            receipt.version = record.version
            receipt.updated_at = utc_now()
            self._commands[command_id] = (fingerprint, receipt)
            self._pending_by_run.pop(record.run_id, None)
            return deepcopy(receipt)

    def fail_command(self, command_id: str, error: dict[str, object]) -> None:
        with self._lock:
            fingerprint, receipt = self._commands[command_id]
            receipt.state = "failed"
            receipt.error = deepcopy(error)
            receipt.updated_at = utc_now()
            self._commands[command_id] = (fingerprint, receipt)
            self._pending_by_run.pop(receipt.run_id, None)

    def _prune_one_terminal_run(self) -> None:
        for run_id, record in self._runs.items():
            if record.status in {"completed", "failed", "cancelled"}:
                del self._runs[run_id]
                self._commands = {
                    command_id: stored
                    for command_id, stored in self._commands.items()
                    if stored[1].run_id != run_id
                }
                self._pending_by_run.pop(run_id, None)
                return
