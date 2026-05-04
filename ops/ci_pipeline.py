import time


def run_lint(c, *, core_dir, task_env, interface_tasks):
    errors = []
    print("=== LINT ===")
    print()

    print("[1/2] go vet ./...")
    with c.cd(str(core_dir)):
        result = c.run("go vet ./...", warn=True, env=task_env())
        if result.exited != 0:
            errors.append("go vet failed")
        else:
            print("  OK")

    print("[2/2] interface lint")
    try:
        interface_tasks.lint.body(c)
    except SystemExit:
        errors.append("next lint failed")
    else:
        print("  OK")

    print()
    if errors:
        print(f"LINT FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("LINT PASSED")


def run_test(c, *, core_dir, task_env, interface_tasks):
    errors = []
    print("=== TEST ===")
    print()

    print("[1/2] go test ./...")
    with c.cd(str(core_dir)):
        result = c.run("go test ./...", warn=True, env=task_env())
        if result.exited != 0:
            errors.append("go tests failed")
        else:
            print("  OK")

    print("[2/2] interface test")
    try:
        interface_tasks.test.body(c)
    except SystemExit:
        errors.append("interface tests failed")
    else:
        print("  OK")

    print()
    if errors:
        print(f"TEST FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("TESTS PASSED")


def run_build(c, *, cache_tasks, core_tasks, interface_tasks):
    errors = []
    print("=== BUILD ===")
    print()
    cache_tasks.ensure_disk_headroom(min_free_gb=10, reason="ci build")

    print("[1/2] core compile")
    try:
        core_tasks.compile.body(c)
    except SystemExit:
        errors.append("go build failed")
    else:
        print("  OK")

    print("[2/2] interface build")
    try:
        interface_tasks.build.body(c)
    except SystemExit:
        errors.append("next build failed")
    else:
        print("  OK")

    print()
    if errors:
        print(f"BUILD FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("BUILD PASSED")


def run_check(c, *, lint_task, test_task, build_task):
    start = time.time()
    print("=" * 60)
    print("  MYCELIS LOCAL CI PIPELINE")
    print("=" * 60)
    print()

    for name, fn in (("LINT", lint_task), ("TEST", test_task), ("BUILD", build_task)):
        stage_start = time.time()
        try:
            fn(c)
        except SystemExit:
            elapsed = time.time() - start
            print()
            print(f"PIPELINE FAILED at stage: {name} ({elapsed:.1f}s)")
            raise SystemExit(1)
        stage_elapsed = time.time() - stage_start
        print(f"  [{name} completed in {stage_elapsed:.1f}s]")
        print()

    elapsed = time.time() - start
    print("=" * 60)
    print(f"  PIPELINE PASSED ({elapsed:.1f}s)")
    print("=" * 60)


def _run_baseline_step(label: str, action, errors: list[str], failure: str):
    print(label)
    try:
        action()
    except SystemExit:
        errors.append(failure)
    else:
        print("  OK")


def run_baseline(
    c,
    *,
    e2e,
    cache_tasks,
    logging_tasks,
    quality,
    core_dir,
    task_env,
    interface_tasks,
):
    errors = []
    print("=== BASELINE ===")
    print()
    cache_tasks.ensure_disk_headroom(min_free_gb=10, reason="ci baseline")

    _run_baseline_step("[1/7] logging.check-schema", lambda: logging_tasks.check_schema.body(c), errors, "logging schema check failed")
    _run_baseline_step("[2/7] logging.check-topics", lambda: logging_tasks.check_topics.body(c), errors, "logging topic check failed")
    _run_baseline_step(
        "[3/7] quality.max-lines --limit=300",
        lambda: quality.max_lines.body(c, limit=300, paths=quality.DEFAULT_SOURCE_PATHS, strict=False),
        errors,
        "quality max-lines check failed",
    )

    print("[4/7] core go test ./... -count=1")
    with c.cd(str(core_dir)):
        result = c.run("go test ./... -count=1", warn=True, hide=True, env=task_env())
        if result.exited != 0:
            errors.append("core go tests failed")
        else:
            print("  OK")

    _run_baseline_step("[5/7] interface build", lambda: interface_tasks.build.body(c), errors, "interface build failed")
    _run_baseline_step("[6/7] interface typecheck", lambda: interface_tasks.typecheck.body(c), errors, "interface typecheck failed")

    print("[7/7] interface test")
    try:
        interface_tasks.stop.body(c)
        interface_tasks.clean.body(c)
        interface_tasks.test.body(c)
    except SystemExit:
        errors.append("interface vitest failed")
    else:
        print("  OK")

    print("[E2E] interface playwright run")
    if not e2e:
        print("  SKIP (--no-e2e)")
    elif errors:
        print("  SKIP (prerequisites failed)")
    else:
        try:
            interface_tasks.build.body(c)
            interface_tasks.e2e.body(c, workers="1", server_mode="start")
        except SystemExit:
            errors.append("interface playwright failed")
        else:
            print("  OK")

    print()
    if errors:
        print(f"BASELINE FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("BASELINE PASSED")
