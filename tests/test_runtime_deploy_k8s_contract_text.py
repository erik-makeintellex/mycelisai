from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
TESTING = ROOT / "docs" / "TESTING.md"
OPERATIONS = ROOT / "docs" / "architecture" / "OPERATIONS.md"
LICENSING = ROOT / "docs" / "licensing.md"
GOVERNANCE_TRUST = ROOT / "docs" / "user" / "governance-trust.md"
LOCAL_DEV_WORKFLOW = ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md"
REMOTE_USER_TESTING = ROOT / "docs" / "REMOTE_USER_TESTING.md"
API_REFERENCE = ROOT / "docs" / "API_REFERENCE.md"
BACKEND_ARCH = ROOT / "docs" / "architecture" / "BACKEND.md"
V8_DEV_STATE = ROOT / ".state/V8_DEV_STATE.md"


def test_k8s_docs_cover_promoted_values_file_contract():
    snippets = [
        (
            README,
            [
                "MYCELIS_K8S_VALUES_FILE",
                "charts/mycelis-core/values-k3d.yaml",
                "charts/mycelis-core/values-enterprise.yaml",
                "charts/mycelis-core/values-enterprise-windows-ai.yaml",
            ],
        ),
        (
            TESTING,
            [
                "MYCELIS_K8S_VALUES_FILE",
                "charts/mycelis-core/values-k3d.yaml",
                "charts/mycelis-core/values-enterprise.yaml",
                "charts/mycelis-core/values-enterprise-windows-ai.yaml",
            ],
        ),
        (
            OPERATIONS,
            [
                "MYCELIS_K8S_VALUES_FILE",
                "values-k3d.yaml",
                "values-enterprise.yaml",
                "values-enterprise-windows-ai.yaml",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Kubernetes preset-values contract is missing from active docs:\n" + "\n".join(missing)


def test_release_preflight_docs_prefer_lane_preset():
    snippets = [
        (README, ["uv run inv ci.release-preflight --lane=release"]),
        (
            TESTING,
            [
                "uv run inv ci.release-preflight --lane=release",
                "lane presets are `baseline`, `runtime`, `service`, and `release`",
            ],
        ),
        (
            OPERATIONS,
            [
                "--lane=baseline|runtime|service|release",
                "`--lane=release` is the recommended full runtime/operator gate",
            ],
        ),
        (LOCAL_DEV_WORKFLOW, ["uv run inv ci.release-preflight --lane=release"]),
        (REMOTE_USER_TESTING, ["uv run inv ci.release-preflight --lane=release"]),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Release-preflight lane contract is missing from active docs:\n" + "\n".join(missing)


def test_remote_user_testing_doc_covers_wsl_deployment_mimic_windows_browser_and_cold_start_notes():
    text = REMOTE_USER_TESTING.read_text(encoding="utf-8")

    required_snippets = [
        "clean WSL deployment-mimic checkout refreshed from git as the validation host",
        "Windows root repo as the dev/staging worktree",
        "verify `http://localhost:3000` from the Windows side with both a simple HTTP probe and a real browser launch",
        "classify it as `cold_start_first_request` instead of a clean first-pass success",
        "do not silently relabel the run as a clean first-pass success",
        "whether the issue is a `cold_start_first_request`, a steady-state regression, or an environment/setup gap",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "Remote user testing doc is missing WSL-proof/browser/cold-start guidance:\n" + "\n".join(missing)


def test_active_docs_reference_guarded_wsl_handoff_tasks():
    snippets = [
        (
            README,
            [
                "`uv run inv wsl.status`",
                "`uv run inv wsl.refresh`",
                "`uv run inv wsl.validate`",
                "`uv run inv wsl.cycle`",
            ],
        ),
        (
            TESTING,
            [
                "`uv run inv wsl.status`",
                "`uv run inv wsl.refresh`",
                "`uv run inv wsl.validate`",
                "`uv run inv wsl.cycle`",
            ],
        ),
        (
            OPERATIONS,
            [
                "`wsl.status`, `wsl.refresh`, `wsl.validate`, `wsl.cycle`",
            ],
        ),
        (
            LOCAL_DEV_WORKFLOW,
            [
                "`uv run inv wsl.status`",
                "`uv run inv wsl.refresh --branch <name>`",
                "`uv run inv wsl.validate`",
                "`uv run inv wsl.cycle --branch <name>`",
            ],
        ),
        (
            REMOTE_USER_TESTING,
            [
                "use a clean WSL deployment-mimic checkout refreshed from git as the validation host",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Guarded WSL handoff/proof tasks are missing from active docs:\n" + "\n".join(missing)


def test_release_proof_sequence_keeps_wsl_validate_before_browser_certification():
    snippets = [
        (
            TESTING,
            [
                "Release-proof sequencing rule:",
                "validate WSL git auth repair/report behavior for `wsl.refresh`",
                "run `uv run inv wsl.validate` from the refreshed WSL proof checkout before trusting browser-gap or certification evidence",
                "that task intentionally runs `ci.release-preflight --lane=runtime --no-e2e` first",
                "keep the newly closed focused browser proof gaps green: `/runs` workflow depth and guided Soma retry/recovery both have focused Chromium proof in production `start` mode",
                "rerun the broader headed Chromium certification pass only after the focused proof-hardening slice is committed and refreshed into WSL",
            ],
        ),
        (
            V8_DEV_STATE,
            [
                "guarded `uv run inv wsl.validate --lane=release` path from the refreshed `mother-brain` checkout",
                "maps `--lane=service` and `--lane=release` to `ci.release-preflight --lane=runtime --no-e2e`",
                "keep the focused `/runs` and guided retry/recovery browser proofs green, refresh the WSL proof checkout from the committed slice, run `wsl.validate`, and then rerun broader headed certification from committed state",
                "run `uv run inv wsl.validate` from the refreshed WSL proof checkout before accepting the new browser-gap evidence as authoritative",
                "live MCP workflow correlation is now green from the refreshed WSL proof checkout",
                "run `wsl.validate` from the refreshed proof checkout, keep the new `/runs` and guided Soma retry/recovery browser workflow proofs green, then rerun the broader headed Chromium certification pass from committed state",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Release-proof sequencing docs drifted from the active Slice 4 order:\n" + "\n".join(missing)


def test_windows_edit_wsl_proof_contract_does_not_turn_wsl_into_day_to_day_worktree():
    text = LOCAL_DEV_WORKFLOW.read_text(encoding="utf-8")

    required_snippets = [
        "For day-to-day Windows development, keep the Windows repo as the edit/review/push surface.",
        "Use this as the canonical deployment-mimic proof path when Windows editing is ready for authoritative build, API, UI, runtime, or release-style validation:",
        "keep the WSL `mother-brain` checkout git-backed and disposable for deployment-mimic proof",
        "These tasks keep the WSL proof checkout git-backed and disposable instead of turning it into a second editing worktree.",
    ]
    missing = [snippet for snippet in required_snippets if snippet not in text]

    assert not missing, "Local dev workflow is missing the Windows-edit/WSL-proof contract:\n" + "\n".join(missing)
    assert "prefer a WSL worktree plus the Compose path above" not in text, (
        "docs/LOCAL_DEV_WORKFLOW.md must not make a WSL worktree the day-to-day Windows edit surface"
    )


def test_identity_docs_describe_deploy_owned_people_access_contract():
    snippets = [
        (
            API_REFERENCE,
            [
                "deploy-owned People & Access contract surfaced read-only",
                "PUT ignores/preserves those deploy-owned fields instead of persisting them",
            ],
        ),
        (
            BACKEND_ARCH,
            [
                "deploy-owned People & Access posture surfaced read-only",
                "PUT ignores/preserves those deploy-owned fields instead of persisting them",
            ],
        ),
        (
            LICENSING,
            [
                "deploy-owned edition/auth posture now resolves from env or a deployment-contract file",
                "settings writes do not persist or override",
            ],
        ),
        (
            GOVERNANCE_TRUST,
            [
                "deploy-owned review state, not an ordinary user preference",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Deploy-owned People & Access contract docs are missing required wording:\n" + "\n".join(missing)
