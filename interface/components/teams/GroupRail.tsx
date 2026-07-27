import {
  type GroupBucket,
  type GroupLifecycleItem,
  type GroupRecordFilters,
  type Group,
} from "./groupWorkspaceTypes";
import { GroupRecordFilterControls } from "./GroupRecordFilterControls";

export function GroupRail({
  buckets,
  filters,
  hiddenSelectedGroup,
  lifecycleByGroupId,
  bulkMode,
  selectedBulkGroupIds,
  bulkActionPending,
  bulkClearOutputs,
  selectedGroupId,
  onFiltersChange,
  onSelectGroup,
  onToggleBulkMode,
  onToggleBulkGroup,
  onSelectAllVisible,
  onClearBulkSelection,
  onBulkClearOutputsChange,
  onBulkClearGroups,
}: {
  buckets: GroupBucket[];
  filters: GroupRecordFilters;
  hiddenSelectedGroup: Group | null;
  lifecycleByGroupId: Map<string, GroupLifecycleItem>;
  bulkMode: boolean;
  selectedBulkGroupIds: Set<string>;
  bulkActionPending: boolean;
  bulkClearOutputs: boolean;
  selectedGroupId: string | null;
  onFiltersChange: (patch: Partial<GroupRecordFilters>) => void;
  onSelectGroup: (groupId: string) => void;
  onToggleBulkMode: () => void;
  onToggleBulkGroup: (groupId: string) => void;
  onSelectAllVisible: () => void;
  onClearBulkSelection: () => void;
  onBulkClearOutputsChange: (value: boolean) => void;
  onBulkClearGroups: () => void;
}) {
  const total = buckets.reduce(
    (count, bucket) => count + bucket.groups.length,
    0,
  );
  const activeVisibleCount = buckets.reduce(
    (count, bucket) =>
      count + bucket.groups.filter((group) => group.status !== "archived").length,
    0,
  );
  const selectedCount = selectedBulkGroupIds.size;
  return (
    <aside className="flex h-full min-h-0 w-full flex-col rounded-2xl border border-cortex-border bg-cortex-surface p-3">
      <div className="flex items-center justify-between border-b border-cortex-border px-1 pb-3">
        <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
          Group records
        </h2>
        <div className="flex items-center gap-2">
          <span className="text-[11px] font-mono text-cortex-text-muted">
            {total}
          </span>
          <button
            type="button"
            onClick={onToggleBulkMode}
            className={`rounded-full border px-2 py-1 text-[11px] font-semibold transition ${
              bulkMode
                ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-primary"
                : "border-cortex-border bg-cortex-bg text-cortex-text-muted hover:text-cortex-text-main"
            }`}
          >
            {bulkMode ? "Done" : "Select"}
          </button>
        </div>
      </div>
      <GroupRecordFilterControls
        filters={filters}
        onFiltersChange={onFiltersChange}
      />
      {bulkMode ? (
        <div
          className="mt-3 rounded-xl border border-cortex-primary/25 bg-cortex-primary/10 p-3"
          data-testid="groups-bulk-actions"
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-xs font-semibold text-cortex-text-main">
              {selectedCount} selected
            </p>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={onSelectAllVisible}
                disabled={activeVisibleCount === 0}
                className="rounded-lg border border-cortex-border bg-cortex-bg px-2 py-1 text-xs font-semibold text-cortex-text-muted hover:text-cortex-text-main disabled:opacity-50"
              >
                Select visible active
              </button>
              <button
                type="button"
                onClick={onClearBulkSelection}
                disabled={selectedCount === 0}
                className="rounded-lg border border-cortex-border bg-cortex-bg px-2 py-1 text-xs font-semibold text-cortex-text-muted hover:text-cortex-text-main disabled:opacity-50"
              >
                Clear selection
              </button>
            </div>
          </div>
          <p className="mt-2 text-xs leading-5 text-cortex-text-muted">
            Bulk actions apply to selected active groups. Retained output files
            stay available unless you choose to remove them here.
          </p>
          <label className="mt-3 flex items-start gap-2 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2">
            <input
              type="checkbox"
              aria-label="Also delete retained output files for selected groups"
              checked={bulkClearOutputs}
              onChange={(event) =>
                onBulkClearOutputsChange(event.currentTarget.checked)
              }
              className="mt-1 h-4 w-4 rounded border-cortex-border bg-cortex-bg accent-cortex-warning"
            />
            <span className="min-w-0">
              <span className="block text-xs font-semibold text-cortex-text-main">
                Also delete retained output files
              </span>
              <span className="mt-1 block text-xs leading-5 text-cortex-text-muted">
                Keep this off when you only want old groups out of active lanes.
              </span>
            </span>
          </label>
          <button
            type="button"
            onClick={onBulkClearGroups}
            disabled={selectedCount === 0 || bulkActionPending}
            className="mt-3 rounded-lg border border-cortex-warning/40 bg-cortex-warning/10 px-3 py-2 text-xs font-semibold text-cortex-warning hover:bg-cortex-warning/15 disabled:opacity-50"
          >
            {bulkActionPending ? "Clearing..." : "Clear selected groups"}
          </button>
        </div>
      ) : null}
      {hiddenSelectedGroup ? (
        <div className="mt-3 rounded-xl border border-cortex-primary/25 bg-cortex-primary/10 p-2">
          <p className="px-1 font-mono text-[10px] uppercase tracking-[0.16em] text-cortex-primary">
            Selected outside filters
          </p>
          <GroupRecordButton
            group={hiddenSelectedGroup}
            lifecycleItem={lifecycleByGroupId.get(hiddenSelectedGroup.group_id)}
            selected
            bulkMode={bulkMode}
            bulkSelected={selectedBulkGroupIds.has(hiddenSelectedGroup.group_id)}
            onSelect={onSelectGroup}
            onToggleBulk={onToggleBulkGroup}
          />
        </div>
      ) : null}
      <div
        className="mt-3 min-h-0 flex-1 space-y-4 overflow-y-auto pr-1"
        data-testid="groups-list"
      >
        {buckets.map((bucket) => (
          <div key={bucket.id}>
            <div className="mb-2 flex items-center justify-between gap-2 px-1">
              <h3 className="text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
                {bucket.title}
              </h3>
              <span className="text-[11px] font-mono text-cortex-text-muted">
                {bucket.groups.length}
              </span>
            </div>
            <div className="space-y-1">
              {bucket.groups.length === 0 ? (
                <p className="px-2 py-1 text-xs text-cortex-text-muted">
                  Nothing here yet.
                </p>
              ) : (
                bucket.groups.map((group) => (
                  <GroupRecordButton
                    key={group.group_id}
                    group={group}
                    lifecycleItem={lifecycleByGroupId.get(group.group_id)}
                    selected={selectedGroupId === group.group_id}
                    bulkMode={bulkMode}
                    bulkSelected={selectedBulkGroupIds.has(group.group_id)}
                    onSelect={onSelectGroup}
                    onToggleBulk={onToggleBulkGroup}
                  />
                ))
              )}
            </div>
          </div>
        ))}
      </div>
    </aside>
  );
}

function GroupRecordButton({
  group,
  lifecycleItem,
  selected,
  bulkMode,
  bulkSelected,
  onSelect,
  onToggleBulk,
}: {
  group: Group;
  lifecycleItem?: GroupLifecycleItem;
  selected: boolean;
  bulkMode: boolean;
  bulkSelected: boolean;
  onSelect: (groupId: string) => void;
  onToggleBulk: (groupId: string) => void;
}) {
  const lifecycleLabel = lifecycleStatusLabel(lifecycleItem);
  const archived = group.status === "archived";
  if (bulkMode) {
    return (
      <label
        data-testid={`groups-list-item-${group.group_id}`}
        className={`flex w-full items-start gap-3 rounded-xl px-3 py-2 text-left transition ${
          bulkSelected
            ? "bg-cortex-primary/10 text-cortex-text-main ring-1 ring-cortex-primary/30"
            : "text-cortex-text-muted hover:bg-cortex-bg hover:text-cortex-text-main"
        } ${archived ? "opacity-60" : ""}`}
      >
        <input
          type="checkbox"
          aria-label={`Select ${group.name}`}
          checked={bulkSelected}
          disabled={archived}
          onChange={() => onToggleBulk(group.group_id)}
          className="mt-1 h-4 w-4 rounded border-cortex-border bg-cortex-bg accent-cortex-primary"
        />
        <GroupRecordText group={group} lifecycleLabel={lifecycleLabel} />
      </label>
    );
  }
  return (
    <button
      type="button"
      data-testid={`groups-list-item-${group.group_id}`}
      aria-current={selected ? "true" : undefined}
      onClick={() => onSelect(group.group_id)}
      className={`w-full rounded-xl px-3 py-2 text-left transition ${selected ? "bg-cortex-primary/10 text-cortex-text-main ring-1 ring-cortex-primary/30" : "text-cortex-text-muted hover:bg-cortex-bg hover:text-cortex-text-main"}`}
    >
      <GroupRecordText group={group} lifecycleLabel={lifecycleLabel} />
    </button>
  );
}

function GroupRecordText({
  group,
  lifecycleLabel,
}: {
  group: Group;
  lifecycleLabel: string | null;
}) {
  return (
    <span className="min-w-0 flex-1">
      <span className="block truncate text-sm font-semibold">{group.name}</span>
      <span className="mt-0.5 flex items-center gap-2 text-[11px] font-mono uppercase tracking-[0.12em]">
        {group.status === "archived" ? "Archived" : group.work_mode}
        <span className="h-1 w-1 rounded-full bg-current opacity-50" />
        {group.team_ids.length} team{group.team_ids.length === 1 ? "" : "s"}
      </span>
      {lifecycleLabel ? (
        <span className="mt-1 inline-flex max-w-full rounded-full border border-cortex-border bg-cortex-bg px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-cortex-text-muted">
          {lifecycleLabel}
        </span>
      ) : null}
    </span>
  );
}

function lifecycleStatusLabel(item?: GroupLifecycleItem) {
  switch (item?.recommendation) {
    case "archive_expired":
      return "Expired";
    case "review_work":
      if (item.output_ready_work_count > 0 && item.output_count === 0) {
        return "Planned only";
      }
      return `${item.active_or_blocked_work_count} work to review`;
    case "archive_completed":
      return "Output ready";
    case "review_standing":
      return "Review owner";
    default:
      return null;
  }
}
