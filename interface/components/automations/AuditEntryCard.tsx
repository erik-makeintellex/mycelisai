"use client";

import { BadgeCheck, ShieldAlert } from "lucide-react";
import type { AuditLogEntry } from "@/store/useCortexStore";

/** Single audit-log entry card, extracted from ApprovalsTab to keep that file under its line cap. */
export default function AuditEntryCard({ entry }: { entry: AuditLogEntry }) {
  const approvalTone =
    entry.approval_status === "approval_required" ||
    entry.approval_status === "cancelled"
      ? "text-amber-300 border-amber-400/30"
      : "text-cortex-success border-cortex-success/30";
  const statusTone =
    entry.result_status === "failed" || entry.result_status === "error"
      ? "text-red-300 border-red-400/30"
      : entry.result_status === "pending"
        ? "text-amber-300 border-amber-400/30"
        : "text-cortex-success border-cortex-success/30";
  const approvalLabel = entry.approval_status
    ? entry.approval_status.replaceAll("_", " ")
    : "logged";
  const resultLabel = entry.result_status.replaceAll("_", " ");
  const operatorSummary =
    typeof entry.details?.operator_summary === "string"
      ? entry.details.operator_summary
      : "";
  const teamID =
    typeof entry.details?.team_id === "string" ? entry.details.team_id : "";
  const askKind =
    typeof entry.details?.ask_kind === "string" ? entry.details.ask_kind : "";
  const laneRole =
    typeof entry.details?.lane_role === "string" ? entry.details.lane_role : "";

  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-cortex-text-main">
              {entry.action.replaceAll("_", " ")}
            </span>
            <span
              className={`rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase ${statusTone}`}
            >
              {resultLabel}
            </span>
            <span
              className={`rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase ${approvalTone}`}
            >
              {approvalLabel}
            </span>
          </div>
          <div className="mt-1 text-xs text-cortex-text-muted">
            {entry.actor} for {entry.user} at{" "}
            {new Date(entry.timestamp).toLocaleString()}
          </div>
        </div>
        <div className="flex items-center gap-2 text-[10px] font-mono text-cortex-text-muted">
          {entry.approval_status === "approval_required" ? (
            <ShieldAlert size={12} className="text-amber-300" />
          ) : (
            <BadgeCheck size={12} className="text-cortex-success" />
          )}
          {entry.template_id ?? "audit"}
        </div>
      </div>

      {operatorSummary ? (
        <p className="mt-3 text-sm leading-6 text-cortex-text-main">
          {operatorSummary}
        </p>
      ) : null}

      <div className="mt-3 flex flex-wrap gap-2 text-[10px] font-mono">
        {entry.capability_used ? (
          <span className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-1 text-cortex-primary">
            Capability: {entry.capability_used}
          </span>
        ) : null}
        {entry.approval_reason ? (
          <span className="rounded border border-amber-400/30 bg-amber-400/10 px-2 py-1 text-amber-300">
            Reason: {entry.approval_reason.replaceAll("_", " ")}
          </span>
        ) : null}
        {entry.run_id ? (
          <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
            Run: {entry.run_id}
          </span>
        ) : null}
        {entry.intent_proof_id ? (
          <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
            Proof: {entry.intent_proof_id}
          </span>
        ) : null}
        {entry.resource ? (
          <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
            Resource: {entry.resource}
          </span>
        ) : null}
        {teamID ? (
          <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
            Team: {teamID}
          </span>
        ) : null}
        {askKind ? (
          <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
            Ask: {askKind.replaceAll("_", " ")}
          </span>
        ) : null}
        {laneRole ? (
          <span className="rounded border border-cortex-border px-2 py-1 text-cortex-text-main">
            Role: {laneRole.replaceAll("_", " ")}
          </span>
        ) : null}
      </div>
    </div>
  );
}
