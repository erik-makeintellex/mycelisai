"""Stale app-container detection for the Compose home runtime.

After a Docker restart a container can keep running an image id that no longer
matches its tag (the tag was rebuilt), so health and journey proof would run on
old code. This compares `docker inspect <container> .Image` with
`docker image inspect <tag> .Id` for the app services (core, interface).
"""

from __future__ import annotations

import subprocess
from dataclasses import dataclass
from typing import Callable

from .config import docker_command

PROJECT = "mycelis-home"
APP_SERVICES = ("core", "interface")
REMEDY = "run `uv run inv compose.down` then `uv run inv compose.up --build`"


@dataclass(frozen=True)
class AppImage:
    service: str
    container: str
    tag: str
    state: str  # current | stale | missing-image | missing-container
    detail: str


def _docker(args: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        docker_command(*args), capture_output=True, text=True, timeout=30, check=False
    )


def _inspect_id(args: list[str]) -> str | None:
    result = _docker(args)
    value = (result.stdout or "").strip()
    return value if result.returncode == 0 and value else None


def check_app_images(project: str = PROJECT) -> list[AppImage]:
    results: list[AppImage] = []
    for service in APP_SERVICES:
        container, tag = f"{project}-{service}-1", f"{project}-{service}"
        running = _inspect_id(["inspect", container, "--format", "{{.Image}}"])
        current = _inspect_id(["image", "inspect", tag, "--format", "{{.Id}}"])
        if running is None:
            state, detail = "missing-container", f"container {container} not found"
        elif current is None:
            state, detail = "missing-image", f"image {tag} not found"
        elif running != current:
            state, detail = "stale", f"runs {running[:19]}, tag {tag} is {current[:19]}"
        else:
            state, detail = "current", f"runs {current[:19]}"
        results.append(AppImage(service, container, tag, state, detail))
    return results


def _print_results(results: list[AppImage], marker: str) -> list[AppImage]:
    bad = [r for r in results if r.state != "current"]
    for r in results:
        label = "OK" if r.state == "current" else marker
        print(f"  [{label}] {r.container:<28} {r.state}: {r.detail}")
    return bad


def require_current(label: str, *, allow_stale: bool = False) -> None:
    """Print per-container image state; exit non-zero unless current (or allowed)."""
    print("App image check:")
    bad = _print_results(check_app_images(), "STALE" if allow_stale else "FAIL")
    if not bad:
        return
    if allow_stale:
        print(f"  --allow-stale set: {label} continues on containers that may run old code.")
        return
    names = ", ".join(r.container for r in bad)
    raise SystemExit(
        f"{label} refused: app container(s) not on their current image ({names}). "
        f"Proof on them would test old code. Remedy: {REMEDY}. "
        f"(Pass --allow-stale only if you accept that.)"
    )


def ensure_current_after_up(recreate: Callable[[list[str]], object]) -> None:
    """After `compose up`: recreate stale services once, re-check, fail if still not current."""
    print("App image check:")
    bad = _print_results(check_app_images(), "STALE")
    stale = [r.service for r in bad if r.state == "stale"]
    if stale:
        print(f"  Recreating stale service(s): {', '.join(stale)}")
        recreate(stale)
        print("Re-checking app images:")
        bad = _print_results(check_app_images(), "FAIL")
    if bad:
        names = ", ".join(r.container for r in bad)
        raise SystemExit(f"Compose up failed: app container(s) not on their current image ({names}). Remedy: {REMEDY}.")
