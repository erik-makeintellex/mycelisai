"use client";

import { useState } from "react";
import { Archive, ArchiveRestore, Trash2 } from "lucide-react";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import type { DeploymentContextEntry } from "./DeploymentContextPanel";

type LifecycleAction = "archive" | "restore" | "delete";
type Blocker = { code: string; httpStatus: number };

const BUTTON_CLASS = "inline-flex items-center gap-1.5 rounded border px-2.5 py-1.5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50";

function lifecycleRequest(entry: DeploymentContextEntry, action: LifecycleAction): [string, RequestInit] {
    const base = `/api/v1/memory/deployment-context/${encodeURIComponent(entry.artifact_id)}`;
    return action === "delete" ? [base, { method: "DELETE" }] : [`${base}/${action}`, { method: "POST" }];
}

// Archive, restore and permanent delete for one saved entry. The server
// decides authority; can_manage only hides buttons that would be refused.
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

    return (
        <div className="space-y-2">
            <div className="flex flex-wrap gap-2">
                {archived ? (
                    <button type="button" disabled={busy} onClick={() => run("restore")} aria-label={`Restore ${entry.title}`}
                        className={`${BUTTON_CLASS} border-cortex-primary/40 text-cortex-primary hover:bg-cortex-primary/10`}>
                        <ArchiveRestore className="h-3.5 w-3.5" />
                        Restore
                    </button>
                ) : (
                    <button type="button" disabled={busy} onClick={() => run("archive")} aria-label={`Archive ${entry.title}`}
                        className={`${BUTTON_CLASS} border-cortex-border text-cortex-text-main hover:bg-cortex-bg`}>
                        <Archive className="h-3.5 w-3.5" />
                        Archive
                    </button>
                )}
                <button type="button" disabled={busy} onClick={() => setConfirmDelete(true)} aria-label={`Delete ${entry.title}`}
                    className={`${BUTTON_CLASS} border-cortex-danger/40 text-cortex-danger hover:bg-cortex-danger/10`}>
                    <Trash2 className="h-3.5 w-3.5" />
                    Delete
                </button>
            </div>
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
                <InlineBlockerNotice code={blocker.code} httpStatus={blocker.httpStatus} viewerIsAdmin={viewerIsAdmin}
                    reason={blocker.code === "admin_required" ? "memory" : undefined} onDismiss={() => setBlocker(null)} />
            ) : null}
        </div>
    );
}
