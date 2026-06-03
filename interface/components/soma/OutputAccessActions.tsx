"use client";

import { useMemo, useState } from "react";
import { Check, ExternalLink, FolderOpen, Loader2 } from "lucide-react";

export function workspacePathFromOutputUrl(url: string | null) {
    if (!url) return null;
    const trimmed = url.trim();
    if (trimmed && !/^(https?:)?\/\//i.test(trimmed) && !trimmed.startsWith("/") && (trimmed.startsWith("workspace/") || trimmed.includes("/") || /\.[a-z0-9]{1,8}$/i.test(trimmed))) {
        return trimmed.replace(/\\/g, "/");
    }
    try {
        const parsed = new URL(trimmed, "http://mycelis.local");
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
    folderLabel = "Open folder",
}: {
    label: string;
    url: string | null;
    storagePath?: string | null;
    openLabel?: string;
    folderLabel?: string;
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
        window.setTimeout(() => setFolderState("idle"), 4000);
    };

    return (
        <span className="inline-flex shrink-0 items-center gap-1">
            {url && (
                <button
                    type="button"
                    onClick={openOutput}
                    className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-[11px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15"
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
                    className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-border/80 bg-cortex-bg/70 px-2.5 text-[11px] font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/45 hover:bg-cortex-primary/10 hover:text-cortex-primary"
                    title={folderState === "failed" ? "Could not open local storage folder" : `Open local storage folder for ${label}`}
                    aria-label={`Open local folder for ${label}`}
                    disabled={folderState === "opening"}
                >
                    {folderState === "opening" ? <Loader2 className="h-3 w-3 animate-spin" /> : null}
                    {folderState === "opened" ? <Check className="h-3 w-3" /> : null}
                    {folderState === "idle" ? <FolderOpen className="h-3 w-3" /> : null}
                    {folderState === "failed" ? <FolderOpen className="h-3 w-3 text-amber-300" /> : null}
                    {folderState === "opened" ? "Folder opened" : folderState === "failed" ? "Folder blocked" : folderLabel}
                </button>
            )}
        </span>
    );
}
