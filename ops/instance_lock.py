from __future__ import annotations

import json
import os
import socket
import uuid
from dataclasses import dataclass
from functools import wraps
from pathlib import Path
from typing import Callable, TypeVar

from .config import ROOT_DIR


LOCK_DIR = ROOT_DIR / "workspace" / "runtime" / "instance-locks"
F = TypeVar("F", bound=Callable)


def _pid_is_running(pid: int) -> bool:
    if pid <= 0:
        return False
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return False
    except PermissionError:
        return True
    return True


@dataclass
class InstanceLease:
    path: Path
    token: str

    def release(self) -> None:
        try:
            payload = json.loads(self.path.read_text(encoding="utf-8"))
        except (FileNotFoundError, OSError, json.JSONDecodeError):
            return
        if payload.get("token") == self.token:
            self.path.unlink(missing_ok=True)


def acquire_instance_lease(name: str, *, lock_dir: Path = LOCK_DIR) -> InstanceLease:
    lock_dir.mkdir(parents=True, exist_ok=True)
    path = lock_dir / f"{name}.json"
    token = str(uuid.uuid4())
    payload = {
        "name": name,
        "pid": os.getpid(),
        "host": socket.gethostname(),
        "token": token,
    }
    encoded = json.dumps(payload, sort_keys=True) + "\n"

    for _ in range(2):
        try:
            fd = os.open(path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
        except FileExistsError:
            try:
                owner = json.loads(path.read_text(encoding="utf-8"))
                owner_pid = int(owner.get("pid") or 0)
            except (OSError, ValueError, TypeError, json.JSONDecodeError):
                owner_pid = 0
                owner = {}
            if owner_pid and _pid_is_running(owner_pid):
                raise RuntimeError(
                    f"{name} is already owned by PID {owner_pid} on "
                    f"{owner.get('host') or 'unknown host'}; wait for that run instead of starting a parallel instance."
                )
            path.unlink(missing_ok=True)
            continue
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            handle.write(encoded)
        return InstanceLease(path=path, token=token)
    raise RuntimeError(f"Unable to acquire exclusive {name} instance lease at {path}.")


def exclusive_instance(name: str):
    def decorate(func: F) -> F:
        @wraps(func)
        def wrapped(*args, **kwargs):
            lease = acquire_instance_lease(name)
            try:
                return func(*args, **kwargs)
            finally:
                lease.release()

        return wrapped  # type: ignore[return-value]

    return decorate
