from __future__ import annotations

import pytest
from invoke import Context
from invoke.exceptions import Exit

from ops import cognitive


@pytest.mark.parametrize(
    "task_func",
    [
        cognitive.install.body,
        cognitive.llm.body,
        cognitive.media.body,
        cognitive.up.body,
        cognitive.status.body,
    ],
)
def test_optional_local_cognitive_tasks_fail_cleanly_on_windows(monkeypatch, task_func):
    monkeypatch.setattr(cognitive, "is_windows", lambda: True)

    with pytest.raises(Exit) as excinfo:
        task_func(Context())

    assert "not supported on Windows hosts" in str(excinfo.value)
