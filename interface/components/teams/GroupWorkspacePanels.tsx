import { useEffect, useState } from "react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
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
import {
  type ApprovalPrompt,
  type Group,
  type GroupBucket,
  type GroupDraft,
  type GroupBroadcastResult,
  type GroupLifecycleItem,
  type GroupLifecycleReport,
  type GroupRecordFilters,
  type Monitor,
  type OutputSummary,
} from "./groupWorkspaceTypes";

type WorkspaceProps = {
  buckets: GroupBucket[];
  monitor: Monitor | null;
  lifecycleReport: GroupLifecycleReport | null;
  lifecycleByGroupId: Map<string, GroupLifecycleItem>;
  recordFilters: GroupRecordFilters;
  selectedGroup: Group | null;
  hiddenSelectedGroup: Group | null;
  selectedGroupId: string | null;
  initialSelectedGroupId: string | null;
  initialPanel: GroupWorkspacePanel | null;
  outputs: Artifact[];
  outputSummary: OutputSummary;
  draft: GroupDraft;
  notice: string | null;
  error: string | null;
  approvalPrompt: ApprovalPrompt | null;
  refreshing: boolean;
  saving: boolean;
  broadcasting: boolean;
  archiving: boolean;
  archivingExpired: boolean;
  clearOutputs: boolean;
  broadcastMessage: string;
  lastBroadcastResult: GroupBroadcastResult | null;
  onRefresh: () => void;
  onArchiveExpired: () => void;
  onRecordFiltersChange: (patch: Partial<GroupRecordFilters>) => void;
  onSelectGroup: (groupId: string) => void;
  onDraftChange: (patch: Partial<GroupDraft>) => void;
  onCreateGroup: () => void;
  onBroadcastMessageChange: (message: string) => void;
  onBroadcast: () => void;
  onArchive: () => void;
  onClearOutputsChange: (value: boolean) => void;
};

export function GroupWorkspacePanels(props: WorkspaceProps) {
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
    onRefresh,
    onArchiveExpired,
    onRecordFiltersChange,
    onSelectGroup,
    onDraftChange,
    onCreateGroup,
    onBroadcastMessageChange,
    onBroadcast,
    onArchive,
    onClearOutputsChange,
  } = props;
  const [activePanel, setActivePanel] = useState<GroupWorkspacePanel>(
    initialPanel ?? "overview",
  );

  useEffect(() => {
    if (initialPanel) setActivePanel(initialPanel);
  }, [initialPanel]);

  const selectPanel = (panel: GroupWorkspacePanel, groupId = selectedGroupId) => {
    setActivePanel(panel);
    updateRouteState(groupId, panel);
  };

  const selectGroup = (groupId: string) => {
    onSelectGroup(groupId);
    selectPanel("overview", groupId);
  };

  return (
    <section
      className="flex h-[calc(100dvh-4.5rem)] min-h-0 flex-col gap-3 overflow-hidden"
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
      <div className="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden rounded-2xl border border-cortex-border bg-cortex-surface p-3 lg:grid-cols-[minmax(260px,340px)_minmax(0,1fr)]">
        <GroupRail
          buckets={buckets}
          filters={recordFilters}
          hiddenSelectedGroup={hiddenSelectedGroup}
          lifecycleByGroupId={lifecycleByGroupId}
          selectedGroupId={selectedGroupId}
          onFiltersChange={onRecordFiltersChange}
          onSelectGroup={selectGroup}
        />
        <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg/30">
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
                className="rounded-2xl border border-cortex-border bg-cortex-surface p-3"
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

function updateRouteState(groupId: string | null, panel: GroupWorkspacePanel) {
  if (typeof window === "undefined") return;
  const nextUrl = new URL(window.location.href);
  if (groupId) nextUrl.searchParams.set("group_id", groupId);
  else nextUrl.searchParams.delete("group_id");
  nextUrl.searchParams.set("panel", panel);
  window.history.replaceState(
    window.history.state,
    "",
    `${nextUrl.pathname}${nextUrl.search}${nextUrl.hash}`,
  );
}
