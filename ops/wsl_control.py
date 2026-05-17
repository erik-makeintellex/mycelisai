from __future__ import annotations

import shutil
import subprocess

from invoke import task


DEFAULT_WSL_DISTRO = "mycelis-root"


def _wsl_available() -> bool:
    return shutil.which("wsl.exe") is not None


@task(
    help={
        "distro": "WSL distro to terminate. Defaults to MYCELIS_WSL_PROOF_DISTRO/MYCELIS_WSL_DISTRO or mycelis-root.",
    }
)
def down(_c, distro=""):
    """Stop the WSL proof distro and localhost-forwarded runtime ports."""
    if not _wsl_available():
        raise SystemExit("wsl.down requires wsl.exe on the Windows host.")

    import os

    selected = (
        distro
        or os.environ.get("MYCELIS_WSL_PROOF_DISTRO")
        or os.environ.get("MYCELIS_WSL_DISTRO")
        or DEFAULT_WSL_DISTRO
    ).strip()
    if not selected:
        raise SystemExit("wsl.down requires a WSL distro name.")

    result = subprocess.run(
        ["wsl.exe", "--terminate", selected],
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        detail = (result.stderr or result.stdout).strip() or f"exit {result.returncode}"
        raise SystemExit(f"Failed to terminate WSL distro '{selected}': {detail}")
    print(f"WSL distro stopped: {selected}")
