import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, RefreshCw, Users } from "lucide-react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { CreateGroupPane } from "./CreateGroupPane";
import { GroupCommunicationPanel } from "./GroupCommunicationPanel";
import { GroupConfigPane } from "./GroupConfigPane";
import { GroupDetailPane } from "./GroupDetailPane";
import { OutputsPanel } from "./GroupOutputsPanel";
import { GroupRail } from "./GroupRail";
import {
  GroupWorkspaceTabs,
  type GroupWorkspacePanel,
} from "./GroupWorkspaceTabs";
import {
  compactButtonClassName,
  type ApprovalPrompt,
  type Group,
  type GroupBucket,
  type GroupDraft,
  type GroupBroadcastResult,
  type GroupRecordFilters,
  type Monitor,
  type OutputSummary,
} from "./groupWorkspaceTypes";

type WorkspaceProps = {
  buckets: GroupBucket[];
  monitor: Monitor | null;
  recordFilters: GroupRecordFilters;
  selectedGroup: Group | null;
  hiddenSelectedGroup: Group | null;
  selectedGroupId: string | null;
  initialSelectedGroupId: string | null;
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
  broadcastMessage: string;
  lastBroadcastResult: GroupBroadcastResult | null;
  onRefresh: () => void;
  onRecordFiltersChange: (patch: Partial<GroupRecordFilters>) => void;
  onSelectGroup: (groupId: string) => void;
  onDraftChange: (patch: Partial<GroupDraft>) => void;
  onCreateGroup: () => void;
  onBroadcastMessageChange: (message: string) => void;
  onBroadcast: () => void;
  onArchive: () => void;
};

export function GroupWorkspacePanels(props: WorkspaceProps) {
  const {
    buckets,
    monitor,
    recordFilters,
    selectedGroup,
    hiddenSelectedGroup,
    selectedGroupId,
    initialSelectedGroupId,
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
    broadcastMessage,
    lastBroadcastResult,
    onRefresh,
    onRecordFiltersChange,
    onSelectGroup,
    onDraftChange,
    onCreateGroup,
    onBroadcastMessageChange,
    onBroadcast,
  onArchive,
  } = props;
  const [activePanel, setActivePanel] = useState<GroupWorkspacePanel>(
    initialSelectedGroupId ? "overview" : "groups",
  );
  const selectGroup = (groupId: string) => {
    onSelectGroup(groupId);
    setActivePanel("overview");
  };

  return (
    <section
      className="flex h-[calc(100vh-2rem)] min-h-[640px] flex-col gap-3 overflow-hidden"
      data-testid="groups-workspace"
    >
      <GroupsHeader
        monitor={monitor}
        refreshing={refreshing}
        onRefresh={onRefresh}
      />
      <div className="flex min-h-0 flex-1 overflow-hidden rounded-2xl border border-cortex-border bg-cortex-surface">
        <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          <GroupWorkspaceTabs
            activePanel={activePanel}
            outputCount={outputSummary.artifactCount}
            onSelect={setActivePanel}
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
          <div className="min-h-0 flex-1 overflow-y-auto p-3">
            {activePanel === "groups" ? (
              <div
                role="tabpanel"
                id="groups-groups-panel"
                aria-labelledby="groups-groups-tab"
              >
                <GroupRail
                  buckets={buckets}
                  filters={recordFilters}
                  hiddenSelectedGroup={hiddenSelectedGroup}
                  selectedGroupId={selectedGroupId}
                  onFiltersChange={onRecordFiltersChange}
                  onSelectGroup={selectGroup}
                />
              </div>
            ) : null}
            {activePanel === "overview" ? (
              <div
                role="tabpanel"
                id="groups-overview-panel"
                aria-labelledby="groups-overview-tab"
              >
                <GroupDetailPane
                  selectedGroup={selectedGroup}
                  outputSummary={outputSummary}
                  archiving={archiving}
                  onArchive={onArchive}
                  onOpenOutputs={() => setActivePanel("outputs")}
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

function GroupsHeader({
  monitor,
  refreshing,
  onRefresh,
}: {
  monitor: Monitor | null;
  refreshing: boolean;
  onRefresh: () => void;
}) {
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
            <Users className="h-3.5 w-3.5" />
            Advanced Group Operations
          </div>
          <h1 className="mt-2 text-lg font-semibold text-cortex-text-main">
            Manage focused collaboration lanes.
          </h1>
          <p className="mt-1 max-w-3xl text-sm leading-5 text-cortex-text-muted">
            Use this page when Soma has created or needs a temporary or standing
            group. New users can stay in Soma; admins can filter, inspect,
            archive, and review retained outputs here.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link href="/dashboard" className={compactButtonClassName}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Open Soma
          </Link>
          <button
            type="button"
            onClick={onRefresh}
            className={compactButtonClassName}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${refreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </button>
        </div>
      </div>
      {monitor ? (
        <p className="mt-2 text-xs text-cortex-text-muted">
          Group lane monitor is {monitor.status || "offline"}.
        </p>
      ) : null}
    </div>
  );
}
