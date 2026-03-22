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
WORKSPACE_DIR = ROOT_DIR / "workspace"
PROJECT_CACHE_ROOT = Path(
    os.environ.get("MYCELIS_PROJECT_CACHE_ROOT", str(WORKSPACE_DIR / "tool-cache"))
)

# -- Host Defaults (env-var overridable) --
# Core API server
# Local tasking talks to the bridged/local port by default, which is 8081 in
# this repo even though the in-cluster service still listens on 8080.
API_HOST = os.environ.get("MYCELIS_API_HOST", "localhost")
API_PORT = int(os.environ.get("MYCELIS_API_PORT", os.environ.get("PORT", "8081")))

# Interface (Next.js) dev server
# Split the bind host from the local probe/base host so the UI can stay LAN-
# reachable by default without forcing local tooling to guess which loopback
# family (IPv4 vs IPv6) the OS picked.
_interface_host_env = os.environ.get("MYCELIS_INTERFACE_HOST")
INTERFACE_HOST = _interface_host_env or "127.0.0.1"
INTERFACE_BIND_HOST = os.environ.get(
    "MYCELIS_INTERFACE_BIND_HOST",
    _interface_host_env or "::",
)
INTERFACE_PORT = int(os.environ.get("MYCELIS_INTERFACE_PORT", "3000"))

def is_windows():
    return platform.system() == "Windows"

def powershell(command: str) -> str:
    """Build a PowerShell invocation that skips the user profile.

    Prevents PSReadLine / profile errors from polluting task output.
    Works on both Windows PowerShell 5.1 and PowerShell 7+.
    """
    return f'powershell -NoProfile -Command "{command}"'


def managed_cache_paths(root: Path | None = None) -> dict[str, Path]:
    cache_root = Path(root or PROJECT_CACHE_ROOT)
    return {
        "root": cache_root,
        "uv": cache_root / "uv",
        "pip": cache_root / "pip",
        "npm": cache_root / "npm",
        "go_build": cache_root / "go-build",
        "go_mod": cache_root / "go-mod",
        "pytest": cache_root / "pytest",
        "playwright": cache_root / "playwright",
        "pycache": cache_root / "pycache",
    }


def ensure_managed_cache_dirs(root: Path | None = None) -> dict[str, Path]:
    paths = managed_cache_paths(root=root)
    for path in paths.values():
        path.mkdir(parents=True, exist_ok=True)
    return paths


def managed_cache_env(
    extra: dict[str, str] | None = None,
    root: Path | None = None,
) -> dict[str, str]:
    paths = managed_cache_paths(root=root)
    env = {
        "MYCELIS_PROJECT_CACHE_ROOT": str(paths["root"]),
        "UV_CACHE_DIR": str(paths["uv"]),
        "PIP_CACHE_DIR": str(paths["pip"]),
        "NPM_CONFIG_CACHE": str(paths["npm"]),
        "npm_config_cache": str(paths["npm"]),
        "GOCACHE": str(paths["go_build"]),
        "GOMODCACHE": str(paths["go_mod"]),
        "PLAYWRIGHT_BROWSERS_PATH": str(paths["playwright"]),
        "NEXT_TELEMETRY_DISABLED": "1",
        "PYTHONPYCACHEPREFIX": str(paths["pycache"]),
    }
    if extra:
        env.update(extra)
    return env
