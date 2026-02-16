"use client";

import { useEffect, useState, useCallback } from "react";
import { DecisionCard } from "@/components/approvals/DecisionCard";
import {
    ShieldCheck,
    ScrollText,
    Settings,
    Plus,
    Trash2,
    Save,
    ChevronDown,
    ChevronRight,
    Loader2,
} from "lucide-react";
import {
    useCortexStore,
    type PolicyConfig,
    type PolicyGroup,
    type PolicyRule,
} from "@/store/useCortexStore";
import TrustSlider from "@/components/workspace/TrustSlider";

type TabId = "approvals" | "policy";

// ── Main Page ────────────────────────────────────────────────

export default function ApprovalsPage() {
    const [activeTab, setActiveTab] = useState<TabId>("approvals");

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            {/* Top Bar */}
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            Governance
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Approval queue and policy configuration
                        </p>
                    </div>
                </div>

                {/* Tab Bar */}
                <div className="flex gap-1 border-b border-cortex-border">
                    <TabButton
                        active={activeTab === "approvals"}
                        onClick={() => setActiveTab("approvals")}
                        icon={<ScrollText size={14} />}
                        label="Approvals Queue"
                    />
                    <TabButton
                        active={activeTab === "policy"}
                        onClick={() => setActiveTab("policy")}
                        icon={<Settings size={14} />}
                        label="Policy Configuration"
                    />
                </div>
            </header>

            {/* Tab Content */}
            <div className="flex-1 overflow-y-auto">
                {activeTab === "approvals" && <ApprovalsTab />}
                {activeTab === "policy" && <PolicyTab />}
            </div>
        </div>
    );
}

// ── Tab Button ───────────────────────────────────────────────

function TabButton({
    active,
    onClick,
    icon,
    label,
}: {
    active: boolean;
    onClick: () => void;
    icon: React.ReactNode;
    label: string;
}) {
    return (
        <button
            onClick={onClick}
            className={`px-4 py-2.5 text-xs font-medium flex items-center gap-2 border-b-2 transition-colors -mb-px ${
                active
                    ? "border-cortex-primary text-cortex-primary"
                    : "border-transparent text-cortex-text-muted hover:text-cortex-text-main"
            }`}
        >
            {icon}
            {label}
        </button>
    );
}

// ── Tab 1: Approvals Queue ───────────────────────────────────

function ApprovalsTab() {
    const pendingApprovals = useCortexStore((s) => s.pendingApprovals);
    const isFetching = useCortexStore((s) => s.isFetchingApprovals);
    const fetchPendingApprovals = useCortexStore((s) => s.fetchPendingApprovals);
    const resolveApproval = useCortexStore((s) => s.resolveApproval);

    // Initial fetch + 5-second polling
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
            {/* Trust Slider */}
            <div className="bg-cortex-surface border border-cortex-border rounded-xl overflow-hidden">
                <TrustSlider />
            </div>

            {/* Status Bar */}
            <div className="flex items-center justify-between">
                <p className="text-xs text-cortex-text-muted font-mono">
                    {pendingApprovals.length} pending request
                    {pendingApprovals.length !== 1 ? "s" : ""}
                </p>
                {isFetching && (
                    <Loader2
                        size={12}
                        className="text-cortex-text-muted animate-spin"
                    />
                )}
            </div>

            {/* Cards or Empty State */}
            {pendingApprovals.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 text-cortex-text-muted">
                    <ShieldCheck
                        size={48}
                        className="mb-4 text-cortex-success/50"
                    />
                    <h2 className="text-lg font-semibold text-cortex-text-main">
                        All Clear
                    </h2>
                    <p className="text-sm">
                        No pending governance requests.
                    </p>
                </div>
            ) : (
                <div className="space-y-4">
                    {pendingApprovals.map((a) => (
                        <DecisionCard
                            key={a.id}
                            approval={a}
                            onResolve={handleResolve}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}

// ── Tab 2: Policy Configuration ──────────────────────────────

function PolicyTab() {
    const policyConfig = useCortexStore((s) => s.policyConfig);
    const isFetchingPolicy = useCortexStore((s) => s.isFetchingPolicy);
    const fetchPolicy = useCortexStore((s) => s.fetchPolicy);
    const updatePolicy = useCortexStore((s) => s.updatePolicy);

    // Local draft for editing
    const [draft, setDraft] = useState<PolicyConfig | null>(null);
    const [isSaving, setIsSaving] = useState(false);

    // Fetch on mount
    useEffect(() => {
        fetchPolicy();
    }, [fetchPolicy]);

    // Sync remote config into local draft
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

    // ── Group CRUD ──

    const addGroup = () => {
        const newGroup: PolicyGroup = {
            name: `group-${config.groups.length + 1}`,
            description: "",
            targets: [],
            rules: [],
        };
        setDraft({
            ...config,
            groups: [...config.groups, newGroup],
        });
    };

    const removeGroup = (idx: number) => {
        setDraft({
            ...config,
            groups: config.groups.filter((_, i) => i !== idx),
        });
    };

    const updateGroup = (idx: number, patch: Partial<PolicyGroup>) => {
        setDraft({
            ...config,
            groups: config.groups.map((g, i) =>
                i === idx ? { ...g, ...patch } : g
            ),
        });
    };

    // ── Rule CRUD ──

    const addRule = (groupIdx: number) => {
        const newRule: PolicyRule = {
            intent: "*",
            condition: "always",
            action: "REQUIRE_APPROVAL",
        };
        const groups = config.groups.map((g, i) =>
            i === groupIdx ? { ...g, rules: [...g.rules, newRule] } : g
        );
        setDraft({ ...config, groups });
    };

    const removeRule = (groupIdx: number, ruleIdx: number) => {
        const groups = config.groups.map((g, gi) =>
            gi === groupIdx
                ? { ...g, rules: g.rules.filter((_, ri) => ri !== ruleIdx) }
                : g
        );
        setDraft({ ...config, groups });
    };

    const updateRule = (
        groupIdx: number,
        ruleIdx: number,
        patch: Partial<PolicyRule>
    ) => {
        const groups = config.groups.map((g, gi) =>
            gi === groupIdx
                ? {
                      ...g,
                      rules: g.rules.map((r, ri) =>
                          ri === ruleIdx ? { ...r, ...patch } : r
                      ),
                  }
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
                <Loader2
                    size={24}
                    className="text-cortex-primary animate-spin"
                />
            </div>
        );
    }

    return (
        <div className="p-6 max-w-4xl mx-auto space-y-4">
            {/* Top Actions */}
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
                    {isSaving ? (
                        <Loader2 size={14} className="animate-spin" />
                    ) : (
                        <Save size={14} />
                    )}
                    Save Policy
                </button>
            </div>

            {/* Policy Groups */}
            {config.groups.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 text-cortex-text-muted">
                    <Settings size={40} className="mb-3 opacity-40" />
                    <p className="text-sm">
                        No policy groups configured. Add one to get started.
                    </p>
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
                            onUpdateRule={(ri, patch) =>
                                updateRule(gi, ri, patch)
                            }
                        />
                    ))}
                </div>
            )}

            {/* Default Action */}
            <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h3 className="text-sm font-semibold text-cortex-text-main">
                            Default Action
                        </h3>
                        <p className="text-xs text-cortex-text-muted mt-0.5">
                            Applied when no policy group rule matches
                        </p>
                    </div>
                    <select
                        value={config.defaults.default_action}
                        onChange={(e) => setDefaultAction(e.target.value)}
                        className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
                    >
                        <option value="ALLOW">ALLOW</option>
                        <option value="DENY">DENY</option>
                        <option value="REQUIRE_APPROVAL">
                            REQUIRE_APPROVAL
                        </option>
                    </select>
                </div>
            </div>
        </div>
    );
}

// ── Policy Group Card ────────────────────────────────────────

function PolicyGroupCard({
    group,
    onUpdate,
    onRemove,
    onAddRule,
    onRemoveRule,
    onUpdateRule,
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
            {/* Group Header */}
            <div
                className="px-4 py-3 flex items-center justify-between cursor-pointer hover:bg-cortex-bg/30 transition-colors"
                onClick={() => setExpanded(!expanded)}
            >
                <div className="flex items-center gap-2 min-w-0">
                    {expanded ? (
                        <ChevronDown
                            size={14}
                            className="text-cortex-text-muted flex-shrink-0"
                        />
                    ) : (
                        <ChevronRight
                            size={14}
                            className="text-cortex-text-muted flex-shrink-0"
                        />
                    )}
                    <div className="min-w-0">
                        <h3 className="text-sm font-semibold text-cortex-text-main truncate">
                            {group.name || "Untitled Group"}
                        </h3>
                        <p className="text-[10px] text-cortex-text-muted font-mono truncate">
                            {group.targets.length > 0
                                ? group.targets.join(", ")
                                : "no targets"}
                            {" -- "}
                            {group.rules.length} rule
                            {group.rules.length !== 1 ? "s" : ""}
                        </p>
                    </div>
                </div>
                <button
                    onClick={(e) => {
                        e.stopPropagation();
                        onRemove();
                    }}
                    className="p-1 text-cortex-text-muted hover:text-cortex-danger rounded transition-colors"
                    title="Delete group"
                >
                    <Trash2 size={14} />
                </button>
            </div>

            {/* Expanded Content */}
            {expanded && (
                <div className="border-t border-cortex-border p-4 space-y-4">
                    {/* Group Meta Fields */}
                    <div className="grid grid-cols-2 gap-3">
                        <div>
                            <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">
                                Name
                            </label>
                            <input
                                type="text"
                                value={group.name}
                                onChange={(e) =>
                                    onUpdate({ name: e.target.value })
                                }
                                className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
                            />
                        </div>
                        <div>
                            <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">
                                Targets (comma-separated)
                            </label>
                            <input
                                type="text"
                                value={group.targets.join(", ")}
                                onChange={(e) =>
                                    onUpdate({
                                        targets: e.target.value
                                            .split(",")
                                            .map((t) => t.trim())
                                            .filter(Boolean),
                                    })
                                }
                                className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
                                placeholder="team-*, agent-recon"
                            />
                        </div>
                    </div>
                    <div>
                        <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">
                            Description
                        </label>
                        <input
                            type="text"
                            value={group.description}
                            onChange={(e) =>
                                onUpdate({ description: e.target.value })
                            }
                            className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
                            placeholder="What this policy group controls..."
                        />
                    </div>

                    {/* Rules Table */}
                    <div>
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-[10px] font-mono uppercase text-cortex-text-muted">
                                Rules
                            </span>
                            <button
                                onClick={onAddRule}
                                className="text-[10px] font-medium text-cortex-primary hover:text-cortex-primary/80 flex items-center gap-1 transition-colors"
                            >
                                <Plus size={10} />
                                Add Rule
                            </button>
                        </div>

                        {group.rules.length === 0 ? (
                            <p className="text-xs text-cortex-text-muted italic py-2">
                                No rules. Add one to define behavior.
                            </p>
                        ) : (
                            <div className="space-y-2">
                                {/* Header row */}
                                <div className="grid grid-cols-[1fr_1fr_140px_28px] gap-2 text-[9px] font-mono uppercase text-cortex-text-muted px-1">
                                    <span>Intent Pattern</span>
                                    <span>Condition</span>
                                    <span>Action</span>
                                    <span />
                                </div>
                                {group.rules.map((rule, ri) => (
                                    <RuleRow
                                        key={ri}
                                        rule={rule}
                                        onUpdate={(patch) =>
                                            onUpdateRule(ri, patch)
                                        }
                                        onRemove={() => onRemoveRule(ri)}
                                    />
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

// ── Rule Row (Inline Editable) ───────────────────────────────

const ACTION_COLORS: Record<string, string> = {
    ALLOW: "text-cortex-success border-cortex-success/30",
    DENY: "text-cortex-danger border-cortex-danger/30",
    REQUIRE_APPROVAL: "text-cortex-warning border-cortex-warning/30",
};

function RuleRow({
    rule,
    onUpdate,
    onRemove,
}: {
    rule: PolicyRule;
    onUpdate: (patch: Partial<PolicyRule>) => void;
    onRemove: () => void;
}) {
    return (
        <div className="grid grid-cols-[1fr_1fr_140px_28px] gap-2 items-center">
            <input
                type="text"
                value={rule.intent}
                onChange={(e) => onUpdate({ intent: e.target.value })}
                className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded px-2 py-1 focus:outline-none focus:border-cortex-primary"
                placeholder="file.write.*"
            />
            <input
                type="text"
                value={rule.condition}
                onChange={(e) => onUpdate({ condition: e.target.value })}
                className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded px-2 py-1 focus:outline-none focus:border-cortex-primary"
                placeholder="trust < 0.7"
            />
            <select
                value={rule.action}
                onChange={(e) =>
                    onUpdate({
                        action: e.target.value as PolicyRule["action"],
                    })
                }
                className={`bg-cortex-bg border text-xs font-mono font-semibold rounded px-2 py-1 focus:outline-none ${
                    ACTION_COLORS[rule.action] ?? "text-cortex-text-main border-cortex-border"
                }`}
            >
                <option value="ALLOW">ALLOW</option>
                <option value="DENY">DENY</option>
                <option value="REQUIRE_APPROVAL">REQUIRE_APPROVAL</option>
            </select>
            <button
                onClick={onRemove}
                className="p-1 text-cortex-text-muted hover:text-cortex-danger rounded transition-colors"
                title="Delete rule"
            >
                <Trash2 size={12} />
            </button>
        </div>
    );
}
