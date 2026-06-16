import type { ExecutionSummaryData, TeamOutputRef, TeamWorkItem } from "@/store/useCortexStore";
import type { RecoveryDegradationLike } from "@/lib/deliveryRuntimeLanguage";

export function buildTeamOutputRef(overrides: Partial<TeamOutputRef> = {}): TeamOutputRef {
  return {
    output_id: "output-1",
    team_id: "team-alpha",
    work_item_id: "work-1",
    kind: "project_package",
    label: "Generated package",
    storage_ref: "workspace/generated/package",
    entrypoint: "index.html",
    proof_ref: "proof-1",
    ...overrides,
  };
}

export function buildTeamWorkItem(overrides: Partial<TeamWorkItem> = {}): TeamWorkItem {
  return {
    id: "work-1",
    title: "Create a reviewable output",
    state: "output_ready",
    ownerLabel: "Team lead",
    scopeLabel: "Governed work",
    teamIds: ["team-alpha"],
    interactions: [
      { action: "inspect", label: "Open run", href: "/runs/run-1" },
      { action: "archive", label: "Clear from review" },
    ],
    outputRefs: [buildTeamOutputRef()],
    proofRefs: ["proof-1"],
    ...overrides,
  };
}

export function buildMediaDegradation(overrides: Partial<RecoveryDegradationLike> = {}): RecoveryDegradationLike {
  return {
    what_failed: "media engine error: local/private ComfyUI engine unreachable",
    trusted_state: "approval and run record remain",
    invalidated_proof: "no completed proof",
    safe_continuation: "retry after repair",
    ...overrides,
  };
}

export function buildExecutionSummary(overrides: Partial<ExecutionSummaryData> = {}): ExecutionSummaryData {
  return {
    execution: {
      status: "completed",
      summary: "Soma completed the approved proposal.",
    },
    outputs: [
      {
        kind: "project_package",
        label: "Generated package",
        folder: "workspace/generated/package",
        entrypoint: "index.html",
        retained: true,
      },
    ],
    proof: [
      {
        label: "Run run-1",
        run_id: "run-1",
        proof_class: "execution_proof",
        verified: true,
      },
    ],
    ...overrides,
  };
}
