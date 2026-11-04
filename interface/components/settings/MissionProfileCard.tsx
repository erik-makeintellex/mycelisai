"use client";

import React from "react";
import { Pencil, Trash2, ChevronRight, Radio, CheckCircle, Zap } from "lucide-react";
import { type MissionProfile } from "@/store/useCortexStore";

export function ProfileCard({
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

                {roleChips.length > 0 && (
                    <div className="flex flex-wrap gap-1.5">
                        {roleChips.map(([role, prov]) => (
                            <span key={role} className="text-[10px] px-1.5 py-0.5 rounded bg-cortex-surface border border-cortex-border text-cortex-text-muted font-mono">
                                {role} <ChevronRight className="w-2.5 h-2.5 inline" /> {prov}
                            </span>
                        ))}
                    </div>
                )}

                {subCount > 0 && (
                    <div className="flex items-center gap-1 text-[10px] text-cortex-text-muted">
                        <Radio className="w-3 h-3" />
                        {subCount} NATS watch{subCount !== 1 ? "es" : ""}
                    </div>
                )}

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
