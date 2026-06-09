"use client";

import { useState } from "react";
import { Check, FolderOpen, Loader2 } from "lucide-react";

export default function WorkspaceFolderAccessCard({
    currentPath,
    onStatus,
}: {
    currentPath: string;
    onStatus: (status: string) => void;
}) {
    const [folderState, setFolderState] = useState<"idle" | "opening" | "opened" | "failed">("idle");
    const buttonLabel = folderState === "opening"
        ? "Opening..."
        : folderState === "opened"
            ? "Folder opened"
            : folderState === "failed"
                ? "Open failed"
                : "Open folder";

    const openCurrentFolder = async () => {
        if (folderState === "opening") return;
        setFolderState("opening");
        onStatus(`Opening local folder for ${currentPath}...`);
        try {
            const res = await fetch(`/api/v1/workspace/files/reveal?path=${encodeURIComponent(currentPath)}`, {
                method: "POST",
            });
            setFolderState(res.ok ? "opened" : "failed");
            onStatus(res.ok ? `Opened local folder for ${currentPath}` : `Could not open the local folder for ${currentPath}. You can still browse it here.`);
        } catch {
            setFolderState("failed");
            onStatus(`Could not open the local folder for ${currentPath}. You can still browse it here.`);
        }
        window.setTimeout(() => setFolderState("idle"), 1800);
    };

    return (
        <div className="mb-3 rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-3">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                    <p className="text-xs font-semibold text-cortex-text-main">Open generated output on this machine</p>
                    <p className="mt-1 text-[11px] leading-5 text-cortex-text-muted">
                        Jump to the current retained-output folder, or stay in Mycelis to browse and preview files.
                    </p>
                </div>
                <button
                    type="button"
                    onClick={() => void openCurrentFolder()}
                    disabled={folderState === "opening"}
                    className="inline-flex shrink-0 items-center justify-center gap-1.5 rounded border border-cortex-primary/35 bg-cortex-surface px-3 py-1.5 text-xs font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/10 disabled:opacity-60"
                    title={folderState === "failed" ? "Could not open the local folder. You can still browse this workspace path below." : `Open local folder for ${currentPath}`}
                    aria-label={`Open current folder ${currentPath}`}
                >
                    {folderState === "opening" ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : null}
                    {folderState === "opened" ? <Check className="h-3.5 w-3.5" /> : null}
                    {folderState === "idle" || folderState === "failed" ? <FolderOpen className="h-3.5 w-3.5" /> : null}
                    {buttonLabel}
                </button>
            </div>
        </div>
    );
}
