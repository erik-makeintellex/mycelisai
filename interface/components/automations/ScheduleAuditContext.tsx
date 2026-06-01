import type React from "react";
import type { AuditLogEntry } from "@/store/useCortexStore";
import {
  formatScheduleHandoffState,
  getScheduleHandoffState,
  scheduleHandoffTone,
} from "@/components/automations/scheduleHandoffState";

function detailText(entry: AuditLogEntry, key: string) {
  const value = entry.details?.[key];
  return typeof value === "string" ? value : "";
}

export default function ScheduleAuditContext({ entry }: { entry: AuditLogEntry }) {
  const scheduleRuleID = detailText(entry, "schedule_rule_id");
  const scheduleRuleName = detailText(entry, "schedule_rule_name");
  const scheduleExecutionID = detailText(entry, "schedule_execution_id");
  const scheduleProposedAt = detailText(entry, "schedule_proposed_at");
  const proofExpectations = detailText(entry, "proof_expectations");
  const recoveryBehavior = detailText(entry, "recovery_behavior");
  const handoffState = getScheduleHandoffState(entry.details);
  const isScheduleOrigin =
    entry.details?.trigger_kind === "schedule" ||
    Boolean(scheduleRuleID || scheduleRuleName || scheduleExecutionID || handoffState);

  if (!isScheduleOrigin) return null;

  return (
    <div className="mt-3 rounded-lg border border-cortex-primary/20 bg-cortex-primary/5 px-3 py-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-0.5 text-[10px] font-mono uppercase text-cortex-primary">
          schedule origin
        </span>
        <span className="min-w-0 truncate text-xs font-medium text-cortex-text-main">
          {scheduleRuleName || scheduleRuleID || "Scheduled proposal"}
        </span>
      </div>
      <div className="mt-2 flex flex-wrap gap-2 text-[10px] font-mono">
        {scheduleExecutionID ? <ScheduleChip>Schedule execution: {scheduleExecutionID}</ScheduleChip> : null}
        {handoffState ? <HandoffChip state={handoffState} /> : null}
        {scheduleProposedAt ? <ScheduleChip>Proposed: {new Date(scheduleProposedAt).toLocaleString()}</ScheduleChip> : null}
        {proofExpectations ? <ProofChip>Proof: {proofExpectations}</ProofChip> : null}
        {recoveryBehavior ? <RecoveryChip>Recovery: {recoveryBehavior}</RecoveryChip> : null}
      </div>
    </div>
  );
}

function HandoffChip({ state }: { state: string }) {
  const tone = scheduleHandoffTone(state);
  const className =
    tone === "success"
      ? "border-cortex-success/30 bg-cortex-success/10 text-cortex-success"
      : tone === "danger"
        ? "border-red-400/30 bg-red-400/10 text-red-300"
        : tone === "pending"
          ? "border-amber-400/30 bg-amber-400/10 text-amber-300"
          : "border-cortex-border bg-cortex-bg/40 text-cortex-text-main";

  return (
    <span className={`rounded border px-2 py-1 ${className}`}>
      Handoff: {formatScheduleHandoffState(state)}
    </span>
  );
}

function ScheduleChip({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
      {children}
    </span>
  );
}

function ProofChip({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-1 text-cortex-primary">
      {children}
    </span>
  );
}

function RecoveryChip({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded border border-amber-400/30 bg-amber-400/10 px-2 py-1 text-amber-300">
      {children}
    </span>
  );
}
