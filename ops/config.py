from pathlib import Path
import os
import platform

# -- Environment Sanitization --
# The Windows Store Python sets a global VIRTUAL_ENV that conflicts with uv's
# local .venv. Clear it on import so all `c.run()` calls inherit a clean env.
_sys_venv = os.environ.get("VIRTUAL_ENV", "")
if _sys_venv and ".venv" not in _sys_venv:
    os.environ.pop("VIRTUAL_ENV", None)

# -- Config --
CLUSTER_NAME = "mycelis-cluster"
NAMESPACE = "mycelis"
ROOT_DIR = Path(__file__).parent.parent.resolve() # tasks/../ -> root
CORE_DIR = ROOT_DIR / "core"
SDK_DIR = ROOT_DIR / "sdk/python"

def is_windows():
    return platform.system() == "Windows"

def powershell(command: str) -> str:
    """Build a PowerShell invocation that skips the user profile.

    Prevents PSReadLine / profile errors from polluting task output.
    Works on both Windows PowerShell 5.1 and PowerShell 7+.
    """
    return f'powershell -NoProfile -Command "{command}"'
