from __future__ import annotations

from tests.test_docs_links import README, ROOT


def test_docs_cover_env_override_contract_for_deployment_automation():
    expectations = {
        README: [
            "MYCELIS_PROVIDER_<PROVIDER_ID>_*",
            "MYCELIS_PROFILE_<PROFILE>_PROVIDER",
            "MYCELIS_MEDIA_*",
            "retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP`",
            "Bundle -> Instantiated Organization -> Inheritance -> Routing",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID",
            "MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT",
            "MYCELIS_PROFILE_<PROFILE>_PROVIDER",
            "MYCELIS_MEDIA_ENDPOINT",
            "retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP`",
            "Bundle -> Instantiated Organization -> Inheritance -> Routing",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID",
            "MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT",
            "MYCELIS_PROFILE_<PROFILE>_PROVIDER",
            "MYCELIS_MEDIA_MODEL_ID",
            "retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP`",
            "Bundle -> Instantiated Organization -> Inheritance -> Routing",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing the deployment env-override contract:\n" + "\n".join(missing)


def test_docs_keep_env_overrides_separate_from_runtime_organization_truth():
    expectations = {
        README: [
            "env overrides are deployment-time infrastructure wiring, not runtime organization behavior",
            "do not replace bundle-defined runtime organization truth",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "These env overrides are for deployment-time provider definition",
            "They are not runtime organization behavior",
            "do not replace bundle-defined truth",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "deployment-time infrastructure configuration only",
            "do not treat env overrides as runtime organization behavior",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing the env/runtime architecture boundary:\n" + "\n".join(missing)


def test_docs_cover_deployment_guidance_across_host_architectures():
    expectations = {
        README: [
            "Deployment guidance by host architecture:",
            "Windows x86_64",
            "Linux x86_64",
            "Linux arm64",
            "Mixed-architecture deployments",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "## Deployment Guidance By Host Architecture",
            "Windows x86_64",
            "Linux x86_64",
            "Linux arm64",
            "Mixed-architecture deployments",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "### Deployment Guidance Across Host Architectures",
            "Windows x86_64",
            "Linux x86_64",
            "Linux arm64",
            "Mixed-architecture deployment rule",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing deployment guidance across host architectures:\n" + "\n".join(missing)


def test_docs_cover_core_runtime_config_mount_alignment():
    expectations = {
        README: [
            "deployed Core image resolves runtime config from `/core/config`",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "deployed Core image resolves those files from `/core/config`",
            "Helm chart mounts the runtime config volume there",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "mounts this runtime config tree at `/core/config`",
        ],
        ROOT / "ops" / "README.md": [
            "startup config from `/core/config`",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing the Core runtime config mount alignment:\n" + "\n".join(missing)


def test_docs_cover_lifecycle_up_database_bootstrap_alignment():
    expectations = {
        ROOT / "docs" / "TESTING.md": [
            "ensures the `cortex` database exists",
            "managed built server",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "ensures the `cortex` database exists before Core starts",
        ],
        ROOT / "ops" / "README.md": [
            "ensures the `cortex` database exists before Core starts",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing lifecycle/database bootstrap alignment:\n" + "\n".join(missing)


def test_canonical_task_docs_cover_required_invoke_commands():
    expected = {
        README: [
            "uv run inv ci.entrypoint-check",
            "uv run inv ci.service-check",
            "uv run inv lifecycle.memory-restart",
            "uv run inv team.architecture-sync",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "uv run inv auth.dev-key",
            "uv run inv auth.break-glass-key",
            "uv run inv auth.posture",
            "uv run inv core.compile",
            "uv run inv interface.typecheck",
            "uv run inv lifecycle.memory-restart",
            "uv run inv logging.check-schema",
            "uv run inv quality.max-lines --limit 300",
            "uv run inv ci.service-check",
            "uv run inv ci.entrypoint-check",
            "uv run inv team.architecture-sync",
            "--live-backend",
        ],
        ROOT / "ops" / "README.md": [
            "uv run inv core.build",
            "uv run inv core.compile",
            "uv run inv interface.typecheck",
            "uv run inv auth.dev-key",
            "uv run inv auth.break-glass-key",
            "uv run inv auth.posture",
            "uv run inv lifecycle.health",
            "uv run inv ci.service-check",
            "uv run inv ci.entrypoint-check",
            "uv run inv team.architecture-sync",
        ],
    }

    missing: list[str] = []
    for path, commands in expected.items():
        text = path.read_text(encoding="utf-8")
        for command in commands:
            if command not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{command}`")

    assert not missing, "Canonical task docs are missing required invoke commands:\n" + "\n".join(missing)


def test_playwright_docs_reflect_owned_server_and_browser_matrix():
    expectations = {
        README: [
            "owns the local Next.js server lifecycle",
            "built production Interface server path",
            "chromium firefox webkit",
            "uv run inv ci.baseline` now includes Playwright by default",
            "uv run inv ci.service-check",
        ],
        ROOT / "docs" / "TESTING.md": [
            "starts/stops the managed Next.js app",
            "built production Interface server path",
            "mobile-chromium",
            "@axe-core/playwright",
            "workspace-live-backend.spec.ts",
            "--live-backend",
            "uv run inv ci.baseline --no-e2e",
            "uv run inv ci.service-check",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "Invoke manages the Next.js server lifecycle",
            "built production Interface server path",
            "mocked Chromium homepage smoke",
            "uv run inv ci.service-check",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

        if "requires running servers" in text:
            missing.append(f"{path.relative_to(ROOT)} still says Playwright requires running servers")

    assert not missing, "Playwright docs are missing required lifecycle/browser guidance:\n" + "\n".join(missing)


def test_active_docs_require_docs_review_as_part_of_slice_completion():
    required_snippets = {
        ROOT / "AGENTS.md": [
            "Every implementation slice that changes product behavior, runtime behavior, operator workflow, API contract, governance posture, or canonical terminology must include a documentation review in the same slice.",
            "When API behavior or payload meaning changes, review `docs/API_REFERENCE.md` in the same slice.",
        ],
        README: [
            "every implementation slice must include a docs review for the touched surface, even when the result is \"reviewed, no content change required\"",
            "`docs/API_REFERENCE.md` when API behavior, payload meaning, or endpoint contract changes",
        ],
        ROOT / "docs" / "TESTING.md": [
            "feature work is also not done until the touched docs are reviewed and updated where meaning changed",
            "end-of-slice reporting should name both the evidence commands run and the docs updated or reviewed unchanged for the touched scope",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "Implementation slices that change runtime, tasking, validation, API meaning, or operator behavior must review and update the owning docs in the same change rather than leaving docs drift for later cleanup.",
        ],
        ROOT / "ops" / "README.md": [
            "Task, runtime, or validation changes are not complete until the matching docs are reviewed and updated in the same slice.",
            "Delivery reporting should include the commands run plus the docs changed and the touched docs reviewed unchanged.",
        ],
    }

    missing: list[str] = []
    for path, snippets in required_snippets.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, (
        "Active docs are missing the docs-review/update contract for implementation slices:\n"
        + "\n".join(missing)
    )
