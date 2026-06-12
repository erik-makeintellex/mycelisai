import Link from "next/link";
import OutputAccessActions from "@/components/soma/OutputAccessActions";
import {
  compactButtonClassName,
  groupKindLabel,
  linkClassName,
  type Group,
  type OutputSummary,
} from "./groupWorkspaceTypes";

export function GroupDetailPane({
  selectedGroup,
  outputSummary,
  archiving,
  onArchive,
  onOpenOutputs,
}: {
  selectedGroup: Group | null;
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
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-cortex-border pb-4">
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <Badge>{groupKindLabel(selectedGroup)}</Badge>
            <Badge muted>{selectedGroup.work_mode}</Badge>
          </div>
          <h2 className="mt-3 break-words text-2xl font-semibold text-cortex-text-main">
            {selectedGroup.name}
          </h2>
          <p
            className="mt-2 max-h-32 max-w-4xl overflow-y-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm leading-6 text-cortex-text-muted [overflow-wrap:anywhere]"
            data-testid="groups-goal-summary"
          >
            {selectedGroup.goal_statement}
          </p>
        </div>
        {selectedGroup.expiry && !archived ? (
          <button
            type="button"
            onClick={onArchive}
            disabled={archiving}
            className={compactButtonClassName}
          >
            {archiving ? "Archiving..." : "Archive temporary group"}
          </button>
        ) : null}
      </div>
      <div className="mt-4 grid gap-4 2xl:grid-cols-[minmax(0,0.8fr)_minmax(260px,1.2fr)]">
        <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            Work summary
          </h3>
          <div
            className="mt-2 flex flex-wrap gap-2"
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
          </div>
          <button
            type="button"
            onClick={onOpenOutputs}
            className={`${compactButtonClassName} mt-3`}
          >
            Review retained outputs
          </button>
        </section>
        <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            Open work surfaces
          </h3>
          <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
            Jump into Soma or the group’s team leads after reviewing scope and
            outputs.
          </p>
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
          <div className="mt-4 flex flex-col gap-2">
            <Link href="/dashboard" className={linkClassName}>
              Open Soma admin home
            </Link>
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
