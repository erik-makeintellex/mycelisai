import {
  inputClassName,
  type GroupKindFilter,
  type GroupRecordFilters,
  type GroupStateFilter,
} from "./groupWorkspaceTypes";

export function GroupRecordFilterControls({
  filters,
  onFiltersChange,
}: {
  filters: GroupRecordFilters;
  onFiltersChange: (patch: Partial<GroupRecordFilters>) => void;
}) {
  return (
    <details className="mt-3 rounded-xl border border-cortex-border bg-cortex-bg">
      <summary className="cursor-pointer px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
        Filters
      </summary>
      <div className="space-y-3 border-t border-cortex-border p-3">
        <label className="block text-xs">
          <span className="font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            Search
          </span>
          <input
            aria-label="Search group records"
            value={filters.query}
            onChange={(event) => onFiltersChange({ query: event.target.value })}
            placeholder="Name, goal, team..."
            suppressHydrationWarning
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
            ["current", "Current"],
            ["completed", "Completed"],
            ["archived", "Archived"],
          ]}
          onChange={(state) => onFiltersChange({ state })}
        />
        {filters.state === "completed" ? <label className="block text-xs">
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
              suppressHydrationWarning
              onChange={(event) =>
                onFiltersChange({ retentionDays: Number(event.target.value) })
              }
              className={`${inputClassName} max-w-24`}
            />
            <span className="text-xs text-cortex-text-muted">days</span>
          </div>
        </label> : null}
      </div>
    </details>
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
