from __future__ import annotations

import subprocess

import pytest

from ops import compose_images


def _current_docker(args: list[str]) -> subprocess.CompletedProcess[str]:
    """Default for unit tests: never touch Docker; every app image is current."""
    return subprocess.CompletedProcess(args, 0, stdout="sha256:current\n", stderr="")


@pytest.fixture(autouse=True)
def _no_live_docker_for_image_check(monkeypatch):
    monkeypatch.setattr(compose_images, "_docker", _current_docker)
