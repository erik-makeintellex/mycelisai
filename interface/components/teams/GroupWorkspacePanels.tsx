import { useEffect, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { CreateGroupPane } from "./CreateGroupPane";
import { GroupCommunicationPanel } from "./GroupCommunicationPanel";
import { GroupConfigPane } from "./GroupConfigPane";
import { GroupDetailPane } from "./GroupDetailPane";
import { GroupsHeader } from "./GroupsHeader";
import { OutputsPanel } from "./GroupOutputsPanel";
import { GroupRail } from "./GroupRail";
import { GroupWorkflowLog } from "./GroupWorkflowLog";
import {
  GroupWorkspaceTabs,
  type GroupWorkspacePanel,
} from "./GroupWorkspaceTabs";
import type { GroupWorkspacePanelsProps } from "./GroupWorkspacePanels.types";

export function GroupWorkspacePanels(props: GroupWorkspacePanelsProps) {
  const {
    buckets,
    monitor,
    lifecycleReport,
    lifecycleByGroupId,
    recordFilters,
    selectedGroup,
    hiddenSelectedGroup,
    selectedGroupId,
    initialPanel,
    outputs,
    outputSummary,
    includeInternalOutputs,
    draft,
    notice,
    error,
    approvalPrompt,
    refreshing,
    saving,
    broadcasting,
    archiving,
    archivingExpired,
    clearOutputs,
    broadcastMessage,
    lastBroadcastResult,
    bulkMode,
    bulkSelectedGroupIds,
    bulkActionPending,
    bulkClearOutputs,
    onRefresh,
    onArchiveExpired,
    onRecordFiltersChange,
    onToggleBulkMode,
    onToggleBulkGroup,
    onSelectAllVisibleBulkGroups,
    onClearBulkSelection,
    onBulkClearOutputsChange,
    onClearSelectedGroups,
    onSelectGroup,
    onDraftChange,
    onCreateGroup,
    onBroadcastMessageChange,
    onBroadcast,
    onArchive,
    onClearOutputsChange,
    onIncludeInternalOutputsChange,
  } = props;
  const [activePanel, setActivePanel] = useState<GroupWorkspacePanel>(
    initialPanel ?? "overview",
  );
  const [compactWorkspaceOpen, setCompactWorkspaceOpen] = useState(
    Boolean(selectedGroupId || initialPanel === "create"),
  );

  const selectPanel = (panel: GroupWorkspacePanel, groupId = selectedGroupId) => {
    setActivePanel(panel);
    if (panel === "create") setCompactWorkspaceOpen(true);
    updateRouteState(groupId, panel, "push");
  };

  const selectGroup = (groupId: string) => {
    onSelectGroup(groupId);
    setActivePanel("overview");
    setCompactWorkspaceOpen(true);
    updateRouteState(groupId, "overview", "push");
  };

  const showGroupList = () => {
    setCompactWorkspaceOpen(false);
    updateRouteState(null, null, "push");
  };

  useEffect(() => {
    const restoreRouteState = () => {
      const params = new URL(window.location.href).searchParams;
      const groupId = params.get("group_id");
      const panel = parseRoutePanel(params.get("panel"));
      if (groupId) onSelectGroup(groupId);
      setActivePanel(panel);
      setCompactWorkspaceOpen(Boolean(groupId) || panel === "create");
    };
    window.addEventListener("popstate", restoreRouteState);
    return () => window.removeEventListener("popstate", restoreRouteState);
  }, [onSelectGroup]);

  return (
    <section
      className="flex h-full min-h-0 flex-col gap-3 overflow-hidden"
      data-testid="groups-workspace"
    >
      <GroupsHeader
        monitor={monitor}
        lifecycleReport={lifecycleReport}
        refreshing={refreshing}
        archivingExpired={archivingExpired}
        onArchiveExpired={onArchiveExpired}
        onCreate={() => selectPanel("create")}
        onRefresh={onRefresh}
      />
      <div className="grid min-h-0 flex-1 grid-cols-1 grid-rows-[minmax(0,1fr)] gap-3 overflow-hidden rounded-2xl border border-cortex-border bg-cortex-surface p-3 lg:grid-cols-[minmax(260px,340px)_minmax(0,1fr)]">
        <div className={`${compactWorkspaceOpen ? "hidden" : "flex"} min-h-0 lg:flex`}>
          <GroupRail
            buckets={buckets}
            filters={recordFilters}
            hiddenSelectedGroup={hiddenSelectedGroup}
            lifecycleByGroupId={lifecycleByGroupId}
            bulkMode={bulkMode}
            selectedBulkGroupIds={bulkSelectedGroupIds}
            bulkActionPending={bulkActionPending}
            bulkClearOutputs={bulkClearOutputs}
            selectedGroupId={selectedGroupId}
            onFiltersChange={onRecordFiltersChange}
            onSelectGroup={selectGroup}
            onToggleBulkMode={onToggleBulkMode}
            onToggleBulkGroup={onToggleBulkGroup}
            onSelectAllVisible={onSelectAllVisibleBulkGroups}
            onClearBulkSelection={onClearBulkSelection}
            onBulkClearOutputsChange={onBulkClearOutputsChange}
            onBulkClearGroups={onClearSelectedGroups}
          />
        </div>
        <div className={`${compactWorkspaceOpen ? "flex" : "hidden"} min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg/30 lg:flex`}>
          <button
            type="button"
            onClick={showGroupList}
            className="flex min-h-11 items-center gap-2 border-b border-cortex-border px-3 text-sm font-semibold text-cortex-primary lg:hidden"
          >
            <ArrowLeft className="h-4 w-4" />
            All groups
          </button>
          <GroupWorkspaceTabs
            activePanel={activePanel}
            outputCount={outputSummary.artifactCount}
            onSelect={selectPanel}
          />
          {notice || error ? (
            <div className="border-b border-cortex-border px-3 py-2">
              {notice ? (
                <p
                  className="rounded-xl border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-sm text-cortex-primary"
                  data-testid="groups-notice"
                  role="status"
                  aria-live="polite"
                >
                  {notice}
                </p>
              ) : null}
              {error ? (
                <p
                  className="rounded-xl border border-cortex-danger/30 bg-cortex-danger/10 px-3 py-2 text-sm text-cortex-danger"
                  data-testid="groups-error"
                  role="alert"
                >
                  {error}
                </p>
              ) : null}
            </div>
          ) : null}
          <div className="min-h-0 min-w-0 flex-1 overflow-y-auto overflow-x-hidden p-3">
            {activePanel === "overview" ? (
              <div
                role="tabpanel"
                id="groups-overview-panel"
                aria-labelledby="groups-overview-tab"
              >
                <GroupDetailPane
                  selectedGroup={selectedGroup}
                  lifecycleItem={
                    selectedGroup
                      ? lifecycleByGroupId.get(selectedGroup.group_id)
                      : undefined
                  }
                  outputSummary={outputSummary}
                  archiving={archiving}
                  clearOutputs={clearOutputs}
                  onArchive={onArchive}
                  onClearOutputsChange={onClearOutputsChange}
                  onOpenOutputs={() => selectPanel("outputs")}
                />
              </div>
            ) : null}
            {activePanel === "outputs" ? (
              <div
                role="tabpanel"
                id="groups-outputs-panel"
                aria-labelledby="groups-outputs-tab"
              >
                <OutputsPanel
                  archived={selectedGroup?.status === "archived"}
                  outputs={outputs}
                  outputSummary={outputSummary}
                  includeInternalOutputs={includeInternalOutputs}
                  onIncludeInternalOutputsChange={
                    onIncludeInternalOutputsChange
                  }
                />
              </div>
            ) : null}
            {activePanel === "workflow" ? (
              <div
                role="tabpanel"
                id="groups-workflow-panel"
                aria-labelledby="groups-workflow-tab"
              >
                <GroupWorkflowLog
                  selectedGroup={selectedGroup}
                  lifecycleItem={
                    selectedGroup
                      ? lifecycleByGroupId.get(selectedGroup.group_id)
                      : undefined
                  }
                  outputs={outputs}
                  monitor={monitor}
                  lastBroadcastResult={lastBroadcastResult}
                  onOpenOutputs={() => selectPanel("outputs")}
                  onOpenMessage={() => selectPanel("message")}
                />
              </div>
            ) : null}
            {activePanel === "message" ? (
              <div
                role="tabpanel"
                id="groups-message-panel"
                aria-labelledby="groups-message-tab"
              >
                <GroupCommunicationPanel
                  monitor={monitor}
                  selectedGroup={selectedGroup}
                  broadcastMessage={broadcastMessage}
                  lastBroadcastResult={lastBroadcastResult}
                  broadcasting={broadcasting}
                  onBroadcastMessageChange={onBroadcastMessageChange}
                  onBroadcast={onBroadcast}
                  onRefresh={onRefresh}
                />
              </div>
            ) : null}
            {activePanel === "settings" ? (
              <div
                role="tabpanel"
                id="groups-settings-panel"
                aria-labelledby="groups-settings-tab"
              >
                <GroupConfigPane selectedGroup={selectedGroup} />
              </div>
            ) : null}
            {activePanel === "create" ? (
              <div
                role="tabpanel"
                id="groups-create-panel"
                aria-labelledby="groups-create-tab"
                className="min-w-0"
              >
                <CreateGroupPane
                  draft={draft}
                  approvalPrompt={approvalPrompt}
                  saving={saving}
                  onDraftChange={onDraftChange}
                  onCreateGroup={onCreateGroup}
                />
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </section>
  );
}

function parseRoutePanel(panel: string | null): GroupWorkspacePanel {
  if (
    panel === "workflow" ||
    panel === "outputs" ||
    panel === "message" ||
    panel === "settings" ||
    panel === "create"
  ) {
    return panel;
  }
  return "overview";
}

function updateRouteState(
  groupId: string | null,
  panel: GroupWorkspacePanel | null,
  mode: "push" | "replace",
) {
  if (typeof window === "undefined") return;
  const nextUrl = new URL(window.location.href);
  if (groupId) nextUrl.searchParams.set("group_id", groupId);
  else nextUrl.searchParams.delete("group_id");
  if (panel) nextUrl.searchParams.set("panel", panel);
  else nextUrl.searchParams.delete("panel");
  const updateHistory =
    mode === "push" ? window.history.pushState : window.history.replaceState;
  updateHistory.call(
    window.history,
    window.history.state,
    "",
    `${nextUrl.pathname}${nextUrl.search}${nextUrl.hash}`,
  );
}
