from __future__ import annotations

from tests.test_docs_links import (
    DOCS_MANIFEST,
    README,
    ROOT,
    V8_1_LIVING_ARCHITECTURE,
    V8_2_FULL_ARCHITECTURE,
    V8_DEV_STATE,
    V8_UI_API_CONTRACT,
)


def test_readme_exposes_layered_architecture_truth():
    text = README.read_text(encoding="utf-8")

    required_sections = [
        "## Active Delivery Target (V8.2 B2+)",
        "## V8.1 Foundation And Compatibility Baseline",
        "## Current Implementation State",
    ]
    missing_sections = [section for section in required_sections if section not in text]
    assert not missing_sections, f"README is missing required layered-truth sections: {missing_sections}"

    required_snippets = [
        "architecture/v8-2.md",
        "full actuation architecture",
        "V8.2 B2+ is the active delivery target",
        "V8.1 remains the foundation and compatibility baseline",
        "Actual implementation state lives in [.state/V8_DEV_STATE.md](.state/V8_DEV_STATE.md).",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, f"README is missing required layered-truth snippets: {missing_snippets}"


def test_v8_docs_cover_managed_exchange_security_foundation():
    expectations = {
        README: [
            "managed exchange is permissioned",
            "normalization into managed exchange does not imply unrestricted trust",
            "free-node release now includes foundational security boundaries",
        ],
        V8_1_LIVING_ARCHITECTURE: [
            "channels, threads, and exchange items carry explicit readers, writers, reviewers, participants, sensitivity classes, and downstream allowed-consumer metadata",
            "capability-producing outputs carry a capability id, risk class, trust class, and audit-ready publication metadata",
            "normalization into exchange does not imply unrestricted trust",
        ],
        V8_UI_API_CONTRACT: [
            "inspect-only trust, sensitivity, and review labels for managed exchange items where they matter operationally",
            "managed exchange security labels stay inspectable in advanced surfaces without leaking forbidden/internal-only security implementation detail into the default Soma-first UX",
        ],
        V8_2_FULL_ARCHITECTURE: [
            "managed exchange remains permissioned",
            "normalization into exchange does not imply unrestricted trust",
        ],
        V8_DEV_STATE: [
            "managed exchange security foundation now exists",
            "enterprise identity, approval workflows, and multi-user access management remain deferred",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Managed exchange security foundation docs are out of sync:\n" + "\n".join(missing)


def test_readme_defines_default_and_advanced_surface_contract():
    text = README.read_text(encoding="utf-8")

    required_sections = [
        "## Default And Advanced Surfaces",
    ]
    missing_sections = [section for section in required_sections if section not in text]
    assert not missing_sections, f"README is missing the default/advanced surface section: {missing_sections}"

    required_snippets = [
        "Default Operator Surface:",
        "Advanced Architecture / Runtime Surface:",
        "source-of-truth layers remain separate:",
        "the advanced architecture/runtime surface is now defined as a contract, but it is not fully implemented yet",
        "the default UX must stay simple and intent-first",
        "the advanced architecture/runtime surface must stay separate, make inheritance legible, and make config origin legible",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, (
        "README is missing required default-vs-advanced surface guidance: "
        f"{missing_snippets}"
    )


def test_readme_declares_development_contract():
    text = README.read_text(encoding="utf-8")

    required_sections = [
        "## Development Contract",
    ]
    missing_sections = [section for section in required_sections if section not in text]
    assert not missing_sections, f"README is missing the development contract section: {missing_sections}"

    required_snippets = [
        "A slice is not complete unless:",
        "- tests pass",
        "- documentation is updated where meaning changed",
        "- architecture alignment is verified across the layered truth surfaces",
        "`README.md` is the primary architecture inception document",
        "`architecture/v8-2.md` is the canonical full architecture",
        "`.state/V8_DEV_STATE.md` is the source of actual implementation truth",
        "all slices must update these surfaces when implementation, release posture, or target meaning changes",
        "end-of-slice reporting must explicitly state which tests ran, which docs changed, and which scoped docs were reviewed but left unchanged",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, f"README is missing required development-contract guidance: {missing_snippets}"


def test_v8_2_full_architecture_is_indexed_exposed_and_canonical():
    index_text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    manifest_text = DOCS_MANIFEST.read_text(encoding="utf-8")
    architecture_text = V8_2_FULL_ARCHITECTURE.read_text(encoding="utf-8")

    required_index_refs = [
        "V8.2 Full Production Architecture",
        "../../architecture/v8-2.md",
        "Full Production Architecture (Canonical Target)",
    ]
    missing_index_refs = [snippet for snippet in required_index_refs if snippet not in index_text]
    assert not missing_index_refs, (
        "Architecture library index is missing the V8.2 full-architecture references: "
        f"{missing_index_refs}"
    )

    required_manifest_refs = [
        'slug: "v8-2-full-production-architecture"',
        'label: "Full Architecture (V8.2)"',
        'path: "architecture/v8-2.md"',
    ]
    missing_manifest_refs = [snippet for snippet in required_manifest_refs if snippet not in manifest_text]
    assert not missing_manifest_refs, (
        "docsManifest is missing the V8.2 full-architecture entry: "
        f"{missing_manifest_refs}"
    )

    required_architecture_snippets = [
        "canonical full production architecture",
        "distributed, learning, capability-enabled, and acting system",
        "distributed execution target state",
        "learning, capability, and actuation target state",
        "Not all elements described here are implemented yet.",
        "V8.2/B2+ is the active delivery frame",
        "V8.1 living organization architecture remains the Soma-primary compatibility baseline",
        "default operator experience remains Soma-primary, intent-first, and simple by default",
        "advanced architecture/runtime control exists as a separate non-default layer for deep users",
    ]
    missing_architecture_snippets = [snippet for snippet in required_architecture_snippets if snippet not in architecture_text]
    assert not missing_architecture_snippets, (
        "V8.2 full architecture doc is missing required canonical-target coverage: "
        f"{missing_architecture_snippets}"
    )


def test_layered_docs_align_on_advanced_surface_contract():
    surfaces = {
        README: [
            "Advanced Architecture / Runtime Surface:",
            "source-of-truth layers remain separate:",
        ],
        V8_UI_API_CONTRACT: [
            "Advanced Architecture / Runtime Surface:",
            "The product must keep these layers distinct:",
            "advanced UI may explain deployment/env influence, but it must not become a second source of runtime truth",
        ],
        V8_1_LIVING_ARCHITECTURE: [
            "### 8.3 Advanced UI boundaries",
            "organization defaults and inheritance visibility",
            "sensitive deployment secrets stay file/env/config driven",
        ],
        V8_2_FULL_ARCHITECTURE: [
            "advanced architecture/runtime control exists as a separate non-default layer for deep users",
        ],
        V8_DEV_STATE: [
            "`NEXT` advanced architecture/runtime configuration remains a separate contract and implementation lane",
        ],
    }

    missing: list[str] = []
    for path, snippets in surfaces.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Layered docs are missing the advanced-surface contract:\n" + "\n".join(missing)


def test_v8_docs_keep_v8_1_baseline_and_v8_2_b2_target_distinct():
    index_text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    v8_1_text = V8_1_LIVING_ARCHITECTURE.read_text(encoding="utf-8")
    state_text = V8_DEV_STATE.read_text(encoding="utf-8")

    required_index_snippets = [
        "foundation and compatibility baseline",
        "canonical full production architecture, full actuation target, and current B2+ delivery frame",
    ]
    missing_index_snippets = [snippet for snippet in required_index_snippets if snippet not in index_text]
    assert not missing_index_snippets, (
        "Architecture library index is missing the baseline-vs-B2+ target distinction: "
        f"{missing_index_snippets}"
    )

    required_v8_1_snippets = [
        "This is the V8.1 foundation and compatibility baseline.",
        "active V8.2/B2+ delivery frame, canonical full production architecture, and full actuation target",
    ]
    missing_v8_1_snippets = [snippet for snippet in required_v8_1_snippets if snippet not in v8_1_text]
    assert not missing_v8_1_snippets, (
        "V8.1 living architecture doc is missing the current-release distinction: "
        f"{missing_v8_1_snippets}"
    )

    required_state_snippets = ["Development is now operating in the V8.2/B2+ delivery frame while preserving the V8.1 Soma-primary compatibility baseline.", "V8.2/B2+ is the active delivery target.", "V8.1 remains the foundation and compatibility baseline for the default Soma-primary operator surface.", "## Parallel Execution Overlay (2026-04-27)", "Backend/Auth Architecture", "Frontend/Auth UX", "Workflow Testing", "Runtime/MCP/Web Capability"]
    missing_state_snippets = [snippet for snippet in required_state_snippets if snippet not in state_text]
    assert not missing_state_snippets, (
        "V8 dev state is missing the current-release vs full-target distinction: "
        f"{missing_state_snippets}"
    )


def test_v8_dev_state_declares_architecture_synchronization_rule():
    text = V8_DEV_STATE.read_text(encoding="utf-8")

    required_snippets = [
        "## Architecture Synchronization Rule",
        "Every slice must:",
        "- update state",
        "- verify README alignment",
        "- verify V8.2 alignment",
        "a slice is not complete unless tests pass, documentation is updated where meaning changed, and architecture alignment is verified",
        "`COMPLETE` records accepted delivered work",
        "`ACTIVE` records work in progress",
        "`NEXT` records the next committed follow-on slices",
        "slice close-out should explicitly report tests run, docs changed, and docs reviewed unchanged for the touched scope",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, (
        "V8 dev state is missing the architecture synchronization rule: "
        f"{missing_snippets}"
    )


def test_v8_2_declares_development_alignment():
    text = V8_2_FULL_ARCHITECTURE.read_text(encoding="utf-8")

    required_snippets = [
        "## Development Alignment",
        "V8.2 is the full target.",
        "Implementation progresses incrementally from the V8.1 compatibility baseline through bounded B2+ modules toward this architecture.",
        "Documentation must always reflect:",
        "- what is implemented",
        "- what is planned",
        "no silent divergence is allowed between implementation, release contract, and full architecture target",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, (
        "V8.2 full architecture doc is missing the development-alignment contract: "
        f"{missing_snippets}"
    )
