import {
  DEPLOYMENT_KIND_LABELS,
  compositeDeploymentId,
  relativeTime,
  uniqueProviderModels,
  type DeploymentRecord,
} from "./deploymentTypes";

export function DeploymentList({
  deployments,
  selectedId,
  onSelect,
}: {
  deployments: DeploymentRecord[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}) {
  if (deployments.length === 0) {
    return (
      <p className="px-2 py-3 text-sm text-cortex-text-muted">
        No deployments match these filters.
      </p>
    );
  }
  return (
    <div className="min-h-0 flex-1 space-y-1 overflow-y-auto pr-1" data-testid="deployments-list">
      {deployments.map((record) => {
        const id = compositeDeploymentId(record);
        const models = uniqueProviderModels(record.members);
        return (
          <button
            key={id}
            type="button"
            data-testid={`deployments-list-item-${id}`}
            onClick={() => onSelect(id)}
            className={`w-full rounded-xl border px-3 py-2.5 text-left transition ${
              selectedId === id
                ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                : "border-cortex-border bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main"
            }`}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="truncate text-sm font-semibold">{record.name}</span>
              <KindBadge kind={record.kind} />
            </div>
            <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] font-mono uppercase tracking-[0.1em]">
              <StateBadge state={record.lifecycle_state} />
              <span className="opacity-60">
                {record.members.length} member{record.members.length === 1 ? "" : "s"}
              </span>
              {models.length > 0 ? (
                <span className="truncate opacity-60" title={models.join(", ")}>
                  {models.join(", ")}
                </span>
              ) : null}
              <span className="ml-auto opacity-60">{relativeTime(record.last_activity)}</span>
            </div>
          </button>
        );
      })}
    </div>
  );
}

export function KindBadge({ kind }: { kind: DeploymentRecord["kind"] }) {
  return (
    <span
      data-testid="deployments-kind-badge"
      className="flex-shrink-0 rounded-full border border-cortex-border bg-cortex-bg px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.1em] text-cortex-text-muted"
    >
      {DEPLOYMENT_KIND_LABELS[kind] ?? kind}
    </span>
  );
}

const STATE_COLORS: Record<string, string> = {
  running: "text-cortex-success",
  idle: "text-cortex-text-muted",
  planned: "text-cortex-primary",
  paused: "text-cortex-warning",
  archived: "text-cortex-text-muted",
  expired: "text-cortex-text-muted",
  error: "text-cortex-danger",
};

export function StateBadge({ state }: { state: string }) {
  return (
    <span data-testid="deployments-state-badge" className={STATE_COLORS[state] ?? "text-cortex-text-muted"}>
      {state}
    </span>
  );
}
