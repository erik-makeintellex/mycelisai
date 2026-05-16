"use client";

import type {
  TeamLeadExecutionContract,
  TeamLeadGuidanceResponse,
} from "@/lib/organizations";
import type { GuidedExecutionContract } from "@/components/organizations/teamLeadGuidanceNormalization";

type TeamLogEntry = {
  source: string;
  label: string;
  message: string;
  details?: string[];
  meta?: string[];
  tone?: "operator" | "soma" | "execution" | "team" | "output" | "next";
};

export function TeamEventLog({
  guidance,
  requestContext,
  requestScope,
}: {
  guidance: TeamLeadGuidanceResponse;
  requestContext: string | null;
  requestScope: "compact" | "broad" | null;
}) {
  const entries = buildTeamLogEntries(guidance, requestContext, requestScope);
  return (
    <section className="rounded-2xl border border-cortex-border bg-cortex-surface">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-cortex-border px-4 py-3">
        <div>
          <p className="text-sm font-semibold text-cortex-text-main">
            Team event log
          </p>
          <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
            Messages and execution events from Soma, team lanes, outputs, and
            next steps appear in one scrollable record.
          </p>
        </div>
        <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
          {entries.length} events
        </span>
      </div>
      <div className="max-h-[520px] overflow-y-auto px-3 py-3">
        <div className="space-y-2">
          {entries.map((entry, index) => (
            <TeamLogRow
              key={`${entry.label}:${entry.source}:${index}`}
              entry={entry}
            />
          ))}
        </div>
      </div>
    </section>
  );
}

function buildTeamLogEntries(
  guidance: TeamLeadGuidanceResponse,
  requestContext: string | null,
  requestScope: "compact" | "broad" | null,
): TeamLogEntry[] {
  const entries: TeamLogEntry[] = [];
  if (requestContext) {
    entries.push({
      source: "Operator",
      label: "You asked Soma to help with",
      message: requestContext,
      tone: "operator",
    });
  }
  entries.push({
    source: "Soma",
    label: "Current request",
    message: guidance.request_label,
    tone: "soma",
  });
  entries.push({
    source: "Soma",
    label: "Soma guidance",
    message: guidance.headline,
    details: [guidance.summary],
    tone: "soma",
  });
  entries.push({
    source: "Soma",
    label: "Compact team default",
    message:
      requestScope === "broad"
        ? "This ask looks broad enough for several small teams or lane bundles."
        : "This ask looks compact enough for one focused team.",
    details: [
      requestScope === "broad"
        ? "Default to lead-only teams and add temporary specialists only after a lead names the missing capability and proof needed."
        : "Default to a lead-only team and add specialists only when the lead can name the missing capability and proof needed.",
    ],
    tone: "team",
  });
  if (guidance.execution_contract) {
    entries.push(...executionContractEntries(guidance.execution_contract));
  }
  entries.push(
    ...guidance.priority_steps.map((step, index) => ({
      source: "Soma",
      label: index === 0 ? "Priority steps" : "Priority step",
      message: step,
      tone: "next" as const,
    })),
  );
  entries.push(
    ...guidance.suggested_follow_ups.map((followUp, index) => ({
      source: "Soma",
      label: index === 0 ? "Keep moving with" : "Follow-up option",
      message: followUp,
      tone: "next" as const,
    })),
  );
  return entries;
}

function executionContractEntries(
  contract: TeamLeadExecutionContract,
): TeamLogEntry[] {
  const richContract = contract as GuidedExecutionContract;
  const title =
    contract.execution_mode === "native_team"
      ? "Native Mycelis team"
      : contract.execution_mode === "external_workflow_contract"
        ? "External workflow contract"
        : contract.execution_mode === "continuity_resume"
          ? "Retained package continuity"
          : "Guided execution path";
  const entries: TeamLogEntry[] = [
    {
      source: contract.owner_label,
      label: "Execution path",
      message: contract.summary,
      meta: [title, contract.team_name, contract.external_target].filter(
        Boolean,
      ) as string[],
      tone: "execution",
    },
  ];
  if (contract.continuity_summary || contract.resume_checkpoint) {
    entries.push({
      source: contract.continuity_label || "Continuity",
      label: "Retained package continuity",
      message:
        contract.continuity_summary ||
        contract.resume_checkpoint ||
        "Retained package continuity is available.",
      details:
        contract.continuity_summary && contract.resume_checkpoint
          ? [contract.resume_checkpoint]
          : undefined,
      tone: "execution",
    });
  }
  if (contract.target_outputs.length > 0) {
    entries.push({
      source: "Outputs",
      label: "Target outputs",
      message: "Outputs expected from this team path.",
      meta: contract.target_outputs,
      tone: "output",
    });
  }
  for (const [index, workstream] of (
    richContract.workstreams ?? []
  ).entries()) {
    entries.push({
      source: workstream.owner_label,
      label: index === 0 ? "Working together now" : "Team lane update",
      message: workstream.label,
      details: [workstream.summary, `Next step: ${workstream.next_step}`],
      meta: [workstream.status, ...(workstream.target_outputs ?? [])].filter(
        Boolean,
      ) as string[],
      tone: "team",
    });
  }
  const hints = [
    richContract.recommended_team_shape,
    richContract.coordination_model,
    typeof richContract.recommended_team_count === "number"
      ? `${richContract.recommended_team_count} team${richContract.recommended_team_count === 1 ? "" : "s"}`
      : null,
    typeof richContract.initial_member_count === "number"
      ? `starts with ${richContract.initial_member_count} lead`
      : null,
    typeof richContract.recommended_team_member_limit === "number"
      ? `expansion cap ${richContract.recommended_team_member_limit}`
      : null,
    richContract.expansion_policy,
  ].filter(Boolean) as string[];
  if (hints.length > 0) {
    entries.push({
      source: "Soma",
      label: "Compact orchestration hints",
      message: hints.join(", "),
      meta: hints,
      tone: "team",
    });
  }
  return entries;
}

function TeamLogRow({ entry }: { entry: TeamLogEntry }) {
  return (
    <article className="grid gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-3 text-sm sm:grid-cols-[120px_minmax(0,1fr)]">
      <div className="flex items-center gap-2 sm:block">
        <span
          className={`inline-flex h-2 w-2 rounded-full ${toneDotClass(entry.tone)}`}
        />
        <p
          className="truncate text-[11px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted"
          title={entry.source}
        >
          {entry.source}
        </p>
      </div>
      <div className="min-w-0">
        <p className="text-xs font-semibold text-cortex-text-main">
          {entry.label}
        </p>
        <p className="mt-1 text-sm leading-6 text-cortex-text-main">
          {entry.message}
        </p>
        {entry.details?.map((detail) => (
          <p
            key={detail}
            className="mt-1 text-sm leading-6 text-cortex-text-muted"
          >
            {detail}
          </p>
        ))}
        {entry.meta && entry.meta.length > 0 ? (
          <div className="mt-2 flex flex-wrap gap-2">
            {entry.meta.map((item) => (
              <span
                key={item}
                className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-1 text-[11px] font-mono text-cortex-text-muted"
              >
                {item}
              </span>
            ))}
          </div>
        ) : null}
      </div>
    </article>
  );
}

function toneDotClass(tone: TeamLogEntry["tone"]) {
  if (tone === "operator") return "bg-cortex-info";
  if (tone === "execution") return "bg-cortex-primary";
  if (tone === "team") return "bg-cortex-success";
  if (tone === "output") return "bg-cortex-warning";
  if (tone === "next") return "bg-cortex-text-muted";
  return "bg-cortex-primary";
}
