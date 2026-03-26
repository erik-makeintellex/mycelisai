"use client";

import { useEffect, useCallback, useState } from "react";
import { DecisionCard } from "@/components/approvals/DecisionCard";
import {
    ShieldCheck,
    Loader2,
    Settings,
    Plus,
    Trash2,
    Save,
    ChevronDown,
    ChevronRight,
    ScrollText,
    BadgeCheck,
    ShieldAlert,
} from "lucide-react";
import {
    useCortexStore,
    type AuditLogEntry,
    type PolicyConfig,
    type PolicyGroup,
    type PolicyRule,
} from "@/store/useCortexStore";
import TrustSlider from "@/components/workspace/TrustSlider";
import ManifestationPanel from "@/components/dashboard/ManifestationPanel";

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

    useEffect(() => {
        fetchPendingApprovals();
        const interval = setInterval(fetchPendingApprovals, 5000);
        return () => clearInterval(interval);
    }, [fetchPendingApprovals]);

    const handleResolve = useCallback(
        (id: string, approved: boolean) => {
            resolveApproval(id, approved);
        },
        [resolveApproval]
    );

    return (
        <div className="p-6 max-w-3xl mx-auto space-y-4">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl overflow-hidden">
                <TrustSlider />
            </div>

            <div className="flex items-center justify-between">
                <p className="text-xs text-cortex-text-muted font-mono">
                    {pendingApprovals.length} pending request{pendingApprovals.length !== 1 ? "s" : ""}
                </p>
                {isFetching && <Loader2 size={12} className="text-cortex-text-muted animate-spin" />}
            </div>

            {pendingApprovals.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 text-cortex-text-muted">
                    <ShieldCheck size={48} className="mb-4 text-cortex-success/50" />
                    <h2 className="text-lg font-semibold text-cortex-text-main">All Clear</h2>
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

    useEffect(() => {
        void fetchAuditLog();
    }, [fetchAuditLog]);

    return (
        <div className="p-6 max-w-4xl mx-auto space-y-4">
            <div className="flex items-start justify-between gap-4 rounded-xl border border-cortex-border bg-cortex-surface p-4">
                <div>
                    <h2 className="text-sm font-semibold text-cortex-text-main">Activity Log</h2>
                    <p className="mt-1 text-xs text-cortex-text-muted">
                        Inspect recent approvals, execution outcomes, capability use, and audit-ready governance events without dropping raw backend logs into the default workflow.
                    </p>
                </div>
                {isFetchingAuditLog ? <Loader2 size={16} className="mt-0.5 animate-spin text-cortex-text-muted" /> : null}
            </div>

            {auditLog.length === 0 ? (
                <div className="flex flex-col items-center justify-center rounded-xl border border-cortex-border bg-cortex-surface px-6 py-16 text-cortex-text-muted">
                    <ScrollText size={40} className="mb-3 opacity-50" />
                    <h3 className="text-base font-semibold text-cortex-text-main">No recent audit activity</h3>
                    <p className="mt-1 text-sm text-center">
                        Proposal generation, approvals, capability usage, and execution runs will appear here as they happen.
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
    const approvalTone = entry.approval_status === "approval_required" || entry.approval_status === "cancelled"
        ? "text-amber-300 border-amber-400/30"
        : "text-cortex-success border-cortex-success/30";
    const statusTone = entry.result_status === "failed" || entry.result_status === "error"
        ? "text-red-300 border-red-400/30"
        : entry.result_status === "pending"
            ? "text-amber-300 border-amber-400/30"
            : "text-cortex-success border-cortex-success/30";
    const approvalLabel = entry.approval_status ? entry.approval_status.replaceAll("_", " ") : "logged";
    const resultLabel = entry.result_status.replaceAll("_", " ");

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                    <div className="flex items-center gap-2">
                        <span className="text-sm font-semibold text-cortex-text-main">{entry.action.replaceAll("_", " ")}</span>
                        <span className={`rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase ${statusTone}`}>
                            {resultLabel}
                        </span>
                        <span className={`rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase ${approvalTone}`}>
                            {approvalLabel}
                        </span>
                    </div>
                    <div className="mt-1 text-xs text-cortex-text-muted">
                        {entry.actor} for {entry.user} at {new Date(entry.timestamp).toLocaleString()}
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
            </div>
        </div>
    );
}

// ── Policy Configuration Tab ──────────────────────────────

function PolicyTab() {
    const policyConfig = useCortexStore((s) => s.policyConfig);
    const isFetchingPolicy = useCortexStore((s) => s.isFetchingPolicy);
    const fetchPolicy = useCortexStore((s) => s.fetchPolicy);
    const updatePolicy = useCortexStore((s) => s.updatePolicy);

    const [draft, setDraft] = useState<PolicyConfig | null>(null);
    const [isSaving, setIsSaving] = useState(false);

    useEffect(() => {
        fetchPolicy();
    }, [fetchPolicy]);

    useEffect(() => {
        if (policyConfig && !draft) {
            setDraft(structuredClone(policyConfig));
        }
    }, [policyConfig, draft]);

    const config = draft ?? {
        groups: [],
        defaults: { default_action: "DENY" },
    };

    const handleSave = async () => {
        setIsSaving(true);
        await updatePolicy(config);
        setIsSaving(false);
    };

    const addGroup = () => {
        const newGroup: PolicyGroup = {
            name: `group-${config.groups.length + 1}`,
            description: "",
            targets: [],
            rules: [],
        };
        setDraft({ ...config, groups: [...config.groups, newGroup] });
    };

    const removeGroup = (idx: number) => {
        setDraft({ ...config, groups: config.groups.filter((_, i) => i !== idx) });
    };

    const updateGroup = (idx: number, patch: Partial<PolicyGroup>) => {
        setDraft({ ...config, groups: config.groups.map((g, i) => (i === idx ? { ...g, ...patch } : g)) });
    };

    const addRule = (groupIdx: number) => {
        const newRule: PolicyRule = { intent: "*", condition: "always", action: "REQUIRE_APPROVAL" };
        const groups = config.groups.map((g, i) =>
            i === groupIdx ? { ...g, rules: [...g.rules, newRule] } : g
        );
        setDraft({ ...config, groups });
    };

    const removeRule = (groupIdx: number, ruleIdx: number) => {
        const groups = config.groups.map((g, gi) =>
            gi === groupIdx ? { ...g, rules: g.rules.filter((_, ri) => ri !== ruleIdx) } : g
        );
        setDraft({ ...config, groups });
    };

    const updateRule = (groupIdx: number, ruleIdx: number, patch: Partial<PolicyRule>) => {
        const groups = config.groups.map((g, gi) =>
            gi === groupIdx
                ? { ...g, rules: g.rules.map((r, ri) => (ri === ruleIdx ? { ...r, ...patch } : r)) }
                : g
        );
        setDraft({ ...config, groups });
    };

    const setDefaultAction = (action: string) => {
        setDraft({ ...config, defaults: { default_action: action } });
    };

    if (isFetchingPolicy && !draft) {
        return (
            <div className="flex items-center justify-center py-20">
                <Loader2 size={24} className="text-cortex-primary animate-spin" />
            </div>
        );
    }

    return (
        <div className="p-6 max-w-4xl mx-auto space-y-4">
            <div className="flex items-center justify-between">
                <button
                    onClick={addGroup}
                    className="px-3 py-1.5 text-xs font-medium text-cortex-primary border border-cortex-primary/30 hover:bg-cortex-primary/10 rounded-lg flex items-center gap-1.5 transition-colors"
                >
                    <Plus size={14} />
                    Add Group
                </button>
                <button
                    onClick={handleSave}
                    disabled={isSaving}
                    className="px-4 py-1.5 text-xs font-medium text-cortex-bg bg-cortex-primary hover:bg-cortex-primary/90 rounded-lg flex items-center gap-1.5 shadow-sm transition-colors disabled:opacity-50"
                >
                    {isSaving ? <Loader2 size={14} className="animate-spin" /> : <Save size={14} />}
                    Save Policy
                </button>
            </div>

            {config.groups.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 text-cortex-text-muted">
                    <Settings size={40} className="mb-3 opacity-40" />
                    <p className="text-sm">No policy groups configured. Add one to get started.</p>
                </div>
            ) : (
                <div className="space-y-3">
                    {config.groups.map((group, gi) => (
                        <PolicyGroupCard
                            key={gi}
                            group={group}
                            onUpdate={(patch) => updateGroup(gi, patch)}
                            onRemove={() => removeGroup(gi)}
                            onAddRule={() => addRule(gi)}
                            onRemoveRule={(ri) => removeRule(gi, ri)}
                            onUpdateRule={(ri, patch) => updateRule(gi, ri, patch)}
                        />
                    ))}
                </div>
            )}

            <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h3 className="text-sm font-semibold text-cortex-text-main">Default Action</h3>
                        <p className="text-xs text-cortex-text-muted mt-0.5">Applied when no policy group rule matches</p>
                    </div>
                    <select
                        value={config.defaults.default_action}
                        onChange={(e) => setDefaultAction(e.target.value)}
                        className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
                    >
                        <option value="ALLOW">ALLOW</option>
                        <option value="DENY">DENY</option>
                        <option value="REQUIRE_APPROVAL">REQUIRE_APPROVAL</option>
                    </select>
                </div>
            </div>
        </div>
    );
}

// ── Policy Group Card ────────────────────────────────────

function PolicyGroupCard({
    group, onUpdate, onRemove, onAddRule, onRemoveRule, onUpdateRule,
}: {
    group: PolicyGroup;
    onUpdate: (patch: Partial<PolicyGroup>) => void;
    onRemove: () => void;
    onAddRule: () => void;
    onRemoveRule: (ruleIdx: number) => void;
    onUpdateRule: (ruleIdx: number, patch: Partial<PolicyRule>) => void;
}) {
    const [expanded, setExpanded] = useState(true);

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl overflow-hidden">
            <div
                className="px-4 py-3 flex items-center justify-between cursor-pointer hover:bg-cortex-bg/30 transition-colors"
                onClick={() => setExpanded(!expanded)}
            >
                <div className="flex items-center gap-2 min-w-0">
                    {expanded ? <ChevronDown size={14} className="text-cortex-text-muted flex-shrink-0" /> : <ChevronRight size={14} className="text-cortex-text-muted flex-shrink-0" />}
                    <div className="min-w-0">
                        <h3 className="text-sm font-semibold text-cortex-text-main truncate">{group.name || "Untitled Group"}</h3>
                        <p className="text-[10px] text-cortex-text-muted font-mono truncate">
                            {group.targets.length > 0 ? group.targets.join(", ") : "no targets"}{" -- "}{group.rules.length} rule{group.rules.length !== 1 ? "s" : ""}
                        </p>
                    </div>
                </div>
                <button onClick={(e) => { e.stopPropagation(); onRemove(); }} className="p-1 text-cortex-text-muted hover:text-cortex-danger rounded transition-colors" title="Delete group">
                    <Trash2 size={14} />
                </button>
            </div>

            {expanded && (
                <div className="border-t border-cortex-border p-4 space-y-4">
                    <div className="grid grid-cols-2 gap-3">
                        <div>
                            <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">Name</label>
                            <input type="text" value={group.name} onChange={(e) => onUpdate({ name: e.target.value })} className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary" />
                        </div>
                        <div>
                            <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">Targets (comma-separated)</label>
                            <input type="text" value={group.targets.join(", ")} onChange={(e) => onUpdate({ targets: e.target.value.split(",").map((t) => t.trim()).filter(Boolean) })} className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary" placeholder="team-*, agent-recon" />
                        </div>
                    </div>
                    <div>
                        <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">Description</label>
                        <input type="text" value={group.description} onChange={(e) => onUpdate({ description: e.target.value })} className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary" placeholder="What this policy group controls..." />
                    </div>

                    <div>
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-[10px] font-mono uppercase text-cortex-text-muted">Rules</span>
                            <button onClick={onAddRule} className="text-[10px] font-medium text-cortex-primary hover:text-cortex-primary/80 flex items-center gap-1 transition-colors">
                                <Plus size={10} />
                                Add Rule
                            </button>
                        </div>

                        {group.rules.length === 0 ? (
                            <p className="text-xs text-cortex-text-muted italic py-2">No rules. Add one to define behavior.</p>
                        ) : (
                            <div className="space-y-2">
                                <div className="grid grid-cols-[1fr_1fr_140px_28px] gap-2 text-[9px] font-mono uppercase text-cortex-text-muted px-1">
                                    <span>Intent Pattern</span>
                                    <span>Condition</span>
                                    <span>Action</span>
                                    <span />
                                </div>
                                {group.rules.map((rule, ri) => (
                                    <RuleRow key={ri} rule={rule} onUpdate={(patch) => onUpdateRule(ri, patch)} onRemove={() => onRemoveRule(ri)} />
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

// ── Rule Row ────────────────────────────────────────────

const ACTION_COLORS: Record<string, string> = {
    ALLOW: "text-cortex-success border-cortex-success/30",
    DENY: "text-cortex-danger border-cortex-danger/30",
    REQUIRE_APPROVAL: "text-cortex-warning border-cortex-warning/30",
};

function RuleRow({ rule, onUpdate, onRemove }: { rule: PolicyRule; onUpdate: (patch: Partial<PolicyRule>) => void; onRemove: () => void }) {
    return (
        <div className="grid grid-cols-[1fr_1fr_140px_28px] gap-2 items-center">
            <input type="text" value={rule.intent} onChange={(e) => onUpdate({ intent: e.target.value })} className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded px-2 py-1 focus:outline-none focus:border-cortex-primary" placeholder="file.write.*" />
            <input type="text" value={rule.condition} onChange={(e) => onUpdate({ condition: e.target.value })} className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded px-2 py-1 focus:outline-none focus:border-cortex-primary" placeholder="trust < 0.7" />
            <select value={rule.action} onChange={(e) => onUpdate({ action: e.target.value as PolicyRule["action"] })} className={`bg-cortex-bg border text-xs font-mono font-semibold rounded px-2 py-1 focus:outline-none ${ACTION_COLORS[rule.action] ?? "text-cortex-text-main border-cortex-border"}`}>
                <option value="ALLOW">ALLOW</option>
                <option value="DENY">DENY</option>
                <option value="REQUIRE_APPROVAL">REQUIRE_APPROVAL</option>
            </select>
            <button onClick={onRemove} className="p-1 text-cortex-text-muted hover:text-cortex-danger rounded transition-colors" title="Delete rule">
                <Trash2 size={12} />
            </button>
        </div>
    );
}
