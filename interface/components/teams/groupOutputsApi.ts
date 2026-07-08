import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { getData } from "./groupWorkspaceTypes";

export async function loadGroupOutputs(
  groupId: string,
  includeInternalOutputs: boolean,
) {
  const params = new URLSearchParams({ limit: "8" });
  if (includeInternalOutputs) params.set("include_internal", "true");
  const res = await fetch(
    `/api/v1/groups/${encodeURIComponent(groupId)}/outputs?${params.toString()}`,
    { cache: "no-store" },
  );
  if (!res.ok) return [];
  const items = await getData<Artifact[]>(res);
  return Array.isArray(items) ? items : [];
}
