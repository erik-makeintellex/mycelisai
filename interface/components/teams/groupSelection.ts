import type { Group } from "./groupWorkspaceTypes";

export function pickSelectedGroupId(
  groups: Group[],
  current: string | null,
  requested: string | null,
  fallbackToFirst = true,
) {
  if (requested && groups.some((group) => group.group_id === requested))
    return requested;
  if (current && groups.some((group) => group.group_id === current))
    return current;
  return fallbackToFirst ? groups[0]?.group_id ?? null : null;
}
