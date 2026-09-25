import {
  DEPLOYMENT_KIND_OPTIONS,
  DEPLOYMENT_STATE_OPTIONS,
  type DeploymentKindFilter,
  type DeploymentStateFilter,
} from "./deploymentTypes";

export function DeploymentFilters({
  kindFilter,
  stateFilter,
  onKindChange,
  onStateChange,
}: {
  kindFilter: DeploymentKindFilter;
  stateFilter: DeploymentStateFilter;
  onKindChange: (kind: DeploymentKindFilter) => void;
  onStateChange: (state: DeploymentStateFilter) => void;
}) {
  return (
    <div className="flex flex-col gap-3 rounded-xl border border-cortex-border bg-cortex-bg p-3">
      <ChipGroup
        label="Kind"
        value={kindFilter}
        options={DEPLOYMENT_KIND_OPTIONS}
        onChange={onKindChange}
        testIdPrefix="deployments-filter-kind"
      />
      <ChipGroup
        label="State"
        value={stateFilter}
        options={DEPLOYMENT_STATE_OPTIONS}
        onChange={onStateChange}
        testIdPrefix="deployments-filter-state"
      />
    </div>
  );
}

function ChipGroup<T extends string>({
  label,
  value,
  options,
  onChange,
  testIdPrefix,
}: {
  label: string;
  value: T;
  options: Array<[T, string]>;
  onChange: (value: T) => void;
  testIdPrefix: string;
}) {
  return (
    <div>
      <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-cortex-text-muted">
        {label}
      </p>
      <div className="mt-2 flex flex-wrap gap-1.5">
        {options.map(([option, optionLabel]) => (
          <button
            key={option}
            type="button"
            data-testid={`${testIdPrefix}-${option}`}
            aria-pressed={value === option}
            onClick={() => onChange(option)}
            className={`rounded-full border px-3 py-1 text-xs font-semibold transition ${
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
