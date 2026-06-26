from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CANONICAL_PRD = ROOT / "docs" / "architecture-library" / "MYCELIS_CANONICAL_PRD.md"
V8_DEV_STATE = ROOT / ".state/V8_DEV_STATE.md"


def test_canonical_prd_tracks_trusted_outcome_recovery_ownership_gates():
    prd = CANONICAL_PRD.read_text(encoding="utf-8")
    state = V8_DEV_STATE.read_text(encoding="utf-8")

    required_prd = [
        "Ask\n-> Understand\n-> Approve\n-> Execute\n-> Deliver\n-> Trust\n-> Recover\n-> Revisit",
        "OutcomeProject",
        "TeamRegistryEntry",
        "The defining product abstraction is the Outcome.",
        "The Outcome never serves the runtime.",
        "authority remains with approved Outcomes",
        "Every user-facing output package should expose",
        "what remains trusted",
        "what proof is invalid",
        "what requires operator attention",
        "MVP is complete when one canonical workflow feels excellent",
    ]
    required_state = [
        "deterministic P0.8 full journey plus Outcome Ownership/default recovery ownership proof",
        "non-empty run-event readback green",
        "MYCELIS_CANONICAL_PRD.md",
    ]

    missing = [snippet for snippet in required_prd if snippet not in prd]
    missing.extend([snippet for snippet in required_state if snippet not in state])

    assert not missing, "Trusted Outcome Journey docs/state are missing recovery ownership gates: " + str(missing)
