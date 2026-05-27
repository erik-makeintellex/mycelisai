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

    const openCurrentFolder = async () => {
        if (folderState === "opening") return;
        setFolderState("opening");
        onStatus(`Opening local folder for ${currentPath}...`);
        try {
            const res = await fetch(`/api/v1/workspace/files/reveal?path=${encodeURIComponent(currentPath)}`, {
                method: "POST",
            });
            setFolderState(res.ok ? "opened" : "failed");
            onStatus(res.ok ? `Opened local folder for ${currentPath}` : "Could not open the local workspace folder");
        } catch {
            setFolderState("failed");
            onStatus("Could not open the local workspace folder");
        }
        window.setTimeout(() => setFolderState("idle"), 1800);
    };

    return (
        <div className="mb-3 rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-3">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                    <p className="text-xs font-semibold text-cortex-text-main">Generated content lives here</p>
                    <p className="mt-1 text-[11px] leading-5 text-cortex-text-muted">
                        Open the current workspace folder on this machine, or browse files below inside Mycelis.
                    </p>
                </div>
                <button
                    type="button"
                    onClick={() => void openCurrentFolder()}
                    disabled={folderState === "opening"}
                    className="inline-flex shrink-0 items-center justify-center gap-1.5 rounded border border-cortex-primary/35 bg-cortex-surface px-3 py-1.5 text-xs font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/10 disabled:opacity-60"
                    title={`Open local folder for ${currentPath}`}
                    aria-label={`Open current folder ${currentPath}`}
                >
                    {folderState === "opening" ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : null}
                    {folderState === "opened" ? <Check className="h-3.5 w-3.5" /> : null}
                    {folderState === "idle" || folderState === "failed" ? <FolderOpen className="h-3.5 w-3.5" /> : null}
                    Open folder
                </button>
            </div>
        </div>
    );
}
