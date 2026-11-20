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

function isReviewerRole(role: string, id: string): boolean {
  const hay = `${role} ${id}`.toLowerCase();
  return ["review", "verifier", "sentry", "qa", "audit", "validator"].some((k) => hay.includes(k));
}

export function DeploymentDetailPane({ record, advancedMode }: { record: DeploymentRecord | null; advancedMode: boolean }) {
  const [configViewerOpen, setConfigViewerOpen] = useState(false);

  if (!record) {
    return (
      <section className="min-h-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4 text-sm text-cortex-text-muted">
        Select a deployment to see what it runs, who is on it, and what it has produced.
      </section>
    );
  }
  return (
    <section className="relative min-h-0 min-w-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4" data-testid="deployments-detail">
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-cortex-border pb-4">
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <KindBadge kind={record.kind} />
            <span className="rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-0.5 text-[11px] font-mono uppercase tracking-[0.1em]">
              <StateBadge state={record.lifecycle_state} />
            </span>
          </div>
          <h2 className="mt-3 text-2xl font-semibold text-cortex-text-main">{record.name}</h2>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">{record.what_runs}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Link
            href={`/dashboard?team_id=${encodeURIComponent(record.id)}&team_name=${encodeURIComponent(record.name)}&team_action=plan_execute`}
            data-testid="deployments-plan-and-execute"
            className="inline-flex items-center justify-center rounded-xl border border-cortex-primary/50 bg-cortex-primary/15 px-3 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/25"
          >
            Plan &amp; Execute Next Steps
          </Link>
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
              onClick={() => { useCortexStore.getState().toggleAdvancedMode(); setConfigViewerOpen(true); }}
              data-testid="deployments-enable-advanced-config"
              className="inline-flex items-center justify-center rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-muted hover:border-cortex-primary/25 hover:text-cortex-text-main"
            >
              Inspect Config (Advanced)
            </button>
          )}
        </div>
      </div>

      {advancedMode && configViewerOpen ? <ConfigViewer configRef={record.config_ref} onClose={() => setConfigViewerOpen(false)} /> : null}

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
          <ChannelsAndChainingPanel record={record} />
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
          {members.map((member) => {
            const isReviewer = isReviewerRole(member.role, member.id);
            return (
              <div key={member.id} data-testid={`deployments-member-${member.id}`} className="rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold text-cortex-text-main">{member.role}</span>
                    {isReviewer ? (
                      <span data-testid="deployments-reviewer-badge" className="rounded border border-cortex-primary/40 bg-cortex-primary/15 px-1.5 py-0.5 font-mono text-[10px] font-semibold uppercase tracking-wider text-cortex-primary">
                        IDEAL REVIEWER · DELIVERY GATE
                      </span>
                    ) : null}
                  </div>
                  <span className="text-[11px] font-mono uppercase tracking-[0.1em] text-cortex-text-muted">
                    {member.provider ?? "unassigned"}{member.model ? `/${member.model}` : ""}{advancedMode && member.provider_origin ? ` (${member.provider_origin})` : ""}
                  </span>
                </div>
                {member.tools.length > 0 ? <p className="mt-1 truncate text-xs text-cortex-text-muted">Tools: {member.tools.map((tool) => tool.name).join(", ")}</p> : null}
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}

function PostureAndOutputsPanel({ record }: { record: DeploymentRecord }) {
  const recentOutputs = record.recent_outputs ?? [];
  const reviewerNames = record.members.filter((m) => isReviewerRole(m.role, m.id)).map((m) => m.role);
  const reviewerLabel = reviewerNames.length > 0 ? reviewerNames.join(", ") : "council-sentry (Ideal Delivery Reviewer)";
  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
      <h3 className="text-sm font-semibold text-cortex-text-main">Work mode and outputs</h3>
      <p className="mt-2 text-sm leading-6 text-cortex-text-muted" data-testid="deployments-approval-posture">
        {approvalPostureSentence(record)}
      </p>
      <p className="mt-2 text-xs leading-5 text-cortex-text-muted" data-testid="deployments-review-gate">
        <span className="font-semibold text-cortex-text-main">Delivery Review &amp; Verification Gate:</span>{" "}
        {reviewerLabel} — complex deliveries chain through <code className="font-mono text-[11px]">{record.id}-review-team</code> (<code className="font-mono text-[11px]">swarm.team.{record.id}.signal.result</code>) for SHA-256 <code className="font-mono text-[11px]">OutputProofEnvelope</code> verification before final release.
      </p>
      <p className="mt-2 text-sm text-cortex-text-muted" data-testid="deployments-outputs-summary">
        Retained outputs: {outputsSummaryLabel(record.outputs_summary)}
      </p>
      <div className="mt-3 space-y-2" data-testid="deployments-team-outputs">
        <div className="text-xs font-semibold uppercase tracking-wider text-cortex-text-muted">
          Generated Team Outputs
        </div>
        {recentOutputs.length > 0 ? (
          <div className="space-y-1.5">
            {recentOutputs.map((item) => (
              <div
                key={item.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs"
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate font-semibold text-cortex-text-main">{item.title}</div>
                  <div className="truncate font-mono text-[11px] text-cortex-text-muted">
                    {item.file_path || item.content_type || "output"}
                    {item.bytes != null ? ` · ${item.bytes} B` : ""}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <a
                    href={item.download_url || `/api/v1/artifacts/${item.id}/download`}
                    download
                    data-testid={`team-output-download-${item.id}`}
                    className="rounded border border-cortex-primary/35 bg-cortex-primary/10 px-2 py-1 font-mono text-[11px] font-medium text-cortex-primary hover:bg-cortex-primary/20"
                  >
                    Download
                  </a>
                  <Link
                    href={item.view_url || `/resources?team=${encodeURIComponent(record.id)}`}
                    data-testid={`team-output-view-${item.id}`}
                    className="rounded border border-cortex-border bg-cortex-bg px-2 py-1 font-mono text-[11px] text-cortex-text-main hover:border-cortex-primary/35"
                  >
                    View / Open
                  </Link>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-xs text-cortex-text-muted">No recent outputs indexed yet.</p>
        )}
      </div>
      <Link href={`/resources?team=${encodeURIComponent(record.id)}`} className="mt-3 inline-flex items-center text-xs font-medium text-cortex-primary hover:underline">
        Browse all generated files in Deliverables &rarr;
      </Link>
    </section>
  );
}

function ChannelsAndChainingPanel({ record }: { record: DeploymentRecord }) {
  const inputChannels = record.input_channels?.length
    ? record.input_channels
    : [`swarm.team.${record.id}.internal.command`, "swarm.global.broadcast"];
  const outputChannels = record.output_channels?.length
    ? record.output_channels
    : [`swarm.team.${record.id}.signal.result`, `swarm.team.${record.id}.signal.status`];
  const primarySpeakChannel = outputChannels[0] || `swarm.team.${record.id}.signal.result`;

  return (
    <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4" data-testid="deployments-team-channels">
      <h3 className="text-sm font-semibold text-cortex-text-main">Channels &amp; Inter-Team Chaining (Listen / Speak)</h3>
      <div className="mt-2 space-y-2 text-xs">
        <div>
          <span className="font-semibold text-cortex-text-muted">Listens on (Inputs):</span>
          <div className="mt-1 flex flex-wrap gap-1.5">
            {inputChannels.map((ch) => (
              <span key={ch} className="rounded border border-cortex-border bg-cortex-surface px-2 py-0.5 font-mono text-[11px] text-cortex-text-main">
                {ch}
              </span>
            ))}
          </div>
        </div>
        <div>
          <span className="font-semibold text-cortex-text-muted">Speaks on (Outputs / Deliveries):</span>
          <div className="mt-1 flex flex-wrap gap-1.5">
            {outputChannels.map((ch) => (
              <span key={ch} className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 font-mono text-[11px] text-cortex-primary">
                {ch}
              </span>
            ))}
          </div>
        </div>
      </div>
      <Link
        href={`/dashboard?team_id=${encodeURIComponent(record.id)}&chain_channel=${encodeURIComponent(primarySpeakChannel)}`}
        data-testid="deployments-chain-output"
        className="mt-3 inline-flex items-center rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-1.5 text-xs font-medium text-cortex-primary hover:bg-cortex-primary/20"
      >
        Chain This Team&apos;s Output to Another Team in Soma &rarr;
      </Link>
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
