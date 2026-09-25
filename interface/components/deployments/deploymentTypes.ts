// Types and small helpers for the Deployments page (RC-004E).
// Mirrors core/pkg/protocol/deployments.go and docs/API_REFERENCE.md#deployments.

export type DeploymentKind =
  | "standing_team"
  | "mission_team"
  | "group"
  | "planned_department";

export type DeploymentState =
  | "running"
  | "idle"
  | "planned"
  | "paused"
  | "archived"
  | "expired"
  | "error";

export type DeploymentScope = {
  kind: string;
  id?: string;
  name?: string;
};

export type DeploymentTool = {
  name: string;
  source: "internal" | "mcp" | "capability";
};

export type DeploymentMember = {
  id: string;
  role: string;
  profile?: string;
  provider?: string;
  model?: string;
  provider_origin?: string;
  provider_status?: string;
  engine_summary?: string;
  status?: string;
  last_heartbeat_at?: string | null;
  tools: DeploymentTool[];
};

export type DeploymentOutputsSummary = {
  retained: number;
  by_status: Record<string, number>;
  proof_status: string;
  last_output_at?: string | null;
};

export type DeploymentConfigRef = {
  kind: string;
  id: string;
  path: string;
};

export type DeploymentRecord = {
  id: string;
  kind: DeploymentKind;
  name: string;
  purpose?: string;
  what_runs: string;
  scope: DeploymentScope;
  origin: string;
  runtime_id?: string;
  lifecycle_state: DeploymentState;
  since?: string | null;
  expires_at?: string | null;
  last_activity?: string | null;
  members: DeploymentMember[];
  tools: DeploymentTool[];
  work_mode?: string;
  approval_policy_ref?: string;
  approval_posture?: string;
  outputs_summary?: DeploymentOutputsSummary | null;
  config_ref: DeploymentConfigRef;
};

export type DeploymentSourceStatus = {
  source: string;
  status: "ok" | "degraded" | "unavailable" | "restricted";
  detail?: string;
  count: number;
};

export type DeploymentList = {
  deployments: DeploymentRecord[];
  sources: DeploymentSourceStatus[];
  generated_at: string;
};

export type DeploymentKindFilter = "all" | DeploymentKind;
export type DeploymentStateFilter = "all" | DeploymentState;

export const DEPLOYMENT_KIND_LABELS: Record<DeploymentKind, string> = {
  standing_team: "Standing team",
  mission_team: "Mission team",
  group: "Group",
  planned_department: "Planned",
};

export const DEPLOYMENT_KIND_OPTIONS: Array<[DeploymentKindFilter, string]> = [
  ["all", "All"],
  ["standing_team", "Standing team"],
  ["mission_team", "Mission team"],
  ["group", "Group"],
  ["planned_department", "Planned"],
];

export const DEPLOYMENT_STATE_OPTIONS: Array<[DeploymentStateFilter, string]> = [
  ["all", "All"],
  ["running", "Running"],
  ["idle", "Idle"],
  ["planned", "Planned"],
  ["paused", "Paused"],
  ["archived", "Archived"],
  ["expired", "Expired"],
  ["error", "Error"],
];

/** Composite key for a deployment record, since ids are unique only within a kind. */
export const compositeDeploymentId = (record: { kind: string; id: string }): string =>
  `${record.kind}:${record.id}`;

export const getData = async <T>(res: Response): Promise<T> => {
  const payload = await res.json();
  return (
    payload && typeof payload === "object" && "data" in payload
      ? (payload as { data: T }).data
      : payload
  ) as T;
};

export const relativeTime = (value?: string | null) => {
  if (!value) return "not yet";
  const diff = Date.now() - new Date(value).getTime();
  if (!Number.isFinite(diff)) return "not yet";
  if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return `${Math.floor(diff / 86_400_000)}d ago`;
};

/** Unique "provider/model" strings actually in use across a deployment's members. */
export const uniqueProviderModels = (members: DeploymentMember[]): string[] => {
  const seen = new Set<string>();
  for (const member of members) {
    if (member.provider && member.model) seen.add(`${member.provider}/${member.model}`);
    else if (member.model) seen.add(member.model);
    else if (member.provider) seen.add(member.provider);
  }
  return Array.from(seen);
};

export const approvalPostureSentence = (record: DeploymentRecord): string => {
  if (record.approval_posture) return record.approval_posture;
  if (record.work_mode) return `Work mode: ${record.work_mode}.`;
  return "No approval posture recorded.";
};

export const outputsSummaryLabel = (summary?: DeploymentOutputsSummary | null): string => {
  if (!summary) return "unknown";
  const count = summary.retained;
  return `${count} retained output${count === 1 ? "" : "s"} (${summary.proof_status})`;
};

export const sourceNoticeText = (source: DeploymentSourceStatus): string => {
  if (source.detail) return source.detail;
  switch (source.status) {
    case "unavailable":
      return `${source.source} unavailable`;
    case "restricted":
      return `${source.source} hidden: requires admin access`;
    default:
      return `${source.source} degraded`;
  }
};
