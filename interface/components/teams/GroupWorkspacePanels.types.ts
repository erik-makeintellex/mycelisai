import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import type { GroupWorkspacePanel } from "./GroupWorkspaceTabs";
import type {
  ApprovalPrompt,
  Group,
  GroupBroadcastResult,
  GroupBucket,
  GroupDraft,
  GroupLifecycleItem,
  GroupLifecycleReport,
  GroupRecordFilters,
  Monitor,
} from "./groupWorkspaceTypes";
import type { OutputSummary } from "./groupOutputClassification";

export type GroupWorkspacePanelsProps = {
  buckets: GroupBucket[];
  monitor: Monitor | null;
  lifecycleReport: GroupLifecycleReport | null;
  lifecycleByGroupId: Map<string, GroupLifecycleItem>;
  recordFilters: GroupRecordFilters;
  bulkMode: boolean;
  bulkSelectedGroupIds: Set<string>;
  bulkActionPending: boolean;
  selectedGroup: Group | null;
  hiddenSelectedGroup: Group | null;
  selectedGroupId: string | null;
  initialSelectedGroupId: string | null;
  initialPanel: GroupWorkspacePanel | null;
  outputs: Artifact[];
  outputSummary: OutputSummary;
  includeInternalOutputs: boolean;
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
  onToggleBulkMode: () => void;
  onToggleBulkGroup: (groupId: string) => void;
  onSelectAllVisibleBulkGroups: () => void;
  onClearBulkSelection: () => void;
  onClearSelectedGroups: () => void;
  onSelectGroup: (groupId: string) => void;
  onDraftChange: (patch: Partial<GroupDraft>) => void;
  onCreateGroup: () => void;
  onBroadcastMessageChange: (message: string) => void;
  onBroadcast: () => void;
  onArchive: () => void;
  onClearOutputsChange: (value: boolean) => void;
  onIncludeInternalOutputsChange: (value: boolean) => void;
};
