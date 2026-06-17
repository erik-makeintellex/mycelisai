import Link from "next/link";
import OutputAccessActions from "@/components/soma/OutputAccessActions";
import {
  compactButtonClassName,
  groupKindLabel,
  linkClassName,
  type Group,
  type GroupLifecycleItem,
  type OutputSummary,
} from "./groupWorkspaceTypes";

export function GroupDetailPane({
  selectedGroup,
  lifecycleItem,
  outputSummary,
  archiving,
  onArchive,
  onOpenOutputs,
}: {
  selectedGroup: Group | null;
  lifecycleItem?: GroupLifecycleItem;
  outputSummary: OutputSummary;
  archiving: boolean;
  onArchive: () => void;
  onOpenOutputs: () => void;
}) {
  if (!selectedGroup) {
    return (
      <section className="min-h-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4 text-sm text-cortex-text-muted">
        Create or select a group to review its lane, config, and outputs.
      </section>
    );
  }
  const archived = selectedGroup.status === "archived";
  return (
    <section className="min-h-0 min-w-0 rounded-2xl border border-cortex-border bg-cortex-surface p-4">
      <div className="grid gap-4 border-b border-cortex-border pb-4 xl:grid-cols-[minmax(0,1fr)_minmax(320px,0.55fr)]">
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <Badge>{groupKindLabel(selectedGroup)}</Badge>
            <Badge muted>{selectedGroup.work_mode}</Badge>
          </div>
          <h2 className="mt-3 break-words text-xl font-semibold text-cortex-text-main">
            {selectedGroup.name}
          </h2>
          <p
            className="mt-2 max-h-24 max-w-4xl overflow-y-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm leading-6 text-cortex-text-muted [overflow-wrap:anywhere]"
            data-testid="groups-goal-summary"
          >
            {selectedGroup.goal_statement}
          </p>
        </div>
        <div className="min-w-0 rounded-xl border border-cortex-border bg-cortex-bg p-3">
          <div
            className="flex flex-wrap gap-2"
            data-testid="groups-output-summary"
          >
            <Badge muted>
              {outputSummary.artifactCount} output
              {outputSummary.artifactCount === 1 ? "" : "s"}
            </Badge>
            <Badge muted>
              {outputSummary.agentCount} contributing lead
              {outputSummary.agentCount === 1 ? "" : "s"}
            </Badge>
            {lifecycleItem ? (
              <Badge muted>
                {lifecycleSummaryLabel(lifecycleItem)}
              </Badge>
            ) : null}
          </div>
          {lifecycleItem ? (
            <p className="mt-3 rounded-lg border border-cortex-border bg-cortex-surface p-3 text-xs leading-5 text-cortex-text-muted">
              <span className="font-semibold text-cortex-text-main">
                Recommended:
              </span>{" "}
              {lifecycleActionLabel(lifecycleItem)}. {lifecycleItem.reason}
            </p>
          ) : null}
          <div className="mt-3 flex flex-wrap gap-2">
            <button
              type="button"
              onClick={onOpenOutputs}
              className={compactButtonClassName}
            >
              Review outputs
            </button>
            <Link href="/dashboard" className={linkClassName}>
              Open Soma
            </Link>
            {selectedGroup.expiry && !archived ? (
              <button
                type="button"
                aria-label="Archive temporary group"
                onClick={onArchive}
                disabled={archiving}
                className={compactButtonClassName}
              >
                {archiving ? "Archiving..." : "Archive"}
              </button>
            ) : null}
          </div>
        </div>
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(320px,0.75fr)]">
        <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            Output access
          </h3>
          {selectedGroup.workspace_folder ? (
            <div className="mt-3 rounded-lg border border-cortex-border bg-cortex-surface p-3">
              <p className="text-xs font-semibold uppercase tracking-[0.12em] text-cortex-text-muted">
                Group output folder
              </p>
              <p className="mt-1 break-all font-mono text-xs text-cortex-text-main">
                {selectedGroup.workspace_folder}
              </p>
              <div className="mt-2">
                <OutputAccessActions
                  label={`${selectedGroup.name} output folder`}
                  url={null}
                  storagePath={selectedGroup.workspace_folder}
                />
              </div>
            </div>
          ) : null}
        </section>
        <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            Team leads
          </h3>
          <div className="mt-3 flex flex-col gap-2">
            {selectedGroup.team_ids.map((teamId) => (
              <Link
                key={teamId}
                href={`/dashboard?team_id=${encodeURIComponent(teamId)}`}
                title={`Open ${teamId} lead`}
                className={`${linkClassName} min-w-0 truncate`}
              >
                Open lead workspace: {teamId}
              </Link>
            ))}
          </div>
        </section>
      </div>
    </section>
  );
}

function lifecycleSummaryLabel(item: GroupLifecycleItem) {
  if (item.active_or_blocked_work_count > 0)
    return `${item.active_or_blocked_work_count} work waiting`;
  if (item.output_ready_work_count > 0 || item.output_count > 0)
    return "output retained";
  if (item.expired) return "expired";
  return item.recommendation.replaceAll("_", " ");
}

function lifecycleActionLabel(item: GroupLifecycleItem) {
  switch (item.recommendation) {
    case "archive_expired":
      return "archive expired temporary lane";
    case "review_work":
      return "review linked team work";
    case "archive_completed":
      return "archive completed lane after output review";
    case "review_standing":
      return "confirm this standing group is still useful";
    case "retained":
      return "keep retained record available";
    default:
      return "keep active";
  }
}

export function Badge({
  children,
  muted = false,
}: {
  children: React.ReactNode;
  muted?: boolean;
}) {
  return (
    <span
      className={`rounded-full border px-3 py-1 text-[11px] font-mono uppercase tracking-[0.14em] ${muted ? "border-cortex-border bg-cortex-bg text-cortex-text-muted" : "border-cortex-primary/25 bg-cortex-primary/10 text-cortex-primary"}`}
    >
      {children}
    </span>
  );
}
