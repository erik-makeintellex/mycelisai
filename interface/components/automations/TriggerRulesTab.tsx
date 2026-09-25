"use client";

import { useEffect, useState } from "react";
import {
    ScrollText,
    Plus,
    Trash2,
    Power,
    PowerOff,
    ChevronDown,
    ChevronRight,
    Loader2,
    Clock,
    Shield,
    Layers,
    Zap,
} from "lucide-react";
import { useCortexStore, type TriggerRule } from "@/store/useCortexStore";
import LoadFailureState from "@/components/shared/LoadFailureState";
import CreateRuleForm from "@/components/automations/CreateRuleForm";

export default function TriggerRulesTab() {
    const rules = useCortexStore((s) => s.triggerRules);
    const isFetching = useCortexStore((s) => s.isFetchingTriggers);
    const fetchTriggerRules = useCortexStore((s) => s.fetchTriggerRules);
    const triggerRulesLoadError = useCortexStore((s) => s.triggerRulesLoadError);
    const createTriggerRule = useCortexStore((s) => s.createTriggerRule);
    const deleteTriggerRule = useCortexStore((s) => s.deleteTriggerRule);
    const toggleTriggerRule = useCortexStore((s) => s.toggleTriggerRule);
    const [showCreate, setShowCreate] = useState(false);
    const [expandedRule, setExpandedRule] = useState<string | null>(null);

    useEffect(() => {
        fetchTriggerRules();
    }, [fetchTriggerRules]);

    return (
        <div className="h-full flex flex-col p-6 gap-4">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <ScrollText size={18} className="text-cortex-primary" />
                    <div>
                        <h2 className="text-sm font-semibold text-cortex-text-main">
                            Trigger Rules
                        </h2>
                        <p className="text-xs text-cortex-text-muted">
                            Declarative IF/THEN rules evaluated on event ingest
                        </p>
                    </div>
                </div>
                <button
                    onClick={() => setShowCreate(!showCreate)}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md bg-cortex-primary/10 text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                >
                    <Plus size={14} />
                    New Rule
                </button>
            </div>

            {/* Create Form */}
            {showCreate && (
                <CreateRuleForm
                    onCreate={async (rule) => {
                        await createTriggerRule(rule);
                        setShowCreate(false);
                    }}
                    onCancel={() => setShowCreate(false)}
                />
            )}

            {/* Rules List */}
            <div className="flex-1 overflow-y-auto space-y-2">
                {isFetching && rules.length === 0 ? (
                    <div className="flex items-center justify-center h-32">
                        <Loader2 size={18} className="animate-spin text-cortex-text-muted" />
                    </div>
                ) : !isFetching && triggerRulesLoadError && rules.length === 0 ? (
                    <LoadFailureState
                        title="Could not load trigger rules"
                        detail={triggerRulesLoadError}
                        onRetry={() => void fetchTriggerRules()}
                        testId="trigger-rules-load-failure"
                    />
                ) : rules.length === 0 ? (
                    <EmptyState />
                ) : (
                    rules.map((rule) => (
                        <RuleCard
                            key={rule.id}
                            rule={rule}
                            expanded={expandedRule === rule.id}
                            onToggleExpand={() =>
                                setExpandedRule(expandedRule === rule.id ? null : rule.id)
                            }
                            onToggleActive={() =>
                                toggleTriggerRule(rule.id, !rule.is_active)
                            }
                            onDelete={() => deleteTriggerRule(rule.id)}
                        />
                    ))
                )}
            </div>

            {/* Footer summary */}
            {rules.length > 0 && (
                <div className="text-xs text-cortex-text-muted border-t border-cortex-border/50 pt-3">
                    {rules.filter((r) => r.is_active).length} active /{" "}
                    {rules.length} total rules
                </div>
            )}
        </div>
    );
}

// ── Rule Card ────────────────────────────────────────────────

function RuleCard({
    rule,
    expanded,
    onToggleExpand,
    onToggleActive,
    onDelete,
}: {
    rule: TriggerRule;
    expanded: boolean;
    onToggleExpand: () => void;
    onToggleActive: () => void;
    onDelete: () => void;
}) {
    const [confirmDelete, setConfirmDelete] = useState(false);

    return (
        <div
            className={`rounded-lg border transition-colors ${
                rule.is_active
                    ? "border-cortex-border bg-cortex-surface"
                    : "border-cortex-border/50 bg-cortex-surface/50 opacity-60"
            }`}
        >
            {/* Header row */}
            <div
                className="flex items-center gap-3 px-4 py-3 cursor-pointer"
                onClick={onToggleExpand}
            >
                {expanded ? (
                    <ChevronDown size={14} className="text-cortex-text-muted flex-shrink-0" />
                ) : (
                    <ChevronRight size={14} className="text-cortex-text-muted flex-shrink-0" />
                )}

                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-cortex-text-main truncate">
                            {rule.name}
                        </span>
                        <span
                            className={`px-1.5 py-0.5 text-[10px] font-mono rounded ${
                                rule.mode === "auto_execute"
                                    ? "bg-amber-500/10 text-amber-400"
                                    : "bg-cortex-primary/10 text-cortex-primary"
                            }`}
                        >
                            {rule.mode === "auto_execute" ? "auto" : "propose"}
                        </span>
                        {!rule.is_active && (
                            <span className="px-1.5 py-0.5 text-[10px] font-mono rounded bg-cortex-border text-cortex-text-muted">
                                disabled
                            </span>
                        )}
                    </div>
                    <div className="flex items-center gap-3 mt-0.5">
                        <span className="text-xs text-cortex-text-muted font-mono">
                            {rule.event_pattern}
                        </span>
                        <span className="text-xs text-cortex-text-muted">
                            &rarr; {rule.target_mission_id.slice(0, 8)}...
                        </span>
                    </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                    <button
                        onClick={onToggleActive}
                        className={`p-1.5 rounded transition-colors ${
                            rule.is_active
                                ? "text-cortex-success hover:bg-cortex-success/10"
                                : "text-cortex-text-muted hover:bg-cortex-border"
                        }`}
                        title={rule.is_active ? "Deactivate" : "Activate"}
                    >
                        {rule.is_active ? <Power size={14} /> : <PowerOff size={14} />}
                    </button>
                    {confirmDelete ? (
                        <div className="flex items-center gap-1">
                            <button
                                onClick={() => {
                                    onDelete();
                                    setConfirmDelete(false);
                                }}
                                className="px-2 py-1 text-[10px] font-medium rounded bg-red-500/10 text-red-400 hover:bg-red-500/20"
                            >
                                Confirm
                            </button>
                            <button
                                onClick={() => setConfirmDelete(false)}
                                className="px-2 py-1 text-[10px] font-medium rounded text-cortex-text-muted hover:bg-cortex-border"
                            >
                                Cancel
                            </button>
                        </div>
                    ) : (
                        <button
                            onClick={() => setConfirmDelete(true)}
                            className="p-1.5 rounded text-cortex-text-muted hover:text-red-400 hover:bg-red-500/10 transition-colors"
                            title="Delete"
                        >
                            <Trash2 size={14} />
                        </button>
                    )}
                </div>
            </div>

            {/* Expanded details */}
            {expanded && (
                <div className="px-4 pb-3 border-t border-cortex-border/50 pt-3 space-y-2">
                    {rule.description && (
                        <p className="text-xs text-cortex-text-muted">{rule.description}</p>
                    )}
                    <div className="grid grid-cols-3 gap-3">
                        <GuardBadge
                            icon={<Clock size={12} />}
                            label="Cooldown"
                            value={`${rule.cooldown_seconds}s`}
                        />
                        <GuardBadge
                            icon={<Layers size={12} />}
                            label="Max Depth"
                            value={String(rule.max_depth)}
                        />
                        <GuardBadge
                            icon={<Shield size={12} />}
                            label="Max Runs"
                            value={String(rule.max_active_runs)}
                        />
                    </div>
                    {rule.last_fired_at && (
                        <p className="text-[10px] text-cortex-text-muted font-mono">
                            Last fired: {new Date(rule.last_fired_at).toLocaleString()}
                        </p>
                    )}
                    <p className="text-[10px] text-cortex-text-muted font-mono">
                        Target: {rule.target_mission_id}
                    </p>
                </div>
            )}
        </div>
    );
}

function GuardBadge({
    icon,
    label,
    value,
}: {
    icon: React.ReactNode;
    label: string;
    value: string;
}) {
    return (
        <div className="flex items-center gap-1.5 text-xs text-cortex-text-muted">
            {icon}
            <span>{label}:</span>
            <span className="font-mono text-cortex-text-main">{value}</span>
        </div>
    );
}

// ── Empty State ──────────────────────────────────────────────

function EmptyState() {
    return (
        <div className="flex flex-col items-center justify-center h-48 text-center gap-3">
            <div className="w-12 h-12 rounded-full bg-cortex-primary/5 flex items-center justify-center">
                <Zap size={20} className="text-cortex-primary/40" />
            </div>
            <div>
                <p className="text-sm text-cortex-text-main font-medium">
                    No trigger rules defined
                </p>
                <p className="text-xs text-cortex-text-muted mt-1 max-w-xs">
                    Trigger rules fire automatically when specific mission events occur.
                    Create your first rule to get started.
                </p>
            </div>
        </div>
    );
}

