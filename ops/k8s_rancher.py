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


def ensure_host_ollama_running(port: int = 11434, timeout_seconds: int = 6) -> bool:
    """Ensure local Ollama daemon is listening on port 11434 on Windows host when installed."""
    try:
        with socket.create_connection(("127.0.0.1", port), timeout=1.0):
            return True
    except OSError:
        pass

    home = Path.home()
    local_app_data = Path(os.environ.get("LOCALAPPDATA", str(home / "AppData" / "Local")))
    candidates = [
        local_app_data / "Programs" / "Ollama" / "ollama.exe",
        Path(r"C:\Program Files\Ollama\ollama.exe"),
    ]
    ollama_bin = next((c for c in candidates if _safe_is_file(c)), None)
    if not ollama_bin:
        return False

    try:
        subprocess.run(
            ["setx", "OLLAMA_HOST", f"0.0.0.0:{port}"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        task_name = "MycelisHostOllamaDaemon"
        cmd_line = f"{ollama_bin} serve"
        subprocess.run(
            ["schtasks", "/Create", "/TN", task_name, "/TR", cmd_line, "/SC", "ONCE", "/ST", "00:00", "/F"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        subprocess.run(
            ["schtasks", "/Run", "/TN", task_name],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        import time

        deadline = time.time() + timeout_seconds
        while time.time() < deadline:
            try:
                with socket.create_connection(("127.0.0.1", port), timeout=1.0):
                    return True
            except OSError:
                time.sleep(0.5)

        env = os.environ.copy()
        env.setdefault("OLLAMA_HOST", f"0.0.0.0:{port}")
        creationflags = getattr(subprocess, "CREATE_NEW_PROCESS_GROUP", 0) | getattr(subprocess, "DETACHED_PROCESS", 0)
        subprocess.Popen(
            [str(ollama_bin), "serve"],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            env=env,
            creationflags=creationflags,
        )
        deadline = time.time() + timeout_seconds
        while time.time() < deadline:
            try:
                with socket.create_connection(("127.0.0.1", port), timeout=1.0):
                    return True
            except OSError:
                time.sleep(0.5)
    except Exception:
        return False
    return False


MAX_RETAINED_IMAGE_VERSIONS = 2
MANAGED_IMAGE_REPOS = (
    "mycelis/core",
    "mycelis/interface",
    "mycelis-home-core",
    "mycelis-home-interface",
)


def prune_old_mycelis_images(
    active_tag: str = "",
    max_versions: int = MAX_RETAINED_IMAGE_VERSIONS,
) -> list[str]:
    """Retain at most `max_versions` (default: 2) image tags per Mycelis repository and prune older + dangling images."""
    if active_tag == "dev":
        return []
    _ensure_rancher_desktop_path()
    keep_limit = max(1, int(max_versions))
    removed: list[str] = []
    try:
        listed = subprocess.run(
            ["docker", "images", "--format", "{{.Repository}}\t{{.Tag}}"],
            capture_output=True,
            text=True,
            timeout=15,
            check=False,
        )
        if listed.returncode == 0:
            by_repo: dict[str, list[str]] = {repo: [] for repo in MANAGED_IMAGE_REPOS}
            for raw_line in listed.stdout.splitlines():
                parts = raw_line.strip().split("\t")
                if len(parts) != 2:
                    continue
                repo, tag = parts[0].strip(), parts[1].strip()
                if repo not in by_repo or not tag or tag == "<none>":
                    continue
                if tag not in by_repo[repo]:
                    by_repo[repo].append(tag)
            stale: list[str] = []
            for repo, tags in by_repo.items():
                ordered = list(tags)
                if active_tag and active_tag in ordered:
                    ordered.remove(active_tag)
                    ordered.insert(0, active_tag)
                for old_tag in ordered[keep_limit:]:
                    stale.append(f"{repo}:{old_tag}")
            if stale:
                subprocess.run(
                    ["docker", "rmi", "-f", *stale],
                    capture_output=True,
                    text=True,
                    timeout=30,
                    check=False,
                )
                removed.extend(stale)
        subprocess.run(
            ["docker", "image", "prune", "-f"],
            capture_output=True,
            text=True,
            timeout=30,
            check=False,
        )
    except Exception:
        return removed
    return removed

