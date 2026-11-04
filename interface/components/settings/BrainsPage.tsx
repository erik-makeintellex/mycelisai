"use client";

import React, { useEffect, useState, useCallback } from "react";
import { RefreshCw, AlertTriangle, Plus, Loader2 } from "lucide-react";
import RemoteEnableModal from "./RemoteEnableModal";
import { BrainEntry, ProviderFormData, blankForm, toProviderFormData } from "./brainsTypes";
import { Modal, ProviderForm } from "./BrainsProviderModal";
import BrainsProviderTable from "./BrainsProviderTable";

export default function BrainsPage() {
    const [brains, setBrains] = useState<BrainEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [confirmRemote, setConfirmRemote] = useState<BrainEntry | null>(null);
    const [showAdd, setShowAdd] = useState(false);
    const [addForm, setAddForm] = useState<ProviderFormData>(blankForm());
    const [addSaving, setAddSaving] = useState(false);
    const [addError, setAddError] = useState<string | null>(null);
    const [editTarget, setEditTarget] = useState<BrainEntry | null>(null);
    const [editForm, setEditForm] = useState<ProviderFormData>(blankForm());
    const [editSaving, setEditSaving] = useState(false);
    const [editError, setEditError] = useState<string | null>(null);
    const [probing, setProbing] = useState(false);
    const [probeResult, setProbeResult] = useState<"alive" | "dead" | null>(null);
    const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
    const [deleting, setDeleting] = useState(false);
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

    const toggleBrain = async (brain: BrainEntry) => {
        if (!brain.enabled && brain.location === "remote") {
            setConfirmRemote(brain);
            return;
        }
        await doToggle(brain.id, !brain.enabled);
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
        setEditForm(toProviderFormData(b));
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
            if (!res.ok || !body.ok) console.error("[BRAINS] Delete failed:", body.error);
            else fetchBrains();
        } catch { /* ignore */ }
        setDeleteTarget(null);
        setDeleting(false);
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Provider Management</h3>
                <div className="flex items-center gap-2">
                    <button onClick={fetchBrains} className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors" title="Refresh">
                        <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                    </button>
                    <button onClick={openAdd} className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors">
                        <Plus className="w-3.5 h-3.5" />
                        Add Provider
                    </button>
                </div>
            </div>

            <BrainsProviderTable
                brains={brains}
                loading={loading}
                rowProbing={rowProbing}
                rowProbeResult={rowProbeResult}
                deleteTarget={deleteTarget}
                deleting={deleting}
                onUpdatePolicy={updatePolicy}
                onToggleBrain={toggleBrain}
                onOpenEdit={openEdit}
                onProbeRow={probeRow}
                onSelectDelete={setDeleteTarget}
                onConfirmDelete={confirmDelete}
            />

            <div className="flex items-start gap-2 p-3 rounded border border-amber-400/20 bg-amber-400/5 text-xs text-amber-400">
                <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                <div>
                    <strong>Data Boundary Notice:</strong> Enabling remote providers means data may leave your local environment.
                    Review each provider&apos;s data boundary before enabling.
                </div>
            </div>

            {confirmRemote && (
                <RemoteEnableModal
                    provider={confirmRemote}
                    onConfirm={() => { doToggle(confirmRemote.id, true); setConfirmRemote(null); }}
                    onCancel={() => setConfirmRemote(null)}
                />
            )}

            {showAdd && (
                <Modal title="Add Provider" onClose={() => setShowAdd(false)}>
                    <div className="space-y-4">
                        <ProviderForm form={addForm} onChange={setAddForm} isEdit={false} probeResult={null} probing={false} showPresets />
                        {addError && <p className="text-red-400 text-xs">{addError}</p>}
                        <div className="flex justify-end gap-2 pt-2 border-t border-cortex-border">
                            <button onClick={() => setShowAdd(false)} className="px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                                Cancel
                            </button>
                            <button onClick={submitAdd} disabled={addSaving || !addForm.id} className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors disabled:opacity-50">
                                {addSaving && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                                Add Provider
                            </button>
                        </div>
                    </div>
                </Modal>
            )}

            {editTarget && (
                <Modal title={`Edit Provider — ${editTarget.id}`} onClose={() => setEditTarget(null)}>
                    <div className="space-y-4">
                        <ProviderForm form={editForm} onChange={setEditForm} isEdit probeResult={probeResult} probing={probing} onProbe={probeInModal} showPresets={false} />
                        {editError && <p className="text-red-400 text-xs">{editError}</p>}
                        <div className="flex justify-end gap-2 pt-2 border-t border-cortex-border">
                            <button onClick={() => setEditTarget(null)} className="px-3 py-1.5 rounded border border-cortex-border text-xs text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                                Cancel
                            </button>
                            <button onClick={submitEdit} disabled={editSaving} className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 text-cortex-primary text-xs hover:bg-cortex-primary/20 transition-colors disabled:opacity-50">
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
