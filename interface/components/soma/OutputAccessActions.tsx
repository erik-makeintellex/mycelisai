"use client";

import { useMemo, useState } from "react";
import { Check, ExternalLink, FolderOpen, Loader2 } from "lucide-react";

export function workspacePathFromOutputUrl(url: string | null) {
    if (!url) return null;
    try {
        const parsed = new URL(url, "http://mycelis.local");
        if (!parsed.pathname.endsWith("/api/v1/workspace/files/view")) return null;
        return parsed.searchParams.get("path");
    } catch {
        return null;
    }
}

export default function OutputAccessActions({
    label,
    url,
    storagePath,
    openLabel = "Open",
}: {
    label: string;
    url: string | null;
    storagePath?: string | null;
    openLabel?: string;
}) {
    const [folderState, setFolderState] = useState<"idle" | "opening" | "opened" | "failed">("idle");
    const workspacePath = useMemo(() => storagePath?.trim() || workspacePathFromOutputUrl(url), [storagePath, url]);
    if (!url && !workspacePath) return null;

    const openOutput = () => {
        if (!url) return;
        window.open(url, "_blank", "noopener,noreferrer");
    };

    const openFolder = async () => {
        if (!workspacePath || folderState === "opening") return;
        setFolderState("opening");
        try {
            const response = await fetch(`/api/v1/workspace/files/reveal?path=${encodeURIComponent(workspacePath)}`, {
                method: "POST",
            });
            setFolderState(response.ok ? "opened" : "failed");
        } catch {
            setFolderState("failed");
        }
        window.setTimeout(() => setFolderState("idle"), 1800);
    };

    return (
        <span className="inline-flex shrink-0 items-center gap-1">
            {url && (
                <button
                    type="button"
                    onClick={openOutput}
                    className="inline-flex h-5 items-center gap-1 rounded border border-cortex-border/70 px-1.5 text-[9px] font-mono uppercase text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
                    title={`${openLabel} ${label} in a new browser window`}
                    aria-label={`${openLabel} ${label} in a new browser window`}
                >
                    <ExternalLink className="h-3 w-3" />
                    {openLabel}
                </button>
            )}
            {workspacePath && (
                <button
                    type="button"
                    onClick={() => void openFolder()}
                    className="inline-flex h-5 items-center gap-1 rounded border border-cortex-border/70 px-1.5 text-[9px] font-mono uppercase text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
                    title={folderState === "failed" ? "Could not open local storage folder" : `Open local storage folder for ${label}`}
                    aria-label={`Open local folder for ${label}`}
                    disabled={folderState === "opening"}
                >
                    {folderState === "opening" ? <Loader2 className="h-3 w-3 animate-spin" /> : null}
                    {folderState === "opened" ? <Check className="h-3 w-3" /> : null}
                    {folderState === "idle" || folderState === "failed" ? <FolderOpen className="h-3 w-3" /> : null}
                    Storage
                </button>
            )}
        </span>
    );
}
