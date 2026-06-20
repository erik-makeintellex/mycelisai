from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
V8_DEV_STATE = ROOT / ".state/V8_DEV_STATE.md"
V8_3_OPERATIONAL_EMBODIMENT = ROOT / "docs" / "architecture-library" / "V8_3_OPERATIONAL_EMBODIMENT_PRD.md"
V8_3_RELEASE_BRIEF = ROOT / "docs" / "architecture-library" / "V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.md"
V8_3_MVP_UI_RUNTIME_PLAN = ROOT / "docs" / "architecture-library" / "V8_3_MVP_UI_RUNTIME_DELIVERY_PLAN.md"


def test_v8_3_docs_track_trusted_outcome_recovery_ownership_gates():
    surfaces = {
        V8_3_OPERATIONAL_EMBODIMENT: [
            "Ask\n-> Understand\n-> Approve\n-> Execute\n-> Deliver\n-> Trust\n-> Recover\n-> Revisit",
            "help a non-technical user own the outcome and trust the result later",
            "failure degrades safely, recovery is actionable",
        ],
        V8_3_RELEASE_BRIEF: [
            "Outcome Ownership Index proves a returning user can identify active outcomes, delivered outcomes, incomplete outcomes, attention needed, trust state, and next step within 15 seconds",
            "remaining gate is live run-event reconstruction",
        ],
        V8_3_MVP_UI_RUNTIME_PLAN: [
            "trusted-outcome-journey.spec.ts",
            "trusted-outcome-journey-live.spec.ts",
            "live confirm-action run events may still be empty",
            "default returned-output recovery ownership",
            "returning-user Outcome Ownership check answers active workflow, delivered output, incomplete/recovery work, attention/review count, trust/evidence, and next step within 15 seconds",
        ],
        V8_DEV_STATE: [
            "deterministic P0.8 full journey plus Outcome Ownership/default recovery ownership proof",
            "live confirm-action run events can still be empty",
            "Default returned-output recovery ownership is green in deterministic P0.8",
        ],
    }

    missing = [
        f"{path.relative_to(ROOT)} missing `{snippet}`"
        for path, snippets in surfaces.items()
        for snippet in snippets
        if snippet not in path.read_text(encoding="utf-8")
    ]

    assert not missing, (
        "V8.3 Trusted Outcome Journey docs/state are missing recovery ownership gates:\n"
        + "\n".join(missing)
    )
