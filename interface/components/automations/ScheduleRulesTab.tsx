"use client";

import { useEffect, useMemo, useState } from "react";
import { CalendarClock, Clock3, History, Loader2, Plus, ShieldCheck } from "lucide-react";
import { useCortexStore, type TriggerRule, type TriggerRuleCreate } from "@/store/useCortexStore";
import {
    getScheduleHandoffState,
} from "@/components/automations/scheduleHandoffState";
import {
    ScheduleHandoffActions,
    ScheduleHandoffBadge,
} from "@/components/automations/ScheduleHandoffControls";

function formatDate(value?: string) {
    if (!value) return "Awaiting first schedule";
    return new Date(value).toLocaleString();
}

function formatInterval(seconds?: number) {
    if (!seconds) return "Not configured";
    if (seconds % 3600 === 0) return `${seconds / 3600}h cadence`;
    if (seconds % 60 === 0) return `${seconds / 60}m cadence`;
    return `${seconds}s cadence`;
}

function getRuleHandoffState(rule: TriggerRule) {
    return (
        getScheduleHandoffState(rule) ||
        getScheduleHandoffState(rule.latest_execution) ||
        getScheduleHandoffState(rule.last_execution) ||
        getScheduleHandoffState(rule.recent_execution)
    );
}

function getRuleHandoffExecution(rule: TriggerRule) {
    for (const execution of [rule.latest_execution, rule.last_execution, rule.recent_execution]) {
        if (execution?.id && getScheduleHandoffState(execution)) return execution;
    }
    return null;
}

export default function ScheduleRulesTab() {
    const rules = useCortexStore((s) => s.triggerRules);
    const isFetching = useCortexStore((s) => s.isFetchingTriggers);
    const fetchTriggerRules = useCortexStore((s) => s.fetchTriggerRules);
    const createTriggerRule = useCortexStore((s) => s.createTriggerRule);
    const toggleTriggerRule = useCortexStore((s) => s.toggleTriggerRule);
    const resolveScheduleHandoff = useCortexStore((s) => s.resolveScheduleHandoff);
    const [showCreate, setShowCreate] = useState(false);
    const [pendingToggleRuleID, setPendingToggleRuleID] = useState<string | null>(null);
    const [pendingHandoffAction, setPendingHandoffAction] = useState<string | null>(null);

    useEffect(() => {
        fetchTriggerRules();
    }, [fetchTriggerRules]);

    const schedules = useMemo(
        () => rules.filter((rule) => rule.trigger_kind === "schedule"),
        [rules],
    );

    return (
        <div className="h-full flex flex-col p-6 gap-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex min-w-0 items-center gap-3">
                    <CalendarClock size={18} className="text-cortex-primary flex-shrink-0" />
                    <div className="min-w-0">
                        <h2 className="text-sm font-semibold text-cortex-text-main">Schedule Rules</h2>
                        <p className="text-xs text-cortex-text-muted">
                            Cadence proposals with cooldown, proof expectations, and recovery text.
                        </p>
                    </div>
                </div>
                <button
                    type="button"
                    onClick={() => setShowCreate((value) => !value)}
                    className="flex flex-shrink-0 items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md bg-cortex-primary/10 text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                >
                    <Plus size={14} />
                    New Schedule
                </button>
            </div>

            {showCreate && (
                <CreateScheduleForm
                    onCancel={() => setShowCreate(false)}
                    onCreate={async (rule) => {
                        await createTriggerRule(rule);
                        setShowCreate(false);
                    }}
                />
            )}

            <div className="flex-1 overflow-y-auto space-y-3">
                {isFetching && schedules.length === 0 ? (
                    <div className="h-32 flex items-center justify-center">
                        <Loader2 size={18} className="animate-spin text-cortex-text-muted" />
                    </div>
                ) : schedules.length === 0 ? (
                    <div className="h-48 flex flex-col items-center justify-center text-center gap-3 rounded-md border border-dashed border-cortex-border">
                        <CalendarClock size={22} className="text-cortex-primary/50" />
                        <div>
                            <p className="text-sm font-medium text-cortex-text-main">No schedule rules yet</p>
                            <p className="text-xs text-cortex-text-muted max-w-sm mt-1">
                                Add propose-only cadence when a mission should be reviewed on a predictable rhythm.
                            </p>
                        </div>
                    </div>
                ) : (
                    schedules.map((rule) => {
                        const isToggling = pendingToggleRuleID === rule.id;
                        const toggleLabel = isToggling ? "Updating..." : rule.is_active ? "Pause" : "Resume";
                        const handoffState = getRuleHandoffState(rule);
                        const handoffExecution = getRuleHandoffExecution(rule);
                        const canResolveHandoff =
                            Boolean(handoffExecution?.id) && handoffState === "awaiting_approval";
                        return (
                            <article key={rule.id} className="rounded-md border border-cortex-border bg-cortex-surface p-4 space-y-3">
                                <div className="flex flex-wrap items-start justify-between gap-3">
                                    <div className="min-w-0 flex-1">
                                        <div className="flex min-w-0 flex-wrap items-center gap-2">
                                            <h3
                                                data-testid={`schedule-rule-name-${rule.id}`}
                                                className="min-w-0 max-w-full truncate text-sm font-semibold text-cortex-text-main"
                                                title={rule.name}
                                            >
                                                {rule.name}
                                            </h3>
                                            <span
                                                data-testid={`schedule-rule-badge-propose-${rule.id}`}
                                                className="flex-shrink-0 px-1.5 py-0.5 text-[10px] rounded bg-cortex-primary/10 text-cortex-primary font-mono"
                                            >
                                                propose only
                                            </span>
                                            {!rule.is_active && (
                                                <span
                                                    data-testid={`schedule-rule-badge-disabled-${rule.id}`}
                                                    className="flex-shrink-0 px-1.5 py-0.5 text-[10px] rounded bg-cortex-border text-cortex-text-muted font-mono"
                                                >
                                                    disabled
                                                </span>
                                            )}
                                            {handoffState && (
                                                <ScheduleHandoffBadge ruleID={rule.id} state={handoffState} />
                                            )}
                                        </div>
                                        {rule.description && <p className="mt-1 truncate text-xs text-cortex-text-muted" title={rule.description}>{rule.description}</p>}
                                    </div>
                                    <button
                                        type="button"
                                        onClick={async () => {
                                            setPendingToggleRuleID(rule.id);
                                            try {
                                                await toggleTriggerRule(rule.id, !rule.is_active);
                                            } finally {
                                                setPendingToggleRuleID(null);
                                            }
                                        }}
                                        disabled={isToggling}
                                        className="flex flex-shrink-0 items-center gap-1.5 px-2.5 py-1.5 text-xs rounded border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main disabled:cursor-wait disabled:opacity-50"
                                    >
                                        {isToggling && <Loader2 size={12} className="animate-spin" />}
                                        {toggleLabel}
                                    </button>
                                </div>

                                <div className="grid gap-3 md:grid-cols-3">
                                    <ScheduleFact icon={<Clock3 size={13} />} label="Cadence" value={formatInterval(rule.schedule_interval_seconds)} />
                                    <ScheduleFact icon={<History size={13} />} label="Next proposal" value={formatDate(rule.next_run_at)} />
                                    <ScheduleFact icon={<ShieldCheck size={13} />} label="Cooldown" value={`${rule.cooldown_seconds}s`} />
                                </div>
                                <div className="grid gap-3 md:grid-cols-2">
                                    <TextBlock label="Proof expected" value={rule.proof_expectations || "Operator-visible result, audit event, and retained proof."} />
                                    <TextBlock label="Recovery" value={rule.recovery_behavior || "Pause the schedule and review the last proposed run."} />
                                </div>
                                {canResolveHandoff && handoffExecution?.id && (
                                    <ScheduleHandoffActions
                                        disabled={pendingHandoffAction === handoffExecution.id}
                                        onResolve={async (status) => {
                                            setPendingHandoffAction(handoffExecution.id);
                                            try {
                                                await resolveScheduleHandoff(rule.id, handoffExecution.id, status);
                                            } finally {
                                                setPendingHandoffAction(null);
                                            }
                                        }}
                                    />
                                )}
                            </article>
                        );
                    })
                )}
            </div>
        </div>
    );
}

function ScheduleFact({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
    return (
        <div className="flex items-center gap-2 rounded border border-cortex-border/60 px-3 py-2">
            <span className="text-cortex-primary">{icon}</span>
            <div>
                <p className="text-[10px] uppercase tracking-wide text-cortex-text-muted">{label}</p>
                <p className="text-xs font-medium text-cortex-text-main">{value}</p>
            </div>
        </div>
    );
}

function TextBlock({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded border border-cortex-border/60 px-3 py-2">
            <p className="text-[10px] uppercase tracking-wide text-cortex-text-muted">{label}</p>
            <p className="text-xs leading-5 text-cortex-text-main mt-1">{value}</p>
        </div>
    );
}

function CreateScheduleForm({ onCreate, onCancel }: { onCreate: (rule: TriggerRuleCreate) => Promise<void>; onCancel: () => void }) {
    const [name, setName] = useState("");
    const [targetMissionId, setTargetMissionId] = useState("");
    const [intervalMinutes, setIntervalMinutes] = useState(60);
    const [proof, setProof] = useState("Operator-visible result, audit event, and retained proof.");
    const [recovery, setRecovery] = useState("Pause the schedule and review the last proposed run.");
    const [saving, setSaving] = useState(false);

    async function submit() {
        if (!name.trim() || !targetMissionId.trim()) return;
        setSaving(true);
        await onCreate({
            name: name.trim(),
            trigger_kind: "schedule",
            event_pattern: "scheduler.due",
            target_mission_id: targetMissionId.trim(),
            mode: "propose",
            cooldown_seconds: Math.max(60, intervalMinutes * 60),
            schedule_interval_seconds: intervalMinutes * 60,
            proof_expectations: proof.trim(),
            recovery_behavior: recovery.trim(),
            is_active: true,
        });
        setSaving(false);
    }

    return (
        <div className="rounded-md border border-cortex-primary/30 bg-cortex-surface p-4 space-y-3">
            <h3 className="text-xs font-semibold text-cortex-primary uppercase tracking-wide">New propose-only schedule</h3>
            <div className="grid gap-3 md:grid-cols-3">
                <Field label="Name" value={name} onChange={setName} placeholder="Weekly evidence review" />
                <Field label="Target mission ID" value={targetMissionId} onChange={setTargetMissionId} placeholder="mission-review" />
                <label className="text-xs text-cortex-text-muted">
                    Cadence minutes
                    <input type="number" min={1} value={intervalMinutes} onChange={(event) => setIntervalMinutes(Number(event.target.value))} className="mt-1 w-full px-3 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                </label>
            </div>
            <Field label="Proof expected" value={proof} onChange={setProof} placeholder="What must be visible after proposal" />
            <Field label="Recovery" value={recovery} onChange={setRecovery} placeholder="What the operator should do if this degrades" />
            <div className="flex justify-end gap-2">
                <button type="button" onClick={onCancel} className="px-3 py-1.5 text-xs rounded text-cortex-text-muted hover:bg-cortex-border">Cancel</button>
                <button type="button" onClick={submit} disabled={saving || !name.trim() || !targetMissionId.trim()} className="px-3 py-1.5 text-xs rounded bg-cortex-primary text-cortex-bg disabled:opacity-40">
                    {saving ? "Saving..." : "Create Schedule"}
                </button>
            </div>
        </div>
    );
}

function Field({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (value: string) => void; placeholder: string }) {
    return (
        <label className="text-xs text-cortex-text-muted">
            {label}
            <input value={value} onChange={(event) => onChange(event.target.value)} placeholder={placeholder} className="mt-1 w-full px-3 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main placeholder:text-cortex-text-muted/40" />
        </label>
    );
}
