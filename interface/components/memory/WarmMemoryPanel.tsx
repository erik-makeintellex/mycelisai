"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Database, FileText, Package, ChevronDown, ChevronRight, Search } from "lucide-react";
import { useCortexStore, type Artifact } from "@/store/useCortexStore";

// ── Types ────────────────────────────────────────────────────

interface SitRep {
    id: string;
    mission_id: string;
    summary: string;
    raw_count: number;
    created_at: string;
}

type TabId = "sitreps" | "artifacts";

// ── Artifact Type Colors ─────────────────────────────────────

function artifactTypeBadge(type: string): string {
    switch (type) {
        case "code":
            return "bg-cortex-primary/20 text-cortex-primary";
        case "document":
            return "bg-cortex-info/20 text-cortex-info";
        case "image":
            return "bg-cortex-warning/20 text-cortex-warning";
        case "data":
        case "chart":
            return "bg-cortex-success/20 text-cortex-success";
        default:
            return "bg-cortex-text-muted/20 text-cortex-text-muted";
    }
}

function statusBadge(status: string): string {
    switch (status) {
        case "approved":
            return "bg-cortex-success/20 text-cortex-success";
        case "rejected":
            return "bg-cortex-danger/20 text-cortex-danger";
        case "pending":
            return "bg-cortex-warning/20 text-cortex-warning";
        case "archived":
            return "bg-cortex-text-muted/20 text-cortex-text-muted";
        default:
            return "bg-cortex-text-muted/20 text-cortex-text-muted";
    }
}

// ── Timestamp Formatting ─────────────────────────────────────

function formatTimestamp(ts: string): string {
    try {
        const d = new Date(ts);
        if (isNaN(d.getTime())) return ts;
        return d.toLocaleDateString("en-GB", {
            day: "2-digit",
            month: "short",
            hour: "2-digit",
            minute: "2-digit",
        });
    } catch {
        return ts;
    }
}

// ── SitRep Card ──────────────────────────────────────────────

function SitRepCard({
    sitrep,
    onSearchRelated,
}: {
    sitrep: SitRep;
    onSearchRelated?: (query: string) => void;
}) {
    const [expanded, setExpanded] = useState(false);
    const summary = sitrep.summary ?? "";
    const needsTruncation = summary.length > 120;
    const displayText = expanded || !needsTruncation ? summary : summary.slice(0, 120) + "...";

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4 space-y-2">
            {/* Summary */}
            <button
                onClick={() => setExpanded(!expanded)}
                className="text-left w-full flex items-start gap-2"
            >
                {needsTruncation && (
                    expanded ? (
                        <ChevronDown className="w-3 h-3 text-cortex-text-muted flex-shrink-0 mt-0.5" />
                    ) : (
                        <ChevronRight className="w-3 h-3 text-cortex-text-muted flex-shrink-0 mt-0.5" />
                    )
                )}
                <span className="text-xs text-cortex-text-main leading-relaxed">
                    {displayText}
                </span>
            </button>

            {/* Meta row */}
            <div className="flex items-center gap-2 flex-wrap">
                <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary">
                    {sitrep.mission_id}
                </span>
                <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-info/15 text-cortex-info">
                    {sitrep.raw_count} raw
                </span>
                <span className="text-[9px] font-mono text-cortex-text-muted ml-auto">
                    {formatTimestamp(sitrep.created_at)}
                </span>
            </div>

            {/* Find Related */}
            {onSearchRelated && (
                <button
                    onClick={() => onSearchRelated(summary.slice(0, 100))}
                    className="flex items-center gap-1 text-[10px] font-mono text-cortex-info hover:text-cortex-primary transition-colors"
                >
                    <Search className="w-3 h-3" />
                    Find Related
                </button>
            )}
        </div>
    );
}

// ── Artifact Row ─────────────────────────────────────────────

function ArtifactRow({ artifact }: { artifact: Artifact }) {
    return (
        <div className="flex items-center gap-2 py-2 px-3 border-b border-cortex-border/30 hover:bg-cortex-surface/40 transition-colors">
            {/* Type badge */}
            <span
                className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded flex-shrink-0 ${artifactTypeBadge(artifact.artifact_type)}`}
            >
                {artifact.artifact_type}
            </span>

            {/* Title */}
            <span className="text-[11px] text-cortex-text-main truncate flex-1">
                {artifact.title}
            </span>

            {/* Status */}
            <span
                className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded flex-shrink-0 ${statusBadge(artifact.status)}`}
            >
                {artifact.status}
            </span>

            {/* Agent ID */}
            <span className="text-[9px] font-mono text-cortex-text-muted truncate max-w-[80px] flex-shrink-0">
                {artifact.agent_id}
            </span>
        </div>
    );
}

// ── WarmMemoryPanel ──────────────────────────────────────────

interface WarmMemoryPanelProps {
    onSearchRelated?: (query: string) => void;
}

export default function WarmMemoryPanel({ onSearchRelated }: WarmMemoryPanelProps) {
    const [activeTab, setActiveTab] = useState<TabId>("sitreps");
    const [sitreps, setSitreps] = useState<SitRep[]>([]);
    const [loading, setLoading] = useState(false);

    const artifacts = useCortexStore((s) => s.artifacts);
    const isFetchingArtifacts = useCortexStore((s) => s.isFetchingArtifacts);
    const fetchArtifacts = useCortexStore((s) => s.fetchArtifacts);

    // Fetch sitreps when tab activates
    const loadSitreps = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetch("/api/v1/memory/sitreps");
            if (res.ok) {
                const data = await res.json();
                setSitreps(Array.isArray(data) ? data : data.sitreps ?? []);
            } else {
                setSitreps([]);
            }
        } catch {
            setSitreps([]);
        }
        setLoading(false);
    }, []);

    useEffect(() => {
        if (activeTab === "sitreps") {
            loadSitreps();
        } else if (activeTab === "artifacts") {
            fetchArtifacts({ limit: 50 });
        }
    }, [activeTab, loadSitreps, fetchArtifacts]);

    const isLoading = activeTab === "sitreps" ? loading : isFetchingArtifacts;

    return (
        <div className="h-full flex flex-col">
            {/* Sub-header */}
            <div className="h-10 flex items-center px-3 border-b border-cortex-border bg-cortex-surface/50 flex-shrink-0">
                <div className="flex items-center gap-2 mr-4">
                    <Database className="w-3.5 h-3.5 text-cortex-warning" />
                    <span className="text-[10px] font-bold uppercase tracking-widest text-cortex-text-muted">
                        Warm
                    </span>
                </div>

                {/* Tab buttons */}
                <div className="flex items-center gap-1">
                    <button
                        onClick={() => setActiveTab("sitreps")}
                        className={`text-[9px] font-mono uppercase px-2 py-1 rounded transition-colors flex items-center gap-1 ${
                            activeTab === "sitreps"
                                ? "bg-cortex-warning/20 text-cortex-warning"
                                : "text-cortex-text-muted hover:text-cortex-text-main"
                        }`}
                    >
                        <FileText className="w-3 h-3" />
                        SitReps
                    </button>
                    <button
                        onClick={() => setActiveTab("artifacts")}
                        className={`text-[9px] font-mono uppercase px-2 py-1 rounded transition-colors flex items-center gap-1 ${
                            activeTab === "artifacts"
                                ? "bg-cortex-warning/20 text-cortex-warning"
                                : "text-cortex-text-muted hover:text-cortex-text-main"
                        }`}
                    >
                        <Package className="w-3 h-3" />
                        Artifacts
                    </button>
                </div>
            </div>

            {/* Tab content */}
            <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border min-h-0">
                {isLoading ? (
                    <div className="flex items-center justify-center h-full">
                        <span className="text-[10px] font-mono text-cortex-text-muted animate-pulse">
                            Loading...
                        </span>
                    </div>
                ) : activeTab === "sitreps" ? (
                    sitreps.length === 0 ? (
                        <div className="flex flex-col items-center justify-center h-full gap-2">
                            <FileText className="w-8 h-8 text-cortex-text-muted opacity-20" />
                            <span className="text-[10px] font-mono text-cortex-text-muted">
                                No sitreps recorded yet.
                            </span>
                        </div>
                    ) : (
                        <div className="p-3 space-y-2">
                            {sitreps.map((sitrep) => (
                                <SitRepCard
                                    key={sitrep.id}
                                    sitrep={sitrep}
                                    onSearchRelated={onSearchRelated}
                                />
                            ))}
                        </div>
                    )
                ) : artifacts.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full gap-2">
                        <Package className="w-8 h-8 text-cortex-text-muted opacity-20" />
                        <span className="text-[10px] font-mono text-cortex-text-muted">
                            No artifacts stored.
                        </span>
                    </div>
                ) : (
                    <div>
                        {artifacts.map((artifact) => (
                            <ArtifactRow key={artifact.id} artifact={artifact} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
