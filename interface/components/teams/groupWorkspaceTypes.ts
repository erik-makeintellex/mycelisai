import type { Artifact } from "@/store/cortexStoreTypesPlanning";

export type WorkMode =
  | "read_only"
  | "propose_only"
  | "execute_with_approval"
  | "execute_bounded";

export type Group = {
  group_id: string;
  name: string;
  goal_statement: string;
  work_mode: WorkMode;
  allowed_capabilities?: string[];
  member_user_ids: string[];
  team_ids: string[];
  coordinator_profile: string;
  approval_policy_ref: string;
  status: "active" | "paused" | "archived";
  expiry?: string | null;
  created_by: string;
  created_at: string;
};

export type Monitor = {
  status?: string;
  published_count?: number;
  last_group_id?: string;
  last_message?: string;
  last_published_at?: string;
  last_error?: string;
};

export type ApprovalPrompt = { confirm_token?: { token?: string } };

export type GroupBucket = {
  id: string;
  title: string;
  groups: Group[];
};

export type GroupKindFilter = "all" | "standing" | "temporary";

export type GroupStateFilter = "all" | "running" | "complete";

export type GroupRecordFilters = {
  query: string;
  kind: GroupKindFilter;
  state: GroupStateFilter;
  retentionDays: number;
};

export type OutputSummary = {
  artifactCount: number;
  agentCount: number;
};

export type GroupDraft = {
  name: string;
  goalStatement: string;
  workMode: WorkMode;
  expiry: string;
  teamIDs: string;
  memberIDs: string;
  coordinatorProfile: string;
  approvalPolicyRef: string;
  allowedCapabilities: string;
};

export const emptyGroupDraft: GroupDraft = {
  name: "",
  goalStatement: "",
  workMode: "propose_only",
  expiry: "",
  teamIDs: "",
  memberIDs: "",
  coordinatorProfile: "",
  approvalPolicyRef: "",
  allowedCapabilities: "",
};

export const defaultGroupRecordFilters: GroupRecordFilters = {
  query: "",
  kind: "all",
  state: "all",
  retentionDays: 30,
};

export const inputClassName =
  "w-full rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main outline-none placeholder:text-cortex-text-muted";

export const compactButtonClassName =
  "inline-flex items-center justify-center rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main hover:border-cortex-primary/25 disabled:opacity-60";

export const linkClassName =
  "rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm text-cortex-text-main hover:border-cortex-primary/25";

export const splitList = (value: string) =>
  value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

export const getData = async <T>(res: Response): Promise<T> => {
  const payload = await res.json();
  return (
    payload && typeof payload === "object" && "data" in payload
      ? payload.data
      : payload
  ) as T;
};

export const relativeTime = (value?: string | null) => {
  if (!value) return "not yet";
  const diff = Date.now() - new Date(value).getTime();
  if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
  return `${Math.floor(diff / 86_400_000)}d ago`;
};

export const groupKindLabel = (group: Group) => {
  if (group.status === "archived" && group.expiry)
    return "Archived temporary group";
  if (group.expiry) return "Temporary group";
  return "Standing group";
};

export const isCompleteGroup = (group: Group, now = Date.now()) => {
  if (group.status === "archived") return true;
  if (!group.expiry) return false;
  const expiryTime = new Date(group.expiry).getTime();
  return Number.isFinite(expiryTime) && expiryTime <= now;
};

export const isGroupWithinRetention = (
  group: Group,
  retentionDays: number,
  now = Date.now(),
) => {
  const days =
    Number.isFinite(retentionDays) && retentionDays > 0 ? retentionDays : 30;
  if (!isCompleteGroup(group, now)) return true;
  const reference = group.expiry || group.created_at;
  const referenceTime = new Date(reference).getTime();
  if (!Number.isFinite(referenceTime)) return true;
  return now - referenceTime <= days * 86_400_000;
};

export const summarizeOutputs = (outputs: Artifact[]): OutputSummary => {
  const uniqueAgents = new Set(
    outputs.map((artifact) => artifact.agent_id).filter(Boolean),
  );
  return {
    artifactCount: outputs.length,
    agentCount: uniqueAgents.size,
  };
};
