export type ScheduleHandoffStateTone = "pending" | "success" | "danger" | "neutral";

const HANDOFF_STATE_KEYS = [
  "schedule_handoff_state",
  "schedule_handoff_status",
  "handoff_state",
  "handoff_status",
  "approval_state",
  "proposal_status",
] as const;

export function getScheduleHandoffState(
  source?: object | null,
): string {
  if (!source) return "";
  const record = source as Record<string, unknown>;
  for (const key of HANDOFF_STATE_KEYS) {
    const value = record[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return "";
}

export function formatScheduleHandoffState(state: string): string {
  return state.replace(/[_-]+/g, " ");
}

export function scheduleHandoffTone(state: string): ScheduleHandoffStateTone {
  const normalized = state.toLowerCase().replace(/[\s-]+/g, "_");
  if (
    normalized === "approved" ||
    normalized === "executed" ||
    normalized === "confirmed" ||
    normalized === "confirmed_pending_execution"
  ) {
    return "success";
  }
  if (
    normalized === "rejected" ||
    normalized === "cancelled" ||
    normalized === "failed"
  ) {
    return "danger";
  }
  if (
    normalized === "pending" ||
    normalized === "proposed" ||
    normalized === "awaiting_approval" ||
    normalized === "approval_required"
  ) {
    return "pending";
  }
  return "neutral";
}
