import os
from dataclasses import dataclass
from pathlib import PurePosixPath

import ops.wsl_runtime as _w


DEFAULT_WSL_DISTRO = os.environ.get("MYCELIS_WSL_PROOF_DISTRO") or os.environ.get("MYCELIS_WSL_DISTRO") or "mother-brain"
DEFAULT_WSL_CHECKOUT = os.environ.get("MYCELIS_WSL_PROOF_REPO", "/home/erik/Projects/mycelisai/scratch")
DEFAULT_WSL_REMOTE = os.environ.get("MYCELIS_WSL_PROOF_REMOTE", "origin")
DEFAULT_GUI_URL = os.environ.get("MYCELIS_WSL_PROOF_GUI_URL", "http://localhost:3000")
DEFAULT_RELEASE_LANE = os.environ.get("MYCELIS_WSL_PROOF_RELEASE_LANE", "runtime")
DEFAULT_COMPOSE_WAIT_TIMEOUT = os.environ.get("MYCELIS_WSL_PROOF_COMPOSE_WAIT_TIMEOUT", "240")
WSL_PROOF_LANES = {"baseline", "runtime", "service", "release"}
WSL_COMPOSE_OWNED_LANES = {"service", "release"}
WSL_REFRESH_CLEAN_EXCLUDES = (
    "workspace/tool-cache/",
    "workspace/logs/",
    "workspace/docker-compose/",
)
GIT_AUTH_FAILURE_MARKERS = (
    "authentication failed",
    "could not read username",
    "could not read password",
    "terminal prompts disabled",
    "permission denied (publickey)",
    "could not read from remote repository",
    "repository not found",
    "support for password authentication was removed",
    "no such device or address",
)


@dataclass
class CommandResult:
    command: list[str]
    returncode: int
    stdout: str
    stderr: str


def _require_windows_dev_host() -> None:
    if not _w.is_windows():
        raise SystemExit("WSL proof tasks must be launched from the Windows dev checkout.")
    if not _w.shutil.which("wsl.exe"):
        raise SystemExit("WSL proof tasks require wsl.exe on the Windows host.")


def _configured_distro(distro: str = "") -> str:
    return (distro or _w.DEFAULT_WSL_DISTRO).strip()


def _configured_checkout(checkout: str = "") -> str:
    raw = (checkout or _w.DEFAULT_WSL_CHECKOUT).strip()
    if not raw:
        raise SystemExit("WSL proof repo path is empty. Set MYCELIS_WSL_PROOF_REPO.")
    normalized = PurePosixPath(raw)
    if not normalized.is_absolute():
        raise SystemExit(f"WSL proof repo path must be absolute: {raw}")
    if str(normalized) == "/" or str(normalized).startswith("/mnt/"):
        raise SystemExit(
            f"WSL proof repo must live on the Linux filesystem, not under /mnt: {normalized}"
        )
    return str(normalized)


def _configured_remote(remote: str = "") -> str:
    return (remote or _w.DEFAULT_WSL_REMOTE).strip() or "origin"


def _wsl_command(*args: str, distro: str = "", checkout: str = "") -> list[str]:
    command = ["wsl.exe"]
    selected_distro = _w._configured_distro(distro)
    if selected_distro:
        command.extend(["-d", selected_distro])
    selected_checkout = (checkout or "").strip()
    if selected_checkout:
        command.extend(["--cd", selected_checkout])
    command.extend(["--exec", *args])
    return command


def _run_command(
    command: list[str],
    *,
    cwd=None,
    capture_output: bool = True,
    check: bool = False,
) -> CommandResult:
    completed = _w.subprocess.run(
        command,
        cwd=cwd,
        capture_output=capture_output,
        text=True,
        check=False,
    )
    result = _w.CommandResult(
        command=command,
        returncode=completed.returncode,
        stdout=completed.stdout or "",
        stderr=completed.stderr or "",
    )
    if check and result.returncode != 0:
        detail = (result.stderr or result.stdout).strip() or f"exit {result.returncode}"
        joined = " ".join(command)
        raise SystemExit(f"Command failed: {joined}\n{detail}")
    return result


def _run_local_git(*args: str, check: bool = True) -> CommandResult:
    return _w._run_command(["git", *args], cwd=_w.ROOT_DIR, check=check)


def _run_wsl_git(*args: str, distro: str = "", checkout: str = "", check: bool = True) -> CommandResult:
    return _w._run_command(
        _w._wsl_command("git", *args, distro=distro, checkout=checkout),
        check=check,
    )


def _run_wsl_git_noninteractive(
    *args: str,
    distro: str = "",
    checkout: str = "",
    check: bool = True,
) -> CommandResult:
    return _w._run_command(
        _w._wsl_command(
            "env",
            "GIT_TERMINAL_PROMPT=0",
            "git",
            *args,
            distro=distro,
            checkout=checkout,
        ),
        check=check,
    )


def _run_wsl_shell(command: str, *, distro: str = "", checkout: str = "", check: bool = True) -> None:
    completed = _w.subprocess.run(
        _w._wsl_command("bash", "-lc", command, distro=distro, checkout=checkout),
        text=True,
        check=False,
    )
    if check and completed.returncode != 0:
        raise SystemExit(f"WSL command failed ({completed.returncode}): {command}")


def _git_output(result: CommandResult) -> str:
    return (result.stdout or "").strip()


def _command_detail(result: CommandResult) -> str:
    return (result.stderr or result.stdout).strip() or f"exit {result.returncode}"


def _is_git_auth_failure(result: CommandResult) -> bool:
    if result.returncode == 0:
        return False
    text = f"{result.stderr}\n{result.stdout}".lower()
    return any(marker in text for marker in _w.GIT_AUTH_FAILURE_MARKERS)


def _is_github_https_remote(remote_url: str) -> bool:
    normalized = remote_url.strip().lower()
    return normalized.startswith("https://") and "github.com" in normalized


def _escaped_git_helper_path(path: str) -> str:
    return path.replace(" ", "\\ ")


def _wsl_remote_url(remote: str, *, distro: str = "", checkout: str = "") -> str:
    result = _w._run_wsl_git("remote", "get-url", remote, distro=distro, checkout=checkout, check=False)
    remote_url = _w._git_output(result)
    if result.returncode != 0 or not remote_url:
        detail = _w._command_detail(result)
        raise SystemExit(f"WSL proof checkout cannot resolve git remote '{remote}'.\n{detail}")
    return remote_url


def _find_windows_gcm_helper(*, distro: str = "", checkout: str = "") -> str:
    command = r"""
for candidate in \
  "/mnt/c/Program Files/Git/mingw64/bin/git-credential-manager.exe" \
  "/mnt/c/Program Files/Git/cmd/git-credential-manager.exe" \
  "/mnt/c/Program Files (x86)/Git/mingw64/bin/git-credential-manager.exe"
do
  if [ -x "$candidate" ]; then
    printf '%s\n' "$candidate"
    exit 0
  fi
done
command -v git-credential-manager 2>/dev/null || true
"""
    result = _w._run_command(
        _w._wsl_command("bash", "-lc", command, distro=distro, checkout=checkout),
        check=False,
    )
    if result.returncode != 0:
        return ""
    return _w._git_output(result).splitlines()[0].strip() if _w._git_output(result) else ""


def _maybe_configure_wsl_gcm_helper(
    remote_url: str,
    *,
    distro: str = "",
    checkout: str = "",
) -> bool:
    if not _w._is_github_https_remote(remote_url):
        return False

    helper_path = _w._find_windows_gcm_helper(distro=distro, checkout=checkout)
    if not helper_path:
        return False

    helper = _w._escaped_git_helper_path(helper_path)
    _w._run_wsl_git("config", "--local", "credential.helper", helper, distro=distro, checkout=checkout)
    print(f"WSL git auth: configured repo-local Git Credential Manager helper: {helper_path}")
    return True


def _wsl_git_auth_guidance(remote: str, remote_url: str, result: CommandResult) -> str:
    detail = _w._command_detail(result)
    if remote_url.strip().lower().startswith("git@") or remote_url.strip().lower().startswith("ssh://"):
        repair = (
            "This WSL checkout uses an SSH remote. Make a GitHub key available inside WSL "
            "and verify it with `ssh -T git@github.com`, or switch only the WSL proof checkout remote "
            "to an HTTPS GitHub URL that can use Git Credential Manager."
        )
    else:
        repair = (
            "This WSL checkout uses an HTTPS remote. Sign in through Git Credential Manager from Windows "
            "with `git fetch {remote}` in the Windows repo, then rerun `uv run inv wsl.refresh`. "
            "If WSL still cannot see Git for Windows, configure the proof checkout with "
            "`git config --local credential.helper /mnt/c/Program\\ Files/Git/mingw64/bin/git-credential-manager.exe`."
        ).format(remote=remote)

    return (
        f"WSL git auth is not ready for remote '{remote}' ({remote_url}).\n"
        f"{repair}\n"
        "Keep the Windows-dev -> WSL-proof boundary git-backed: push from Windows, fetch/reset in WSL, "
        "and do not copy source trees across the host boundary.\n"
        f"Original git output:\n{detail}"
    )


def _fetch_wsl_remote_with_auth_repair(remote: str, *, distro: str = "", checkout: str = "") -> None:
    remote_url = _w._wsl_remote_url(remote, distro=distro, checkout=checkout)
    result = _w._run_wsl_git_noninteractive(
        "fetch",
        "--prune",
        remote,
        distro=distro,
        checkout=checkout,
        check=False,
    )
    if result.returncode == 0:
        return

    if _w._is_git_auth_failure(result) and _w._maybe_configure_wsl_gcm_helper(remote_url, distro=distro, checkout=checkout):
        print("WSL git auth: retrying fetch after helper repair...")
        result = _w._run_wsl_git_noninteractive(
            "fetch",
            "--prune",
            remote,
            distro=distro,
            checkout=checkout,
            check=False,
        )
        if result.returncode == 0:
            return

    if _w._is_git_auth_failure(result):
        raise SystemExit(_w._wsl_git_auth_guidance(remote, remote_url, result))

    detail = _w._command_detail(result)
    raise SystemExit(f"WSL git fetch failed for remote '{remote}'.\n{detail}")
