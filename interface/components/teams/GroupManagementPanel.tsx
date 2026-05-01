"use client";

import { useEffect, useMemo, useState } from "react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { GroupWorkspacePanels } from "./GroupWorkspacePanels";
import {
  emptyGroupDraft,
  getData,
  isCompleteGroup,
  splitList,
  summarizeOutputs,
  type ApprovalPrompt,
  type Group,
  type GroupBucket,
  type GroupDraft,
  type Monitor,
} from "./groupWorkspaceTypes";
import { pickSelectedGroupId } from "./groupSelection";
import { filterGroups, useGroupRecordFilters } from "./useGroupRecordFilters";

export default function GroupManagementPanel({
  initialSelectedGroupId = null,
}: {
  initialSelectedGroupId?: string | null;
}) {
  const [groups, setGroups] = useState<Group[]>([]);
  const [monitor, setMonitor] = useState<Monitor | null>(null);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [outputs, setOutputs] = useState<Artifact[]>([]);
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [approvalPrompt, setApprovalPrompt] = useState<ApprovalPrompt | null>(
    null,
  );
  const [refreshing, setRefreshing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [broadcasting, setBroadcasting] = useState(false);
  const [archiving, setArchiving] = useState(false);
  const [draft, setDraft] = useState<GroupDraft>(emptyGroupDraft);
  const [broadcastMessage, setBroadcastMessage] = useState("");
  const { recordFilters, updateRecordFilters } = useGroupRecordFilters();

  const selectedGroup = useMemo(
    () => groups.find((group) => group.group_id === selectedGroupId) ?? null,
    [groups, selectedGroupId],
  );
  const filteredGroups = useMemo(
    () => filterGroups(groups, recordFilters),
    [groups, recordFilters],
  );
  const selectedGroupHiddenByFilters = Boolean(
    selectedGroup &&
      !filteredGroups.some((group) => group.group_id === selectedGroup.group_id),
  );
  const buckets = useMemo<GroupBucket[]>(
    () => [
      {
        id: "standing",
        title: "Standing groups",
        groups: filteredGroups.filter((group) => !group.expiry),
      },
      {
        id: "temporary",
        title: "Temporary groups",
        groups: filteredGroups.filter(
          (group) => !!group.expiry && !isCompleteGroup(group),
        ),
      },
      {
        id: "archived",
        title: "Completed records",
        groups: filteredGroups.filter((group) => isCompleteGroup(group)),
      },
    ],
    [filteredGroups],
  );
  const selectedGroupIsArchived = selectedGroup?.status === "archived";
  const outputSummary = useMemo(() => summarizeOutputs(outputs), [outputs]);

  const loadGroups = async () => {
    setRefreshing(true);
    setError(null);
    try {
      const [groupsRes, monitorRes] = await Promise.all([
        fetch("/api/v1/groups", { cache: "no-store" }),
        fetch("/api/v1/groups/monitor", { cache: "no-store" }),
      ]);
      if (!groupsRes.ok) throw new Error("Could not load groups.");
      const nextGroups = await getData<Group[]>(groupsRes);
      setGroups(nextGroups);
      setSelectedGroupId((current) =>
        pickSelectedGroupId(nextGroups, current, initialSelectedGroupId),
      );
      if (monitorRes.ok) setMonitor(await getData<Monitor>(monitorRes));
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : "Could not load groups.",
      );
    } finally {
      setRefreshing(false);
    }
  };

  useEffect(() => {
    void loadGroups();
  }, []);

  useEffect(() => {
    if (!initialSelectedGroupId) return;
    if (!groups.some((group) => group.group_id === initialSelectedGroupId))
      return;
    setSelectedGroupId(initialSelectedGroupId);
  }, [groups, initialSelectedGroupId]);

  useEffect(() => {
    if (selectedGroupId && groups.some((group) => group.group_id === selectedGroupId)) {
      return;
    }
    if (filteredGroups.length === 0) {
      setSelectedGroupId(null);
      return;
    }
    if (!selectedGroupId) {
      setSelectedGroupId(filteredGroups[0].group_id);
    }
  }, [filteredGroups, groups, selectedGroupId]);

  useEffect(() => {
    if (!selectedGroup) {
      setOutputs([]);
      return;
    }
    let cancelled = false;
    const loadOutputs = async () => {
      const res = await fetch(
        `/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/outputs?limit=8`,
        { cache: "no-store" },
      );
      if (cancelled) return;
      if (!res.ok) {
        setOutputs([]);
        return;
      }
      const items = await getData<Artifact[]>(res);
      if (!cancelled) setOutputs(Array.isArray(items) ? items : []);
    };
    void loadOutputs();
    return () => {
      cancelled = true;
    };
  }, [selectedGroup]);

  const createGroup = async () => {
    setSaving(true);
    setNotice(null);
    setError(null);
    try {
      const res = await fetch("/api/v1/groups", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: draft.name.trim(),
          goal_statement: draft.goalStatement.trim(),
          work_mode: draft.workMode,
          team_ids: splitList(draft.teamIDs),
          member_user_ids: splitList(draft.memberIDs),
          coordinator_profile: draft.coordinatorProfile.trim(),
          approval_policy_ref: draft.approvalPolicyRef.trim(),
          allowed_capabilities: splitList(draft.allowedCapabilities),
          expiry: draft.expiry ? new Date(draft.expiry).toISOString() : null,
          confirm_token: approvalPrompt?.confirm_token?.token ?? "",
        }),
      });
      const payload = await getData<ApprovalPrompt | Group>(res);
      if (res.status === 202) {
        setApprovalPrompt(payload as ApprovalPrompt);
        setNotice("Approval required before the group can be created.");
        return;
      }
      if (!res.ok) throw new Error("Could not create the group.");
      setApprovalPrompt(null);
      setNotice("Group created successfully.");
      setDraft(emptyGroupDraft);
      await loadGroups();
    } catch (createError) {
      setError(
        createError instanceof Error
          ? createError.message
          : "Could not create the group.",
      );
    } finally {
      setSaving(false);
    }
  };

  const broadcastToGroup = async () => {
    if (!selectedGroup || !broadcastMessage.trim()) return;
    setBroadcasting(true);
    setNotice(null);
    setError(null);
    try {
      const res = await fetch(
        `/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/broadcast`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ message: broadcastMessage.trim() }),
        },
      );
      if (!res.ok)
        throw new Error("Could not broadcast to the selected group.");
      setNotice("Broadcast queued for the selected group.");
      setBroadcastMessage("");
      await loadGroups();
    } catch (broadcastError) {
      setError(
        broadcastError instanceof Error
          ? broadcastError.message
          : "Could not broadcast to the selected group.",
      );
    } finally {
      setBroadcasting(false);
    }
  };

  const archiveSelectedGroup = async () => {
    if (!selectedGroup || !selectedGroup.expiry || selectedGroupIsArchived)
      return;
    setArchiving(true);
    setNotice(null);
    setError(null);
    try {
      const res = await fetch(
        `/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/status`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ status: "archived" }),
        },
      );
      if (!res.ok)
        throw new Error("Could not archive the selected temporary group.");
      const archivedGroup = await getData<Group>(res);
      setGroups((current) =>
        current.map((group) =>
          group.group_id === archivedGroup.group_id ? archivedGroup : group,
        ),
      );
      setSelectedGroupId(archivedGroup.group_id);
      setBroadcastMessage("");
      setNotice(
        "Temporary group archived. Retained outputs are still available for review.",
      );
    } catch (archiveError) {
      setError(
        archiveError instanceof Error
          ? archiveError.message
          : "Could not archive the selected temporary group.",
      );
    } finally {
      setArchiving(false);
    }
  };

  return (
    <GroupWorkspacePanels
      buckets={buckets}
      monitor={monitor}
      recordFilters={recordFilters}
      selectedGroup={selectedGroup}
      hiddenSelectedGroup={selectedGroupHiddenByFilters ? selectedGroup : null}
      selectedGroupId={selectedGroupId}
      outputs={outputs}
      outputSummary={outputSummary}
      draft={draft}
      notice={notice}
      error={error}
      approvalPrompt={approvalPrompt}
      refreshing={refreshing}
      saving={saving}
      broadcasting={broadcasting}
      archiving={archiving}
      broadcastMessage={broadcastMessage}
      onRefresh={() => void loadGroups()}
      onRecordFiltersChange={updateRecordFilters}
      onSelectGroup={setSelectedGroupId}
      onDraftChange={(patch) =>
        setDraft((current) => ({ ...current, ...patch }))
      }
      onCreateGroup={() => void createGroup()}
      onBroadcastMessageChange={setBroadcastMessage}
      onBroadcast={() => void broadcastToGroup()}
      onArchive={() => void archiveSelectedGroup()}
    />
  );
}
