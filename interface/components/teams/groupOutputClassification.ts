import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import type { GroupLifecycleItem } from "./groupWorkspaceTypes";

export type OutputSummary = {
  artifactCount: number;
  agentCount: number;
  deliveredCount: number;
  internalCount: number;
};

export type GroupArtifactClass =
  | "user_deliverable"
  | "planning"
  | "proof"
  | "internal_handoff";

export const summarizeOutputs = (outputs: Artifact[]): OutputSummary => {
  const uniqueAgents = new Set(
    outputs.map((artifact) => artifact.agent_id).filter(Boolean),
  );
  const deliveredCount = outputs.filter(isDeliveredArtifact).length;
  return {
    artifactCount: outputs.length,
    agentCount: uniqueAgents.size,
    deliveredCount,
    internalCount: outputs.length - deliveredCount,
  };
};

export const classifyGroupArtifact = (
  artifact: Artifact,
): GroupArtifactClass => {
  const metadata = artifact.metadata ?? {};
  const explicit =
    normalizeClass(artifact.output_class) ||
    metadataString(metadata, "output_class") ||
    metadataString(metadata, "delivery_status") ||
    metadataString(metadata, "visibility");
  if (explicit) {
    if (["planning", "plan", "draft"].includes(explicit)) return "planning";
    if (["proof", "evidence", "trust"].includes(explicit)) return "proof";
    if (["internal", "internal_handoff", "handoff"].includes(explicit)) {
      return "internal_handoff";
    }
    if (["source", "source_material", "support"].includes(explicit)) {
      return "internal_handoff";
    }
    if (["deliverable", "user_deliverable", "output"].includes(explicit)) {
      return "user_deliverable";
    }
  }
  if (metadata.internal === true) return "internal_handoff";
  const path = normalizeArtifactPath(
    artifact.file_path ||
      metadataString(metadata, "file_path") ||
      metadataString(metadata, "path") ||
      artifact.title,
  );
  const base = path.split("/").pop() ?? path;
  if (
    path.includes("/planning/") ||
    base === "team_evocation.md" ||
    base === "research_council_handoff.md"
  ) {
    return "planning";
  }
  if (path.includes("/proof/")) return "proof";
  if (
    path.includes("/watch/") ||
    path.includes("/source/") ||
    path.includes("/support/")
  ) {
    return "internal_handoff";
  }
  return "user_deliverable";
};

export const isDeliveredArtifact = (artifact: Artifact) =>
  classifyGroupArtifact(artifact) === "user_deliverable";

export const groupDeliveryStatusLabel = (
  lifecycleItem: GroupLifecycleItem | undefined,
  outputSummary: OutputSummary,
) => {
  if (outputSummary.deliveredCount > 0) return "Delivered";
  if ((lifecycleItem?.active_or_blocked_work_count ?? 0) > 0)
    return "Queued no delivery";
  if ((lifecycleItem?.output_ready_work_count ?? 0) > 0)
    return "Planned only";
  return "No delivery yet";
};

function normalizeArtifactPath(value: string | undefined) {
  return (value ?? "")
    .trim()
    .replaceAll("\\", "/")
    .replace(/^\/+|\/+$/g, "")
    .toLowerCase();
}

function normalizeClass(value: string | undefined) {
  return typeof value === "string" ? value.trim().toLowerCase() : "";
}

function metadataString(metadata: Record<string, any>, key: string) {
  const value = metadata[key];
  return typeof value === "string" ? value.trim().toLowerCase() : "";
}
