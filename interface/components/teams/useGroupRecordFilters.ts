import { useEffect, useState } from "react";
import {
  defaultGroupRecordFilters,
  isCompleteGroup,
  isGroupWithinRetention,
  type Group,
  type GroupRecordFilters,
} from "./groupWorkspaceTypes";

const storageKey = "mycelis.groups.recordFilters";

export function useGroupRecordFilters() {
  const [recordFilters, setRecordFilters] = useState<GroupRecordFilters>(
    defaultGroupRecordFilters,
  );

  useEffect(() => {
    try {
      const raw = window.localStorage.getItem(storageKey);
      if (!raw) return;
      setRecordFilters(parseRecordFilters(JSON.parse(raw)));
    } catch {
      setRecordFilters(defaultGroupRecordFilters);
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem(storageKey, JSON.stringify(recordFilters));
  }, [recordFilters]);

  return {
    recordFilters,
    updateRecordFilters: (patch: Partial<GroupRecordFilters>) =>
      setRecordFilters((current) =>
        parseRecordFilters({ ...current, ...patch }),
      ),
  };
}

export function filterGroups(groups: Group[], filters: GroupRecordFilters) {
  const query = filters.query.trim().toLowerCase();
  return groups.filter((group) => {
    const searchable =
      `${group.name} ${group.goal_statement} ${group.team_ids.join(" ")} ${group.coordinator_profile}`.toLowerCase();
    if (query && !searchable.includes(query)) return false;
    if (filters.kind === "standing" && group.expiry) return false;
    if (filters.kind === "temporary" && !group.expiry) return false;
    const complete = isCompleteGroup(group);
    if (filters.state === "running" && complete) return false;
    if (filters.state === "complete" && !complete) return false;
    return isGroupWithinRetention(group, filters.retentionDays);
  });
}

function parseRecordFilters(
  value: Partial<GroupRecordFilters>,
): GroupRecordFilters {
  const kind =
    value.kind === "standing" || value.kind === "temporary"
      ? value.kind
      : "all";
  const state =
    value.state === "running" || value.state === "complete"
      ? value.state
      : "all";
  const retentionDays = Number(value.retentionDays);
  return {
    query: typeof value.query === "string" ? value.query.slice(0, 120) : "",
    kind,
    state,
    retentionDays:
      Number.isFinite(retentionDays) && retentionDays > 0
        ? Math.min(Math.round(retentionDays), 3650)
        : 30,
  };
}
