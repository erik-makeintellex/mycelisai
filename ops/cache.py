from __future__ import annotations

import os
import stat
import shutil
import time
from contextlib import suppress
from pathlib import Path

from invoke import Collection, task

from .config import (
    PROJECT_CACHE_ROOT,
    ROOT_DIR,
    ensure_managed_cache_dirs,
    is_windows,
    managed_cache_paths,
)

ns = Collection("cache")

PROJECT_CACHE_ARTIFACTS = (
    ROOT_DIR / ".next",
    ROOT_DIR / "interface" / ".next",
    ROOT_DIR / "interface" / "coverage",
    ROOT_DIR / "interface" / "playwright-report",
    ROOT_DIR / "interface" / "test-results",
    ROOT_DIR / "interface" / "tsconfig.tsbuildinfo",
    ROOT_DIR / "core" / "bin",
)


def _username() -> str:
    return (
        os.environ.get("USERNAME")
        or os.environ.get("USER")
        or os.environ.get("LOGNAME")
        or "user"
    )


def _default_user_cache_root() -> Path:
    if is_windows():
        drive = Path(ROOT_DIR.drive + "\\") if ROOT_DIR.drive else Path.home().anchor
        drive_root = str(drive)
        if drive_root.upper().startswith("C:"):
            return Path.home() / "AppData" / "Local" / "MycelisCache"
        return drive / "Users" / _username() / "AppData" / "Local" / "MycelisCache"
    return Path.home() / ".cache" / "mycelis"


def _user_cache_root(root: str | None = None) -> Path:
    explicit = root or os.environ.get("MYCELIS_USER_CACHE_ROOT")
    if explicit:
        return Path(explicit)
    return _default_user_cache_root()


def _user_cache_paths(root: str | None = None) -> dict[str, Path]:
    cache_root = _user_cache_root(root=root)
    return {
        "root": cache_root,
        "uv": cache_root / "uv",
        "pip": cache_root / "pip",
        "npm": cache_root / "npm",
        "go_build": cache_root / "go-build",
        "go_mod": cache_root / "go-mod",
        "playwright": cache_root / "playwright",
    }


def _iter_files(path: Path):
    if not path.exists():
        return
    if path.is_file():
        yield path
        return
    for child in path.rglob("*"):
        if child.is_file():
            yield child


def _path_size_bytes(path: Path) -> int:
    return sum(item.stat().st_size for item in _iter_files(path))


def _format_size(num_bytes: int) -> str:
    units = ("B", "KB", "MB", "GB", "TB")
    size = float(num_bytes)
    for unit in units:
        if size < 1024 or unit == units[-1]:
            return f"{size:.2f} {unit}"
        size /= 1024
    return f"{num_bytes} B"


def _delete_path(path: Path) -> int:
    def _handle_remove_error(function, target, excinfo):
        error = excinfo[1] if isinstance(excinfo, tuple) else excinfo
        if isinstance(error, FileNotFoundError):
            return
        if isinstance(error, PermissionError):
            with suppress(OSError):
                os.chmod(target, stat.S_IWRITE | stat.S_IREAD)
            function(target)
            return
        raise error

    reclaimed = _path_size_bytes(path)
    if not path.exists():
        return reclaimed
    if path.is_file():
        try:
            path.unlink()
        except PermissionError:
            with suppress(OSError):
                os.chmod(path, stat.S_IWRITE | stat.S_IREAD)
            path.unlink()
    else:
        attempts = 0
        while True:
            try:
                shutil.rmtree(path, ignore_errors=False, onexc=_handle_remove_error)
                break
            except OSError as error:
                if getattr(error, "winerror", None) == 145 and attempts < 2:
                    attempts += 1
                    time.sleep(0.2)
                    continue
                raise
    return reclaimed


def _set_windows_user_env(name: str, value: str):
    import winreg

    with winreg.OpenKey(
        winreg.HKEY_CURRENT_USER,
        r"Environment",
        0,
        winreg.KEY_SET_VALUE,
    ) as key:
        winreg.SetValueEx(key, name, 0, winreg.REG_EXPAND_SZ, value)


def _broadcast_windows_env_change():
    import ctypes

    HWND_BROADCAST = 0xFFFF
    WM_SETTINGCHANGE = 0x001A
    SMTO_ABORTIFHUNG = 0x0002
    ctypes.windll.user32.SendMessageTimeoutW(
        HWND_BROADCAST,
        WM_SETTINGCHANGE,
        0,
        "Environment",
        SMTO_ABORTIFHUNG,
        5000,
        None,
    )


@task
def status(c):
    """Report managed project and user cache sizes."""
    del c
    project_paths = ensure_managed_cache_dirs(root=PROJECT_CACHE_ROOT)
    user_paths = _user_cache_paths()

    print("=== CACHE STATUS ===")
    print(f"Project cache root: {project_paths['root']}")
    for name, path in project_paths.items():
        if name == "root":
            continue
        print(f"  - {name}: {_format_size(_path_size_bytes(path))} ({path})")

    print(f"User cache root: {user_paths['root']}")
    for name, path in user_paths.items():
        if name == "root":
            continue
        print(f"  - {name}: {_format_size(_path_size_bytes(path))} ({path})")

    print("Project build artifacts:")
    for path in PROJECT_CACHE_ARTIFACTS:
        print(f"  - {_format_size(_path_size_bytes(path))} ({path})")


@task(
    help={
        "project": "Clean project-managed caches and build artifacts (default: True).",
        "user": "Clean user-level tool caches configured by cache.apply-user-policy (default: False).",
    }
)
def clean(c, project=True, user=False):
    """Clean managed project caches and optional user-level tool caches."""
    del c
    reclaimed = 0

    if project:
        project_paths = ensure_managed_cache_dirs(root=PROJECT_CACHE_ROOT)
        for name, path in project_paths.items():
            if name == "root":
                continue
            reclaimed += _delete_path(path)
            path.mkdir(parents=True, exist_ok=True)
        for path in PROJECT_CACHE_ARTIFACTS:
            if path.exists():
                reclaimed += _delete_path(path)

    if user:
        for name, path in _user_cache_paths().items():
            if name == "root":
                continue
            if path.exists():
                reclaimed += _delete_path(path)
            path.mkdir(parents=True, exist_ok=True)

    print(f"CACHE CLEAN COMPLETE: reclaimed approximately {_format_size(reclaimed)}")


@task(name="apply-user-policy", help={"root": "Optional cache root override for persistent user-level cache env vars."})
def apply_user_policy(c, root=""):
    """Persist user-level cache env vars so tools stop defaulting back to C:."""
    del c
    if not is_windows():
        raise SystemExit("cache.apply-user-policy currently supports Windows only.")

    paths = _user_cache_paths(root=root)
    for path in paths.values():
        path.mkdir(parents=True, exist_ok=True)

    assignments = {
        "MYCELIS_USER_CACHE_ROOT": str(paths["root"]),
        "UV_CACHE_DIR": str(paths["uv"]),
        "PIP_CACHE_DIR": str(paths["pip"]),
        "NPM_CONFIG_CACHE": str(paths["npm"]),
        "GOCACHE": str(paths["go_build"]),
        "GOMODCACHE": str(paths["go_mod"]),
        "PLAYWRIGHT_BROWSERS_PATH": str(paths["playwright"]),
    }

    for name, value in assignments.items():
        _set_windows_user_env(name, value)
        os.environ[name] = value

    _broadcast_windows_env_change()

    print("User cache policy applied:")
    for name, value in assignments.items():
        print(f"  - {name}={value}")
    print("New shells will inherit these values automatically.")


ns.add_task(status)
ns.add_task(clean)
ns.add_task(apply_user_policy)
