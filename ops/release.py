from __future__ import annotations

from invoke import Collection, task

from . import quality
from .version import read_semver, validate_release_tag

PROTECTED_BRANCHES = ("main", "dev")


def _parse_prunable_branches(merged_output: str, current_branch: str = "") -> list[str]:
    """Return local branches merged into main, excluding protected and checked-out branches."""
    prunable: list[str] = []
    for raw_line in (merged_output or "").splitlines():
        stripped = raw_line.strip()
        if not stripped:
            continue
        is_checked_out = stripped.startswith("*")
        branch_name = stripped.lstrip("*+ ").strip()
        if (
            not branch_name
            or is_checked_out
            or branch_name in PROTECTED_BRANCHES
            or (current_branch and branch_name == current_branch)
        ):
            continue
        prunable.append(branch_name)
    return prunable


@task
def status(c):
    """
    Report current branch, working tree state, VERSION, latest main tag, main...dev divergence,
    and merged local branches eligible for pruning.
    """
    branch_res = c.run("git rev-parse --abbrev-ref HEAD", hide=True)
    current_branch = branch_res.stdout.strip()

    porcelain_res = c.run("git status --porcelain", hide=True)
    tree_state = "clean" if not porcelain_res.stdout.strip() else "dirty"

    semver = read_semver()

    tag_res = c.run("git describe --tags --abbrev=0 main", hide=True, warn=True)
    latest_tag = tag_res.stdout.strip() if getattr(tag_res, "exited", 0) == 0 and tag_res.stdout.strip() else "none"

    div_res = c.run("git rev-list --left-right --count main...dev", hide=True, warn=True)
    divergence = div_res.stdout.strip() if getattr(div_res, "exited", 0) == 0 and div_res.stdout.strip() else "unknown"

    merged_res = c.run("git branch --merged main", hide=True)
    prunable = _parse_prunable_branches(merged_res.stdout, current_branch=current_branch)

    print(f"Branch: {current_branch} ({tree_state})")
    print(f"Tracked VERSION: {semver}")
    print(f"Latest tag on main: {latest_tag}")
    print(f"Divergence (main...dev): {divergence}")
    if prunable:
        print(f"Merged branches eligible for pruning ({len(prunable)}): {', '.join(prunable)}")
    else:
        print("Merged branches eligible for pruning (0): none")


@task(
    help={
        "remote": "Also prune stale remote-tracking refs via git fetch --prune (default: False).",
        "dry_run": "List merged branches without deleting them (default: False).",
    }
)
def prune_branches(c, remote=False, dry_run=False):
    """
    Delete local feature/archive branches already merged into main while protecting main, dev,
    and the currently checked-out branch.
    """
    merged_res = c.run("git branch --merged main", hide=True)
    prunable = _parse_prunable_branches(merged_res.stdout)

    if not prunable:
        print("No merged local branches to prune.")
    else:
        for branch_name in prunable:
            if dry_run:
                print(f"[DRY-RUN] Would delete merged branch: {branch_name}")
            else:
                c.run(f"git branch -d {branch_name}")

    if remote and not dry_run:
        c.run("git fetch --prune origin")


@task(
    help={
        "tag": "Annotated release or RC tag to create on main (e.g. v0.8.3-rc1 or v0.8.3).",
        "message": "Annotation message for the tag.",
        "source_branch": "Verified integration branch to promote into main (default: dev).",
        "prune": "Delete merged local feature branches after staging onto main (default: True).",
        "dry_run": "Validate gates and print planned git commands without mutating refs (default: False).",
    }
)
def stage(c, tag, message="", source_branch="dev", prune=True, dry_run=False):
    """
    Fast-forward stage a verified integration branch onto main, create an annotated release/RC tag,
    optionally prune merged feature branches, and return to the source branch.
    """
    validated_tag = validate_release_tag(tag)
    porcelain_res = c.run("git status --porcelain", hide=True)
    if porcelain_res.stdout.strip():
        raise SystemExit("Working tree must be clean before staging a release onto main.")

    quality.max_lines.body(c, limit=300)
    c.run("git diff --check", hide=True)

    tag_message = message.strip() if message and message.strip() else f"Release {validated_tag} staged on main from {source_branch}"

    if dry_run:
        print("[DRY-RUN] Planned release staging sequence:")
        print("  git checkout main")
        print(f"  git merge --ff-only {source_branch}")
        print(f'  git tag -a {validated_tag} -m "{tag_message}"')
        if prune:
            print("  uv run inv release.prune-branches")
        print(f"  git checkout {source_branch}")
        return

    c.run("git checkout main")
    c.run(f"git merge --ff-only {source_branch}")
    c.run(f'git tag -a {validated_tag} -m "{tag_message}"')
    if prune:
        prune_branches.body(c, remote=False, dry_run=False)
    c.run(f"git checkout {source_branch}")


ns = Collection("release")
ns.add_task(status, name="status")
ns.add_task(prune_branches, name="prune-branches")
ns.add_task(stage, name="stage")
