from __future__ import annotations

import json
import os

import pytest

from ops import instance_lock


def test_instance_lease_rejects_live_owner(tmp_path):
    first = instance_lock.acquire_instance_lease("interface-e2e", lock_dir=tmp_path)
    try:
        with pytest.raises(RuntimeError, match=f"already owned by PID {os.getpid()}"):
            instance_lock.acquire_instance_lease("interface-e2e", lock_dir=tmp_path)
    finally:
        first.release()


def test_instance_lease_reclaims_stale_owner(tmp_path, monkeypatch):
    path = tmp_path / "interface-e2e.json"
    path.write_text(json.dumps({"pid": 987654321, "host": "stale"}), encoding="utf-8")
    monkeypatch.setattr(instance_lock, "_pid_is_running", lambda _pid: False)

    lease = instance_lock.acquire_instance_lease("interface-e2e", lock_dir=tmp_path)

    assert lease.path == path
    lease.release()
    assert not path.exists()


def test_instance_lease_release_cannot_remove_new_owner(tmp_path):
    lease = instance_lock.acquire_instance_lease("interface-e2e", lock_dir=tmp_path)
    lease.path.write_text(json.dumps({"pid": os.getpid(), "token": "replacement"}), encoding="utf-8")

    lease.release()

    assert lease.path.exists()
