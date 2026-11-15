from pathlib import PurePosixPath

import ops.wsl_runtime as _w


def _clean_wsl_proof_checkout(*, distro: str = "", checkout: str = "") -> None:
    exclude_args: list[str] = []
    for path in _w.WSL_REFRESH_CLEAN_EXCLUDES:
        exclude_args.extend(["-e", path])
    _w._run_wsl_git("clean", "-fdx", *exclude_args, distro=distro, checkout=checkout)


def _current_local_branch(*, require_attached: bool = True) -> str:
    branch = _w._git_output(_w._run_local_git("branch", "--show-current"))
    if not branch and require_attached:
        raise SystemExit("Windows dev checkout is detached. Pass --branch or --ref explicitly.")
    return branch or "(detached)"


def _local_upstream() -> str:
    result = _w._run_local_git("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}", check=False)
    if result.returncode != 0:
        return ""
    return _w._git_output(result)


def _rev_list_counts(left: str, right: str, *, cwd=None, wsl: bool = False, distro: str = "", checkout: str = "") -> tuple[int, int]:
    if wsl:
        result = _w._run_wsl_git("rev-list", "--left-right", "--count", f"{left}...{right}", distro=distro, checkout=checkout)
    else:
        result = _w._run_command(["git", "rev-list", "--left-right", "--count", f"{left}...{right}"], cwd=cwd or _w.ROOT_DIR, check=True)
    text = (result.stdout or "").strip()
    parts = text.split()
    if len(parts) != 2 or not all(part.isdigit() for part in parts):
        return (0, 0)
    return (int(parts[0]), int(parts[1]))


def _wsl_repo_guard(*, distro: str = "", checkout: str = "") -> None:
    selected_checkout = _w._configured_checkout(checkout)
    top_level = _w._git_output(
        _w._run_wsl_git("rev-parse", "--show-toplevel", distro=distro, checkout=selected_checkout, check=False)
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
    ref = _w._remote_branch_ref(remote, branch)
    result = _w._run_local_git("rev-parse", "--verify", ref, check=False)
    if result.returncode != 0:
        raise SystemExit(f"Remote branch not found locally after fetch: {remote}/{branch}")


def _ensure_remote_contains_commit(remote: str, ref: str, *, distro: str = "", checkout: str = "") -> None:
    result = _w._run_wsl_git("branch", "-r", "--contains", ref, distro=distro, checkout=checkout, check=False)
    lines = [line.strip() for line in result.stdout.splitlines() if line.strip()]
    if not any(line == f"{remote}/HEAD" or line.startswith(f"{remote}/") for line in lines):
        raise SystemExit(f"Requested ref is not present on remote {remote}: {ref}")


def _collect_local_state() -> dict[str, str | int]:
    branch = _w._current_local_branch(require_attached=False)
    commit = _w._git_output(_w._run_local_git("rev-parse", "HEAD"))
    upstream = _w._local_upstream()
    ahead = behind = 0
    if upstream:
        ahead, behind = _w._rev_list_counts("HEAD", upstream, cwd=_w.ROOT_DIR)
    dirty = len([line for line in _w._run_local_git("status", "--porcelain").stdout.splitlines() if line.strip()])
    return {
        "branch": branch,
        "commit": commit,
        "upstream": upstream or "(none)",
        "ahead": ahead,
        "behind": behind,
        "dirty": dirty,
    }


def _collect_wsl_state(*, distro: str = "", checkout: str = "") -> dict[str, str | int]:
    selected_checkout = _w._configured_checkout(checkout)
    _w._wsl_repo_guard(distro=distro, checkout=selected_checkout)
    branch = _w._git_output(
        _w._run_wsl_git("branch", "--show-current", distro=distro, checkout=selected_checkout, check=False)
    )
    commit = _w._git_output(_w._run_wsl_git("rev-parse", "HEAD", distro=distro, checkout=selected_checkout))
    dirty = len(
        [line for line in _w._run_wsl_git("status", "--porcelain", distro=distro, checkout=selected_checkout).stdout.splitlines() if line.strip()]
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
    _w._run_local_git("fetch", "--prune", remote)
    _w._ensure_remote_branch_exists(remote, branch)
    ahead, _behind = _w._rev_list_counts("HEAD", _w._remote_branch_ref(remote, branch), cwd=_w.ROOT_DIR)
    if ahead > 0:
        raise SystemExit(
            f"Windows dev branch '{branch}' is ahead of {remote}/{branch} by {ahead} commit(s). Push first, then refresh WSL."
        )


def _assert_windows_dev_checkout_is_ready_for_refresh() -> None:
    local_state = _w._collect_local_state()
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

    target_branch = (branch or "").strip() or _w._current_local_branch(require_attached=True)
    _w._assert_current_branch_published(remote, target_branch)
    remote_ref = f"{remote}/{target_branch}"
    return (target_branch, remote_ref, remote_ref)


def _probe_windows_gui(url: str, timeout_seconds: float = 10.0) -> None:
    try:
        request = _w.urllib.request.Request(url, headers={"Accept": "text/html"})
        with _w.urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            status = getattr(response, "status", 0)
    except _w.urllib.error.HTTPError as exc:
        status = exc.code
    except Exception as exc:  # pragma: no cover - network/runtime dependent
        raise SystemExit(f"Windows-side GUI probe failed for {url}: {exc}") from exc

    if status not in {200, 304}:
        raise SystemExit(f"Windows-side GUI probe failed for {url}: HTTP {status}")

    print(f"Windows GUI probe: {url} [{status}]")


def _ensure_wsl_compose_env(*, distro: str = "", checkout: str = "") -> None:
    selected_checkout = _w._configured_checkout(checkout)
    exists = lambda name: _w._run_command(_w._wsl_command("bash", "-lc", f"test -f {name}", distro=distro, checkout=selected_checkout), check=False).returncode == 0
    if exists(".env"):
        print("WSL secret env: using existing .env")
    else:
        if not exists(".env.example"):
            raise SystemExit("WSL proof checkout is missing .env and .env.example.")
        print("WSL secret env: bootstrapping .env from .env.example")
        _w._run_wsl_shell("cp .env.example .env", distro=distro, checkout=selected_checkout)

    if exists(".env.compose"):
        print("WSL compose env: using existing .env.compose")
        return

    if not exists(".env.compose.example"):
        raise SystemExit("WSL proof checkout is missing .env.compose and .env.compose.example.")
    print("WSL compose env: bootstrapping .env.compose from .env.compose.example")
    _w._run_wsl_shell("cp .env.compose.example .env.compose", distro=distro, checkout=selected_checkout)


def _ensure_wsl_output_block_path(*, distro: str = "", checkout: str = "") -> None:
    selected_checkout = _w._configured_checkout(checkout)
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
    _w._run_wsl_shell(command, distro=distro, checkout=selected_checkout)


def _wsl_preflight_lane(selected_lane: str) -> str:
    if selected_lane in _w.WSL_COMPOSE_OWNED_LANES:
        return "runtime"
    return selected_lane
