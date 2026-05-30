"use client";

import { useEffect, useMemo, useState } from "react";
import { CalendarClock, Clock3, History, Loader2, Plus, ShieldCheck } from "lucide-react";
import { useCortexStore, type TriggerRuleCreate } from "@/store/useCortexStore";

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

export default function ScheduleRulesTab() {
    const rules = useCortexStore((s) => s.triggerRules);
    const isFetching = useCortexStore((s) => s.isFetchingTriggers);
    const fetchTriggerRules = useCortexStore((s) => s.fetchTriggerRules);
    const createTriggerRule = useCortexStore((s) => s.createTriggerRule);
    const toggleTriggerRule = useCortexStore((s) => s.toggleTriggerRule);
    const [showCreate, setShowCreate] = useState(false);

    useEffect(() => {
        fetchTriggerRules();
    }, [fetchTriggerRules]);

    const schedules = useMemo(
        () => rules.filter((rule) => rule.trigger_kind === "schedule"),
        [rules],
    );

    return (
        <div className="h-full flex flex-col p-6 gap-4">
            <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <CalendarClock size={18} className="text-cortex-primary" />
                    <div>
                        <h2 className="text-sm font-semibold text-cortex-text-main">Schedule Rules</h2>
                        <p className="text-xs text-cortex-text-muted">
                            Cadence proposals with cooldown, proof expectations, and recovery text.
                        </p>
                    </div>
                </div>
                <button
                    type="button"
                    onClick={() => setShowCreate((value) => !value)}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md bg-cortex-primary/10 text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
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
                    schedules.map((rule) => (
                        <article key={rule.id} className="rounded-md border border-cortex-border bg-cortex-surface p-4 space-y-3">
                            <div className="flex items-start justify-between gap-3">
                                <div>
                                    <div className="flex items-center gap-2">
                                        <h3 className="text-sm font-semibold text-cortex-text-main">{rule.name}</h3>
                                        <span className="px-1.5 py-0.5 text-[10px] rounded bg-cortex-primary/10 text-cortex-primary font-mono">propose only</span>
                                        {!rule.is_active && <span className="px-1.5 py-0.5 text-[10px] rounded bg-cortex-border text-cortex-text-muted font-mono">disabled</span>}
                                    </div>
                                    {rule.description && <p className="text-xs text-cortex-text-muted mt-1">{rule.description}</p>}
                                </div>
                                <button
                                    type="button"
                                    onClick={() => toggleTriggerRule(rule.id, !rule.is_active)}
                                    className="px-2.5 py-1.5 text-xs rounded border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main"
                                >
                                    {rule.is_active ? "Pause" : "Resume"}
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
                        </article>
                    ))
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
