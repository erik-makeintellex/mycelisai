import type { Group } from "./groupWorkspaceTypes";

export function pickSelectedGroupId(
  groups: Group[],
  current: string | null,
  requested: string | null,
) {
  if (requested && groups.some((group) => group.group_id === requested))
    return requested;
  if (current && groups.some((group) => group.group_id === current))
    return current;
  return groups[0]?.group_id ?? null;
}
