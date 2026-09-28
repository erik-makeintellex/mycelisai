"use client";

import React, { useEffect, useState, useCallback } from "react";
import {
    Brain, Globe, Server, RefreshCw, AlertTriangle, CheckCircle,
    XCircle, Power, Plus, Pencil, Trash2, Wifi, WifiOff, Loader2,
} from "lucide-react";
import RemoteEnableModal from "./RemoteEnableModal";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import ProviderForm from "@/components/settings/BrainsPageProviderForm";
import TokenBudgetsPanel from "@/components/settings/TokenBudgetsPanel";
import { blankForm, Modal, TOKEN_BUDGET_PRESETS, type BrainEntry, type ProviderFormData } from "@/components/settings/BrainsPageFormShared";


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

    // Toggle/policy blockers: an engine in use (409 provider_bound) or an
    // audit-unavailable policy change must explain what happened instead of
    // the control silently snapping back with no message.
    const [toggleBlocker, setToggleBlocker] = useState<{ code: string; httpStatus?: number; boundProfiles: string[] } | null>(null);
    const [policyBlocker, setPolicyBlocker] = useState<{ code: string; httpStatus?: number } | null>(null);

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
        if (!brain.enabled && brain.location === "remote") {
            setConfirmRemote(brain);
            return;
        }
        await doToggle(brain.id, !brain.enabled);
    };

    const doToggle = async (id: string, enabled: boolean) => {
        try {
            const res = await fetch(`/api/v1/brains/${id}/toggle`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ enabled }),
            });
            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                const code: string = body?.data?.code ?? (res.status === 409 ? "provider_bound" : "request_failed");
                const boundProfiles: string[] = Array.isArray(body?.data?.profiles) ? body.data.profiles : [];
                setToggleBlocker({ code, httpStatus: res.status, boundProfiles });
                return;
            }
            setToggleBlocker(null);
            fetchBrains();
        } catch {
            setToggleBlocker({ code: "request_failed", boundProfiles: [] });
        }
    };

    const resetProfileOverride = async (profile: string) => {
        try {
            await fetch(`/api/v1/cognitive/profiles/${profile}/override`, { method: "DELETE" });
        } catch { /* the banner stays open so the admin can retry */ }
        setToggleBlocker(null);
        fetchBrains();
    };

    const updatePolicy = async (id: string, policy: string) => {
        try {
            const res = await fetch(`/api/v1/brains/${id}/policy`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ usage_policy: policy }),
            });
            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                setPolicyBlocker({ code: body?.data?.code ?? "request_failed", httpStatus: res.status });
                return;
            }
            setPolicyBlocker(null);
            fetchBrains();
        } catch {
            setPolicyBlocker({ code: "request_failed" });
        }
    };

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
            token_budget_profile: b.token_budget_profile || "standard",
            max_output_tokens: b.max_output_tokens || 1024,
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

            {/* Toggle blocker (e.g. provider_bound): explain, don't snap back silently */}
            {toggleBlocker && (
                <InlineBlockerNotice code={toggleBlocker.code} httpStatus={toggleBlocker.httpStatus} onDismiss={() => setToggleBlocker(null)}>
                    {toggleBlocker.boundProfiles.length > 0 && (
                        <div className="mt-2 flex flex-wrap gap-1.5">
                            {toggleBlocker.boundProfiles.map((profile) => (
                                <button
                                    key={profile}
                                    onClick={() => resetProfileOverride(profile)}
                                    className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-2 py-1 text-xs text-cortex-primary hover:bg-cortex-primary/20"
                                >
                                    Use default engine for {profile}
                                </button>
                            ))}
                        </div>
                    )}
                </InlineBlockerNotice>
            )}

            {policyBlocker && (
                <InlineBlockerNotice code={policyBlocker.code} httpStatus={policyBlocker.httpStatus} onDismiss={() => setPolicyBlocker(null)} />
            )}

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
                                            onChange={(e) => updatePolicy(b.id, e.target.value)}
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
                                        <td colSpan={9} className="px-4 py-2.5">
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

            {/* Advanced: B1 token budgets (effective policy, overrides for admins) */}
            <div className="pt-2 space-y-3">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Advanced</h3>
                <TokenBudgetsPanel />
            </div>
        </div>
    );
}
