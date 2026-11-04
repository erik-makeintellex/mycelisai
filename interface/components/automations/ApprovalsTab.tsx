"use client";

import { useEffect, useCallback, useState } from "react";
import { DecisionCard } from "@/components/approvals/DecisionCard";
import { ShieldCheck, Loader2, ScrollText } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import TrustSlider from "@/components/workspace/TrustSlider";
import ManifestationPanel from "@/components/dashboard/ManifestationPanel";
import LoadFailureState from "@/components/shared/LoadFailureState";
import AuditEntryCard from "@/components/automations/AuditEntryCard";
import { PolicyTab } from "./ApprovalsPolicyTab";

type SubTab = "queue" | "policy" | "proposals" | "audit";

export default function ApprovalsTab() {
  const [subTab, setSubTab] = useState<SubTab>("queue");

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 pt-4 flex gap-4 border-b border-cortex-border/50">
        <button
          onClick={() => setSubTab("queue")}
          className={`pb-2 text-xs font-medium border-b-2 transition-colors -mb-px ${subTab === "queue" ? "border-cortex-primary text-cortex-primary" : "border-transparent text-cortex-text-muted"}`}
        >
          Queue
        </button>
        <button
          onClick={() => setSubTab("policy")}
          className={`pb-2 text-xs font-medium border-b-2 transition-colors -mb-px ${subTab === "policy" ? "border-cortex-primary text-cortex-primary" : "border-transparent text-cortex-text-muted"}`}
        >
          Policy
        </button>
        <button
          onClick={() => setSubTab("proposals")}
          className={`pb-2 text-xs font-medium border-b-2 transition-colors -mb-px ${subTab === "proposals" ? "border-cortex-primary text-cortex-primary" : "border-transparent text-cortex-text-muted"}`}
        >
          Proposals
        </button>
        <button
          onClick={() => setSubTab("audit")}
          className={`pb-2 text-xs font-medium border-b-2 transition-colors -mb-px ${subTab === "audit" ? "border-cortex-primary text-cortex-primary" : "border-transparent text-cortex-text-muted"}`}
        >
          Audit
        </button>
      </div>
      <div className="flex-1 overflow-y-auto">
        {subTab === "queue" && <ApprovalsQueue />}
        {subTab === "policy" && <PolicyTab />}
        {subTab === "proposals" && (
          <div className="p-6 max-w-3xl mx-auto">
            <ManifestationPanel />
          </div>
        )}
        {subTab === "audit" && <AuditTab />}
      </div>
    </div>
  );
}

function ApprovalsQueue() {
  const pendingApprovals = useCortexStore((s) => s.pendingApprovals);
  const isFetching = useCortexStore((s) => s.isFetchingApprovals);
  const fetchPendingApprovals = useCortexStore((s) => s.fetchPendingApprovals);
  const resolveApproval = useCortexStore((s) => s.resolveApproval);
  const approvalsLoadError = useCortexStore((s) => s.approvalsLoadError);

  useEffect(() => {
    fetchPendingApprovals();
    const interval = setInterval(fetchPendingApprovals, 5000);
    return () => clearInterval(interval);
  }, [fetchPendingApprovals]);

  const handleResolve = useCallback(
    (id: string, approved: boolean) => {
      resolveApproval(id, approved);
    },
    [resolveApproval],
  );

  return (
    <div className="p-6 max-w-3xl mx-auto space-y-4">
      <div className="bg-cortex-surface border border-cortex-border rounded-xl overflow-hidden">
        <TrustSlider />
      </div>

      <div className="flex items-center justify-between">
        <p className="text-xs text-cortex-text-muted font-mono">
          {pendingApprovals.length} pending request
          {pendingApprovals.length !== 1 ? "s" : ""}
        </p>
        {isFetching && (
          <Loader2 size={12} className="text-cortex-text-muted animate-spin" />
        )}
      </div>

      {!isFetching && approvalsLoadError && pendingApprovals.length === 0 ? (
        <LoadFailureState
          title="Could not load approvals"
          detail={approvalsLoadError}
          onRetry={() => void fetchPendingApprovals()}
          testId="approvals-load-failure"
        />
      ) : pendingApprovals.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-cortex-text-muted">
          <ShieldCheck size={48} className="mb-4 text-cortex-success/50" />
          <h2 className="text-lg font-semibold text-cortex-text-main">
            All Clear
          </h2>
          <p className="text-sm">No pending governance requests.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {pendingApprovals.map((a) => (
            <DecisionCard key={a.id} approval={a} onResolve={handleResolve} />
          ))}
        </div>
      )}
    </div>
  );
}

function AuditTab() {
  const auditLog = useCortexStore((s) => s.auditLog);
  const isFetchingAuditLog = useCortexStore((s) => s.isFetchingAuditLog);
  const fetchAuditLog = useCortexStore((s) => s.fetchAuditLog);
  const auditLogLoadError = useCortexStore((s) => s.auditLogLoadError);

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

      {!isFetchingAuditLog && auditLogLoadError && auditLog.length === 0 ? (
        <LoadFailureState
          title="Could not load the audit log"
          detail={auditLogLoadError}
          onRetry={() => void fetchAuditLog()}
          testId="audit-log-load-failure"
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
