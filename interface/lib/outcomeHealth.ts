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
  if (health === "completed") return "Ready";
  if (health === "blocked") return "Needs recovery";
  if (health === "degraded") return "Needs review";
  if (health === "running") return "Working";
  if (health === "waiting") return "Waiting";
  if (health === "archived") return "Archived";
  return "Healthy";
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
