import Link from "next/link";
import { Download } from "lucide-react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import {
  compactButtonClassName,
  groupKindLabel,
  linkClassName,
  relativeTime,
  type Group,
  type OutputSummary,
} from "./groupWorkspaceTypes";

export function GroupDetailPane({
  selectedGroup,
  outputs,
  outputSummary,
  archiving,
  onArchive,
}: {
  selectedGroup: Group | null;
  outputs: Artifact[];
  outputSummary: OutputSummary;
  archiving: boolean;
  onArchive: () => void;
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
    <section className="min-h-0 min-w-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4">
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-cortex-border pb-4">
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <Badge>{groupKindLabel(selectedGroup)}</Badge>
            <Badge muted>{selectedGroup.work_mode}</Badge>
          </div>
          <h2 className="mt-3 text-2xl font-semibold text-cortex-text-main">
            {selectedGroup.name}
          </h2>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">
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
      <div className="mt-4 grid gap-4 2xl:grid-cols-[minmax(0,1.25fr)_minmax(260px,0.75fr)]">
        <OutputsPanel
          archived={archived}
          outputs={outputs}
          outputSummary={outputSummary}
        />
        <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            Open work surfaces
          </h3>
          <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
            Jump into Soma or the group’s team leads after reviewing scope and
            outputs.
          </p>
          <div className="mt-4 flex flex-col gap-2">
            <Link href="/dashboard" className={linkClassName}>
              Open Soma admin home
            </Link>
            {selectedGroup.team_ids.map((teamId) => (
              <Link
                key={teamId}
                href={`/dashboard?team_id=${encodeURIComponent(teamId)}`}
                title={`Open ${teamId} lead`}
                className={linkClassName}
              >
                Open lead workspace
              </Link>
            ))}
          </div>
        </section>
      </div>
    </section>
  );
}

function OutputsPanel({
  archived,
  outputs,
  outputSummary,
}: {
  archived: boolean;
  outputs: Artifact[];
  outputSummary: OutputSummary;
}) {
  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
      <h3 className="text-sm font-semibold text-cortex-text-main">
        {archived ? "Retained outputs" : "Recent outputs"}
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
      {archived ? (
        <p
          className="mt-2 text-sm leading-6 text-cortex-text-muted"
          data-testid="groups-retained-outputs-note"
        >
          Review the outputs this archived temporary group already produced.
          Downloads remain available so the work can still be inspected after
          the collaboration window closes.
        </p>
      ) : null}
      <div className="mt-3 max-h-[420px] space-y-3 overflow-y-auto pr-1">
        {outputs.length === 0 ? (
          <p className="text-sm text-cortex-text-muted">
            No recent team outputs found for this group yet.
          </p>
        ) : (
          outputs
            .slice(0, 8)
            .map((artifact) => (
              <ArtifactRow key={artifact.id} artifact={artifact} />
            ))
        )}
      </div>
    </section>
  );
}

function Badge({
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

function ArtifactRow({ artifact }: { artifact: Artifact }) {
  const readable =
    artifact.artifact_type === "code" ||
    artifact.artifact_type === "document" ||
    artifact.artifact_type === "data" ||
    artifact.artifact_type === "chart";
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-cortex-text-main">
            {artifact.title}
          </p>
          <p className="mt-1 text-[11px] font-mono uppercase tracking-[0.12em] text-cortex-text-muted">
            {artifact.artifact_type} | {artifact.agent_id} |{" "}
            {relativeTime(artifact.created_at)}
          </p>
        </div>
        <a
          href={`/api/v1/artifacts/${encodeURIComponent(artifact.id)}/download`}
          download
          className="inline-flex items-center gap-1 rounded-xl border border-cortex-primary/30 px-3 py-2 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"
        >
          <Download className="h-3.5 w-3.5" />
          Download
        </a>
      </div>
      {readable && artifact.content ? (
        <pre className="mt-3 max-h-48 overflow-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-6 text-cortex-text-muted">
          {artifact.content}
        </pre>
      ) : artifact.file_path ? (
        <p className="mt-3 text-sm text-cortex-text-muted">
          Saved path: {artifact.file_path}
        </p>
      ) : null}
    </div>
  );
}
