"use client";

import Link from "next/link";
import type {
  TeamLeadExecutionContract,
  TeamLeadWorkflowGroupDraft,
} from "@/lib/organizations";
import type { GuidedExecutionContract } from "@/components/organizations/teamLeadGuidanceNormalization";

export type LaunchedGroupState = {
  groupId: string;
  name: string;
};

export function CompactTeamGuidanceCard({
  requestScope,
}: {
  requestScope: "compact" | "broad" | null;
}) {
  const broad = requestScope === "broad";
  return (
    <div
      className={`rounded-2xl border px-4 py-4 ${broad ? "border-cortex-primary/30 bg-cortex-primary/10" : "border-cortex-border bg-cortex-surface/60"}`}
    >
      <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
        Compact team default
      </p>
      <p className="mt-2 text-sm font-semibold text-cortex-text-main">
        {broad
          ? "This ask looks broad enough for several small teams or lane bundles."
          : "This ask looks compact enough for one focused team."}
      </p>
      <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
        Default to the smallest useful team: one accountable lead first. If the
        request crosses marketing, product, operations, or media at once, Soma
        should split it into a few compact teams and add temporary specialists
        only after a lead names the missing capability and proof needed.
      </p>
    </div>
  );
}

export function ExecutionContractCard({
  contract,
}: {
  contract: TeamLeadExecutionContract;
}) {
  const richContract = contract as GuidedExecutionContract;
  const isNativeTeam = contract.execution_mode === "native_team";
  const isContinuityResume = contract.execution_mode === "continuity_resume";
  const title = isNativeTeam
    ? "Native Mycelis team"
    : contract.execution_mode === "external_workflow_contract"
      ? "External workflow contract"
      : isContinuityResume
        ? "Retained package continuity"
        : "Guided execution path";

  return (
    <div className="rounded-2xl border border-cortex-primary/25 bg-cortex-primary/5 px-4 py-4">
      <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
        Execution path
      </p>
      <div className="mt-3 flex flex-wrap items-center gap-2">
        <span className="rounded-full border border-cortex-primary/25 bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-primary">
          {title}
        </span>
        <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
          {contract.owner_label}
        </span>
        {contract.team_name ? (
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            {contract.team_name}
          </span>
        ) : null}
        {contract.external_target ? (
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            {contract.external_target}
          </span>
        ) : null}
      </div>
      <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
        {contract.summary}
      </p>
      {isContinuityResume &&
      (contract.continuity_summary || contract.resume_checkpoint) ? (
        <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
          {contract.continuity_label ? (
            <p className="text-xs font-mono uppercase tracking-[0.18em] text-cortex-primary">
              {contract.continuity_label}
            </p>
          ) : null}
          {contract.continuity_summary ? (
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
              {contract.continuity_summary}
            </p>
          ) : null}
          {contract.resume_checkpoint ? (
            <p className="mt-3 text-sm font-medium text-cortex-text-main">
              {contract.resume_checkpoint}
            </p>
          ) : null}
        </div>
      ) : null}
      {contract.target_outputs.length > 0 ? (
        <div className="mt-4">
          <p className="text-sm font-medium text-cortex-text-main">
            Target outputs
          </p>
          <div className="mt-2 flex flex-wrap gap-2">
            {contract.target_outputs.map((output) => (
              <span
                key={output}
                className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-2 text-sm text-cortex-text-main"
              >
                {output}
              </span>
            ))}
          </div>
        </div>
      ) : null}
      {richContract.workstreams && richContract.workstreams.length > 0 ? (
        <WorkstreamList contract={richContract} />
      ) : null}
      <CompactOrchestrationHints contract={richContract} />
    </div>
  );
}

function WorkstreamList({ contract }: { contract: GuidedExecutionContract }) {
  return (
    <div className="mt-4">
      <p className="text-sm font-medium text-cortex-text-main">
        Working together now
      </p>
      <div className="mt-3 space-y-3">
        {contract.workstreams?.map((workstream) => (
          <div
            key={`${workstream.label}:${workstream.owner_label}`}
            className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3"
          >
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full border border-cortex-primary/25 bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-primary">
                {workstream.label}
              </span>
              {workstream.status ? (
                <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                  {workstream.status}
                </span>
              ) : null}
              <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                {workstream.owner_label}
              </span>
            </div>
            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
              {workstream.summary}
            </p>
            <p className="mt-3 text-sm font-medium text-cortex-text-main">
              Next step
            </p>
            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
              {workstream.next_step}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}

function CompactOrchestrationHints({
  contract,
}: {
  contract: GuidedExecutionContract;
}) {
  if (
    !contract.recommended_team_shape &&
    !contract.coordination_model &&
    typeof contract.recommended_team_count !== "number" &&
    typeof contract.initial_member_count !== "number" &&
    typeof contract.recommended_team_member_limit !== "number" &&
    !contract.expansion_policy &&
    !contract.temporary_addition_guidance
  ) {
    return null;
  }
  return (
    <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
      <p className="text-xs font-mono uppercase tracking-[0.18em] text-cortex-primary">
        Compact orchestration hints
      </p>
      <div className="mt-3 flex flex-wrap gap-2">
        {contract.recommended_team_shape ? (
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            {contract.recommended_team_shape}
          </span>
        ) : null}
        {contract.coordination_model ? (
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            {contract.coordination_model}
          </span>
        ) : null}
        {typeof contract.recommended_team_count === "number" ? (
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            {contract.recommended_team_count} team
            {contract.recommended_team_count === 1 ? "" : "s"}
          </span>
        ) : null}
        {typeof contract.initial_member_count === "number" ? (
          <span className="rounded-full border border-cortex-primary/25 bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-primary">
            starts with {contract.initial_member_count} lead
          </span>
        ) : null}
        {typeof contract.recommended_team_member_limit === "number" ? (
          <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            expansion cap {contract.recommended_team_member_limit}
          </span>
        ) : null}
      </div>
      {contract.expansion_policy ? (
        <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
          {contract.expansion_policy}
        </p>
      ) : null}
      {contract.temporary_addition_guidance ? (
        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
          {contract.temporary_addition_guidance}
        </p>
      ) : null}
    </div>
  );
}

export function TemporaryWorkflowLaunchCard({
  draft,
  launching,
  launchedGroup,
  error,
  onLaunch,
}: {
  draft: TeamLeadWorkflowGroupDraft;
  launching: boolean;
  launchedGroup: LaunchedGroupState | null;
  error: string | null;
  onLaunch: () => void;
}) {
  const launchable = draft.work_mode !== "resume_continuity";
  const continuityHref =
    !launchable && draft.group_id
      ? `/groups?group_id=${encodeURIComponent(draft.group_id)}`
      : null;
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
      <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
        {launchable
          ? "Launch temporary workflow group"
          : "Retained package continuity"}
      </p>
      <p className="mt-3 text-sm font-semibold text-cortex-text-main">
        {draft.name}
      </p>
      <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
        {draft.summary}
      </p>
      <div className="mt-3 flex flex-wrap gap-2">
        <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
          {draft.work_mode}
        </span>
        <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
          {draft.coordinator_profile}
        </span>
        {typeof draft.initial_member_count === "number" ? (
          <span className="rounded-full border border-cortex-primary/25 bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-primary">
            starts with {draft.initial_member_count} lead
          </span>
        ) : null}
        {typeof draft.recommended_member_limit === "number" ? (
          <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            expansion cap {draft.recommended_member_limit}
          </span>
        ) : null}
        {typeof draft.expiry_hours === "number" && draft.expiry_hours > 0 ? (
          <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
            expires in {draft.expiry_hours}h
          </span>
        ) : null}
      </div>
      {draft.expansion_policy ? (
        <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
          {draft.expansion_policy}
        </p>
      ) : null}
      {draft.temporary_addition_guidance ? (
        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
          {draft.temporary_addition_guidance}
        </p>
      ) : null}
      {launchable ? (
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={onLaunch}
            disabled={launching}
            className="rounded-2xl border border-cortex-primary/35 bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg disabled:opacity-60"
          >
            {launching
              ? "Creating workflow group..."
              : "Create temporary workflow group"}
          </button>
          {launchedGroup ? (
            <Link
              href={`/groups?group_id=${encodeURIComponent(launchedGroup.groupId)}`}
              className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-2 text-sm font-medium text-cortex-text-main hover:border-cortex-primary/20"
            >
              Open {launchedGroup.name}
            </Link>
          ) : null}
        </div>
      ) : (
        <div className="mt-4 space-y-3">
          <p className="text-sm text-cortex-text-muted">
            This continuity path is review-first. Reopen the retained package
            and continue from the recorded checkpoint instead of launching a new
            temporary group.
          </p>
          {continuityHref ? (
            <Link
              href={continuityHref}
              className="inline-flex rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-2 text-sm font-medium text-cortex-text-main hover:border-cortex-primary/20"
            >
              Open retained package
            </Link>
          ) : null}
        </div>
      )}
      {launchedGroup ? (
        <p className="mt-3 text-sm text-cortex-primary">
          Soma launched {launchedGroup.name}. The workflow group is ready for
          focused coordination and retained output review.
        </p>
      ) : null}
      {error ? (
        <p className="mt-3 text-sm text-cortex-danger">{error}</p>
      ) : null}
    </div>
  );
}
