import { useState } from "react";
import Link from "next/link";
import { useCortexStore } from "@/store/useCortexStore";
import {
  approvalPostureSentence,
  outputsSummaryLabel,
  relativeTime,
  type DeploymentMember,
  type DeploymentRecord,
} from "./deploymentTypes";
import { KindBadge, StateBadge } from "./DeploymentList";
import { ConfigViewer } from "./ConfigViewer";

export function DeploymentDetailPane({
  record,
  advancedMode,
}: {
  record: DeploymentRecord | null;
  advancedMode: boolean;
}) {
  const [configViewerOpen, setConfigViewerOpen] = useState(false);

  if (!record) {
    return (
      <section className="min-h-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4 text-sm text-cortex-text-muted">
        Select a deployment to see what it runs, who is on it, and what it has produced.
      </section>
    );
  }
  return (
    <section
      className="relative min-h-0 min-w-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4"
      data-testid="deployments-detail"
    >
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-cortex-border pb-4">
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <KindBadge kind={record.kind} />
            <span className="rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-0.5 text-[11px] font-mono uppercase tracking-[0.1em]">
              <StateBadge state={record.lifecycle_state} />
            </span>
          </div>
          <h2 className="mt-3 text-2xl font-semibold text-cortex-text-main">{record.name}</h2>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">
            {record.what_runs}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Link
            href={`/dashboard?team_id=${encodeURIComponent(record.id)}`}
            data-testid="deployments-engage-soma"
            className="inline-flex items-center justify-center rounded-xl border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-2 text-sm font-medium text-cortex-primary hover:bg-cortex-primary/20"
          >
            Work with Team in Soma
          </Link>
          <Link
            href={`/resources?team=${encodeURIComponent(record.id)}`}
            data-testid="deployments-open-deliverables"
            className="inline-flex items-center justify-center rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm font-medium text-cortex-text-main hover:border-cortex-primary/35 hover:text-cortex-primary"
          >
            View Deliverables
          </Link>
          {advancedMode ? (
            <button
              type="button"
              onClick={() => setConfigViewerOpen(true)}
              data-testid="deployments-view-config"
              className="inline-flex items-center justify-center rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main hover:border-cortex-primary/25"
            >
              View config
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                useCortexStore.getState().toggleAdvancedMode();
                setConfigViewerOpen(true);
              }}
              data-testid="deployments-enable-advanced-config"
              className="inline-flex items-center justify-center rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-muted hover:border-cortex-primary/25 hover:text-cortex-text-main"
            >
              Inspect Config (Advanced)
            </button>
          )}
        </div>
      </div>

      {advancedMode && configViewerOpen ? (
        <ConfigViewer configRef={record.config_ref} onClose={() => setConfigViewerOpen(false)} />
      ) : null}

      {advancedMode ? (
        <div className="mt-3 flex flex-wrap gap-3 text-[11px] font-mono uppercase tracking-[0.1em] text-cortex-text-muted">
          <span>Origin: {record.origin}</span>
          <span>Scope: {record.scope.kind}{record.scope.name ? ` / ${record.scope.name}` : ""}</span>
        </div>
      ) : null}

      <div className="mt-4 grid gap-4 2xl:grid-cols-[minmax(0,1.25fr)_minmax(260px,0.75fr)]">
        <div className="space-y-4">
          <MembersPanel members={record.members} advancedMode={advancedMode} />
          <PostureAndOutputsPanel record={record} />
        </div>
        <TimingPanel record={record} />
      </div>
    </section>
  );
}

function MembersPanel({ members, advancedMode }: { members: DeploymentMember[]; advancedMode: boolean }) {
  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
      <h3 className="text-sm font-semibold text-cortex-text-main">Members</h3>
      {members.length === 0 ? (
        <p className="mt-2 text-sm text-cortex-text-muted">No members recorded.</p>
      ) : (
        <div className="mt-3 space-y-2" data-testid="deployments-members">
          {members.map((member) => (
            <div
              key={member.id}
              data-testid={`deployments-member-${member.id}`}
              className="rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="text-sm font-semibold text-cortex-text-main">{member.role}</span>
                <span className="text-[11px] font-mono uppercase tracking-[0.1em] text-cortex-text-muted">
                  {member.provider ?? "unassigned"}
                  {member.model ? `/${member.model}` : ""}
                  {advancedMode && member.provider_origin ? ` (${member.provider_origin})` : ""}
                </span>
              </div>
              {member.tools.length > 0 ? (
                <p className="mt-1 truncate text-xs text-cortex-text-muted">
                  Tools: {member.tools.map((tool) => tool.name).join(", ")}
                </p>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function PostureAndOutputsPanel({ record }: { record: DeploymentRecord }) {
  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
      <h3 className="text-sm font-semibold text-cortex-text-main">Work mode and outputs</h3>
      <p className="mt-2 text-sm leading-6 text-cortex-text-muted" data-testid="deployments-approval-posture">
        {approvalPostureSentence(record)}
      </p>
      <p className="mt-2 text-sm text-cortex-text-muted" data-testid="deployments-outputs-summary">
        Retained outputs: {outputsSummaryLabel(record.outputs_summary)}
      </p>
      <Link href={`/resources?team=${encodeURIComponent(record.id)}`} className="mt-2 inline-flex items-center text-xs font-medium text-cortex-primary hover:underline">Browse generated files in Deliverables &rarr;</Link>
    </section>
  );
}

function TimingPanel({ record }: { record: DeploymentRecord }) {
  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
      <h3 className="text-sm font-semibold text-cortex-text-main">Since and last activity</h3>
      <dl className="mt-2 space-y-1.5 text-sm text-cortex-text-muted">
        <div className="flex items-center justify-between gap-2">
          <dt>Since</dt>
          <dd>{relativeTime(record.since)}</dd>
        </div>
        <div className="flex items-center justify-between gap-2">
          <dt>Last activity</dt>
          <dd>{relativeTime(record.last_activity)}</dd>
        </div>
        {record.expires_at ? (
          <div className="flex items-center justify-between gap-2">
            <dt>Expires</dt>
            <dd>{relativeTime(record.expires_at)}</dd>
          </div>
        ) : null}
      </dl>
    </section>
  );
}
