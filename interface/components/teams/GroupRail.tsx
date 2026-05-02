import {
  inputClassName,
  type GroupBucket,
  type GroupKindFilter,
  type GroupRecordFilters,
  type GroupStateFilter,
  type Group,
} from "./groupWorkspaceTypes";

export function GroupRail({
  buckets,
  filters,
  hiddenSelectedGroup,
  selectedGroupId,
  onFiltersChange,
  onSelectGroup,
}: {
  buckets: GroupBucket[];
  filters: GroupRecordFilters;
  hiddenSelectedGroup: Group | null;
  selectedGroupId: string | null;
  onFiltersChange: (patch: Partial<GroupRecordFilters>) => void;
  onSelectGroup: (groupId: string) => void;
}) {
  const total = buckets.reduce(
    (count, bucket) => count + bucket.groups.length,
    0,
  );
  return (
    <aside className="flex min-h-0 flex-col rounded-2xl border border-cortex-border bg-cortex-surface p-3">
      <div className="flex items-center justify-between border-b border-cortex-border px-1 pb-3">
        <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
          Group records
        </h2>
        <span className="text-[11px] font-mono text-cortex-text-muted">
          {total}
        </span>
      </div>
      <GroupRecordFilterControls
        filters={filters}
        onFiltersChange={onFiltersChange}
      />
      {hiddenSelectedGroup ? (
        <div className="mt-3 rounded-xl border border-cortex-primary/25 bg-cortex-primary/10 p-2">
          <p className="px-1 font-mono text-[10px] uppercase tracking-[0.16em] text-cortex-primary">
            Selected outside filters
          </p>
          <GroupRecordButton
            group={hiddenSelectedGroup}
            selected
            onSelect={onSelectGroup}
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
                    selected={selectedGroupId === group.group_id}
                    onSelect={onSelectGroup}
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
  selected,
  onSelect,
}: {
  group: Group;
  selected: boolean;
  onSelect: (groupId: string) => void;
}) {
  return (
    <button
      type="button"
      data-testid={`groups-list-item-${group.group_id}`}
      onClick={() => onSelect(group.group_id)}
      className={`w-full rounded-xl px-3 py-2 text-left transition ${selected ? "bg-cortex-primary/10 text-cortex-text-main ring-1 ring-cortex-primary/30" : "text-cortex-text-muted hover:bg-cortex-bg hover:text-cortex-text-main"}`}
    >
      <span className="block truncate text-sm font-semibold">{group.name}</span>
      <span className="mt-0.5 flex items-center gap-2 text-[11px] font-mono uppercase tracking-[0.12em]">
        {group.status === "archived" ? "Archived" : group.work_mode}
        <span className="h-1 w-1 rounded-full bg-current opacity-50" />
        {group.team_ids.length} team{group.team_ids.length === 1 ? "" : "s"}
      </span>
    </button>
  );
}

function GroupRecordFilterControls({
  filters,
  onFiltersChange,
}: {
  filters: GroupRecordFilters;
  onFiltersChange: (patch: Partial<GroupRecordFilters>) => void;
}) {
  return (
    <div className="mt-3 space-y-3 rounded-xl border border-cortex-border bg-cortex-bg p-3">
      <label className="block text-xs">
        <span className="font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
          Search
        </span>
        <input
          aria-label="Search group records"
          value={filters.query}
          onChange={(event) => onFiltersChange({ query: event.target.value })}
          placeholder="Name, goal, team..."
          className={`${inputClassName} mt-2`}
        />
      </label>
      <FilterButtons<GroupKindFilter>
        label="Type"
        value={filters.kind}
        options={[
          ["all", "All"],
          ["standing", "Full time"],
          ["temporary", "Temp"],
        ]}
        onChange={(kind) => onFiltersChange({ kind })}
      />
      <FilterButtons<GroupStateFilter>
        label="State"
        value={filters.state}
        options={[
          ["all", "All"],
          ["running", "Running"],
          ["complete", "Complete"],
        ]}
        onChange={(state) => onFiltersChange({ state })}
      />
      <label className="block text-xs">
        <span className="font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
          Show completed records from last
        </span>
        <div className="mt-2 flex items-center gap-2">
          <input
            aria-label="Completed record retention days"
            type="number"
            min={1}
            max={3650}
            value={filters.retentionDays}
            onChange={(event) =>
              onFiltersChange({ retentionDays: Number(event.target.value) })
            }
            className={`${inputClassName} max-w-24`}
          />
          <span className="text-xs text-cortex-text-muted">days</span>
        </div>
      </label>
    </div>
  );
}

function FilterButtons<T extends string>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: Array<[T, string]>;
  onChange: (value: T) => void;
}) {
  return (
    <div>
      <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-cortex-text-muted">
        {label}
      </p>
      <div className="mt-2 grid grid-cols-3 gap-1">
        {options.map(([option, optionLabel]) => (
          <button
            key={option}
            type="button"
            onClick={() => onChange(option)}
            className={`rounded-lg border px-2 py-1.5 text-xs font-semibold transition ${
              value === option
                ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-primary"
                : "border-cortex-border bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main"
            }`}
          >
            {optionLabel}
          </button>
        ))}
      </div>
    </div>
  );
}
