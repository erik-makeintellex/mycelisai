"use client";

import { BarChart3, ExternalLink, Image as ImageIcon, Music, Video } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";
import { ChartRenderer, type MycelisChartSpec } from "@/components/charts";
import type { ChatArtifactRef } from "@/store/useCortexStore";
import {
    artifactDownloadHref,
    artifactIcon,
    binaryArtifactLabel,
} from "./missionControlChatHelpers";
import ChatCopyButton from "./ChatCopyButton";

function mediaSource(artifact: ChatArtifactRef): string | null {
    if (artifact.url?.trim()) return artifact.url.trim();
    if (!artifact.content?.trim() || !artifact.content_type?.trim()) return null;
    const contentType = artifact.content_type.trim();
    if (!/^(audio|image|video)\//i.test(contentType)) return null;
    return `data:${contentType};base64,${artifact.content.trim()}`;
}

function ArtifactHeader({
    artifact,
    icon,
    children,
}: {
    artifact: ChatArtifactRef;
    icon: React.ReactNode;
    children?: ReactNode;
}) {
    return (
        <div className="flex items-center gap-2 border-b border-cortex-border bg-cortex-surface/50 px-3 py-1.5">
            {icon}
            <span className="flex-1 truncate text-[10px] font-mono font-bold text-cortex-text-main">
                {artifact.title}
            </span>
            {children}
        </div>
    );
}

function SaveBadge({ path, href }: { path: string; href: string | null }) {
    return (
        <span className="text-cortex-success">
            Saved to:{" "}
            {href ? (
                <a href={href} download className="underline underline-offset-2 hover:text-cortex-success/80">
                    {path}
                </a>
            ) : path}
        </span>
    );
}

function MediaArtifact({ artifact }: { artifact: ChatArtifactRef }) {
    const src = mediaSource(artifact);
    const type = artifact.content_type?.toLowerCase() ?? "";
    const isVideo = artifact.type === "video" || type.startsWith("video/");
    const Icon = isVideo ? Video : Music;

    return (
        <div className="mt-2 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg">
            <ArtifactHeader artifact={artifact} icon={<Icon className="h-3 w-3 text-cortex-primary" />} />
            <div className="p-3">
                {src ? (
                    isVideo ? (
                        <video controls className="max-h-80 w-full rounded bg-black" src={src}>
                            Your browser does not support video playback.
                        </video>
                    ) : (
                        <audio controls className="w-full" src={src}>
                            Your browser does not support audio playback.
                        </audio>
                    )
                ) : (
                    <DownloadOnly artifact={artifact} label={`${isVideo ? "Video" : "Audio"} artifact`} />
                )}
            </div>
        </div>
    );
}

function DownloadOnly({ artifact, label }: { artifact: ChatArtifactRef; label: string }) {
    const href = artifactDownloadHref(artifact);
    return (
        <div className="rounded border border-cortex-border bg-cortex-surface/40 px-3 py-2 text-[10px] font-mono">
            {artifact.saved_path && href ? (
                <span className="text-cortex-success">Saved object: <a href={href} download target="_blank" rel="noopener noreferrer" className="underline underline-offset-2 hover:text-cortex-success/80">{artifact.saved_path}</a></span>
            ) : href ? (
                <a href={href} download target="_blank" rel="noopener noreferrer" className="text-cortex-primary underline underline-offset-2">Download {binaryArtifactLabel(artifact)}</a>
            ) : (
                <span className="text-cortex-text-muted">{label} has no playable source or download link.</span>
            )}
        </div>
    );
}

export default function InlineArtifact({ artifact }: { artifact: ChatArtifactRef }) {
    const [expanded, setExpanded] = useState(false);
    const [saving, setSaving] = useState(false);
    const [savedPath, setSavedPath] = useState(artifact.saved_path || "");
    const [saveError, setSaveError] = useState("");
    const Icon = artifactIcon(artifact.type);
    const downloadHref = artifactDownloadHref(artifact);
    const contentType = artifact.content_type?.toLowerCase() ?? "";

    const handleSave = async () => {
        if (!artifact.id || saving) return;
        setSaving(true);
        setSaveError("");
        try {
            const res = await fetch(`/api/v1/artifacts/${encodeURIComponent(artifact.id)}/save`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ folder: "saved-media" }),
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);
            setSavedPath(data?.file_path || "saved-media");
        } catch (err) {
            setSaveError(err instanceof Error ? err.message : "save failed");
        } finally {
            setSaving(false);
        }
    };

    if (artifact.type === "chart" && artifact.content) {
        try {
            const spec = JSON.parse(artifact.content) as MycelisChartSpec;
            if (spec.chart_type && spec.data) {
                return (
                    <div className="mt-2 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg">
                        <ArtifactHeader artifact={artifact} icon={<BarChart3 className="h-3 w-3 text-cortex-primary" />} />
                        <div className="max-h-[280px] p-2"><ChartRenderer spec={spec} /></div>
                    </div>
                );
            }
        } catch {
            /* Render as a generic artifact below. */
        }
    }

    if (artifact.type === "image" || contentType.startsWith("image/")) {
        const src = mediaSource(artifact);
        const cached = artifact.cached || !!artifact.expires_at;
        return (
            <div className="mt-2 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg">
                <ArtifactHeader artifact={artifact} icon={<ImageIcon className="h-3 w-3 text-cortex-primary" />}>
                    {cached && !savedPath ? <span className="rounded border border-cortex-warning/30 bg-cortex-warning/10 px-1.5 py-0.5 text-[8px] font-mono text-cortex-warning">cached</span> : null}
                    {savedPath ? <span className="rounded border border-cortex-success/30 bg-cortex-success/10 px-1.5 py-0.5 text-[8px] font-mono text-cortex-success">saved</span> : null}
                    {artifact.id && !savedPath ? <button onClick={handleSave} disabled={saving} className="rounded border border-cortex-primary/30 bg-cortex-primary/10 px-1.5 py-0.5 text-[8px] font-mono text-cortex-primary hover:bg-cortex-primary/20 disabled:opacity-60" title="Save image to workspace/saved-media">{saving ? "Saving..." : "Save"}</button> : null}
                    {artifact.url ? <a href={artifact.url} target="_blank" rel="noopener noreferrer" className="rounded p-0.5 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-primary"><ExternalLink className="h-3 w-3" /></a> : null}
                </ArtifactHeader>
                {src ? <img src={src} alt={artifact.title} className="max-h-80 w-full bg-cortex-bg object-contain p-2" /> : <div className="flex h-32 items-center justify-center text-cortex-text-muted/40"><ImageIcon className="h-8 w-8" /></div>}
                {(savedPath || saveError) ? <div className="border-t border-cortex-border/60 px-3 py-1.5 text-[9px] font-mono">{savedPath ? <SaveBadge path={savedPath} href={downloadHref} /> : <span className="text-cortex-danger">Save failed: {saveError}</span>}</div> : null}
            </div>
        );
    }

    if (artifact.type === "audio" || artifact.type === "video" || contentType.startsWith("audio/") || contentType.startsWith("video/")) {
        return <MediaArtifact artifact={artifact} />;
    }

    if (artifact.type === "code" && artifact.content) {
        return (
            <div className="mt-2 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg">
                <ArtifactHeader artifact={artifact} icon={<Icon className="h-3 w-3 text-cortex-primary" />}>
                    {artifact.content_type ? <span className="rounded bg-cortex-bg px-1.5 py-0.5 text-[8px] font-mono text-cortex-text-muted">{artifact.content_type.replace("text/", "").replace("application/", "")}</span> : null}
                    <ChatCopyButton text={artifact.content} />
                </ArtifactHeader>
                <pre className="max-h-64 overflow-auto p-3 text-[11px] font-mono leading-relaxed text-cortex-text-main">{artifact.content}</pre>
            </div>
        );
    }

    if ((artifact.type === "data" || artifact.type === "document") && artifact.content) {
        let displayContent = artifact.content;
        if (artifact.content_type?.includes("json")) {
            try { displayContent = JSON.stringify(JSON.parse(artifact.content), null, 2); } catch { /* keep raw */ }
        }
        return (
            <div className="mt-2 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg">
                <ArtifactHeader artifact={artifact} icon={<Icon className="h-3 w-3 text-cortex-primary" />}>
                    <ChatCopyButton text={displayContent} />
                    <button onClick={() => setExpanded(!expanded)} className="text-[8px] font-mono text-cortex-primary hover:text-cortex-primary/80">{expanded ? "collapse" : "expand"}</button>
                </ArtifactHeader>
                <pre className={`overflow-auto p-3 text-[11px] font-mono leading-relaxed text-cortex-text-main ${expanded ? "max-h-96" : "max-h-32"}`}>{displayContent}</pre>
            </div>
        );
    }

    return (
        <div className="mt-2 rounded-lg border border-cortex-border bg-cortex-bg">
            <div className="flex items-center gap-2 px-3 py-2">
                <Icon className="h-3.5 w-3.5 text-cortex-primary" />
                <span className="flex-1 truncate text-[10px] font-mono font-bold text-cortex-text-main">{artifact.title}</span>
                <span className="text-[8px] font-mono uppercase text-cortex-text-muted">{artifact.type}</span>
                {downloadHref ? <a href={downloadHref} download target="_blank" rel="noopener noreferrer" className="rounded p-0.5 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-primary" title={`Download ${binaryArtifactLabel(artifact)}`}><ExternalLink className="h-3 w-3" /></a> : null}
            </div>
            {(artifact.saved_path || downloadHref) ? (
                <div className="border-t border-cortex-border/60 px-3 py-2 text-[9px] font-mono">
                    {artifact.saved_path ? <SaveBadge path={artifact.saved_path} href={downloadHref} /> : <DownloadOnly artifact={artifact} label="Artifact" />}
                </div>
            ) : null}
        </div>
    );
}
