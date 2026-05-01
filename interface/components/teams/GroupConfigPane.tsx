import { SlidersHorizontal } from "lucide-react";
import { relativeTime, type Group } from "./groupWorkspaceTypes";

export function GroupConfigPane({
  selectedGroup,
}: {
  selectedGroup: Group | null;
}) {
  const capabilities = selectedGroup?.allowed_capabilities?.length
    ? selectedGroup.allowed_capabilities.join(", ")
    : "inherits lane policy";
  return (
    <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-3">
      <div className="flex items-center gap-2">
        <SlidersHorizontal className="h-4 w-4 text-cortex-primary" />
        <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
          Group Config
        </h2>
      </div>
      {selectedGroup ? (
        <div className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
          <Info
            label="Focused team lead"
            value={
              selectedGroup.coordinator_profile || `${selectedGroup.name} lead`
            }
            detail="Narrow group lane counterpart."
          />
          <Info
            label="Agent backend model"
            value="Inherits organization AI Engine"
            detail="Per-group model override is not persisted by the groups API yet."
          />
          <Info
            label="Approval policy"
            value={selectedGroup.approval_policy_ref || "default"}
            detail="Used for group creation and governed actions."
          />
          <Info
            label="Capabilities"
            value={capabilities}
            detail="Allowed runtime capabilities for this lane."
          />
          <Info
            label="Created by"
            value={selectedGroup.created_by}
            detail={`Created ${relativeTime(selectedGroup.created_at)}`}
          />
        </div>
      ) : (
        <p className="mt-3 text-sm text-cortex-text-muted">
          Select a group to inspect lane config.
        </p>
      )}
    </section>
  );
}

function Info({
  label,
  value,
  detail,
}: {
  label: string;
  value: string;
  detail: string;
}) {
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2">
      <p className="text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-primary">
        {label}
      </p>
      <p className="mt-1 truncate text-sm font-semibold text-cortex-text-main">
        {value}
      </p>
      <p className="mt-1 line-clamp-2 text-xs leading-4 text-cortex-text-muted">
        {detail}
      </p>
    </div>
  );
}
