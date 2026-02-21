"use client";

import React, { useEffect, useState, useCallback } from "react";
import { Brain, Globe, Server, RefreshCw, AlertTriangle, CheckCircle, XCircle, Power } from "lucide-react";
import RemoteEnableModal from "./RemoteEnableModal";

interface BrainEntry {
    id: string;
    type: string;
    endpoint?: string;
    model_id: string;
    location: string;
    data_boundary: string;
    usage_policy: string;
    roles_allowed: string[];
    enabled: boolean;
    status: string;
}

export default function BrainsPage() {
    const [brains, setBrains] = useState<BrainEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [confirmRemote, setConfirmRemote] = useState<BrainEntry | null>(null);

    const fetchBrains = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetch("/api/v1/brains");
            const body = await res.json();
            if (body.ok) setBrains(body.data || []);
        } catch { /* ignore */ }
        setLoading(false);
    }, []);

    useEffect(() => { fetchBrains(); }, [fetchBrains]);

    const toggleBrain = async (brain: BrainEntry) => {
        // If enabling a remote provider, require confirmation
        if (!brain.enabled && brain.location === "remote") {
            setConfirmRemote(brain);
            return;
        }
        await doToggle(brain.id, !brain.enabled);
    };

    const doToggle = async (id: string, enabled: boolean) => {
        try {
            await fetch(`/api/v1/brains/${id}/toggle`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ enabled }),
            });
            fetchBrains();
        } catch { /* ignore */ }
    };

    const updatePolicy = async (id: string, policy: string) => {
        try {
            await fetch(`/api/v1/brains/${id}/policy`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ usage_policy: policy }),
            });
            fetchBrains();
        } catch { /* ignore */ }
    };

    const statusIcon = (status: string) => {
        if (status === "online") return <CheckCircle className="w-3.5 h-3.5 text-cortex-success" />;
        if (status === "disabled") return <Power className="w-3.5 h-3.5 text-cortex-text-muted" />;
        return <XCircle className="w-3.5 h-3.5 text-red-400" />;
    };

    const locationBadge = (loc: string) => {
        if (loc === "remote") return (
            <span className="flex items-center gap-1 text-amber-400 text-[10px]">
                <Globe className="w-3 h-3" /> Remote
            </span>
        );
        return (
            <span className="flex items-center gap-1 text-cortex-success text-[10px]">
                <Server className="w-3 h-3" /> Local
            </span>
        );
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Provider Management</h3>
                <button
                    onClick={fetchBrains}
                    className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    title="Refresh"
                >
                    <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                </button>
            </div>

            <div className="rounded-lg border border-cortex-border overflow-hidden">
                <table className="w-full text-xs font-mono">
                    <thead>
                        <tr className="bg-cortex-surface/50 text-cortex-text-muted">
                            <th className="text-left px-4 py-2">Provider</th>
                            <th className="text-left px-4 py-2">Location</th>
                            <th className="text-left px-4 py-2">Model</th>
                            <th className="text-left px-4 py-2">Status</th>
                            <th className="text-left px-4 py-2">Policy</th>
                            <th className="text-left px-4 py-2">Data Boundary</th>
                            <th className="text-center px-4 py-2">Enabled</th>
                        </tr>
                    </thead>
                    <tbody>
                        {brains.map((b) => (
                            <tr key={b.id} className="border-t border-cortex-border hover:bg-cortex-surface/30 transition-colors">
                                <td className="px-4 py-2.5">
                                    <div className="flex items-center gap-2">
                                        <Brain className="w-3.5 h-3.5 text-cortex-primary" />
                                        <span className="text-cortex-text-main font-semibold">{b.id}</span>
                                    </div>
                                </td>
                                <td className="px-4 py-2.5">{locationBadge(b.location)}</td>
                                <td className="px-4 py-2.5 text-cortex-text-main">{b.model_id || "\u2014"}</td>
                                <td className="px-4 py-2.5">
                                    <div className="flex items-center gap-1.5">
                                        {statusIcon(b.status)}
                                        <span className="text-cortex-text-muted">{b.status}</span>
                                    </div>
                                </td>
                                <td className="px-4 py-2.5">
                                    <select
                                        value={b.usage_policy}
                                        onChange={(e) => updatePolicy(b.id, e.target.value)}
                                        className="bg-cortex-bg border border-cortex-border rounded px-1.5 py-0.5 text-[10px] text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                                    >
                                        <option value="local_first">Local First</option>
                                        <option value="allow_escalation">Allow Escalation</option>
                                        <option value="require_approval">Require Approval</option>
                                        <option value="disallowed">Disallowed</option>
                                    </select>
                                </td>
                                <td className="px-4 py-2.5">
                                    <span className={`text-[10px] px-2 py-0.5 rounded border ${
                                        b.data_boundary === "leaves_org"
                                            ? "text-amber-400 border-amber-400/30 bg-amber-400/5"
                                            : "text-cortex-success border-cortex-success/30 bg-cortex-success/5"
                                    }`}>
                                        {b.data_boundary === "leaves_org" ? "LEAVES ORG" : "LOCAL ONLY"}
                                    </span>
                                </td>
                                <td className="px-4 py-2.5 text-center">
                                    <button
                                        onClick={() => toggleBrain(b)}
                                        className={`w-8 h-4 rounded-full relative transition-colors border ${
                                            b.enabled
                                                ? "bg-cortex-success/20 border-cortex-success/40"
                                                : "bg-cortex-bg border-cortex-border"
                                        }`}
                                    >
                                        <div className={`absolute top-0 w-4 h-4 rounded-full shadow-sm transition-all ${
                                            b.enabled
                                                ? "right-0 bg-cortex-success"
                                                : "left-0 bg-cortex-text-muted"
                                        }`} />
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
                {brains.length === 0 && !loading && (
                    <div className="px-4 py-8 text-center text-cortex-text-muted text-xs">
                        No providers configured.
                    </div>
                )}
            </div>

            {/* Remote warning */}
            <div className="flex items-start gap-2 p-3 rounded border border-amber-400/20 bg-amber-400/5 text-xs text-amber-400">
                <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                <div>
                    <strong>Data Boundary Notice:</strong> Enabling remote providers means data may leave your local environment.
                    Review each provider&apos;s data boundary before enabling.
                </div>
            </div>

            {/* Remote enable confirmation modal */}
            {confirmRemote && (
                <RemoteEnableModal
                    provider={confirmRemote}
                    onConfirm={() => {
                        doToggle(confirmRemote.id, true);
                        setConfirmRemote(null);
                    }}
                    onCancel={() => setConfirmRemote(null)}
                />
            )}
        </div>
    );
}
