import { useMemo, useState } from "react";
import { errorMessage, type Group } from "./groupWorkspaceTypes";

export function useGroupBulkActions({
  groups,
  filteredGroups,
  loadGroups,
  setNotice,
  setError,
}: {
  groups: Group[];
  filteredGroups: Group[];
  loadGroups: () => Promise<void>;
  setNotice: (message: string | null) => void;
  setError: (message: string | null) => void;
}) {
  const [bulkMode, setBulkMode] = useState(false);
  const [bulkClearing, setBulkClearing] = useState(false);
  const [bulkSelectedGroupIds, setBulkSelectedGroupIds] = useState<Set<string>>(
    () => new Set(),
  );
  const activeFilteredGroupIds = useMemo(
    () =>
      filteredGroups
        .filter((group) => group.status !== "archived")
        .map((group) => group.group_id),
    [filteredGroups],
  );

  const toggleBulkMode = () => {
    setBulkMode((current) => {
      if (current) setBulkSelectedGroupIds(new Set());
      return !current;
    });
  };

  const toggleBulkGroup = (groupId: string) => {
    const group = groups.find((item) => item.group_id === groupId);
    if (!group || group.status === "archived") return;
    setBulkSelectedGroupIds((current) => {
      const next = new Set(current);
      if (next.has(groupId)) {
        next.delete(groupId);
      } else {
        next.add(groupId);
      }
      return next;
    });
  };

  const selectAllVisibleBulkGroups = () => {
    setBulkSelectedGroupIds(new Set(activeFilteredGroupIds));
  };

  const clearBulkSelection = () => {
    setBulkSelectedGroupIds(new Set());
  };

  const clearSelectedGroups = async (includeOutputs = false) => {
    const groupIds = [...bulkSelectedGroupIds].filter((groupId) =>
      groups.some(
        (group) => group.group_id === groupId && group.status !== "archived",
      ),
    );
    if (groupIds.length === 0) return false;
    setBulkClearing(true);
    setNotice(null);
    setError(null);
    try {
      let cleared = 0;
      for (const groupId of groupIds) {
        const res = await fetch(
          `/api/v1/groups/${encodeURIComponent(groupId)}/clear`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ include_outputs: includeOutputs }),
          },
        );
        if (!res.ok) throw new Error("Could not clear selected groups.");
        cleared += 1;
      }
      setBulkSelectedGroupIds(new Set());
      setBulkMode(false);
      const outputCopy = includeOutputs
        ? "Retained output files were also cleared where available."
        : "Retained outputs stayed available.";
      setNotice(`${cleared} selected group${cleared === 1 ? "" : "s"} cleared from active lanes. ${outputCopy}`);
      await loadGroups();
      return true;
    } catch (bulkError) {
      setError(errorMessage(bulkError, "Could not clear selected groups."));
      return false;
    } finally {
      setBulkClearing(false);
    }
  };

  return {
    bulkMode,
    bulkClearing,
    bulkSelectedGroupIds,
    toggleBulkMode,
    toggleBulkGroup,
    selectAllVisibleBulkGroups,
    clearBulkSelection,
    clearSelectedGroups,
  };
}
