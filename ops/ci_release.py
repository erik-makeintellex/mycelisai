import time


RELEASE_PREFLIGHT_LANES = {
    "baseline": {"runtime_posture": False, "service_health": False, "live_backend": False},
    "runtime": {"runtime_posture": True, "service_health": False, "live_backend": False},
    "service": {"runtime_posture": False, "service_health": True, "live_backend": False},
    "release": {"runtime_posture": True, "service_health": True, "live_backend": True},
}


def run_service_check(c, *, live_backend, lifecycle, db_tasks, interface_tasks):
    errors = []
    print("=== SERVICE CHECK ===")
    print()

    if live_backend:
        print("[1/3] lifecycle.up --frontend=false --build=false")
        try:
            lifecycle.up.body(c, frontend=False, build=False)
            print("  OK")
        except SystemExit:
            errors.append("lifecycle up failed")
        print("[1.5/3] db.migrate")
        if db_tasks.schema_bootstrapped():
            print("  SKIP (cortex schema already compatible with the current runtime)")
        else:
            try:
                db_tasks.migrate.body(c)
                print("  OK")
            except SystemExit:
                errors.append("database migrate failed")

        print("[2/3] lifecycle.health")
        try:
            lifecycle.health.body(c)
            print("  OK")
        except SystemExit:
            errors.append("lifecycle health failed")

        print("[3/3] interface live-backend governed playwright")
        if errors:
            print("  SKIP (prerequisites failed)")
        else:
            try:
                interface_tasks.build.body(c)
                time.sleep(3)
                interface_tasks.e2e.body(
                    c,
                    project="chromium",
                    spec="e2e/specs/soma-governance-live.spec.ts",
                    live_backend=True,
                    workers="1",
                    server_mode="start",
                )
            except SystemExit:
                errors.append("interface live-backend governed playwright failed")
            else:
                print("  OK")
    else:
        print("[1/2] lifecycle.health")
        try:
            lifecycle.health.body(c)
            print("  OK")
        except SystemExit:
            errors.append("lifecycle health failed")
        print("[2/2] interface live-backend playwright")
        print("  SKIP (--live-backend not set)")

    print()
    if errors:
        print(f"SERVICE CHECK FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("SERVICE CHECK PASSED")


def run_toolchain_check(c, *, strict):
    print("=== TOOLCHAIN CHECK ===")
    go_result = c.run("go version", hide=True, warn=True)
    node_result = c.run("node -v", hide=True, warn=True)
    npm_result = c.run("npm -v", hide=True, warn=True)

    go_text = (go_result.stdout or "").strip()
    node_text = (node_result.stdout or "").strip()
    npm_text = (npm_result.stdout or "").strip()
    print(f"go:   {go_text or 'unavailable'}")
    print(f"node: {node_text or 'unavailable'}")
    print(f"npm:  {npm_text or 'unavailable'}")

    if go_result.exited != 0:
        raise SystemExit("TOOLCHAIN CHECK FAILED: go is unavailable.")

    locked_go_prefix = "go1.26"
    if locked_go_prefix not in go_text:
        message = f"Go version drift: expected {locked_go_prefix} (locked docs), found '{go_text}'."
        if strict:
            raise SystemExit(f"TOOLCHAIN CHECK FAILED: {message}")
        print(f"WARN: {message}")
    else:
        print("Go version matches lock policy.")


def run_entrypoint_check(c):
    print("=== ENTRYPOINT CHECK ===")
    primary = c.run("uv run inv -l", hide=True, warn=True)
    if primary.exited != 0:
        raise SystemExit("ENTRYPOINT CHECK FAILED: 'uv run inv -l' did not succeed.")
    print("uv run inv -l: OK")

    bare_uvx = c.run("uvx inv -l", hide=True, warn=True)
    bare_uvx_text = f"{bare_uvx.stdout or ''}{bare_uvx.stderr or ''}"
    expected_error = "does not provide any executables"
    if bare_uvx.exited == 0 or expected_error not in bare_uvx_text:
        raise SystemExit(
            "ENTRYPOINT CHECK FAILED: expected bare 'uvx inv -l' to fail with the package-executable message."
        )
    print("uvx inv -l: expected failure confirmed")

    compat = c.run("uvx --from invoke inv -l", hide=True, warn=True)
    if compat.exited != 0:
        raise SystemExit("ENTRYPOINT CHECK FAILED: 'uvx --from invoke inv -l' did not succeed.")
    print("uvx --from invoke inv -l: OK")
    print("ENTRYPOINT CHECK PASSED")


def release_preflight_clean_tree(c):
    status = c.run("git status --porcelain", hide=True, warn=True)
    dirty_lines = [ln for ln in (status.stdout or "").splitlines() if ln.strip()]
    if not dirty_lines:
        return
    print("Working tree is not clean:")
    preview = dirty_lines[:20]
    for ln in preview:
        print(f"  {ln}")
    if len(dirty_lines) > len(preview):
        print(f"  ... and {len(dirty_lines) - len(preview)} more")
    raise SystemExit("RELEASE PREFLIGHT FAILED: clean-tree requirement not met.")


def resolve_release_preflight_lane(lane, *, runtime_posture=False, service_health=False, live_backend=False):
    normalized_lane = (lane or "baseline").strip().lower()
    if normalized_lane not in RELEASE_PREFLIGHT_LANES:
        valid_lanes = ", ".join(sorted(RELEASE_PREFLIGHT_LANES))
        raise SystemExit(f"RELEASE PREFLIGHT FAILED: unsupported lane '{lane}'. Expected one of: {valid_lanes}.")

    resolved = dict(RELEASE_PREFLIGHT_LANES[normalized_lane])
    resolved["runtime_posture"] = resolved["runtime_posture"] or runtime_posture
    resolved["service_health"] = resolved["service_health"] or service_health or live_backend
    resolved["live_backend"] = resolved["live_backend"] or live_backend
    if resolved["live_backend"]:
        resolved["service_health"] = True
    return normalized_lane, resolved


def run_release_preflight(
    c,
    *,
    lane,
    e2e,
    strict_toolchain,
    service_health,
    live_backend,
    runtime_posture,
    runtime_posture_check,
    toolchain_check,
    baseline,
    service_check,
):
    resolved_lane, resolved = resolve_release_preflight_lane(
        lane,
        runtime_posture=runtime_posture,
        service_health=service_health,
        live_backend=live_backend,
    )
    stages = [
        ("clean-tree", lambda: release_preflight_clean_tree(c)),
        ("toolchain-check", lambda: toolchain_check.body(c, strict=strict_toolchain)),
    ]
    if resolved["runtime_posture"]:
        stages.append(("runtime-posture", lambda: runtime_posture_check(c)))
    stages.append(("baseline", lambda: baseline.body(c, e2e=e2e)))
    if resolved["service_health"]:
        stages.append(("service-check", lambda: service_check.body(c, live_backend=resolved["live_backend"])))

    print(f"=== RELEASE PREFLIGHT ({resolved_lane}) ===")
    for index, (stage_name, runner) in enumerate(stages, start=1):
        print(f"[{index}/{len(stages)}] {stage_name}")
        runner()
    print("RELEASE PREFLIGHT PASSED")


def run_deploy(c, *, lint_task, test_task):
    from . import k8s

    print("=== DEPLOY ===")
    print()
    lint_task(c)
    test_task(c)
    k8s.deploy(c)
    print()
    print("DEPLOY COMPLETE")
