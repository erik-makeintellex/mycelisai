"use client";

import React, { useEffect, useState } from "react";
import {
    Layers, Plus, Pencil, Trash2, Loader2, X, ChevronRight,
    Radio, CheckCircle, Server, Globe, Zap,
} from "lucide-react";
import { useCortexStore, MissionProfile, MissionProfileCreate } from "@/store/useCortexStore";
import ContextSwitchModal from "./ContextSwitchModal";

// ── Types ─────────────────────────────────────────────────────────────────────

interface BrainEntry {
    id: string;
    type: string;
    model_id: string;
    location: string;
    enabled: boolean;
    status: string;
}

const COUNCIL_ROLES = ["architect", "coder", "creative", "sentry", "admin"];

const blankProfile = (): MissionProfileCreate => ({
    name: "",
    description: "",
    role_providers: {},
    subscriptions: [],
    context_strategy: "fresh",
    auto_start: false,
});

// ── Role → Provider assignment table ─────────────────────────────────────────

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

// ── Subscriptions editor ─────────────────────────────────────────────────────

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

// ── Profile editor pane ───────────────────────────────────────────────────────

function ProfileEditor({
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
            {/* Name + description */}
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

            {/* Role assignments */}
            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Role → Provider Assignments</label>
                <RoleAssignmentTable
                    roleProviders={form.role_providers}
                    brains={brains}
                    onChange={(rp) => setForm({ ...form, role_providers: rp })}
                />
            </div>

            {/* Context strategy */}
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

            {/* NATS subscriptions */}
            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">NATS Reactive Subscriptions</label>
                <SubscriptionsEditor
                    subs={form.subscriptions}
                    onChange={(s) => setForm({ ...form, subscriptions: s })}
                />
            </div>

            {/* Auto-start toggle */}
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

            {/* Actions */}
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

// ── Profile list card ─────────────────────────────────────────────────────────

function ProfileCard({
    profile,
    isActive,
    onEdit,
    onActivate,
    onDelete,
    deleting,
    deleteConfirm,
    onConfirmDelete,
    onCancelDelete,
}: {
    profile: MissionProfile;
    isActive: boolean;
    onEdit: () => void;
    onActivate: () => void;
    onDelete: () => void;
    deleting: boolean;
    deleteConfirm: boolean;
    onConfirmDelete: () => void;
    onCancelDelete: () => void;
}) {
    const roleChips = Object.entries(profile.role_providers);
    const subCount = profile.subscriptions.length;

    return (
        <div className={`rounded-lg border transition-colors ${
            isActive ? "border-cortex-primary/50 bg-cortex-primary/5" : "border-cortex-border"
        }`}>
            <div className="px-4 py-3 space-y-2">
                {/* Header */}
                <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                            {isActive && <CheckCircle className="w-3.5 h-3.5 text-cortex-primary flex-shrink-0" />}
                            <span className="text-sm font-semibold text-cortex-text-main truncate">{profile.name}</span>
                            {isActive && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded border border-cortex-primary/40 bg-cortex-primary/10 text-cortex-primary">
                                    ACTIVE
                                </span>
                            )}
                            {profile.auto_start && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded border border-cortex-success/30 bg-cortex-success/10 text-cortex-success">
                                    AUTO
                                </span>
                            )}
                        </div>
                        {profile.description && (
                            <p className="text-xs text-cortex-text-muted mt-0.5 truncate">{profile.description}</p>
                        )}
                    </div>
                    {/* Actions */}
                    <div className="flex items-center gap-1 flex-shrink-0">
                        {!isActive && (
                            <button
                                onClick={onActivate}
                                className="flex items-center gap-1 px-2 py-1 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] hover:bg-cortex-primary/10 transition-colors"
                            >
                                <Zap className="w-3 h-3" /> Activate
                            </button>
                        )}
                        <button
                            onClick={onEdit}
                            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors"
                            title="Edit"
                        >
                            <Pencil className="w-3.5 h-3.5" />
                        </button>
                        <button
                            onClick={onDelete}
                            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-red-400 transition-colors"
                            title="Delete"
                        >
                            <Trash2 className="w-3.5 h-3.5" />
                        </button>
                    </div>
                </div>

                {/* Role → provider chips */}
                {roleChips.length > 0 && (
                    <div className="flex flex-wrap gap-1.5">
                        {roleChips.map(([role, prov]) => (
                            <span key={role} className="text-[10px] px-1.5 py-0.5 rounded bg-cortex-surface border border-cortex-border text-cortex-text-muted font-mono">
                                {role} <ChevronRight className="w-2.5 h-2.5 inline" /> {prov}
                            </span>
                        ))}
                    </div>
                )}

                {/* Subscription count */}
                {subCount > 0 && (
                    <div className="flex items-center gap-1 text-[10px] text-cortex-text-muted">
                        <Radio className="w-3 h-3" />
                        {subCount} NATS watch{subCount !== 1 ? "es" : ""}
                    </div>
                )}

                {/* Delete confirmation */}
                {deleteConfirm && (
                    <div className="flex items-center justify-between pt-2 border-t border-red-400/20 mt-2">
                        <span className="text-xs text-red-400">Delete <strong>{profile.name}</strong>?</span>
                        <div className="flex items-center gap-2">
                            <button
                                onClick={onCancelDelete}
                                className="px-2 py-0.5 rounded border border-cortex-border text-[10px] text-cortex-text-muted hover:text-cortex-text-main"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={onConfirmDelete}
                                disabled={deleting}
                                className="px-2 py-0.5 rounded border border-red-400/40 bg-red-400/10 text-red-400 text-[10px] hover:bg-red-400/20 disabled:opacity-50"
                            >
                                {deleting ? "…" : "Delete"}
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function MissionProfilesPage() {
    const {
        missionProfiles,
        activeProfileId,
        fetchMissionProfiles,
        createMissionProfile,
        updateMissionProfile,
        deleteMissionProfile,
    } = useCortexStore();

    const [brains, setBrains] = useState<BrainEntry[]>([]);
    const [loading, setLoading] = useState(true);

    // Editor state
    type EditorMode = { mode: "new" } | { mode: "edit"; profile: MissionProfile };
    const [editor, setEditor] = useState<EditorMode | null>(null);
    const [saving, setSaving] = useState(false);
    const [saveError, setSaveError] = useState<string | null>(null);

    // Activate modal
    const [activateTarget, setActivateTarget] = useState<MissionProfile | null>(null);

    // Delete state
    const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
    const [deleting, setDeleting] = useState(false);

    useEffect(() => {
        fetchMissionProfiles();
        // Fetch available brains for role assignment dropdowns
        fetch("/api/v1/brains")
            .then((r) => r.json())
            .then((body) => { if (body.ok) setBrains(body.data || []); })
            .catch(() => {})
            .finally(() => setLoading(false));
    }, [fetchMissionProfiles]);

    // ── Save (create or update) ─────────────────────────────────────────────

    const handleSave = async (form: MissionProfileCreate) => {
        setSaving(true);
        setSaveError(null);
        try {
            if (editor?.mode === "edit") {
                await updateMissionProfile(editor.profile.id, form);
            } else {
                const created = await createMissionProfile(form);
                if (!created) {
                    setSaveError("Failed to create profile");
                    return;
                }
            }
            setEditor(null);
        } catch (err) {
            setSaveError(err instanceof Error ? err.message : "Save failed");
        } finally {
            setSaving(false);
        }
    };

    // ── Delete ──────────────────────────────────────────────────────────────

    const handleDelete = async (id: string) => {
        setDeleting(true);
        try {
            await deleteMissionProfile(id);
            setDeleteConfirm(null);
        } finally {
            setDeleting(false);
        }
    };

    // ── Derive editor initial form ──────────────────────────────────────────

    const editorInitial = (): MissionProfileCreate => {
        if (editor?.mode === "edit") {
            const p = editor.profile;
            return {
                name: p.name,
                description: p.description ?? "",
                role_providers: { ...p.role_providers },
                subscriptions: p.subscriptions.map((s) => ({ ...s })),
                context_strategy: p.context_strategy,
                auto_start: p.auto_start,
            };
        }
        return blankProfile();
    };

    return (
        <div className="space-y-4">
            {/* Header */}
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Mission Profiles</h3>
                <button
                    onClick={() => { setEditor({ mode: "new" }); setSaveError(null); }}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors"
                >
                    <Plus className="w-3.5 h-3.5" />
                    New Profile
                </button>
            </div>

            {/* Editor panel */}
            {editor && (
                <div className="rounded-lg border border-cortex-primary/30 bg-cortex-surface/50 p-4 space-y-3">
                    <div className="flex items-center gap-2">
                        <Layers className="w-4 h-4 text-cortex-primary" />
                        <h4 className="text-sm font-semibold text-cortex-text-main">
                            {editor.mode === "new" ? "New Profile" : `Edit — ${editor.profile.name}`}
                        </h4>
                    </div>
                    <ProfileEditor
                        profile={editorInitial()}
                        brains={brains}
                        onSave={handleSave}
                        onCancel={() => setEditor(null)}
                        saving={saving}
                        error={saveError}
                    />
                </div>
            )}

            {/* Profile list */}
            {loading ? (
                <div className="flex items-center justify-center py-10 text-cortex-text-muted text-xs">
                    <Loader2 className="w-4 h-4 animate-spin mr-2" /> Loading…
                </div>
            ) : missionProfiles.length === 0 ? (
                <div className="px-4 py-10 text-center text-cortex-text-muted text-xs border border-dashed border-cortex-border rounded-lg">
                    No profiles configured. Create one to assign providers to workflow roles.
                </div>
            ) : (
                <div className="space-y-2">
                    {missionProfiles.map((p) => (
                        <ProfileCard
                            key={p.id}
                            profile={p}
                            isActive={p.id === activeProfileId}
                            onEdit={() => { setEditor({ mode: "edit", profile: p }); setSaveError(null); }}
                            onActivate={() => setActivateTarget(p)}
                            onDelete={() => setDeleteConfirm(p.id)}
                            deleting={deleting && deleteConfirm === p.id}
                            deleteConfirm={deleteConfirm === p.id}
                            onConfirmDelete={() => handleDelete(p.id)}
                            onCancelDelete={() => setDeleteConfirm(null)}
                        />
                    ))}
                </div>
            )}

            {/* Context switch modal */}
            {activateTarget && (
                <ContextSwitchModal
                    profileId={activateTarget.id}
                    profileName={activateTarget.name}
                    onClose={() => setActivateTarget(null)}
                    onActivated={() => fetchMissionProfiles()}
                />
            )}
        </div>
    );
}
