import os
import shutil
import subprocess
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import PurePosixPath

from invoke import Collection, task

from .config import ROOT_DIR, is_windows


ns = Collection("wsl")

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
    if not is_windows():
        raise SystemExit("WSL proof tasks must be launched from the Windows dev checkout.")
    if not shutil.which("wsl.exe"):
        raise SystemExit("WSL proof tasks require wsl.exe on the Windows host.")


def _configured_distro(distro: str = "") -> str:
    return (distro or DEFAULT_WSL_DISTRO).strip()


def _configured_checkout(checkout: str = "") -> str:
    raw = (checkout or DEFAULT_WSL_CHECKOUT).strip()
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
    return (remote or DEFAULT_WSL_REMOTE).strip() or "origin"


def _wsl_command(*args: str, distro: str = "", checkout: str = "") -> list[str]:
    command = ["wsl.exe"]
    selected_distro = _configured_distro(distro)
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
    completed = subprocess.run(
        command,
        cwd=cwd,
        capture_output=capture_output,
        text=True,
        check=False,
    )
    result = CommandResult(
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
    return _run_command(["git", *args], cwd=ROOT_DIR, check=check)


def _run_wsl_git(*args: str, distro: str = "", checkout: str = "", check: bool = True) -> CommandResult:
    return _run_command(
        _wsl_command("git", *args, distro=distro, checkout=checkout),
        check=check,
    )


def _run_wsl_git_noninteractive(
    *args: str,
    distro: str = "",
    checkout: str = "",
    check: bool = True,
) -> CommandResult:
    return _run_command(
        _wsl_command(
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
    completed = subprocess.run(
        _wsl_command("bash", "-lc", command, distro=distro, checkout=checkout),
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
    return any(marker in text for marker in GIT_AUTH_FAILURE_MARKERS)


def _is_github_https_remote(remote_url: str) -> bool:
    normalized = remote_url.strip().lower()
    return normalized.startswith("https://") and "github.com" in normalized


def _escaped_git_helper_path(path: str) -> str:
    return path.replace(" ", "\\ ")


def _wsl_remote_url(remote: str, *, distro: str = "", checkout: str = "") -> str:
    result = _run_wsl_git("remote", "get-url", remote, distro=distro, checkout=checkout, check=False)
    remote_url = _git_output(result)
    if result.returncode != 0 or not remote_url:
        detail = _command_detail(result)
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
    result = _run_command(
        _wsl_command("bash", "-lc", command, distro=distro, checkout=checkout),
        check=False,
    )
    if result.returncode != 0:
        return ""
    return _git_output(result).splitlines()[0].strip() if _git_output(result) else ""


def _maybe_configure_wsl_gcm_helper(
    remote_url: str,
    *,
    distro: str = "",
    checkout: str = "",
) -> bool:
    if not _is_github_https_remote(remote_url):
        return False

    helper_path = _find_windows_gcm_helper(distro=distro, checkout=checkout)
    if not helper_path:
        return False

    helper = _escaped_git_helper_path(helper_path)
    _run_wsl_git("config", "--local", "credential.helper", helper, distro=distro, checkout=checkout)
    print(f"WSL git auth: configured repo-local Git Credential Manager helper: {helper_path}")
    return True


def _wsl_git_auth_guidance(remote: str, remote_url: str, result: CommandResult) -> str:
    detail = _command_detail(result)
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
    remote_url = _wsl_remote_url(remote, distro=distro, checkout=checkout)
    result = _run_wsl_git_noninteractive(
        "fetch",
        "--prune",
        remote,
        distro=distro,
        checkout=checkout,
        check=False,
    )
    if result.returncode == 0:
        return

    if _is_git_auth_failure(result) and _maybe_configure_wsl_gcm_helper(remote_url, distro=distro, checkout=checkout):
        print("WSL git auth: retrying fetch after helper repair...")
        result = _run_wsl_git_noninteractive(
            "fetch",
            "--prune",
            remote,
            distro=distro,
            checkout=checkout,
            check=False,
        )
        if result.returncode == 0:
            return

    if _is_git_auth_failure(result):
        raise SystemExit(_wsl_git_auth_guidance(remote, remote_url, result))

    detail = _command_detail(result)
    raise SystemExit(f"WSL git fetch failed for remote '{remote}'.\n{detail}")


def _clean_wsl_proof_checkout(*, distro: str = "", checkout: str = "") -> None:
    exclude_args: list[str] = []
    for path in WSL_REFRESH_CLEAN_EXCLUDES:
        exclude_args.extend(["-e", path])
    _run_wsl_git("clean", "-fdx", *exclude_args, distro=distro, checkout=checkout)


def _current_local_branch(*, require_attached: bool = True) -> str:
    branch = _git_output(_run_local_git("branch", "--show-current"))
    if not branch and require_attached:
        raise SystemExit("Windows dev checkout is detached. Pass --branch or --ref explicitly.")
    return branch or "(detached)"


def _local_upstream() -> str:
    result = _run_local_git("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}", check=False)
    if result.returncode != 0:
        return ""
    return _git_output(result)


def _rev_list_counts(left: str, right: str, *, cwd=None, wsl: bool = False, distro: str = "", checkout: str = "") -> tuple[int, int]:
    if wsl:
        result = _run_wsl_git("rev-list", "--left-right", "--count", f"{left}...{right}", distro=distro, checkout=checkout)
    else:
        result = _run_command(["git", "rev-list", "--left-right", "--count", f"{left}...{right}"], cwd=cwd or ROOT_DIR, check=True)
    text = (result.stdout or "").strip()
    parts = text.split()
    if len(parts) != 2 or not all(part.isdigit() for part in parts):
        return (0, 0)
    return (int(parts[0]), int(parts[1]))


def _wsl_repo_guard(*, distro: str = "", checkout: str = "") -> None:
    selected_checkout = _configured_checkout(checkout)
    top_level = _git_output(
        _run_wsl_git("rev-parse", "--show-toplevel", distro=distro, checkout=selected_checkout, check=False)
    )
    if not top_level:
        raise SystemExit(f"Configured WSL proof repo is not a git checkout: {selected_checkout}")
    normalized_top = str(PurePosixPath(top_level))
    if normalized_top != str(PurePosixPath(selected_checkout)):
        raise SystemExit(
            f"Refusing to operate on unexpected WSL repo root. Expected {selected_checkout}, found {normalized_top}."
        )


def _remote_branch_ref(remote: str, branch: str) -> str:
    return f"refs/remotes/{remote}/{branch}"


def _ensure_remote_branch_exists(remote: str, branch: str) -> None:
    ref = _remote_branch_ref(remote, branch)
    result = _run_local_git("rev-parse", "--verify", ref, check=False)
    if result.returncode != 0:
        raise SystemExit(f"Remote branch not found locally after fetch: {remote}/{branch}")


def _ensure_remote_contains_commit(remote: str, ref: str, *, distro: str = "", checkout: str = "") -> None:
    result = _run_wsl_git("branch", "-r", "--contains", ref, distro=distro, checkout=checkout, check=False)
    lines = [line.strip() for line in result.stdout.splitlines() if line.strip()]
    if not any(line == f"{remote}/HEAD" or line.startswith(f"{remote}/") for line in lines):
        raise SystemExit(f"Requested ref is not present on remote {remote}: {ref}")


def _collect_local_state() -> dict[str, str | int]:
    branch = _current_local_branch(require_attached=False)
    commit = _git_output(_run_local_git("rev-parse", "HEAD"))
    upstream = _local_upstream()
    ahead = behind = 0
    if upstream:
        ahead, behind = _rev_list_counts("HEAD", upstream, cwd=ROOT_DIR)
    dirty = len([line for line in _run_local_git("status", "--porcelain").stdout.splitlines() if line.strip()])
    return {
        "branch": branch,
        "commit": commit,
        "upstream": upstream or "(none)",
        "ahead": ahead,
        "behind": behind,
        "dirty": dirty,
    }


def _collect_wsl_state(*, distro: str = "", checkout: str = "") -> dict[str, str | int]:
    selected_checkout = _configured_checkout(checkout)
    _wsl_repo_guard(distro=distro, checkout=selected_checkout)
    branch = _git_output(
        _run_wsl_git("branch", "--show-current", distro=distro, checkout=selected_checkout, check=False)
    )
    commit = _git_output(_run_wsl_git("rev-parse", "HEAD", distro=distro, checkout=selected_checkout))
    dirty = len(
        [line for line in _run_wsl_git("status", "--porcelain", distro=distro, checkout=selected_checkout).stdout.splitlines() if line.strip()]
    )
    return {
        "branch": branch or "(detached)",
        "commit": commit,
        "dirty": dirty,
    }


def _print_repo_state(title: str, state: dict[str, str | int], *, path: str = "", distro: str = "") -> None:
    print(title)
    if path:
        print(f"  path: {path}")
    if distro:
        print(f"  distro: {distro}")
    print(f"  branch: {state['branch']}")
    print(f"  commit: {state['commit']}")
    if "upstream" in state:
        print(f"  upstream: {state['upstream']}")
        print(f"  ahead/behind: {state['ahead']}/{state['behind']}")
    print(f"  dirty paths: {state['dirty']}")


def _assert_current_branch_published(remote: str, branch: str) -> None:
    _run_local_git("fetch", "--prune", remote)
    _ensure_remote_branch_exists(remote, branch)
    ahead, _behind = _rev_list_counts("HEAD", _remote_branch_ref(remote, branch), cwd=ROOT_DIR)
    if ahead > 0:
        raise SystemExit(
            f"Windows dev branch '{branch}' is ahead of {remote}/{branch} by {ahead} commit(s). Push first, then refresh WSL."
        )


def _assert_windows_dev_checkout_is_ready_for_refresh() -> None:
    local_state = _collect_local_state()
    dirty_paths = int(local_state["dirty"])
    if dirty_paths > 0:
        raise SystemExit(
            f"Windows dev checkout has {dirty_paths} dirty path(s). Commit/push or discard them before refreshing WSL."
        )


def _resolve_target(branch: str, ref: str, remote: str) -> tuple[str, str, str]:
    if branch and ref:
        raise SystemExit("Use either --branch or --ref, not both.")
    if ref:
        return ("detached", ref.strip(), ref.strip())

    target_branch = (branch or "").strip() or _current_local_branch(require_attached=True)
    _assert_current_branch_published(remote, target_branch)
    remote_ref = f"{remote}/{target_branch}"
    return (target_branch, remote_ref, remote_ref)


def _probe_windows_gui(url: str, timeout_seconds: float = 10.0) -> None:
    try:
        request = urllib.request.Request(url, headers={"Accept": "text/html"})
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            status = getattr(response, "status", 0)
    except urllib.error.HTTPError as exc:
        status = exc.code
    except Exception as exc:  # pragma: no cover - network/runtime dependent
        raise SystemExit(f"Windows-side GUI probe failed for {url}: {exc}") from exc

    if status not in {200, 304}:
        raise SystemExit(f"Windows-side GUI probe failed for {url}: HTTP {status}")

    print(f"Windows GUI probe: {url} [{status}]")


def _ensure_wsl_compose_env(*, distro: str = "", checkout: str = "") -> None:
    selected_checkout = _configured_checkout(checkout)
    exists = lambda name: _run_command(_wsl_command("bash", "-lc", f"test -f {name}", distro=distro, checkout=selected_checkout), check=False).returncode == 0
    if exists(".env"):
        print("WSL secret env: using existing .env")
    else:
        if not exists(".env.example"):
            raise SystemExit("WSL proof checkout is missing .env and .env.example.")
        print("WSL secret env: bootstrapping .env from .env.example")
        _run_wsl_shell("cp .env.example .env", distro=distro, checkout=selected_checkout)

    if exists(".env.compose"):
        print("WSL compose env: using existing .env.compose")
        return

    if not exists(".env.compose.example"):
        raise SystemExit("WSL proof checkout is missing .env.compose and .env.compose.example.")
    print("WSL compose env: bootstrapping .env.compose from .env.compose.example")
    _run_wsl_shell("cp .env.compose.example .env.compose", distro=distro, checkout=selected_checkout)


def _ensure_wsl_output_block_path(*, distro: str = "", checkout: str = "") -> None:
    selected_checkout = _configured_checkout(checkout)
    command = """python3 - <<'PY'
from pathlib import Path
import os

values = {}
for raw_line in Path('.env.compose').read_text(encoding='utf-8').splitlines():
    line = raw_line.strip()
    if not line or line.startswith('#') or '=' not in raw_line:
        continue
    key, value = raw_line.split('=', 1)
    values[key.strip()] = value.strip().strip('"').strip("'")

raw_path = values.get('MYCELIS_OUTPUT_HOST_PATH', './workspace/docker-compose/data').strip()
path = Path(os.path.expandvars(os.path.expanduser(raw_path)))
if not path.is_absolute():
    path = (Path.cwd() / path).resolve()
path.mkdir(parents=True, exist_ok=True)
print(f"WSL output block path: {path}")
PY"""
    _run_wsl_shell(command, distro=distro, checkout=selected_checkout)


def _wsl_preflight_lane(selected_lane: str) -> str:
    if selected_lane in WSL_COMPOSE_OWNED_LANES:
        return "runtime"
    return selected_lane


@task
def status(_c, distro="", checkout="", remote=""):
    """
    Report the Windows dev checkout state and the WSL proof checkout state.
    """
    _require_windows_dev_host()
    selected_distro = _configured_distro(distro)
    selected_checkout = _configured_checkout(checkout)
    selected_remote = _configured_remote(remote)

    print("=== WSL PROOF STATUS ===")
    print(f"Windows dev repo: {ROOT_DIR}")
    print(f"Configured WSL proof repo: {selected_checkout}")
    print(f"Configured WSL distro: {selected_distro}")
    print(f"Configured remote: {selected_remote}")

    local_state = _collect_local_state()
    wsl_state = _collect_wsl_state(distro=selected_distro, checkout=selected_checkout)

    print()
    _print_repo_state("Windows dev checkout", local_state, path=str(ROOT_DIR))
    print()
    _print_repo_state("WSL proof checkout", wsl_state, path=selected_checkout, distro=selected_distro)
    print()

    if local_state["commit"] == wsl_state["commit"]:
        print("Commit sync: Windows dev HEAD matches the WSL proof checkout.")
    else:
        print("Commit sync: Windows dev HEAD differs from the WSL proof checkout.")


@task(
    help={
        "branch": "Remote branch to refresh in the WSL proof checkout. Defaults to the current Windows dev branch.",
        "ref": "Specific remote commit/ref to refresh in the WSL proof checkout (detached HEAD).",
        "distro": "WSL distro name. Defaults to MYCELIS_WSL_PROOF_DISTRO or mother-brain.",
        "checkout": "Linux-path checkout to refresh. Defaults to MYCELIS_WSL_PROOF_REPO.",
        "remote": "Git remote name (default: origin).",
    }
)
def refresh(_c, branch="", ref="", distro="", checkout="", remote=""):
    """
    Refresh the WSL proof checkout from git only. Never copies from the Windows dev tree.
    """
    _require_windows_dev_host()
    selected_distro = _configured_distro(distro)
    selected_checkout = _configured_checkout(checkout)
    selected_remote = _configured_remote(remote)
    _wsl_repo_guard(distro=selected_distro, checkout=selected_checkout)

    print("=== WSL PROOF REFRESH ===")
    print(f"Windows dev repo: {ROOT_DIR}")
    print(f"WSL proof repo: {selected_checkout}")
    print(f"WSL distro: {selected_distro}")
    print(f"Git remote: {selected_remote}")

    _assert_windows_dev_checkout_is_ready_for_refresh()
    target_branch, checkout_ref, report_ref = _resolve_target(branch, ref, selected_remote)

    print("Fetching remote refs in the Windows dev repo...")
    _run_local_git("fetch", "--prune", selected_remote)

    print("Fetching remote refs in the WSL proof repo...")
    _fetch_wsl_remote_with_auth_repair(selected_remote, distro=selected_distro, checkout=selected_checkout)

    if target_branch == "detached":
        _ensure_remote_contains_commit(selected_remote, checkout_ref, distro=selected_distro, checkout=selected_checkout)
        print(f"Refreshing WSL proof checkout to detached ref {checkout_ref}...")
        _run_wsl_git("checkout", "--detach", checkout_ref, distro=selected_distro, checkout=selected_checkout)
        _run_wsl_git("reset", "--hard", checkout_ref, distro=selected_distro, checkout=selected_checkout)
    else:
        _ensure_remote_branch_exists(selected_remote, target_branch)
        print(f"Refreshing WSL proof checkout to {selected_remote}/{target_branch}...")
        _run_wsl_git(
            "checkout",
            "-B",
            target_branch,
            f"{selected_remote}/{target_branch}",
            distro=selected_distro,
            checkout=selected_checkout,
        )
        _run_wsl_git(
            "reset",
            "--hard",
            f"{selected_remote}/{target_branch}",
            distro=selected_distro,
            checkout=selected_checkout,
        )

    _clean_wsl_proof_checkout(distro=selected_distro, checkout=selected_checkout)

    refreshed_state = _collect_wsl_state(distro=selected_distro, checkout=selected_checkout)
    print()
    print(f"Requested proof ref: {report_ref}")
    _print_repo_state("WSL proof checkout", refreshed_state, path=selected_checkout, distro=selected_distro)


@task(
    help={
        "lane": "WSL proof lane. service/release run runtime preflight, then task-owned Compose/browser proof (default: runtime).",
        "distro": "WSL distro name. Defaults to MYCELIS_WSL_PROOF_DISTRO or mother-brain.",
        "checkout": "Linux-path checkout to validate. Defaults to MYCELIS_WSL_PROOF_REPO.",
        "gui_url": "Windows-side URL to probe after the WSL stack is healthy (default: http://localhost:3000).",
        "compose_wait_timeout": "Compose wait timeout in seconds (default: 240).",
    }
)
def validate(_c, lane="", distro="", checkout="", gui_url="", compose_wait_timeout=""):
    """
    Run the proof-only WSL validation flow against the Linux-native checkout.
    """
    _require_windows_dev_host()
    selected_distro = _configured_distro(distro)
    selected_checkout = _configured_checkout(checkout)
    selected_gui_url = (gui_url or DEFAULT_GUI_URL).strip() or DEFAULT_GUI_URL
    selected_lane = (lane or DEFAULT_RELEASE_LANE).strip() or DEFAULT_RELEASE_LANE
    selected_wait_timeout = (compose_wait_timeout or DEFAULT_COMPOSE_WAIT_TIMEOUT).strip() or DEFAULT_COMPOSE_WAIT_TIMEOUT

    if selected_lane not in WSL_PROOF_LANES:
        raise SystemExit(
            f"Unsupported WSL proof validation lane '{selected_lane}'. Expected baseline, runtime, service, or release."
        )
    preflight_lane = _wsl_preflight_lane(selected_lane)

    wsl_state = _collect_wsl_state(distro=selected_distro, checkout=selected_checkout)
    if int(wsl_state["dirty"]) > 0:
        raise SystemExit(
            "WSL proof checkout is dirty. Refresh it from git first so validation runs against a clean deployment-mimic tree."
        )

    print("=== WSL PROOF VALIDATE ===")
    _print_repo_state("WSL proof checkout", wsl_state, path=selected_checkout, distro=selected_distro)
    print(f"Requested WSL proof lane: {selected_lane}")
    if preflight_lane != selected_lane:
        print(
            "Release-preflight lane inside WSL: runtime "
            f"(the {selected_lane} service/browser gates are owned by the Compose proof sequence below)"
        )
    else:
        print(f"Release-preflight lane inside WSL: {preflight_lane}")
    print(f"Windows GUI probe URL: {selected_gui_url}")
    print()

    _ensure_wsl_compose_env(distro=selected_distro, checkout=selected_checkout)
    _ensure_wsl_output_block_path(distro=selected_distro, checkout=selected_checkout)

    commands = (
        "uv run inv install",
        f"uv run inv ci.release-preflight --lane={preflight_lane} --no-e2e",
        "uv run inv auth.posture --compose",
        f"uv run inv compose.up --build --wait-timeout={selected_wait_timeout}",
        "uv run inv compose.health",
        "uv run inv compose.storage-health",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts --live-backend --workers=1 --server-mode=start",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/team-creation.spec.ts --live-backend --workers=1 --server-mode=start",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts --live-backend --workers=1 --server-mode=start",
        "uv run inv interface.e2e --project=chromium --spec=e2e/specs/workspace-live-backend.spec.ts --live-backend --workers=1 --server-mode=start",
    )
    for command in commands:
        print(f"[WSL] {command}")
        _run_wsl_shell(command, distro=selected_distro, checkout=selected_checkout)

    _probe_windows_gui(selected_gui_url)
    print("WSL proof validation PASSED")


@task(
    help={
        "branch": "Remote branch to refresh in the WSL proof checkout before validation.",
        "ref": "Specific remote commit/ref to refresh before validation.",
        "lane": "Release-preflight lane to run inside WSL before compose/browser proof (default: runtime).",
        "distro": "WSL distro name. Defaults to MYCELIS_WSL_PROOF_DISTRO or mother-brain.",
        "checkout": "Linux-path checkout to refresh/validate. Defaults to MYCELIS_WSL_PROOF_REPO.",
        "remote": "Git remote name (default: origin).",
        "gui_url": "Windows-side URL to probe after the WSL stack is healthy (default: http://localhost:3000).",
        "compose_wait_timeout": "Compose wait timeout in seconds (default: 240).",
    }
)
def cycle(_c, branch="", ref="", lane="", distro="", checkout="", remote="", gui_url="", compose_wait_timeout=""):
    """
    Refresh the WSL proof checkout from git, then run the WSL validation flow.
    """
    refresh.body(
        _c,
        branch=branch,
        ref=ref,
        distro=distro,
        checkout=checkout,
        remote=remote,
    )
    print()
    validate.body(
        _c,
        lane=lane,
        distro=distro,
        checkout=checkout,
        gui_url=gui_url,
        compose_wait_timeout=compose_wait_timeout,
    )


ns.add_task(status)
ns.add_task(refresh)
ns.add_task(validate)
ns.add_task(cycle)
