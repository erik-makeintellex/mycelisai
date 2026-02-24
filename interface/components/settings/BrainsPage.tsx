"use client";

import React, { useEffect, useState, useCallback } from "react";
import {
    Brain, Globe, Server, RefreshCw, AlertTriangle, CheckCircle,
    XCircle, Power, Plus, Pencil, Trash2, Wifi, WifiOff, Loader2, X,
} from "lucide-react";
import RemoteEnableModal from "./RemoteEnableModal";

// ── Types ─────────────────────────────────────────────────────────────────────

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

interface ProviderFormData {
    id: string;
    type: string;
    endpoint: string;
    model_id: string;
    api_key: string;
    location: string;
    data_boundary: string;
    usage_policy: string;
    roles_allowed: string[];
    enabled: boolean;
}

// ── Type presets ──────────────────────────────────────────────────────────────

const PROVIDER_PRESETS: Record<string, Partial<ProviderFormData>> = {
    ollama: {
        type: "openai_compatible",
        endpoint: "http://localhost:11434/v1",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    vllm: {
        type: "openai_compatible",
        endpoint: "http://localhost:8000/v1",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    lmstudio: {
        type: "openai_compatible",
        endpoint: "http://localhost:1234/v1",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
    openai: {
        type: "openai",
        endpoint: "https://api.openai.com/v1",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        roles_allowed: ["all"],
    },
    anthropic: {
        type: "anthropic",
        endpoint: "",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        roles_allowed: ["all"],
    },
    google: {
        type: "google",
        endpoint: "",
        location: "remote",
        data_boundary: "leaves_org",
        usage_policy: "require_approval",
        roles_allowed: ["all"],
    },
    custom: {
        type: "openai_compatible",
        endpoint: "",
        location: "local",
        data_boundary: "local_only",
        usage_policy: "local_first",
        roles_allowed: ["all"],
    },
};

const PRESET_LABELS: Record<string, string> = {
    ollama: "Ollama",
    vllm: "vLLM",
    lmstudio: "LM Studio",
    openai: "OpenAI",
    anthropic: "Anthropic",
    google: "Google",
    custom: "Custom",
};

const COUNCIL_ROLES = ["all", "architect", "coder", "creative", "sentry", "admin"];

const blankForm = (): ProviderFormData => ({
    id: "",
    type: "openai_compatible",
    endpoint: "",
    model_id: "",
    api_key: "",
    location: "local",
    data_boundary: "local_only",
    usage_policy: "local_first",
    roles_allowed: ["all"],
    enabled: true,
});

// ── Provider form (shared between add and edit) ───────────────────────────────

function ProviderForm({
    form,
    onChange,
    isEdit,
    probeResult,
    probing,
    onProbe,
    showPresets,
}: {
    form: ProviderFormData;
    onChange: (f: ProviderFormData) => void;
    isEdit: boolean;
    probeResult: "alive" | "dead" | null;
    probing: boolean;
    onProbe?: () => void;
    showPresets: boolean;
}) {
    const field = (key: keyof ProviderFormData, label: string, node: React.ReactNode) => (
        <div className="space-y-1">
            <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">{label}</label>
            {node}
        </div>
    );

    const inputCls = "w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary placeholder:text-cortex-text-muted/50";

    const toggleRole = (role: string) => {
        const cur = form.roles_allowed;
        if (role === "all") {
            onChange({ ...form, roles_allowed: ["all"] });
            return;
        }
        const without = cur.filter((r) => r !== "all" && r !== role);
        const next = cur.includes(role) ? without : [...without, role];
        onChange({ ...form, roles_allowed: next.length ? next : ["all"] });
    };

    return (
        <div className="space-y-3">
            {/* Preset selector — add mode only */}
            {showPresets && (
                <div className="space-y-1">
                    <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Quick Preset</label>
                    <div className="flex flex-wrap gap-1.5">
                        {Object.entries(PRESET_LABELS).map(([key, label]) => (
                            <button
                                key={key}
                                type="button"
                                onClick={() => {
                                    const preset = PROVIDER_PRESETS[key];
                                    const suggestedId = key === "custom" ? "" : key;
                                    onChange({ ...blankForm(), ...preset, id: form.id || suggestedId });
                                }}
                                className="px-2 py-0.5 rounded text-[10px] border border-cortex-border text-cortex-text-muted hover:border-cortex-primary hover:text-cortex-primary transition-colors"
                            >
                                {label}
                            </button>
                        ))}
                    </div>
                </div>
            )}

            <div className="grid grid-cols-2 gap-3">
                {/* ID */}
                {field("id", "Provider ID", (
                    <input
                        value={form.id}
                        onChange={(e) => onChange({ ...form, id: e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, "") })}
                        placeholder="e.g. my-ollama"
                        disabled={isEdit}
                        className={`${inputCls} ${isEdit ? "opacity-50 cursor-not-allowed" : ""}`}
                    />
                ))}
                {/* Type */}
                {field("type", "Provider Type", (
                    <input
                        value={form.type}
                        onChange={(e) => onChange({ ...form, type: e.target.value })}
                        placeholder="openai_compatible"
                        className={inputCls}
                    />
                ))}
                {/* Endpoint */}
                {field("endpoint", "Endpoint URL", (
                    <input
                        value={form.endpoint}
                        onChange={(e) => onChange({ ...form, endpoint: e.target.value })}
                        placeholder="http://localhost:11434/v1"
                        className={inputCls}
                    />
                ))}
                {/* Model ID */}
                {field("model_id", "Model ID", (
                    <input
                        value={form.model_id}
                        onChange={(e) => onChange({ ...form, model_id: e.target.value })}
                        placeholder="llama3:8b"
                        className={inputCls}
                    />
                ))}
                {/* API Key */}
                {field("api_key", isEdit ? "API Key (blank = keep existing)" : "API Key", (
                    <input
                        type="password"
                        value={form.api_key}
                        onChange={(e) => onChange({ ...form, api_key: e.target.value })}
                        placeholder={isEdit ? "leave blank to keep existing" : "optional"}
                        className={inputCls}
                    />
                ))}
                {/* Usage Policy */}
                {field("usage_policy", "Usage Policy", (
                    <select
                        value={form.usage_policy}
                        onChange={(e) => onChange({ ...form, usage_policy: e.target.value })}
                        className={inputCls}
                    >
                        <option value="local_first">Local First</option>
                        <option value="allow_escalation">Allow Escalation</option>
                        <option value="require_approval">Require Approval</option>
                        <option value="disallowed">Disallowed</option>
                    </select>
                ))}
                {/* Location */}
                {field("location", "Location", (
                    <select
                        value={form.location}
                        onChange={(e) => {
                            const loc = e.target.value;
                            onChange({
                                ...form,
                                location: loc,
                                data_boundary: loc === "remote" ? "leaves_org" : "local_only",
                            });
                        }}
                        className={inputCls}
                    >
                        <option value="local">Local</option>
                        <option value="remote">Remote</option>
                    </select>
                ))}
                {/* Data Boundary */}
                {field("data_boundary", "Data Boundary", (
                    <select
                        value={form.data_boundary}
                        onChange={(e) => onChange({ ...form, data_boundary: e.target.value })}
                        className={inputCls}
                    >
                        <option value="local_only">Local Only</option>
                        <option value="leaves_org">Leaves Org</option>
                    </select>
                ))}
            </div>

            {/* Roles */}
            <div className="space-y-1">
                <label className="text-[10px] uppercase tracking-wider text-cortex-text-muted">Roles Allowed</label>
                <div className="flex flex-wrap gap-1.5">
                    {COUNCIL_ROLES.map((role) => {
                        const active = form.roles_allowed.includes(role);
                        return (
                            <button
                                key={role}
                                type="button"
                                onClick={() => toggleRole(role)}
                                className={`px-2 py-0.5 rounded text-[10px] border transition-colors ${
                                    active
                                        ? "border-cortex-primary text-cortex-primary bg-cortex-primary/10"
                                        : "border-cortex-border text-cortex-text-muted hover:border-cortex-primary/50"
                                }`}
                            >
                                {role}
                            </button>
                        );
                    })}
                </div>
            </div>

            {/* Enabled toggle */}
            <label className="flex items-center gap-2 cursor-pointer text-xs text-cortex-text-muted select-none">
                <div
                    onClick={() => onChange({ ...form, enabled: !form.enabled })}
                    className={`w-8 h-4 rounded-full relative transition-colors border cursor-pointer ${
                        form.enabled
                            ? "bg-cortex-success/20 border-cortex-success/40"
                            : "bg-cortex-bg border-cortex-border"
                    }`}
                >
                    <div className={`absolute top-0 w-4 h-4 rounded-full shadow-sm transition-all ${
                        form.enabled ? "right-0 bg-cortex-success" : "left-0 bg-cortex-text-muted"
                    }`} />
                </div>
                Enable on save
            </label>

            {/* Test connection (edit only) */}
            {isEdit && onProbe && (
                <div className="flex items-center gap-2">
                    <button
                        type="button"
                        onClick={onProbe}
                        disabled={probing}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:border-cortex-primary hover:text-cortex-primary transition-colors disabled:opacity-50"
                    >
                        {probing ? <Loader2 className="w-3 h-3 animate-spin" /> : <Wifi className="w-3 h-3" />}
                        Test Connection
                    </button>
                    {probeResult === "alive" && (
                        <span className="flex items-center gap-1 text-cortex-success text-xs">
                            <CheckCircle className="w-3 h-3" /> Online
                        </span>
                    )}
                    {probeResult === "dead" && (
                        <span className="flex items-center gap-1 text-red-400 text-xs">
                            <WifiOff className="w-3 h-3" /> Unreachable
                        </span>
                    )}
                </div>
            )}
        </div>
    );
}

// ── Modal wrapper ─────────────────────────────────────────────────────────────

function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl">
                <div className="flex items-center justify-between px-5 py-3 border-b border-cortex-border">
                    <h3 className="text-sm font-semibold text-cortex-text-main">{title}</h3>
                    <button onClick={onClose} className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                        <X className="w-4 h-4" />
                    </button>
                </div>
                <div className="p-5">{children}</div>
            </div>
        </div>
    );
}

// ── Main component ────────────────────────────────────────────────────────────

export default function BrainsPage() {
    const [brains, setBrains] = useState<BrainEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [confirmRemote, setConfirmRemote] = useState<BrainEntry | null>(null);

    // Add modal
    const [showAdd, setShowAdd] = useState(false);
    const [addForm, setAddForm] = useState<ProviderFormData>(blankForm());
    const [addSaving, setAddSaving] = useState(false);
    const [addError, setAddError] = useState<string | null>(null);

    // Edit modal
    const [editTarget, setEditTarget] = useState<BrainEntry | null>(null);
    const [editForm, setEditForm] = useState<ProviderFormData>(blankForm());
    const [editSaving, setEditSaving] = useState(false);
    const [editError, setEditError] = useState<string | null>(null);
    const [probing, setProbing] = useState(false);
    const [probeResult, setProbeResult] = useState<"alive" | "dead" | null>(null);

    // Delete confirmation
    const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
    const [deleting, setDeleting] = useState(false);

    // Per-row probe
    const [rowProbing, setRowProbing] = useState<string | null>(null);
    const [rowProbeResult, setRowProbeResult] = useState<Record<string, "alive" | "dead">>({});

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

    // ── Toggle (existing) ───────────────────────────────────────────────────

    const toggleBrain = async (brain: BrainEntry) => {
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

    // ── Add provider ────────────────────────────────────────────────────────

    const openAdd = () => {
        setAddForm(blankForm());
        setAddError(null);
        setShowAdd(true);
    };

    const submitAdd = async () => {
        setAddSaving(true);
        setAddError(null);
        try {
            const res = await fetch("/api/v1/brains", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(addForm),
            });
            const body = await res.json();
            if (!res.ok || !body.ok) {
                setAddError(body.error || `Error ${res.status}`);
                return;
            }
            setShowAdd(false);
            fetchBrains();
        } catch (err) {
            setAddError(err instanceof Error ? err.message : "Request failed");
        } finally {
            setAddSaving(false);
        }
    };

    // ── Edit provider ───────────────────────────────────────────────────────

    const openEdit = (b: BrainEntry) => {
        setEditTarget(b);
        setEditForm({
            id: b.id,
            type: b.type,
            endpoint: b.endpoint ?? "",
            model_id: b.model_id,
            api_key: "",
            location: b.location,
            data_boundary: b.data_boundary,
            usage_policy: b.usage_policy,
            roles_allowed: b.roles_allowed?.length ? b.roles_allowed : ["all"],
            enabled: b.enabled,
        });
        setEditError(null);
        setProbeResult(null);
        setProbing(false);
    };

    const submitEdit = async () => {
        if (!editTarget) return;
        setEditSaving(true);
        setEditError(null);
        try {
            const res = await fetch(`/api/v1/brains/${editTarget.id}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(editForm),
            });
            const body = await res.json();
            if (!res.ok || !body.ok) {
                setEditError(body.error || `Error ${res.status}`);
                return;
            }
            setEditTarget(null);
            fetchBrains();
        } catch (err) {
            setEditError(err instanceof Error ? err.message : "Request failed");
        } finally {
            setEditSaving(false);
        }
    };

    const probeInModal = async () => {
        if (!editTarget) return;
        setProbing(true);
        setProbeResult(null);
        try {
            const res = await fetch(`/api/v1/brains/${editTarget.id}/probe`, { method: "POST" });
            const body = await res.json();
            setProbeResult(body?.data?.alive ? "alive" : "dead");
        } catch {
            setProbeResult("dead");
        } finally {
            setProbing(false);
        }
    };

    // ── Row probe ───────────────────────────────────────────────────────────

    const probeRow = async (id: string) => {
        setRowProbing(id);
        try {
            const res = await fetch(`/api/v1/brains/${id}/probe`, { method: "POST" });
            const body = await res.json();
            setRowProbeResult((prev) => ({ ...prev, [id]: body?.data?.alive ? "alive" : "dead" }));
        } catch {
            setRowProbeResult((prev) => ({ ...prev, [id]: "dead" }));
        } finally {
            setRowProbing(null);
        }
    };

    // ── Delete provider ─────────────────────────────────────────────────────

    const confirmDelete = async () => {
        if (!deleteTarget) return;
        setDeleting(true);
        try {
            const res = await fetch(`/api/v1/brains/${deleteTarget}`, { method: "DELETE" });
            const body = await res.json();
            if (!res.ok || !body.ok) {
                console.error("[BRAINS] Delete failed:", body.error);
            } else {
                fetchBrains();
            }
        } catch { /* ignore */ }
        setDeleteTarget(null);
        setDeleting(false);
    };

    // ── Render helpers ──────────────────────────────────────────────────────

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
            {/* Header */}
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Provider Management</h3>
                <div className="flex items-center gap-2">
                    <button
                        onClick={fetchBrains}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                    </button>
                    <button
                        onClick={openAdd}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors"
                    >
                        <Plus className="w-3.5 h-3.5" />
                        Add Provider
                    </button>
                </div>
            </div>

            {/* Table */}
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
                                    <td className="px-4 py-2.5">
                                        <div className="flex items-center justify-center gap-1">
                                            {/* Edit */}
                                            <button
                                                onClick={() => openEdit(b)}
                                                title="Edit"
                                                className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors"
                                            >
                                                <Pencil className="w-3.5 h-3.5" />
                                            </button>
                                            {/* Test */}
                                            <button
                                                onClick={() => probeRow(b.id)}
                                                title="Test connection"
                                                disabled={rowProbing === b.id}
                                                className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors disabled:opacity-50"
                                            >
                                                {rowProbing === b.id
                                                    ? <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                                    : <Wifi className="w-3.5 h-3.5" />
                                                }
                                            </button>
                                            {/* Delete */}
                                            <button
                                                onClick={() => setDeleteTarget(b.id)}
                                                title="Delete"
                                                className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-red-400 transition-colors"
                                            >
                                                <Trash2 className="w-3.5 h-3.5" />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                                {/* Delete confirmation row */}
                                {deleteTarget === b.id && (
                                    <tr className="bg-red-400/5 border-t border-red-400/20">
                                        <td colSpan={8} className="px-4 py-2.5">
                                            <div className="flex items-center justify-between">
                                                <span className="text-xs text-red-400">
                                                    Delete <strong>{b.id}</strong>? This cannot be undone.
                                                </span>
                                                <div className="flex items-center gap-2">
                                                    <button
                                                        onClick={() => setDeleteTarget(null)}
                                                        className="px-2.5 py-1 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                                                    >
                                                        Cancel
                                                    </button>
                                                    <button
                                                        onClick={confirmDelete}
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

            {/* Add Provider Modal */}
            {showAdd && (
                <Modal title="Add Provider" onClose={() => setShowAdd(false)}>
                    <div className="space-y-4">
                        <ProviderForm
                            form={addForm}
                            onChange={setAddForm}
                            isEdit={false}
                            probeResult={null}
                            probing={false}
                            showPresets
                        />
                        {addError && (
                            <p className="text-red-400 text-xs">{addError}</p>
                        )}
                        <div className="flex justify-end gap-2 pt-2 border-t border-cortex-border">
                            <button
                                onClick={() => setShowAdd(false)}
                                className="px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={submitAdd}
                                disabled={addSaving || !addForm.id}
                                className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors disabled:opacity-50"
                            >
                                {addSaving && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                                Add Provider
                            </button>
                        </div>
                    </div>
                </Modal>
            )}

            {/* Edit Provider Modal */}
            {editTarget && (
                <Modal title={`Edit Provider — ${editTarget.id}`} onClose={() => setEditTarget(null)}>
                    <div className="space-y-4">
                        <ProviderForm
                            form={editForm}
                            onChange={setEditForm}
                            isEdit
                            probeResult={probeResult}
                            probing={probing}
                            onProbe={probeInModal}
                            showPresets={false}
                        />
                        {editError && (
                            <p className="text-red-400 text-xs">{editError}</p>
                        )}
                        <div className="flex justify-end gap-2 pt-2 border-t border-cortex-border">
                            <button
                                onClick={() => setEditTarget(null)}
                                className="px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={submitEdit}
                                disabled={editSaving}
                                className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors disabled:opacity-50"
                            >
                                {editSaving && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                                Save Changes
                            </button>
                        </div>
                    </div>
                </Modal>
            )}
        </div>
    );
}
