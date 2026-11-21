"use client";

import React, { useEffect, useState } from "react";
import { Layers, Plus, Loader2 } from "lucide-react";
import { useCortexStore, MissionProfile, MissionProfileCreate } from "@/store/useCortexStore";
import ContextSwitchModal from "./ContextSwitchModal";
import { type BrainEntry, blankProfile, ProfileEditor } from "./MissionProfileEditor";
import { ProfileCard } from "./MissionProfileCard";

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
                <div>
                    <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Mission Profiles</h3>
                    <p className="mt-1 text-xs text-cortex-text-muted">
                        Assign local or web-based agentry engines (OpenAI, Claude, Gemini, OpenRouter, Groq, DeepSeek, Ollama) to Soma and team roles.
                    </p>
                </div>
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
