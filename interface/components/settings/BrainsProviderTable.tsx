"use client";

import React from "react";
import {
    Brain, Globe, Server, CheckCircle, XCircle,
    Power, Pencil, Trash2, Wifi, WifiOff, Loader2,
} from "lucide-react";
import { BrainEntry, TOKEN_BUDGET_PRESETS } from "./brainsTypes";

interface BrainsProviderTableProps {
    brains: BrainEntry[];
    loading: boolean;
    rowProbing: string | null;
    rowProbeResult: Record<string, "alive" | "dead">;
    deleteTarget: string | null;
    deleting: boolean;
    onUpdatePolicy: (id: string, policy: string) => void;
    onToggleBrain: (brain: BrainEntry) => void;
    onOpenEdit: (brain: BrainEntry) => void;
    onProbeRow: (id: string) => void;
    onSelectDelete: (id: string | null) => void;
    onConfirmDelete: () => void;
}

function statusIcon(status: string) {
    if (status === "online") return <CheckCircle className="w-3.5 h-3.5 text-cortex-success" />;
    if (status === "disabled") return <Power className="w-3.5 h-3.5 text-cortex-text-muted" />;
    return <XCircle className="w-3.5 h-3.5 text-red-400" />;
}

function locationBadge(loc: string) {
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
}

export default function BrainsProviderTable({
    brains,
    loading,
    rowProbing,
    rowProbeResult,
    deleteTarget,
    deleting,
    onUpdatePolicy,
    onToggleBrain,
    onOpenEdit,
    onProbeRow,
    onSelectDelete,
    onConfirmDelete,
}: BrainsProviderTableProps) {
    return (
        <div className="rounded-lg border border-cortex-border overflow-hidden">
            <table className="w-full text-xs font-mono">
                <thead>
                    <tr className="bg-cortex-surface/50 text-cortex-text-muted">
                        <th className="text-left px-4 py-2">Provider</th>
                        <th className="text-left px-4 py-2">Location</th>
                        <th className="text-left px-4 py-2">Model</th>
                        <th className="text-left px-4 py-2">Status</th>
                        <th className="text-left px-4 py-2">Policy</th>
                        <th className="text-left px-4 py-2">Token Budget</th>
                        <th className="text-left px-4 py-2">Data Boundary</th>
                        <th className="text-center px-4 py-2">Enabled</th>
                        <th className="text-center px-4 py-2">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {brains.map((b) => (
                        <React.Fragment key={b.id}>
                            <tr className="border-t border-cortex-border hover:bg-cortex-surface/30 transition-colors">
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
                                        {rowProbeResult[b.id] === "alive"
                                            ? <CheckCircle className="w-3.5 h-3.5 text-cortex-success" />
                                            : rowProbeResult[b.id] === "dead"
                                            ? <WifiOff className="w-3.5 h-3.5 text-red-400" />
                                            : statusIcon(b.status)}
                                        <span className="text-cortex-text-muted">
                                            {rowProbeResult[b.id] ?? b.status}
                                        </span>
                                    </div>
                                </td>
                                <td className="px-4 py-2.5">
                                    <select
                                        value={b.usage_policy}
                                        onChange={(e) => onUpdatePolicy(b.id, e.target.value)}
                                        className="bg-cortex-bg border border-cortex-border rounded px-1.5 py-0.5 text-[10px] text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                                    >
                                        <option value="local_first">Local First</option>
                                        <option value="allow_escalation">Allow Escalation</option>
                                        <option value="require_approval">Require Approval</option>
                                        <option value="disallowed">Disallowed</option>
                                    </select>
                                </td>
                                <td className="px-4 py-2.5 text-cortex-text-main">
                                    <div className="flex flex-col">
                                        <span>{TOKEN_BUDGET_PRESETS[b.token_budget_profile || "standard"]?.label ?? "Standard"}</span>
                                        <span className="text-[10px] text-cortex-text-muted">{b.max_output_tokens || 1024} max</span>
                                    </div>
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
                                        onClick={() => onToggleBrain(b)}
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
                                <td className="px-4 py-2.5">
                                    <div className="flex items-center justify-center gap-1">
                                        <button
                                            onClick={() => onOpenEdit(b)}
                                            title="Edit"
                                            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors"
                                        >
                                            <Pencil className="w-3.5 h-3.5" />
                                        </button>
                                        <button
                                            onClick={() => onProbeRow(b.id)}
                                            title="Test connection"
                                            disabled={rowProbing === b.id}
                                            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors disabled:opacity-50"
                                        >
                                            {rowProbing === b.id
                                                ? <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                                : <Wifi className="w-3.5 h-3.5" />
                                            }
                                        </button>
                                        <button
                                            onClick={() => onSelectDelete(b.id)}
                                            title="Delete"
                                            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-red-400 transition-colors"
                                        >
                                            <Trash2 className="w-3.5 h-3.5" />
                                        </button>
                                    </div>
                                </td>
                            </tr>
                            {deleteTarget === b.id && (
                                <tr className="bg-red-400/5 border-t border-red-400/20">
                                    <td colSpan={9} className="px-4 py-2.5">
                                        <div className="flex items-center justify-between">
                                            <span className="text-xs text-red-400">
                                                Delete <strong>{b.id}</strong>? This cannot be undone.
                                            </span>
                                            <div className="flex items-center gap-2">
                                                <button
                                                    onClick={() => onSelectDelete(null)}
                                                    className="px-2.5 py-1 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                                                >
                                                    Cancel
                                                </button>
                                                <button
                                                    onClick={onConfirmDelete}
                                                    disabled={deleting}
                                                    className="px-2.5 py-1 rounded border border-red-400/40 bg-red-400/10 text-red-400 text-xs hover:bg-red-400/20 transition-colors disabled:opacity-50"
                                                >
                                                    {deleting ? "Deleting…" : "Delete"}
                                                </button>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            )}
                        </React.Fragment>
                    ))}
                </tbody>
            </table>
            {brains.length === 0 && !loading && (
                <div className="px-4 py-8 text-center text-cortex-text-muted text-xs">
                    No providers configured. Add one to get started.
                </div>
            )}
        </div>
    );
}
