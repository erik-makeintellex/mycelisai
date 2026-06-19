"use client";

import { useEffect, useMemo, useState } from "react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { GroupWorkspacePanels } from "./GroupWorkspacePanels";
import type { GroupWorkspacePanel } from "./GroupWorkspaceTabs";
import {
  buildGroupBuckets,
  emptyGroupDraft,
  errorMessage,
  getData,
  groupHiddenByFilters,
  splitList,
  summarizeOutputs,
  visibleGroupBroadcastResult,
  type ApprovalPrompt,
  type Group,
  type GroupBroadcastResult,
  type GroupDraft,
  type GroupLifecycleReport,
  type Monitor,
} from "./groupWorkspaceTypes";
import { pickSelectedGroupId } from "./groupSelection";
import { filterGroups, useGroupRecordFilters } from "./useGroupRecordFilters";

export default function GroupManagementPanel({
  initialSelectedGroupId = null,
  initialPanel = null,
}: {
  initialSelectedGroupId?: string | null;
  initialPanel?: GroupWorkspacePanel | null;
}) {
  const [groups, setGroups] = useState<Group[]>([]);
  const [monitor, setMonitor] = useState<Monitor | null>(null);
  const [lifecycleReport, setLifecycleReport] =
    useState<GroupLifecycleReport | null>(null);
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
  const [archivingExpired, setArchivingExpired] = useState(false);
  const [draft, setDraft] = useState<GroupDraft>(emptyGroupDraft);
  const [broadcastMessage, setBroadcastMessage] = useState("");
  const [lastBroadcastResult, setLastBroadcastResult] =
    useState<GroupBroadcastResult | null>(null);
  const { recordFilters, updateRecordFilters } = useGroupRecordFilters();

  const selectedGroup =
    groups.find((group) => group.group_id === selectedGroupId) ?? null;
  const filteredGroups = filterGroups(groups, recordFilters);
  const selectedGroupHiddenByFilters = groupHiddenByFilters(
    selectedGroup,
    filteredGroups,
  );
  const visibleBroadcastResult = visibleGroupBroadcastResult(
    lastBroadcastResult,
    selectedGroup,
  );
  const buckets = useMemo(
    () => buildGroupBuckets(filteredGroups),
    [filteredGroups],
  );
  const lifecycleByGroupId = useMemo(() => {
    const byID = new Map<string, GroupLifecycleReport["items"][number]>();
    lifecycleReport?.items.forEach((item) => byID.set(item.group_id, item));
    return byID;
  }, [lifecycleReport]);
  const loadGroups = async () => {
    setRefreshing(true);
    setError(null);
    try {
      const [groupsRes, monitorRes, lifecycleRes] = await Promise.all([
        fetch("/api/v1/groups", { cache: "no-store" }),
        fetch("/api/v1/groups/monitor", { cache: "no-store" }),
        fetch("/api/v1/groups/lifecycle", { cache: "no-store" }),
      ]);
      if (!groupsRes.ok) throw new Error("Could not load groups.");
      const nextGroups = await getData<Group[]>(groupsRes);
      setGroups(nextGroups);
      setSelectedGroupId((current) =>
        pickSelectedGroupId(nextGroups, current, initialSelectedGroupId),
      );
      if (monitorRes.ok) setMonitor(await getData<Monitor>(monitorRes));
      if (lifecycleRes.ok)
        setLifecycleReport(await getData<GroupLifecycleReport>(lifecycleRes));
    } catch (loadError) {
      setError(errorMessage(loadError, "Could not load groups."));
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
    if (
      selectedGroupId &&
      groups.some((group) => group.group_id === selectedGroupId)
    ) {
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
          workspace_folder: draft.workspaceFolder.trim(),
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
      setError(errorMessage(createError, "Could not create the group."));
    } finally {
      setSaving(false);
    }
  };

  const broadcastToGroup = async () => {
    if (!selectedGroup || !broadcastMessage.trim()) return;
    setBroadcasting(true);
    setLastBroadcastResult(null);
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
      const payload = await getData<GroupBroadcastResult>(res);
      if (!res.ok)
        throw new Error("Could not broadcast to the selected group.");
      setLastBroadcastResult(payload);
      setNotice("Broadcast queued for the selected group.");
      setBroadcastMessage("");
      await loadGroups();
    } catch (broadcastError) {
      setError(
        errorMessage(
          broadcastError,
          "Could not broadcast to the selected group.",
        ),
      );
    } finally {
      setBroadcasting(false);
    }
  };

  const archiveSelectedGroup = async () => {
    if (
      !selectedGroup ||
      !selectedGroup.expiry ||
      selectedGroup.status === "archived"
    )
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
      setLastBroadcastResult(null);
      setNotice(
        "Temporary group archived. Retained outputs are still available for review.",
      );
    } catch (archiveError) {
      setError(
        errorMessage(
          archiveError,
          "Could not archive the selected temporary group.",
        ),
      );
    } finally {
      setArchiving(false);
    }
  };

  const archiveExpiredGroups = async () => {
    setArchivingExpired(true);
    setNotice(null);
    setError(null);
    try {
      const res = await fetch("/api/v1/groups/lifecycle/archive-expired", {
        method: "POST",
      });
      const payload = await getData<{
        archived_count?: number;
        report?: GroupLifecycleReport;
      }>(res);
      if (!res.ok) throw new Error("Could not archive expired groups.");
      const archivedCount = payload.archived_count ?? 0;
      if (payload.report) setLifecycleReport(payload.report);
      setNotice(
        archivedCount > 0
          ? `${archivedCount} expired temporary group${archivedCount === 1 ? "" : "s"} archived. Retained outputs remain reviewable.`
          : "No expired temporary groups needed cleanup.",
      );
      await loadGroups();
    } catch (archiveError) {
      setError(
        errorMessage(archiveError, "Could not archive expired groups."),
      );
    } finally {
      setArchivingExpired(false);
    }
  };

  return (
    <GroupWorkspacePanels
      buckets={buckets}
      monitor={monitor}
      lifecycleReport={lifecycleReport}
      lifecycleByGroupId={lifecycleByGroupId}
      recordFilters={recordFilters}
      selectedGroup={selectedGroup}
      hiddenSelectedGroup={selectedGroupHiddenByFilters ? selectedGroup : null}
      selectedGroupId={selectedGroupId}
      initialSelectedGroupId={initialSelectedGroupId}
      initialPanel={initialPanel}
      outputs={outputs}
      outputSummary={summarizeOutputs(outputs)}
      draft={draft}
      notice={notice}
      error={error}
      approvalPrompt={approvalPrompt}
      refreshing={refreshing}
      saving={saving}
      broadcasting={broadcasting}
      archiving={archiving}
      archivingExpired={archivingExpired}
      broadcastMessage={broadcastMessage}
      lastBroadcastResult={visibleBroadcastResult}
      onRefresh={() => void loadGroups()}
      onArchiveExpired={() => void archiveExpiredGroups()}
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
