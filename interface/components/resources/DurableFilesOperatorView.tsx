"use client";

import React, { useEffect, useMemo, useState } from "react";
import {
    Download,
    Eye,
    EyeOff,
    FileCode2,
    FileText,
    FolderOpen,
    RefreshCw,
    Sliders,
} from "lucide-react";
import { useCortexStore, type Artifact } from "@/store/useCortexStore";
import { councilLabel } from "@/lib/labels";

interface DeliverableItem {
    id: string;
    title: string;
    contentType: string;
    filePath?: string;
    content?: string;
    producer: string;
    createdAt?: string;
    saved: boolean;
}

function triggerDownload(item: DeliverableItem) {
    const body = item.content ?? `Saved workspace file: ${item.filePath ?? item.title}\n`;
    const blob = new Blob([body], { type: item.contentType || "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = item.filePath?.split("/").pop() || item.title || "deliverable.txt";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

export default function DurableFilesOperatorView() {
    const artifacts = useCortexStore((s) => s.artifacts) ?? [];
    const isFetchingArtifacts = useCortexStore((s) => s.isFetchingArtifacts) ?? false;
    const fetchArtifacts = useCortexStore((s) => s.fetchArtifacts);
    const missionChat = useCortexStore((s) => s.missionChat) ?? [];
    const assistantName = useCortexStore((s) => s.assistantName) || "Soma";
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);

    const [previewId, setPreviewId] = useState<string | null>(null);
    const [openedNotice, setOpenedNotice] = useState<string | null>(null);

    useEffect(() => {
        fetchArtifacts?.({ limit: 50 });
    }, [fetchArtifacts]);

    const items = useMemo<DeliverableItem[]>(() => {
        const list: DeliverableItem[] = artifacts.map((art: Artifact) => ({
            id: art.id,
            title: art.title || art.file_path || "Generated Deliverable",
            contentType: art.content_type || "text/plain",
            filePath: art.file_path,
            content: art.content,
            producer: councilLabel(art.agent_id || "admin", assistantName).name,
            createdAt: art.created_at,
            saved: Boolean(art.file_path) || art.status === "approved",
        }));

        const seenIds = new Set(list.map((i) => i.id));
        for (const msg of missionChat) {
            if (!Array.isArray(msg.artifacts)) continue;
            for (const chatArt of msg.artifacts) {
                const artId = chatArt.id || chatArt.saved_path || chatArt.title || "workspace-artifact";
                if (seenIds.has(artId)) continue;
                seenIds.add(artId);
                list.push({
                    id: artId,
                    title: chatArt.title || chatArt.saved_path || "Workspace Artifact",
                    contentType: chatArt.content_type || "text/plain",
                    filePath: chatArt.saved_path,
                    content: chatArt.content,
                    producer: assistantName,
                    saved: Boolean(chatArt.saved_path),
                });
            }
        }
        return list;
    }, [artifacts, missionChat, assistantName]);

    return (
        <div className="h-full overflow-y-auto bg-cortex-bg px-6 py-6" data-testid="durable-files-operator-view">
            <div className="max-w-5xl mx-auto space-y-6">
                <header className="flex flex-wrap items-start justify-between gap-4 border-b border-cortex-border pb-4">
                    <div>
                        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-wider text-cortex-primary">
                            <FolderOpen className="w-3.5 h-3.5" />
                            Durable Files &amp; Deliverables
                        </div>
                        <h1 className="mt-3 text-2xl font-bold text-cortex-text-main tracking-tight">
                            Durable Files &amp; Deliverables
                        </h1>
                        <p className="mt-1 text-sm text-cortex-text-muted">
                            Browse, preview, open, and download files and deliverables produced across your workspace.
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            type="button"
                            onClick={() => fetchArtifacts({ limit: 50 })}
                            className="inline-flex items-center gap-1.5 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-mono text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        >
                            <RefreshCw className={`w-3.5 h-3.5 ${isFetchingArtifacts ? "animate-spin" : ""}`} />
                            Refresh
                        </button>
                        <button
                            type="button"
                            onClick={toggleAdvancedMode}
                            data-testid="resources-switch-advanced"
                            className="inline-flex items-center gap-1.5 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-mono text-cortex-text-muted hover:border-cortex-primary/40 hover:text-cortex-primary transition-colors"
                        >
                            <Sliders className="w-3.5 h-3.5" />
                            Switch to Advanced Mode to configure AI Engines &amp; Connected MCP Tools
                        </button>
                    </div>
                </header>

                {openedNotice && (
                    <div
                        data-testid="durable-files-open-notice"
                        className="rounded-lg border border-cortex-info/30 bg-cortex-info/10 px-4 py-2.5 text-xs font-mono text-cortex-info flex items-center justify-between"
                    >
                        <span>Workspace path ready: {openedNotice}</span>
                        <button
                            type="button"
                            onClick={() => setOpenedNotice(null)}
                            className="text-cortex-text-muted hover:text-cortex-text-main"
                        >
                            Dismiss
                        </button>
                    </div>
                )}

                {isFetchingArtifacts && items.length === 0 ? (
                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface p-8 text-center text-xs font-mono text-cortex-text-muted">
                        Loading workspace files and deliverables...
                    </div>
                ) : items.length === 0 ? (
                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface p-8 text-center space-y-2">
                        <FileText className="w-8 h-8 text-cortex-primary mx-auto opacity-80" />
                        <h2 className="text-base font-semibold text-cortex-text-main">
                            No durable deliverables saved yet
                        </h2>
                        <p className="text-sm text-cortex-text-muted max-w-lg mx-auto">
                            Ask {assistantName} or a team in your workspace to produce a brief, report, code file, or document and it will appear here for one-click preview and download.
                        </p>
                    </div>
                ) : (
                    <div className="space-y-3">
                        {items.map((item) => {
                            const isPreviewOpen = previewId === item.id;
                            const Icon = item.contentType.includes("code") || item.filePath?.match(/\.(ts|tsx|py|go|json|yaml|md)$/i)
                                ? FileCode2
                                : FileText;
                            return (
                                <div
                                    key={item.id}
                                    data-testid={`durable-file-card-${item.id}`}
                                    className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-3"
                                >
                                    <div className="flex flex-wrap items-start justify-between gap-3">
                                        <div className="flex items-start gap-3 min-w-0">
                                            <div className="p-2 rounded-lg bg-cortex-primary/10 border border-cortex-primary/20 text-cortex-primary">
                                                <Icon className="w-4 h-4" />
                                            </div>
                                            <div className="min-w-0">
                                                <div className="flex items-center gap-2 flex-wrap">
                                                    <h3 className="text-sm font-semibold text-cortex-text-main truncate">
                                                        {item.title}
                                                    </h3>
                                                    <span className="px-2 py-0.5 rounded-full text-[10px] font-mono bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/25">
                                                        Produced by {item.producer}
                                                    </span>
                                                </div>
                                                {item.filePath && (
                                                    <p className="text-xs font-mono text-cortex-text-muted mt-0.5 truncate">
                                                        {item.filePath}
                                                    </p>
                                                )}
                                            </div>
                                        </div>

                                        <div className="flex items-center gap-2">
                                            <button
                                                type="button"
                                                data-testid={`preview-deliverable-${item.id}`}
                                                onClick={() => setPreviewId(isPreviewOpen ? null : item.id)}
                                                className="inline-flex items-center gap-1 rounded-md border border-cortex-border bg-cortex-bg px-2.5 py-1.5 text-xs font-mono text-cortex-text-main hover:border-cortex-primary/40 transition-colors"
                                            >
                                                {isPreviewOpen ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                                                Preview
                                            </button>
                                            <button
                                                type="button"
                                                data-testid={`open-deliverable-${item.id}`}
                                                onClick={() => {
                                                    setPreviewId(item.id);
                                                    setOpenedNotice(item.filePath || item.title);
                                                }}
                                                className="inline-flex items-center gap-1 rounded-md border border-cortex-border bg-cortex-bg px-2.5 py-1.5 text-xs font-mono text-cortex-text-main hover:border-cortex-primary/40 transition-colors"
                                            >
                                                <FolderOpen className="w-3.5 h-3.5" />
                                                Open
                                            </button>
                                            <button
                                                type="button"
                                                onClick={() => triggerDownload(item)}
                                                className="inline-flex items-center gap-1 rounded-md border border-cortex-primary/30 bg-cortex-primary/10 px-2.5 py-1.5 text-xs font-mono text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                                            >
                                                <Download className="w-3.5 h-3.5" />
                                                Download
                                            </button>
                                        </div>
                                    </div>

                                    {isPreviewOpen && (
                                        <div
                                            data-testid={`deliverable-preview-pane-${item.id}`}
                                            className="rounded-lg border border-cortex-border bg-cortex-bg p-3 text-xs font-mono text-cortex-text-main whitespace-pre-wrap max-h-60 overflow-y-auto"
                                        >
                                            {item.content || `Deliverable saved at ${item.filePath || item.title}`}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                )}

                <footer className="pt-2 border-t border-cortex-border flex justify-end">
                    <button
                        type="button"
                        onClick={toggleAdvancedMode}
                        data-testid="resources-switch-advanced-footer"
                        className="inline-flex items-center gap-1.5 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-mono text-cortex-text-muted hover:border-cortex-primary/40 hover:text-cortex-primary transition-colors"
                    >
                        <Sliders className="w-3.5 h-3.5" />
                        Switch to Advanced Mode to configure AI Engines &amp; Connected MCP Tools
                    </button>
                </footer>
            </div>
        </div>
    );
}
