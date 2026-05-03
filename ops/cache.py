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
DEFAULT_MIN_FREE_GB = 8


def _username() -> str:
    return os.environ.get("USERNAME") or os.environ.get("USER") or os.environ.get("LOGNAME") or "user"


def _default_user_cache_root() -> Path:
    if is_windows():
        drive = Path(ROOT_DIR.drive + "\\") if ROOT_DIR.drive else Path.home().anchor
        if str(drive).upper().startswith("C:"):
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


def _existing_usage_path(path: Path) -> Path:
    candidate = path
    while not candidate.exists():
        if candidate.parent == candidate:
            return ROOT_DIR
        candidate = candidate.parent
    return candidate


def _disk_targets(paths: tuple[Path, ...] | list[Path] | None = None) -> list[tuple[str, Path]]:
    requested = paths or [PROJECT_CACHE_ROOT, ROOT_DIR]
    deduped: dict[int, tuple[str, Path]] = {}
    for raw_path in requested:
        usage_path = _existing_usage_path(Path(raw_path))
        try:
            device_id = usage_path.stat().st_dev
        except OSError:
            device_id = hash(str(usage_path))
        deduped.setdefault(device_id, (str(raw_path), usage_path))
    return list(deduped.values())


def ensure_disk_headroom(
    *,
    min_free_gb: int = DEFAULT_MIN_FREE_GB,
    paths: tuple[Path, ...] | list[Path] | None = None,
    reason: str = "",
) -> None:
    failures: list[str] = []
    heading = f"=== DISK HEADROOM CHECK{f' ({reason})' if reason else ''} ==="
    print(heading)
    for label, usage_path in _disk_targets(paths):
        usage = shutil.disk_usage(usage_path)
        free_gb = usage.free / float(1024 ** 3)
        print(
            f"  - {label}: free {_format_size(usage.free)} / total {_format_size(usage.total)}"
            f" ({free_gb:.1f} GiB free)"
        )
        if free_gb < float(min_free_gb):
            failures.append(f"{label} has only {free_gb:.1f} GiB free")

    print("  Note: this guard covers the repo/cache volume, not Docker daemon image-layer storage.")
    if failures:
        print("  Suggested recovery order: uv run inv lifecycle.down -> uv run inv cache.status -> uv run inv cache.clean")
        raise SystemExit(
            f"DISK HEADROOM CHECK FAILED: need at least {min_free_gb} GiB free.\n- "
            + "\n- ".join(failures)
        )


def _delete_path(path: Path) -> int:
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


def _set_windows_user_env(name: str, value: str):
    import winreg

    with winreg.OpenKey(winreg.HKEY_CURRENT_USER, r"Environment", 0, winreg.KEY_SET_VALUE) as key:
        winreg.SetValueEx(key, name, 0, winreg.REG_EXPAND_SZ, value)


def _broadcast_windows_env_change():
    import ctypes

    HWND_BROADCAST = 0xFFFF
    WM_SETTINGCHANGE = 0x001A
    SMTO_ABORTIFHUNG = 0x0002
    ctypes.windll.user32.SendMessageTimeoutW(HWND_BROADCAST, WM_SETTINGCHANGE, 0, "Environment", SMTO_ABORTIFHUNG, 5000, None)


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
    print("Disk headroom:")
    for label, usage_path in _disk_targets():
        usage = shutil.disk_usage(usage_path)
        free_gb = usage.free / float(1024 ** 3)
        print(
            f"  - {label}: free {_format_size(usage.free)} / total {_format_size(usage.total)}"
            f" ({free_gb:.1f} GiB free)"
        )
    print("  Note: Docker daemon / WSL image-layer storage is tracked separately from repo-managed cache usage.")


@task(help={"min_free_gb": "Minimum free disk headroom in GiB required before heavy build/test work (default: 8)."})
def guard(c, min_free_gb=DEFAULT_MIN_FREE_GB):
    """Fail fast when the repo/cache volume is too full for repeated build and test churn."""
    del c
    ensure_disk_headroom(min_free_gb=int(min_free_gb), reason="preflight")


@task(help={"project": "Clean project-managed caches and build artifacts (default: True).", "user": "Clean user-level tool caches configured by cache.apply-user-policy (default: False)."})
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
ns.add_task(guard)
ns.add_task(clean)
ns.add_task(apply_user_policy)
