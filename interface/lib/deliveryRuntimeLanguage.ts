import type { TeamWorkItemState } from "@/store/useCortexStore";

export type OutputFolderState = "idle" | "opening" | "opened" | "failed";
export type TeamWorkStateGroup = "not_started" | "running" | "needs_review" | "output_ready" | "needs_recovery" | "archived";

export interface RecoveryDegradationLike {
  what_failed?: string | null;
  trusted_state?: string | null;
  invalidated_proof?: string | null;
  safe_continuation?: string | null;
}

export interface RecoveryTrustCopy {
  failed: string;
  trusted: string;
  invalid: string;
  recovery: string;
}

export const TEAM_WORK_STATE_LABELS: Record<TeamWorkItemState, string> = {
  new: "Ready to brief",
  briefed: "Ready to start",
  queued: "Queued",
  running: "In progress",
  reviewing: "In review",
  paused: "Paused",
  output_ready: "Output ready",
  degraded: "Needs recovery",
  needs_operator: "Needs response",
  archived: "Archived",
};

export const RUN_RECEIPT_SECTION_LABELS = {
  outcome: "Outcome",
  status: "Status",
  output: "Output",
  trust: "Trust",
  proof: "Proof",
  recovery: "Recovery",
  inspect: "Inspect",
} as const;

export const OUTPUT_PACKAGE_ACTION_LABELS = {
  openFile: "Open file",
  openFolder: "Open folder",
  openInResources: "Open in Resources",
  download: "Download",
  copyPath: "Copy path",
  viewProof: "View proof",
} as const;

export const STALE_FAILED_PLAN_REVIEW_COPY = {
  title: "Old proposal cannot run",
  description: "Soma could not find an approved execution plan for this older proposal. Nothing changed, and there is no completed output to trust.",
  nextAction: "Clear this from review. Nothing ran, and there is no output to trust. Start a new Soma ask if you still want this work.",
  reason: "The original proposal is not executable anymore because no approved execution plan was retained for it.",
  trustedState: "Trusted: the failure record and audit trail. Not trusted: any implied output from this attempt.",
  recommendedChoice: "Clear from review if this is old test data. Inspect only if you need the failure details.",
} as const;

export const DEGRADED_TEAM_WORK_REVIEW_COPY = {
  title: "Team work needs recovery",
  description: "The team did not finish this work. The saved work item is still available, but no output should be trusted yet.",
  trustedState: "Trusted: retained context, proof refs, and status history. Not trusted: unfinished output.",
  reason: "The team did not finish cleanly, so the result needs recovery or cleanup before you rely on it.",
  recommendedChoice: "Recover when the runtime dependency is available, or archive if this attempt is no longer useful.",
} as const;

export const NEEDS_OPERATOR_REVIEW_COPY = {
  title: "Team needs your response",
  description: "The team is waiting for missing direction before it can continue.",
  trustedState: "Trusted: retained context, proof refs, and status history. Not trusted: unfinished output.",
  reason: "The team needs a decision or missing direction before it can safely continue.",
  recommendedChoice: "Respond or steer the work with the missing decision.",
} as const;

export const QUEUED_RECOVERY_REVIEW_COPY = {
  title: "Recovery request queued",
  description: "Soma has queued a recovery attempt for this work item. Wait for a new output or proof before trusting the result.",
} as const;

export const OUTPUT_READY_REVIEW_COPY = {
  trustedState: "Trusted after review: retained outputs, proof refs, and run history shown on this item.",
  reason: "The team produced retained output. Review it, then continue or archive the work item.",
  recommendedChoice: "Open the output, verify the proof, then archive or ask for follow-up.",
} as const;

export function teamWorkStateLabel(state: TeamWorkItemState) {
  return TEAM_WORK_STATE_LABELS[state];
}

export function teamWorkStateGroup(state: TeamWorkItemState): TeamWorkStateGroup {
  if (state === "new" || state === "briefed") return "not_started";
  if (state === "queued" || state === "running" || state === "paused") return "running";
  if (state === "reviewing" || state === "needs_operator") return "needs_review";
  if (state === "output_ready") return "output_ready";
  if (state === "degraded") return "needs_recovery";
  return "archived";
}

export function outputFolderButtonLabel(folderState: OutputFolderState, fallbackLabel: string) {
  if (folderState === "opening") return "Opening...";
  if (folderState === "opened") return "Folder opened";
  if (folderState === "failed") return "Open failed";
  return fallbackLabel;
}

export function outputPackageKindLabel(kind?: string | null, entrypoint?: string | null) {
  const normalized = kind?.trim().toLowerCase();
  if (normalized === "project_package" || entrypoint?.trim()) return "Project package";
  if (normalized === "image" || normalized === "media") return "Media output";
  if (normalized === "document" || normalized === "report") return "Document";
  if (normalized === "folder") return "Folder";
  if (normalized === "code" || normalized === "file") return "File";
  return "Output";
}

function compactText(value?: string | null) {
  const trimmed = value?.trim();
  return trimmed || null;
}

export function localMediaDependencyRecovery(degradation?: RecoveryDegradationLike | null): RecoveryTrustCopy | null {
  if (!degradation || typeof degradation !== "object") return null;
  const joined = [
    compactText(degradation.what_failed),
    compactText(degradation.trusted_state),
    compactText(degradation.invalidated_proof),
    compactText(degradation.safe_continuation),
  ].filter(Boolean).join(" ").toLowerCase();
  if (
    !joined.includes("comfyui")
    && !joined.includes("media engine")
    && !joined.includes("media capability")
    && !joined.includes("local/private")
  ) {
    return null;
  }
  return {
    failed: "Local media generation is not reachable, so Soma could not create the image output.",
    trusted: "The approval, request, failed run record, and audit trail remain available for review.",
    invalid: "No completed image output or execution proof should be trusted for this attempt.",
    recovery: "Start or reconnect the configured ComfyUI upstream, then retry. If you only need text/files, ask Soma to rerun without image generation.",
  };
}

export function recoveryTrustLines(copy: RecoveryTrustCopy) {
  return [
    copy.failed,
    `Still available: ${copy.trusted}`,
    `Not reliable: ${copy.invalid}`,
    `Safe next: ${copy.recovery}`,
  ];
}
