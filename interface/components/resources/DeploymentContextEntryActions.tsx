"use client";

import { useState } from "react";
import { Archive, ArchiveRestore, Pencil, Trash2 } from "lucide-react";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import { editDeploymentContextEntry, type DeploymentContextEditPatch } from "@/lib/deploymentContextEdit";
import type { DeploymentContextEntry } from "./DeploymentContextPanel";

type LifecycleAction = "archive" | "restore" | "delete";
type Blocker = { code: string; httpStatus: number };
type EditDraft = { title: string; sourceLabel: string; content: string };

const BUTTON_CLASS = "inline-flex items-center gap-1.5 rounded border px-2.5 py-1.5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50";
const EDIT_INPUT_CLASS = "mt-1 w-full rounded border border-cortex-border bg-cortex-bg px-2.5 py-1.5 text-xs normal-case text-cortex-text-main placeholder:text-cortex-text-muted/60 focus:outline-none focus:border-cortex-primary";
const EDIT_LABEL_CLASS = "block text-[11px] font-mono uppercase tracking-widest text-cortex-text-muted";

function lifecycleRequest(entry: DeploymentContextEntry, action: LifecycleAction): [string, RequestInit] {
    const base = `/api/v1/memory/deployment-context/${encodeURIComponent(entry.artifact_id)}`;
    return action === "delete" ? [base, { method: "DELETE" }] : [`${base}/${action}`, { method: "POST" }];
}

// The PATCH content field replaces the saved text wholesale; the list API
// only ever hands back a 220-char content_preview, which would silently
// truncate a longer entry if it were saved back as-is. The draft starts
// blank so a save never re-sends a preview as the real content.
function emptyDraft(entry: DeploymentContextEntry): EditDraft {
    return { title: entry.title, sourceLabel: entry.source_label, content: "" };
}

// Archive, restore, permanent delete and in-place edit for one saved
// entry. The server decides authority; can_manage only hides buttons
// that would be refused.
export default function DeploymentContextEntryActions({
    entry,
    viewerIsAdmin,
    onChanged,
}: {
    entry: DeploymentContextEntry;
    viewerIsAdmin: boolean;
    onChanged: () => void | Promise<void>;
}) {
    const [busy, setBusy] = useState(false);
    const [confirmDelete, setConfirmDelete] = useState(false);
    const [blocker, setBlocker] = useState<Blocker | null>(null);
    const [editing, setEditing] = useState(false);
    const [draft, setDraft] = useState<EditDraft>(() => emptyDraft(entry));

    if (!entry.can_manage) return null;
    const archived = entry.lifecycle_state === "archived";

    const run = async (action: LifecycleAction) => {
        setBusy(true);
        setBlocker(null);
        try {
            const res = await fetch(...lifecycleRequest(entry, action));
            const payload = await res.json().catch(() => ({}));
            if (!res.ok || payload?.ok === false) {
                const code = typeof payload?.data?.code === "string" ? payload.data.code : "request_failed";
                setBlocker({ code, httpStatus: res.status });
                return;
            }
            setConfirmDelete(false);
            await onChanged();
        } catch {
            setBlocker({ code: "request_failed", httpStatus: 0 });
        } finally {
            setBusy(false);
        }
    };

    const startEdit = () => {
        setBlocker(null);
        setDraft(emptyDraft(entry));
        setEditing(true);
    };

    const cancelEdit = () => {
        setBlocker(null);
        setEditing(false);
    };

    const titleChanged = draft.title.trim().length > 0 && draft.title.trim() !== entry.title;
    const labelChanged = draft.sourceLabel.trim().length > 0 && draft.sourceLabel.trim() !== entry.source_label;
    const contentTyped = draft.content.trim().length > 0;
    const canSave = !busy && (titleChanged || labelChanged || contentTyped);

    const saveEdit = async () => {
        if (!canSave) return;
        setBusy(true);
        setBlocker(null);
        const patch: DeploymentContextEditPatch = {};
        if (titleChanged) patch.title = draft.title.trim();
        if (labelChanged) patch.source_label = draft.sourceLabel.trim();
        if (contentTyped) patch.content = draft.content;
        const result = await editDeploymentContextEntry(entry.artifact_id, patch);
        setBusy(false);
        if (!result.ok) {
            setBlocker({ code: result.code, httpStatus: result.httpStatus });
            return;
        }
        setEditing(false);
        await onChanged();
    };

    // Either race (archived elsewhere, or edited/moved elsewhere) means
    // this draft is stale; refresh the list and close rather than risk
    // resubmitting over a version the operator never saw.
    const reloadAfterConflict = async () => {
        setBlocker(null);
        setEditing(false);
        await onChanged();
    };

    return (
        <div className="space-y-2">
            {editing ? (
                <div className="space-y-2 rounded-lg border border-cortex-border bg-cortex-bg/60 p-3">
                    <label className={EDIT_LABEL_CLASS}>
                        Title
                        <input value={draft.title} disabled={busy} onChange={(e) => setDraft((d) => ({ ...d, title: e.target.value }))} className={EDIT_INPUT_CLASS} />
                    </label>
                    <label className={EDIT_LABEL_CLASS}>
                        Source label
                        <input value={draft.sourceLabel} disabled={busy} onChange={(e) => setDraft((d) => ({ ...d, sourceLabel: e.target.value }))} className={EDIT_INPUT_CLASS} />
                    </label>
                    <label className={EDIT_LABEL_CLASS}>
                        Content
                        <textarea
                            value={draft.content}
                            disabled={busy}
                            onChange={(e) => setDraft((d) => ({ ...d, content: e.target.value }))}
                            placeholder="Leave blank to keep the saved content. Type new text to replace it entirely."
                            className={`${EDIT_INPUT_CLASS} min-h-[100px] resize-y`}
                        />
                    </label>
                    <div className="flex flex-wrap gap-2">
                        <button type="button" disabled={!canSave} onClick={() => void saveEdit()} aria-label={`Save ${entry.title}`}
                            className={`${BUTTON_CLASS} border-cortex-primary bg-cortex-primary text-white hover:brightness-105`}>
                            Save
                        </button>
                        <button type="button" disabled={busy} onClick={cancelEdit}
                            className={`${BUTTON_CLASS} border-cortex-border text-cortex-text-main hover:bg-cortex-bg`}>
                            Cancel
                        </button>
                    </div>
                </div>
            ) : (
                <div className="flex flex-wrap items-center gap-2">
                    {archived ? (
                        <>
                            <button type="button" disabled={busy} onClick={() => run("restore")} aria-label={`Restore ${entry.title}`}
                                className={`${BUTTON_CLASS} border-cortex-primary/40 text-cortex-primary hover:bg-cortex-primary/10`}>
                                <ArchiveRestore className="h-3.5 w-3.5" />
                                Restore
                            </button>
                            <span className="text-[11px] text-cortex-text-muted">Restore to edit</span>
                        </>
                    ) : (
                        <>
                            <button type="button" disabled={busy} onClick={() => run("archive")} aria-label={`Archive ${entry.title}`}
                                className={`${BUTTON_CLASS} border-cortex-border text-cortex-text-main hover:bg-cortex-bg`}>
                                <Archive className="h-3.5 w-3.5" />
                                Archive
                            </button>
                            <button type="button" disabled={busy} onClick={startEdit} aria-label={`Edit ${entry.title}`}
                                className={`${BUTTON_CLASS} border-cortex-border text-cortex-text-main hover:bg-cortex-bg`}>
                                <Pencil className="h-3.5 w-3.5" />
                                Edit
                            </button>
                        </>
                    )}
                    <button type="button" disabled={busy} onClick={() => setConfirmDelete(true)} aria-label={`Delete ${entry.title}`}
                        className={`${BUTTON_CLASS} border-cortex-danger/40 text-cortex-danger hover:bg-cortex-danger/10`}>
                        <Trash2 className="h-3.5 w-3.5" />
                        Delete
                    </button>
                </div>
            )}
            {confirmDelete ? (
                <div className="flex flex-wrap items-center gap-2 rounded-lg border border-cortex-danger/40 bg-cortex-danger/10 px-3 py-2 text-xs text-cortex-text-main">
                    <span>Delete permanently? Soma will no longer use this.</span>
                    <button type="button" disabled={busy} onClick={() => run("delete")}
                        className={`${BUTTON_CLASS} border-cortex-danger bg-cortex-danger text-white hover:brightness-105`}>
                        Delete permanently
                    </button>
                    <button type="button" disabled={busy} onClick={() => setConfirmDelete(false)}
                        className={`${BUTTON_CLASS} border-cortex-border text-cortex-text-main hover:bg-cortex-bg`}>
                        Cancel
                    </button>
                </div>
            ) : null}
            {blocker ? (
                <InlineBlockerNotice
                    code={blocker.code}
                    httpStatus={blocker.httpStatus}
                    viewerIsAdmin={viewerIsAdmin}
                    reason={blocker.code === "admin_required" ? "memory" : undefined}
                    onDismiss={blocker.code === "memory_entry_archived" ? () => void reloadAfterConflict() : () => setBlocker(null)}
                    onRetry={blocker.code === "memory_entry_changed" ? () => void reloadAfterConflict() : undefined}
                />
            ) : null}
        </div>
    );
}
