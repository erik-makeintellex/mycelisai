"use client";

import React, { useState } from "react";
import { Plus, Loader2, X } from "lucide-react";
import { type MissionProfileCreate } from "@/store/useCortexStore";

export interface BrainEntry {
    id: string;
    type: string;
    model_id: string;
    location: string;
    enabled: boolean;
    status: string;
}

export const COUNCIL_ROLES = ["architect", "coder", "creative", "sentry", "admin"];

export const blankProfile = (): MissionProfileCreate => ({
    name: "",
    description: "",
    role_providers: {},
    subscriptions: [],
    context_strategy: "fresh",
    auto_start: false,
});

function RoleAssignmentTable({
    roleProviders,
    brains,
    onChange,
}: {
    roleProviders: Record<string, string>;
    brains: BrainEntry[];
    onChange: (rp: Record<string, string>) => void;
}) {
    const enabledBrains = brains.filter((b) => b.enabled);

    return (
        <div className="rounded border border-cortex-border overflow-hidden">
            <table className="w-full text-xs">
                <thead>
                    <tr className="bg-cortex-surface/50">
                        <th className="text-left px-3 py-1.5 text-cortex-text-muted font-normal uppercase tracking-wider text-[10px]">Role</th>
                        <th className="text-left px-3 py-1.5 text-cortex-text-muted font-normal uppercase tracking-wider text-[10px]">Provider</th>
                    </tr>
                </thead>
                <tbody>
                    {COUNCIL_ROLES.map((role) => (
                        <tr key={role} className="border-t border-cortex-border">
                            <td className="px-3 py-2 text-cortex-text-muted font-mono">{role}</td>
                            <td className="px-3 py-2">
                                <select
                                    value={roleProviders[role] ?? ""}
                                    onChange={(e) => {
                                        const val = e.target.value;
                                        const next = { ...roleProviders };
                                        if (val) {
                                            next[role] = val;
                                        } else {
                                            delete next[role];
                                        }
                                        onChange(next);
                                    }}
                                    className="w-full bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-[11px] text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                                >
                                    <option value="">— default —</option>
                                    {enabledBrains.map((b) => (
                                        <option key={b.id} value={b.id}>
                                            {b.id} ({b.model_id || b.type}){b.location === "remote" ? " ⚠ remote" : ""}
                                        </option>
                                    ))}
                                </select>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}

function SubscriptionsEditor({
    subs,
    onChange,
}: {
    subs: { topic: string; condition?: string }[];
    onChange: (s: { topic: string; condition?: string }[]) => void;
}) {
    const add = () => onChange([...subs, { topic: "", condition: "" }]);
    const remove = (i: number) => onChange(subs.filter((_, idx) => idx !== i));
    const update = (i: number, field: "topic" | "condition", val: string) => {
        const next = subs.map((s, idx) => idx === i ? { ...s, [field]: val } : s);
        onChange(next);
    };

    return (
        <div className="space-y-2">
            {subs.map((s, i) => (
                <div key={i} className="flex items-center gap-2">
                    <input
                        value={s.topic}
                        onChange={(e) => update(i, "topic", e.target.value)}
                        placeholder="swarm.events.>"
                        className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary font-mono placeholder:text-cortex-text-muted/50"
                    />
                    <input
                        value={s.condition ?? ""}
                        onChange={(e) => update(i, "condition", e.target.value)}
                        placeholder="condition (optional)"
                        className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs text-cortex-text-muted focus:outline-none focus:ring-1 focus:ring-cortex-primary placeholder:text-cortex-text-muted/50"
                    />
                    <button
                        onClick={() => remove(i)}
                        className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-red-400 transition-colors"
                    >
                        <X className="w-3.5 h-3.5" />
                    </button>
                </div>
            ))}
            <button
                onClick={add}
                className="flex items-center gap-1.5 text-[10px] text-cortex-text-muted hover:text-cortex-primary transition-colors"
            >
                <Plus className="w-3 h-3" /> Add NATS topic
            </button>
        </div>
    );
}

export function ProfileEditor({
    profile,
    brains,
    onSave,
    onCancel,
    saving,
    error,
}: {
    profile: MissionProfileCreate;
    brains: BrainEntry[];
    onSave: (p: MissionProfileCreate) => void;
    onCancel: () => void;
    saving: boolean;
    error: string | null;
}) {
    const [form, setForm] = useState<MissionProfileCreate>(profile);

    const inputCls = "w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary";

    return (
        <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                    <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Profile Name</label>
                    <input
                        value={form.name}
                        onChange={(e) => setForm({ ...form, name: e.target.value })}
                        placeholder="Research Mode"
                        className={inputCls}
                    />
                </div>
                <div className="space-y-1">
                    <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Description</label>
                    <input
                        value={form.description ?? ""}
                        onChange={(e) => setForm({ ...form, description: e.target.value })}
                        placeholder="Optional description"
                        className={inputCls}
                    />
                </div>
            </div>

            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Role → Provider Assignments</label>
                <RoleAssignmentTable
                    roleProviders={form.role_providers}
                    brains={brains}
                    onChange={(rp) => setForm({ ...form, role_providers: rp })}
                />
            </div>

            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Context Strategy</label>
                <div className="flex gap-3">
                    {(["fresh", "warm"] as const).map((s) => (
                        <label key={s} className="flex items-center gap-1.5 cursor-pointer text-xs text-cortex-text-muted">
                            <input
                                type="radio"
                                name="ctx_strategy"
                                value={s}
                                checked={form.context_strategy === s}
                                onChange={() => setForm({ ...form, context_strategy: s })}
                                className="accent-cyan-500"
                            />
                            {s === "fresh" ? "Start Fresh" : "Warm (carry context)"}
                        </label>
                    ))}
                </div>
            </div>

            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">NATS Reactive Subscriptions</label>
                <SubscriptionsEditor
                    subs={form.subscriptions}
                    onChange={(s) => setForm({ ...form, subscriptions: s })}
                />
            </div>

            <label className="flex items-center gap-2 cursor-pointer text-xs text-cortex-text-muted select-none">
                <div
                    onClick={() => setForm({ ...form, auto_start: !form.auto_start })}
                    className={`w-8 h-4 rounded-full relative transition-colors border cursor-pointer ${
                        form.auto_start
                            ? "bg-cortex-success/20 border-cortex-success/40"
                            : "bg-cortex-bg border-cortex-border"
                    }`}
                >
                    <div className={`absolute top-0 w-4 h-4 rounded-full shadow-sm transition-all ${
                        form.auto_start ? "right-0 bg-cortex-success" : "left-0 bg-cortex-text-muted"
                    }`} />
                </div>
                Auto-start (activate on system boot)
            </label>

            {error && <p className="text-red-400 text-xs">{error}</p>}

            <div className="flex justify-end gap-2 pt-2 border-t border-cortex-border">
                <button
                    onClick={onCancel}
                    className="px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    Cancel
                </button>
                <button
                    onClick={() => onSave(form)}
                    disabled={saving || !form.name}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors disabled:opacity-50"
                >
                    {saving && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                    Save Profile
                </button>
            </div>
        </div>
    );
}
