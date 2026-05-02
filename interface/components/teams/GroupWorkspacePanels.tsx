import { RefreshCw, Users } from "lucide-react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { CreateGroupPane } from "./CreateGroupPane";
import { GroupCommunicationPanel } from "./GroupCommunicationPanel";
import { GroupConfigPane } from "./GroupConfigPane";
import { GroupDetailPane } from "./GroupDetailPane";
import { GroupRail } from "./GroupRail";
import {
  compactButtonClassName,
  type ApprovalPrompt,
  type Group,
  type GroupBucket,
  type GroupDraft,
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
    onRefresh,
    onRecordFiltersChange,
    onSelectGroup,
    onDraftChange,
    onCreateGroup,
    onBroadcastMessageChange,
    onBroadcast,
    onArchive,
  } = props;

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
      <div className="grid min-h-0 flex-1 gap-3 xl:grid-cols-[360px_minmax(0,1fr)]">
        <GroupRail
          buckets={buckets}
          filters={recordFilters}
          hiddenSelectedGroup={hiddenSelectedGroup}
          selectedGroupId={selectedGroupId}
          onFiltersChange={onRecordFiltersChange}
          onSelectGroup={onSelectGroup}
        />
        <div className="min-h-0 min-w-0 space-y-4 overflow-y-auto pr-1">
          <GroupDetailPane
            selectedGroup={selectedGroup}
            outputs={outputs}
            outputSummary={outputSummary}
            archiving={archiving}
            onArchive={onArchive}
          />
          <GroupConfigPane selectedGroup={selectedGroup} />
          <GroupCommunicationPanel
            monitor={monitor}
            selectedGroup={selectedGroup}
            broadcastMessage={broadcastMessage}
            broadcasting={broadcasting}
            onBroadcastMessageChange={onBroadcastMessageChange}
            onBroadcast={onBroadcast}
            onRefresh={onRefresh}
          />
          <details
            className="rounded-2xl border border-cortex-border bg-cortex-surface p-3"
            open={!selectedGroup}
          >
            <summary className="cursor-pointer text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
              Create a new group
            </summary>
            <div className="mt-3">
              <CreateGroupPane
                draft={draft}
                notice={notice}
                error={error}
                approvalPrompt={approvalPrompt}
                saving={saving}
                onDraftChange={onDraftChange}
                onCreateGroup={onCreateGroup}
              />
            </div>
          </details>
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
            Groups Workspace
          </div>
          <h1 className="mt-2 text-lg font-semibold text-cortex-text-main">
            Create, review, and coordinate focused groups.
          </h1>
          <p className="mt-1 max-w-3xl text-sm leading-5 text-cortex-text-muted">
            Use groups as compact collaboration lanes while Soma stays the root
            admin chat.
          </p>
        </div>
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
      {monitor ? (
        <p className="mt-2 text-xs text-cortex-text-muted">
          Group lane monitor is {monitor.status || "offline"}.
        </p>
      ) : null}
    </div>
  );
}
