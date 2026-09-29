"use client";

import { useEffect, useState } from "react";
import { BadgeCheck, Loader2, ScrollText, ShieldAlert } from "lucide-react";
import { useCortexStore, type AuditLogEntry } from "@/store/useCortexStore";
import ManifestationPanel from "@/components/dashboard/ManifestationPanel";
import ScheduleAuditContext from "@/components/automations/ScheduleAuditContext";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import PolicyTab from "@/components/automations/ApprovalsPolicyTab";

// Proposals is the durable approval path, so it is the default view. The old
// in-memory "Queue" was retired (C2-RETIRE): it could only report fake success.
type SubTab = "proposals" | "policy" | "audit";

const SUB_TABS: { id: SubTab; label: string }[] = [
  { id: "proposals", label: "Proposals" },
  { id: "policy", label: "Policy" },
  { id: "audit", label: "Audit" },
];

export default function ApprovalsTab() {
  const [subTab, setSubTab] = useState<SubTab>("proposals");

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 pt-4 flex gap-4 border-b border-cortex-border/50">
        {SUB_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setSubTab(tab.id)}
            className={`pb-2 text-xs font-medium border-b-2 transition-colors -mb-px ${subTab === tab.id ? "border-cortex-primary text-cortex-primary" : "border-transparent text-cortex-text-muted"}`}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="flex-1 overflow-y-auto">
        {subTab === "proposals" && (
          <div className="p-6 max-w-3xl mx-auto">
            <ManifestationPanel />
          </div>
        )}
        {subTab === "policy" && <PolicyTab />}
        {subTab === "audit" && <AuditTab />}
      </div>
    </div>
  );
}

function AuditTab() {
  const auditLog = useCortexStore((s) => s.auditLog);
  const isFetchingAuditLog = useCortexStore((s) => s.isFetchingAuditLog);
  const auditLogError = useCortexStore((s) => s.auditLogError);
  const fetchAuditLog = useCortexStore((s) => s.fetchAuditLog);

  useEffect(() => {
    void fetchAuditLog();
  }, [fetchAuditLog]);

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-4">
      <div className="flex items-start justify-between gap-4 rounded-xl border border-cortex-border bg-cortex-surface p-4">
        <div>
          <h2 className="text-sm font-semibold text-cortex-text-main">
            Activity Log
          </h2>
          <p className="mt-1 text-xs text-cortex-text-muted">
            Inspect recent approvals, execution outcomes, capability use, and
            audit-ready governance events without dropping raw backend logs into
            the default workflow.
          </p>
        </div>
        {isFetchingAuditLog ? (
          <Loader2
            size={16}
            className="mt-0.5 animate-spin text-cortex-text-muted"
          />
        ) : null}
      </div>

      {auditLogError ? (
        <InlineBlockerNotice
          code={auditLogError.code}
          httpStatus={auditLogError.httpStatus}
          retryLabel="Try again"
          onRetry={() => void fetchAuditLog()}
        />
      ) : auditLog.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-cortex-border bg-cortex-surface px-6 py-16 text-cortex-text-muted">
          <ScrollText size={40} className="mb-3 opacity-50" />
          <h3 className="text-base font-semibold text-cortex-text-main">
            No recent audit activity
          </h3>
          <p className="mt-1 text-sm text-center">
            Proposal generation, approvals, capability usage, and execution runs
            will appear here as they happen.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {auditLog.map((entry) => (
            <AuditEntryCard key={entry.id} entry={entry} />
          ))}
        </div>
      )}
    </div>
  );
}

function AuditEntryCard({ entry }: { entry: AuditLogEntry }) {
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

      <ScheduleAuditContext entry={entry} />

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
