"use client";

import React, { useEffect, useState } from "react";
import { X, Archive, RefreshCw, FolderOpen, Loader2, Clock } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

interface Props {
    profileId: string;
    profileName: string;
    onClose: () => void;
    onActivated?: () => void;
}

type Strategy = "transfer" | "fresh" | "snapshot";

export default function ContextSwitchModal({ profileId, profileName, onClose, onActivated }: Props) {
    const {
        contextSnapshots,
        fetchContextSnapshots,
        createContextSnapshot,
        activateMissionProfile,
    } = useCortexStore();

    const [strategy, setStrategy] = useState<Strategy>("transfer");
    const [selectedSnapshot, setSelectedSnapshot] = useState<string>("");
    const [snapshotName, setSnapshotName] = useState<string>(`snapshot-${Date.now()}`);
    const [working, setWorking] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetchContextSnapshots();
    }, [fetchContextSnapshots]);

    const handleConfirm = async () => {
        setWorking(true);
        setError(null);
        try {
            if (strategy === "transfer") {
                // Snapshot current context, then activate
                const snap = await createContextSnapshot(snapshotName);
                if (!snap) {
                    setError("Failed to save current context. Activate anyway?");
                }
            }
            // For "snapshot" strategy: the profile will be activated with context_strategy set upstream
            // For "fresh": just activate — backend starts with empty context
            await activateMissionProfile(profileId);
            onActivated?.();
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Activation failed");
        } finally {
            setWorking(false);
        }
    };

    const strategies: { id: Strategy; icon: React.ReactNode; label: string; desc: string }[] = [
        {
            id: "transfer",
            icon: <Archive className="w-4 h-4" />,
            label: "Cache & Transfer",
            desc: "Save the current conversation context as a snapshot, then activate the new profile.",
        },
        {
            id: "fresh",
            icon: <RefreshCw className="w-4 h-4" />,
            label: "Start Fresh",
            desc: "Activate the profile with no context carry-over. A clean slate.",
        },
        {
            id: "snapshot",
            icon: <FolderOpen className="w-4 h-4" />,
            label: "Load Snapshot",
            desc: "Restore a previously saved context snapshot into the new profile.",
        },
    ];

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl w-full max-w-lg shadow-2xl">
                {/* Header */}
                <div className="flex items-center justify-between px-5 py-3 border-b border-cortex-border">
                    <div>
                        <h3 className="text-sm font-semibold text-cortex-text-main">Switch Profile</h3>
                        <p className="text-[10px] text-cortex-text-muted mt-0.5">
                            Activating <span className="text-cortex-primary">{profileName}</span>
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>

                {/* Body */}
                <div className="p-5 space-y-4">
                    <p className="text-xs text-cortex-text-muted">
                        Choose how to handle the current conversation context before switching.
                    </p>

                    {/* Strategy radios */}
                    <div className="space-y-2">
                        {strategies.map((s) => (
                            <label
                                key={s.id}
                                className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                                    strategy === s.id
                                        ? "border-cortex-primary bg-cortex-primary/5"
                                        : "border-cortex-border hover:border-cortex-primary/40"
                                }`}
                            >
                                <input
                                    type="radio"
                                    name="strategy"
                                    value={s.id}
                                    checked={strategy === s.id}
                                    onChange={() => setStrategy(s.id)}
                                    className="mt-0.5 accent-cyan-500"
                                />
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <span className={strategy === s.id ? "text-cortex-primary" : "text-cortex-text-muted"}>
                                            {s.icon}
                                        </span>
                                        <span className="text-xs font-semibold text-cortex-text-main">{s.label}</span>
                                    </div>
                                    <p className="text-[10px] text-cortex-text-muted mt-0.5 leading-relaxed">{s.desc}</p>
                                </div>
                            </label>
                        ))}
                    </div>

                    {/* Cache & Transfer — snapshot name */}
                    {strategy === "transfer" && (
                        <div className="space-y-1 pl-2 border-l-2 border-cortex-primary/30">
                            <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Snapshot Name</label>
                            <input
                                value={snapshotName}
                                onChange={(e) => setSnapshotName(e.target.value)}
                                className="w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                                placeholder="my-context-snapshot"
                            />
                        </div>
                    )}

                    {/* Load Snapshot — select list */}
                    {strategy === "snapshot" && (
                        <div className="space-y-2 pl-2 border-l-2 border-cortex-primary/30">
                            <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Select Snapshot</label>
                            {contextSnapshots.length === 0 ? (
                                <p className="text-xs text-cortex-text-muted italic">No snapshots saved yet.</p>
                            ) : (
                                <div className="space-y-1 max-h-40 overflow-y-auto">
                                    {contextSnapshots.map((snap) => (
                                        <label
                                            key={snap.id}
                                            className={`flex items-center gap-2.5 p-2 rounded border cursor-pointer transition-colors text-xs ${
                                                selectedSnapshot === snap.id
                                                    ? "border-cortex-primary bg-cortex-primary/5 text-cortex-text-main"
                                                    : "border-cortex-border text-cortex-text-muted hover:border-cortex-primary/40"
                                            }`}
                                        >
                                            <input
                                                type="radio"
                                                name="snapshot"
                                                value={snap.id}
                                                checked={selectedSnapshot === snap.id}
                                                onChange={() => setSelectedSnapshot(snap.id)}
                                                className="accent-cyan-500"
                                            />
                                            <Clock className="w-3 h-3 flex-shrink-0" />
                                            <div className="flex-1 min-w-0">
                                                <div className="font-medium truncate">{snap.name}</div>
                                                <div className="text-[10px] opacity-60">
                                                    {new Date(snap.created_at).toLocaleString()}
                                                    {snap.source_profile && ` · from ${snap.source_profile}`}
                                                </div>
                                            </div>
                                        </label>
                                    ))}
                                </div>
                            )}
                        </div>
                    )}

                    {/* Error */}
                    {error && (
                        <p className="text-red-400 text-xs">{error}</p>
                    )}
                </div>

                {/* Footer */}
                <div className="flex items-center justify-end gap-2 px-5 py-3 border-t border-cortex-border">
                    <button
                        onClick={onClose}
                        className="px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleConfirm}
                        disabled={working || (strategy === "snapshot" && !selectedSnapshot)}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors disabled:opacity-50"
                    >
                        {working && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                        Activate Profile
                    </button>
                </div>
            </div>
        </div>
    );
}
