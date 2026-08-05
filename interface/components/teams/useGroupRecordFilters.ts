import { useMemo, useSyncExternalStore } from "react";
import {
  defaultGroupRecordFilters,
  isCompleteGroup,
  isGroupWithinRetention,
  type Group,
  type GroupRecordFilters,
} from "./groupWorkspaceTypes";

const storageKey = "mycelis.groups.recordFilters";
const filtersChangedEvent = "mycelis:group-record-filters";

function subscribeRecordFilters(onChange: () => void) {
  window.addEventListener("storage", onChange);
  window.addEventListener(filtersChangedEvent, onChange);
  return () => {
    window.removeEventListener("storage", onChange);
    window.removeEventListener(filtersChangedEvent, onChange);
  };
}

function readRecordFiltersSnapshot() {
  return window.localStorage.getItem(storageKey) ?? "";
}

function parseSnapshot(raw: string): GroupRecordFilters {
  if (!raw) return defaultGroupRecordFilters;
  try {
    return parseRecordFilters(JSON.parse(raw));
  } catch {
    return defaultGroupRecordFilters;
  }
}

export function useGroupRecordFilters() {
  const snapshot = useSyncExternalStore(
    subscribeRecordFilters,
    readRecordFiltersSnapshot,
    () => "",
  );
  const recordFilters = useMemo(() => parseSnapshot(snapshot), [snapshot]);

  return {
    recordFilters,
    updateRecordFilters: (patch: Partial<GroupRecordFilters>) => {
      const next = parseRecordFilters({ ...recordFilters, ...patch });
      window.localStorage.setItem(storageKey, JSON.stringify(next));
      window.dispatchEvent(new Event(filtersChangedEvent));
    },
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
    const archived = group.status === "archived";
    const completed = !archived && isCompleteGroup(group);
    if (filters.state === "current" && (archived || completed)) return false;
    if (filters.state === "completed" && !completed) return false;
    if (filters.state === "archived" && !archived) return false;
    return filters.state === "current"
      ? true
      : isGroupWithinRetention(group, filters.retentionDays);
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
    value.state === "completed" || value.state === "archived"
      ? value.state
      : "current";
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
