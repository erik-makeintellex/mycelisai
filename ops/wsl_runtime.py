import os
import shutil
import subprocess
import urllib.error
import urllib.request
from pathlib import PurePosixPath

from invoke import Collection, task

from .config import ROOT_DIR, is_windows
from .wsl_runtime_helpers import (
    DEFAULT_COMPOSE_WAIT_TIMEOUT,
    DEFAULT_GUI_URL,
    DEFAULT_RELEASE_LANE,
    DEFAULT_WSL_CHECKOUT,
    DEFAULT_WSL_DISTRO,
    DEFAULT_WSL_REMOTE,
    GIT_AUTH_FAILURE_MARKERS,
    WSL_COMPOSE_OWNED_LANES,
    WSL_PROOF_LANES,
    WSL_REFRESH_CLEAN_EXCLUDES,
    CommandResult,
    _command_detail,
    _configured_checkout,
    _configured_distro,
    _configured_remote,
    _escaped_git_helper_path,
    _fetch_wsl_remote_with_auth_repair,
    _find_windows_gcm_helper,
    _git_output,
    _is_git_auth_failure,
    _is_github_https_remote,
    _maybe_configure_wsl_gcm_helper,
    _require_windows_dev_host,
    _run_command,
    _run_local_git,
    _run_wsl_git,
    _run_wsl_git_noninteractive,
    _run_wsl_shell,
    _wsl_command,
    _wsl_git_auth_guidance,
    _wsl_remote_url,
)
from .wsl_runtime_prereqs import (
    _assert_current_branch_published,
    _assert_windows_dev_checkout_is_ready_for_refresh,
    _clean_wsl_proof_checkout,
    _collect_local_state,
    _collect_wsl_state,
    _current_local_branch,
    _ensure_remote_branch_exists,
    _ensure_remote_contains_commit,
    _ensure_wsl_compose_env,
    _ensure_wsl_output_block_path,
    _local_upstream,
    _print_repo_state,
    _probe_windows_gui,
    _remote_branch_ref,
    _resolve_target,
    _rev_list_counts,
    _wsl_preflight_lane,
    _wsl_repo_guard,
)


ns = Collection("wsl")


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
