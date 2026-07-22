export type OutcomeHealthState =
  | "healthy"
  | "waiting"
  | "running"
  | "degraded"
  | "blocked"
  | "completed"
  | "archived";

export function normalizeOutcomeHealth(value: unknown): OutcomeHealthState {
  const raw = typeof value === "string" ? value.trim().toLowerCase() : "";
  if (
    raw === "healthy" ||
    raw === "waiting" ||
    raw === "running" ||
    raw === "degraded" ||
    raw === "blocked" ||
    raw === "completed" ||
    raw === "archived"
  ) {
    return raw;
  }
  return "healthy";
}

export function outcomeHealthLabel(health: OutcomeHealthState) {
  return health.charAt(0).toUpperCase() + health.slice(1);
}

export function outcomeHealthFromRunStatus(value: unknown): OutcomeHealthState {
  const raw = typeof value === "string" ? value.trim().toLowerCase() : "";
  if (raw === "archived") return "archived";
  if (["pending", "queued", "new", "briefed", "paused"].includes(raw)) return "waiting";
  if (["running", "reviewing"].includes(raw)) return "running";
  if (["completed", "succeeded", "success", "output_ready"].includes(raw)) return "completed";
  if (["degraded", "needs_attention"].includes(raw)) return "degraded";
  if (["failed", "blocked", "needs_operator", "cancelled", "canceled"].includes(raw)) return "blocked";
  return "healthy";
}

export function aggregateOutcomeHealth(states: OutcomeHealthState[]): OutcomeHealthState {
  if (states.length === 0) return "healthy";
  if (states.every((state) => state === "archived")) return "archived";
  for (const state of ["blocked", "degraded", "running", "completed", "waiting", "healthy"] as const) {
    if (states.includes(state)) return state;
  }
  return "healthy";
}

export function outcomeHealthClassName(health: OutcomeHealthState) {
  if (health === "blocked") return "border-cortex-danger/35 bg-cortex-danger/10 text-cortex-danger";
  if (health === "degraded" || health === "waiting") {
    return "border-cortex-warning/40 bg-cortex-warning/10 text-cortex-warning";
  }
  if (health === "running" || health === "completed") {
    return "border-cortex-primary/35 bg-cortex-primary/10 text-cortex-primary";
  }
  if (health === "archived") return "border-cortex-border bg-cortex-bg text-cortex-text-muted";
  return "border-cortex-success/35 bg-cortex-success/10 text-cortex-success";
}
