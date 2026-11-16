from __future__ import annotations

import os
import socket
import subprocess
from pathlib import Path
from urllib.parse import urlparse


DEFAULT_RANCHER_SERVER_URL = "https://127.0.0.1:6443"
DEFAULT_RANCHER_CONTEXT = "rancher-desktop"


def rancher_target_url() -> str:
    for env_key in ("MYCELIS_RANCHER_URL", "MYCELIS_K8S_SERVER"):
        value = os.environ.get(env_key, "").strip()
        if value:
            return value
    return DEFAULT_RANCHER_SERVER_URL


def rancher_target_context() -> str:
    for env_key in ("MYCELIS_K8S_CONTEXT", "MYCELIS_RANCHER_CONTEXT"):
        value = os.environ.get(env_key, "").strip()
        if value:
            return value
    return DEFAULT_RANCHER_CONTEXT


def _safe_is_dir(path: Path) -> bool:
    try:
        return path.is_dir()
    except OSError:
        return False


def _safe_is_file(path: Path) -> bool:
    try:
        return path.is_file()
    except OSError:
        return False


def _rancher_desktop_candidate_dirs() -> list[Path]:
    home = Path.home()
    local_app_data = Path(os.environ.get("LOCALAPPDATA", str(home / "AppData" / "Local")))
    candidates: list[Path] = [
        home / ".rd" / "bin",
        home / "scoop" / "apps" / "rancher-desktop" / "current" / "resources" / "resources" / "win32" / "bin",
        home / "scoop" / "apps" / "rancher-desktop" / "current" / "resources" / "resources" / "win32" / "docker-cli-plugins",
        home / "scoop" / "apps" / "rancher-desktop" / "current" / "resources" / "resources" / "linux" / "bin",
        Path(r"C:\Program Files\Rancher Desktop\resources\resources\win32\bin"),
        local_app_data / "Programs" / "Rancher Desktop" / "resources" / "resources" / "win32" / "bin",
        Path("/opt/rancher-desktop/resources/resources/linux/bin"),
        Path("/Applications/Rancher Desktop.app/Contents/Resources/resources/darwin/bin"),
    ]
    wsl_users = Path("/mnt/c/Users")
    if _safe_is_dir(wsl_users):
        try:
            for user_dir in wsl_users.iterdir():
                if not _safe_is_dir(user_dir):
                    continue
                candidates.extend(
                    [
                        user_dir / "scoop" / "apps" / "rancher-desktop" / "current" / "resources" / "resources" / "linux" / "bin",
                        user_dir / "scoop" / "apps" / "rancher-desktop" / "current" / "resources" / "resources" / "win32" / "bin",
                    ]
                )
        except OSError:
            pass
    return candidates


def _ensure_shared_kubeconfig() -> None:
    if os.environ.get("KUBECONFIG", "").strip():
        return
    default_kubeconfig = Path.home() / ".kube" / "config"
    if _safe_is_file(default_kubeconfig):
        return
    wsl_users = Path("/mnt/c/Users")
    if not _safe_is_dir(wsl_users):
        return
    try:
        for user_dir in wsl_users.iterdir():
            candidate = user_dir / ".kube" / "config"
            if _safe_is_file(candidate):
                os.environ["KUBECONFIG"] = str(candidate)
                return
    except OSError:
        return


def _ensure_rancher_desktop_path() -> list[str]:
    _ensure_shared_kubeconfig()
    current_path = os.environ.get("PATH", "")
    existing_entries = {entry.strip().lower() for entry in current_path.split(os.pathsep) if entry.strip()}
    added: list[str] = []
    for candidate in _rancher_desktop_candidate_dirs():
        if _safe_is_dir(candidate):
            norm = str(candidate).lower()
            if norm not in existing_entries:
                added.append(str(candidate))
                existing_entries.add(norm)
    if added:
        os.environ["PATH"] = os.pathsep.join(([current_path] if current_path else []) + added)
    return added


def _rancher_endpoint_reachable(server_url: str | None = None) -> bool:
    parsed = urlparse(server_url or rancher_target_url())
    host = parsed.hostname or "127.0.0.1"
    port = parsed.port or (6443 if parsed.scheme == "https" else 80)
    try:
        with socket.create_connection((host, port), timeout=1.0):
            return True
    except OSError:
        return False


def _rancher_desktop_context_available() -> bool:
    target_ctx = rancher_target_context()
    try:
        proc = subprocess.run(
            ["kubectl", "config", "current-context"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        if proc.returncode == 0 and proc.stdout.strip() == target_ctx:
            return True
        contexts = subprocess.run(
            ["kubectl", "config", "get-contexts", "-o", "name"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        if contexts.returncode == 0:
            names = {line.strip() for line in contexts.stdout.splitlines() if line.strip()}
            if target_ctx in names and _rancher_endpoint_reachable():
                return True
    except Exception:
        return False
    return False
