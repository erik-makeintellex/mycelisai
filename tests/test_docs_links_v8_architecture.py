from __future__ import annotations

from tests.test_docs_links import (
    DOCS_MANIFEST,
    LEGACY_V7_DEV_STATE,
    PRD_INDEX,
    README,
    ROOT,
    SOMA_COUNCIL_PROTOCOL,
    UI_OPERATOR_EXPERIENCE,
    V8_BOOTSTRAP_MODEL,
    V8_DEV_STATE,
)


def test_prd_index_points_to_modular_architecture_library():
    text = PRD_INDEX.read_text(encoding="utf-8")
    required_links = [
        "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
        "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
        "docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md",
        "docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md",
        "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
        "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
        "docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md",
    ]

    missing = [link for link in required_links if link not in text]
    assert not missing, f"PRD index is missing modular architecture links: {missing}"


def test_architecture_library_index_omits_temporary_plan_docs():
    text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    forbidden_refs = [
        "NEXT_EXECUTION_SLICES_V7.md",
        "MVP_RELEASE_STRIKE_TEAM_PLAN_V7.md",
        "MVP_INTEGRATION_AND_TOOLSHIP_EXECUTION_PLAN_V7.md",
        "UI_GENERATION_AND_TESTING_EXECUTION_PLAN_V7.md",
        "V8_FULL_TESTING_ACTION_PLAN.md",
        "V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md",
        "V8_UI_WORKFLOW_VERIFICATION_PLAN.md",
        "V8_STRUCTURED_TEAM_ASKS_AND_LANE_ROUTING_PLAN.md",
        "V8_GOVERNED_DEPLOYMENT_CONTEXT_AND_RAG_PLAN.md",
        "V8_RELEASE_PLATFORM_REVIEW_SECURITY_MONITORING_DEBUG.md",
    ]

    present = [ref for ref in forbidden_refs if ref in text]
    assert not present, f"Architecture library index still references temporary plan docs: {present}"


def test_legacy_v7_dev_state_uses_delivery_program_snapshot():
    text = LEGACY_V7_DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "### Delivery Program Snapshot",
        "P0  Operational foundation and gate discipline",
        "P1  Logging, error handling, and hot-path cleanup",
        "### Active Queue In Canonical Order",
        "Slice 1  Launch Crew and workflow onboarding execution contract",
        "Slice 2  P1 logging, error handling, and execution feedback",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Legacy V7 dev state is missing delivery-program framing: {missing}"


def test_ui_operator_experience_covers_direct_first_and_theme_simplification_contract():
    text = UI_OPERATOR_EXPERIENCE.read_text(encoding="utf-8")
    required_snippets = [
        "### 3.5 Soma-First conversation economy rule",
        "Workspace chat should default to Soma-only execution for normal interaction.",
        "### 3.6 Theme simplification rule",
        "diagnostics are progressive-disclosure, not always-on density",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "UI/operator experience doc is missing direct-first or theme-simplification contract snippets: "
        f"{missing}"
    )


def test_soma_council_protocol_defines_consultation_trigger_modes():
    text = SOMA_COUNCIL_PROTOCOL.read_text(encoding="utf-8")
    required_snippets = [
        "### 2.1 Consultation trigger policy (token economy)",
        "Council consultation is required only when at least one condition is true:",
        "- `none`: no council call, Soma responds directly",
        "- `targeted`: one specialist selected by need",
        "- `full_council`: only for explicitly broad architectural/program decisions",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Soma-council protocol is missing consultation trigger contract snippets: {missing}"


def test_legacy_v7_dev_state_tracks_active_slice2_theme_and_routing_priorities():
    text = LEGACY_V7_DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "1. `ACTIVE` Slice 2 sub-track: Workspace theme simplification and conversation-first hierarchy",
        "2. `ACTIVE` Slice 2 sub-track: Soma direct-first routing economy and consultation trigger discipline",
        "Where theme/routing are now encoded in architecture planning:",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Legacy V7 dev state is missing active Slice 2 theme/routing priorities: {missing}"


def test_canonical_surfaces_do_not_expose_purged_planning_docs():
    purged_paths = [
        "docs/UI_FRAMEWORK_V7.md",
        "docs/UI_ELEMENTS_PLANNING_V7.md",
        "docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md",
        "docs/ui-delivery/TEAM_ABCQ_EXECUTION_BOARD.md",
        "docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md",
        "docs/product/SOMA_EXTENSION_OF_SELF_PRD_V7.md",
    ]

    surfaces = {
        README: README.read_text(encoding="utf-8"),
        DOCS_MANIFEST: DOCS_MANIFEST.read_text(encoding="utf-8"),
        LEGACY_V7_DEV_STATE: LEGACY_V7_DEV_STATE.read_text(encoding="utf-8"),
        ROOT / "docs" / "README.md": (ROOT / "docs" / "README.md").read_text(encoding="utf-8"),
    }

    offenders: list[str] = []
    for path, text in surfaces.items():
        for purged in purged_paths:
            if purged in text:
                offenders.append(f"{path.relative_to(ROOT)} still references `{purged}`")

    assert not offenders, "Canonical surfaces still reference purged planning docs:\n" + "\n".join(offenders)


def test_v8_bootstrap_model_template_section_defines_blueprint_contract():
    text = V8_BOOTSTRAP_MODEL.read_text(encoding="utf-8")
    required_snippets = [
        "## Template and instantiation entry points",
        "Template = reusable blueprint",
        "Inception / AI Organization = actual instantiated organization",
        "Templates are reusable organization blueprints, not just UI presets.",
        "- organization type",
        "- default Team Lead / Soma Kernel posture",
        "- default Advisors / Council composition",
        "- default Departments / Teams",
        "- default Specialists / Agents",
        "- default AI Engine Settings / provider policy",
        "- default Memory & Personality settings",
        "- optional beginner-facing labels and descriptions",
        "By default, a template does not contain:",
        "- live runtime state",
        "- execution history",
        "- per-run outcomes",
        "- user-specific secrets",
        "- starter templates for beginners",
        "- domain templates such as Research, Engineering, and Marketing",
        "- executive templates such as CTO, COO, and Product Lead",
        "- personal / continuity templates",
        "- empty / minimal template",
        "### Template validation expectations",
        "Templates should be treated as governed blueprint inputs, not free-form UI metadata blobs.",
        "- declare enough structure to identify its organization type and intended operating posture",
        "- organize defaults along the same conceptual scopes used by bootstrap resolution",
        "- separate beginner-facing labels from runtime-shaping defaults",
        "- be resolvable into an instantiated organization without inventing a second hidden template model",
        "1. create from template",
        "2. create empty",
        "3. create from config/API",
        "4. clone and modify an existing template later",
        "1. the template supplies defaults",
        "2. the instantiated organization becomes its own object after creation",
        "3. later edits to the template do not silently rewrite existing organizations unless that behavior is explicitly designed and governed",
        "A template can:",
        "- define team defaults",
        "- define agent defaults",
        "- define optional advanced overrides",
        "Beginner UI should mainly show:",
        "- template name",
        "- purpose",
        "- a simple summary",
        "Advanced panels may expose:",
        "- internal structure",
        "- routing posture",
        "- scoped defaults",
        "### Target-delivery implication",
        "Templates must support target delivery rather than acting as decorative presets.",
        "- templates are bootstrap inputs for real organizations that should be capable of participating in the governed execution platform",
        "- instantiated organizations created from templates must still resolve into runtime behavior that can produce target product outcomes such as `answer`, `proposal`, `execution_result`, and `blocker`",
        "- the template system should not become a second planning-only layer that is disconnected from execution, governance, or operator-visible delivery behavior",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 bootstrap model is missing required template/instantiation contract snippets: "
        f"{missing}"
    )


def test_v8_dev_state_tracks_bootstrap_completion_and_validation_pass():
    text = V8_DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "### 17. Template and instantiation entry points definition",
        "the `Migration from V7 bootstrap assumptions` section now explains how fixed V7 startup assumptions collapse into explicit V8 configuration sources",
        "Task 004  Config and bootstrap model planning                       [COMPLETE]",
        "Task 008  Planning-integration validation pass                      [COMPLETE]",
        "Task 009  Next-execution/governance guidance migration              [NEXT]",
        "`COMPLETE` run the planning-integration validation pass so README, the architecture-library index, docs manifests, and doc-tests all confirm the new V7-to-V8 bootstrap migration contract.",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 dev state is missing the completed bootstrap plan or the new validation-pass next step: "
        f"{missing}"
    )


def test_v8_bootstrap_model_keeps_template_contract_linked_to_target_delivery():
    bootstrap_text = V8_BOOTSTRAP_MODEL.read_text(encoding="utf-8")
    target_text = (ROOT / "docs" / "architecture-library" / "TARGET_DELIVERABLE_V7.md").read_text(encoding="utf-8")

    required_target_outcomes = ["`answer`", "`proposal`", "`execution_result`", "`blocker`"]
    missing_target_outcomes = [marker for marker in required_target_outcomes if marker not in target_text]
    assert not missing_target_outcomes, (
        "Target deliverable doc is missing required product-outcome markers: "
        f"{missing_target_outcomes}"
    )

    required_bootstrap_linkage = [
        "Templates must support target delivery rather than acting as decorative presets.",
        "`answer`, `proposal`, `execution_result`, and `blocker`",
    ]
    missing_bootstrap_linkage = [snippet for snippet in required_bootstrap_linkage if snippet not in bootstrap_text]
    assert not missing_bootstrap_linkage, (
        "V8 bootstrap model is missing required target-delivery linkage for templates: "
        f"{missing_bootstrap_linkage}"
    )


def test_v8_bootstrap_model_defines_v7_to_v8_migration_contract():
    text = V8_BOOTSTRAP_MODEL.read_text(encoding="utf-8")
    required_snippets = [
        "## Migration from V7 bootstrap assumptions",
        "canonical V7->V8 migration contract",
        "V7 never exposed a single declarative bootstrap contract.",
        "**YAML manifests and sidecar config files**",
        "**Runtime configuration** (env vars, `.env`, CLI arguments)",
        "**Database state** produced by ad hoc migrations or interactive setup scripts",
        "**Operator flows** (UI wizards, CLI prompts) that mutated standing-team rows directly",
        "standing team definitions were hydrated automatically at process start if the database was empty",
        "Soma + Council roles were fixed to a single canonical lineup",
        "runtime state (runs, manifests, NATS registrations) doubled as bootstrap inputs",
        "1. **Templates** provide reusable blueprints but remain separate from instantiated organizations (`Template ≠ instantiated organization`).",
        "2. **Declarative configuration artifacts** (files, APIs, automation payloads) describe organization inputs",
        "3. **Operator flows** submit explicit organization creation or update intents that feed the same instantiation pipeline",
        "4. **Runtime/persistent state** supplies lineage and continuity only after an organization exists",
        "V7 YAML and manifest assets remain valid migration inputs, but they must be translated into V8 template/config packages before use.",
        "Auto-hydrating standing-team rows at process start.",
        "Treating last-run database state as the bootstrap plan.",
        "V8 keeps prior assets useful, but only after they conform to the explicit template + instantiation + inheritance + precedence pipeline.",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 bootstrap model migration section is missing required V7-to-V8 contract details: "
        f"{missing}"
    )
