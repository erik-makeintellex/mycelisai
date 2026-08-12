"use client";

import { useMemo, useState } from "react";
import { Check, ExternalLink, FolderOpen, Loader2 } from "lucide-react";
import { outputFolderButtonLabel, type OutputFolderState } from "@/lib/deliveryRuntimeLanguage";

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
    primary = false,
    showOpen = true,
    showFolder = true,
}: {
    label: string;
    url: string | null;
    storagePath?: string | null;
    openLabel?: string;
    folderLabel?: string;
    primary?: boolean;
    showOpen?: boolean;
    showFolder?: boolean;
}) {
    const [folderState, setFolderState] = useState<OutputFolderState>("idle");
    const workspacePath = useMemo(() => storagePath?.trim() || workspacePathFromOutputUrl(url), [storagePath, url]);
    if ((!url || !showOpen) && (!workspacePath || !showFolder)) return null;

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

    const folderTitle = folderState === "failed"
        ? "Could not open the local folder. Browse it from Resources -> Output Files."
        : `Open the local folder containing ${label}${workspacePath ? ` (${workspacePath})` : ""}`;

    return (
        <span className={`inline-flex min-w-0 items-center gap-1 ${primary ? "w-full sm:w-auto" : "shrink-0"}`}>
            {url && showOpen && (
                <button
                    type="button"
                    onClick={openOutput}
                    className={primary
                        ? "inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-cortex-primary bg-cortex-primary px-4 text-sm font-semibold text-cortex-bg transition-colors hover:bg-cortex-primary/90 sm:w-auto"
                        : "inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-[11px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15"}
                    title={`${openLabel} ${label} in a new browser window`}
                    aria-label={`${openLabel} ${label} in a new browser window`}
                >
                    <ExternalLink className={primary ? "h-4 w-4" : "h-3 w-3"} />
                    {openLabel}
                </button>
            )}
            {workspacePath && showFolder && (
                <button
                    type="button"
                    onClick={() => void openFolder()}
                    className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-border/80 bg-cortex-bg/70 px-2.5 text-[11px] font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/45 hover:bg-cortex-primary/10 hover:text-cortex-primary"
                    title={folderTitle}
                    aria-label={`Open local folder for ${label}${workspacePath ? ` at ${workspacePath}` : ""}`}
                    disabled={folderState === "opening"}
                >
                    {folderState === "opening" ? <Loader2 className="h-3 w-3 animate-spin" /> : null}
                    {folderState === "opened" ? <Check className="h-3 w-3" /> : null}
                    {folderState === "idle" ? <FolderOpen className="h-3 w-3" /> : null}
                    {folderState === "failed" ? <FolderOpen className="h-3 w-3 text-amber-300" /> : null}
                    {outputFolderButtonLabel(folderState, folderLabel)}
                </button>
            )}
        </span>
    );
}
