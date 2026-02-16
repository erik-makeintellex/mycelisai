"use client";

import React, { useEffect, useMemo, useState } from "react";
import {
    Package,
    Code,
    FileText,
    Image,
    Music,
    Database,
    File,
    BarChart3,
    X,
    CheckCircle,
    XCircle,
    Loader2,
    Shield,
} from "lucide-react";
import {
    useCortexStore,
    type Artifact,
    type ArtifactType,
    type ArtifactStatus,
} from "@/store/useCortexStore";
import { ChartRenderer, type MycelisChartSpec } from "@/components/charts";

interface ArtifactViewerProps {
    missionId: string;
}

type FilterTypeOption = "all" | ArtifactType;
type FilterStatusOption = "all" | ArtifactStatus;

function artifactIcon(type: ArtifactType) {
    switch (type) {
        case "code":
            return Code;
        case "document":
            return FileText;
        case "image":
            return Image;
        case "audio":
            return Music;
        case "data":
            return Database;
        case "chart":
            return BarChart3;
        case "file":
        default:
            return File;
    }
}

function statusBadgeClasses(status: ArtifactStatus): string {
    switch (status) {
        case "pending":
            return "bg-cortex-warning/15 text-cortex-warning border border-cortex-warning/30";
        case "approved":
            return "bg-cortex-success/15 text-cortex-success border border-cortex-success/30";
        case "rejected":
            return "bg-cortex-danger/15 text-cortex-danger border border-cortex-danger/30";
        case "archived":
            return "bg-cortex-text-muted/15 text-cortex-text-muted border border-cortex-text-muted/30";
        default:
            return "bg-cortex-border text-cortex-text-muted";
    }
}

export default function ArtifactViewer({ missionId }: ArtifactViewerProps) {
    const artifacts = useCortexStore((s) => s.artifacts);
    const isFetchingArtifacts = useCortexStore((s) => s.isFetchingArtifacts);
    const selectedArtifactDetail = useCortexStore((s) => s.selectedArtifactDetail);
    const fetchArtifacts = useCortexStore((s) => s.fetchArtifacts);
    const getArtifactDetail = useCortexStore((s) => s.getArtifactDetail);
    const updateArtifactStatus = useCortexStore((s) => s.updateArtifactStatus);

    const [typeFilter, setTypeFilter] = useState<FilterTypeOption>("all");
    const [statusFilter, setStatusFilter] = useState<FilterStatusOption>("all");

    useEffect(() => {
        fetchArtifacts({ mission_id: missionId });
    }, [missionId, fetchArtifacts]);

    const filtered = useMemo(() => {
        return artifacts.filter((a) => {
            if (typeFilter !== "all" && a.artifact_type !== typeFilter) return false;
            if (statusFilter !== "all" && a.status !== statusFilter) return false;
            return true;
        });
    }, [artifacts, typeFilter, statusFilter]);

    const handleCardClick = (artifact: Artifact) => {
        getArtifactDetail(artifact.id);
    };

    const handleClose = () => {
        useCortexStore.setState({ selectedArtifactDetail: null });
    };

    const typeOptions: FilterTypeOption[] = [
        "all",
        "code",
        "document",
        "image",
        "audio",
        "data",
        "file",
        "chart",
    ];

    const statusOptions: FilterStatusOption[] = [
        "all",
        "pending",
        "approved",
        "rejected",
        "archived",
    ];

    return (
        <div className="h-full flex flex-col bg-cortex-bg relative">
            {/* Header */}
            <div className="p-3 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between flex-shrink-0">
                <h3 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest flex items-center gap-2">
                    <Package className="w-3.5 h-3.5" />
                    Artifacts
                </h3>
                <span className="text-[10px] font-mono text-cortex-text-muted/60">
                    {filtered.length} of {artifacts.length}
                </span>
            </div>

            {/* Filters */}
            <div className="px-4 py-2 border-b border-cortex-border/50 flex items-center gap-3 flex-shrink-0 overflow-x-auto">
                {/* Type filters */}
                <div className="flex items-center gap-1">
                    {typeOptions.map((opt) => (
                        <button
                            key={opt}
                            onClick={() => setTypeFilter(opt)}
                            className={`px-2 py-0.5 rounded text-[9px] font-mono uppercase transition-all ${
                                typeFilter === opt
                                    ? "bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30"
                                    : "text-cortex-text-muted/60 hover:text-cortex-text-muted hover:bg-cortex-border/30 border border-transparent"
                            }`}
                        >
                            {opt}
                        </button>
                    ))}
                </div>

                <div className="h-3 w-px bg-cortex-border flex-shrink-0" />

                {/* Status filters */}
                <div className="flex items-center gap-1">
                    {statusOptions.map((opt) => (
                        <button
                            key={opt}
                            onClick={() => setStatusFilter(opt)}
                            className={`px-2 py-0.5 rounded text-[9px] font-mono uppercase transition-all ${
                                statusFilter === opt
                                    ? "bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30"
                                    : "text-cortex-text-muted/60 hover:text-cortex-text-muted hover:bg-cortex-border/30 border border-transparent"
                            }`}
                        >
                            {opt}
                        </button>
                    ))}
                </div>
            </div>

            {/* Artifact grid */}
            <div className="flex-1 overflow-y-auto p-4">
                {isFetchingArtifacts ? (
                    <div className="flex flex-col items-center justify-center h-full">
                        <Loader2 className="w-6 h-6 text-cortex-primary animate-spin mb-2" />
                        <p className="font-mono text-xs text-cortex-text-muted/60">
                            Loading artifacts...
                        </p>
                    </div>
                ) : filtered.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full">
                        <Package className="w-10 h-10 opacity-20 text-cortex-text-muted mb-3" />
                        <p className="font-mono text-xs text-cortex-text-muted/60">
                            No artifacts found
                        </p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {filtered.map((artifact) => {
                            const Icon = artifactIcon(artifact.artifact_type);
                            return (
                                <button
                                    key={artifact.id}
                                    onClick={() => handleCardClick(artifact)}
                                    className="bg-cortex-surface border border-cortex-border rounded-xl p-4 text-left hover:border-cortex-primary/40 transition-all group"
                                >
                                    <div className="flex items-start justify-between mb-2">
                                        <div className="p-1.5 rounded bg-cortex-bg group-hover:bg-cortex-primary/10 transition-colors">
                                            <Icon className="w-4 h-4 text-cortex-text-muted group-hover:text-cortex-primary transition-colors" />
                                        </div>
                                        <span
                                            className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded ${statusBadgeClasses(
                                                artifact.status
                                            )}`}
                                        >
                                            {artifact.status}
                                        </span>
                                    </div>

                                    <h4 className="text-sm font-mono font-bold text-cortex-text-main truncate mb-1">
                                        {artifact.title}
                                    </h4>

                                    <div className="flex items-center gap-2 mb-2">
                                        <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-bg text-cortex-text-muted border border-cortex-border">
                                            {artifact.agent_id}
                                        </span>
                                    </div>

                                    {/* Trust score bar */}
                                    {artifact.trust_score != null && (
                                        <div className="mt-2">
                                            <div className="flex items-center justify-between mb-1">
                                                <span className="text-[9px] font-mono text-cortex-text-muted/60">
                                                    Trust
                                                </span>
                                                <span className="text-[9px] font-mono text-cortex-text-muted">
                                                    {(artifact.trust_score * 100).toFixed(0)}%
                                                </span>
                                            </div>
                                            <div className="h-1 bg-cortex-bg rounded-full overflow-hidden">
                                                <div
                                                    className={`h-full rounded-full transition-all ${
                                                        artifact.trust_score >= 0.7
                                                            ? "bg-cortex-success"
                                                            : artifact.trust_score >= 0.4
                                                            ? "bg-cortex-warning"
                                                            : "bg-cortex-danger"
                                                    }`}
                                                    style={{
                                                        width: `${artifact.trust_score * 100}%`,
                                                    }}
                                                />
                                            </div>
                                        </div>
                                    )}

                                    {/* Chart thumbnail preview */}
                                    {artifact.artifact_type === "chart" &&
                                        artifact.content &&
                                        (() => {
                                            try {
                                                const spec = JSON.parse(
                                                    artifact.content,
                                                ) as MycelisChartSpec;
                                                if (
                                                    spec.chart_type &&
                                                    spec.data
                                                ) {
                                                    return (
                                                        <div className="mt-2 h-24 overflow-hidden rounded bg-cortex-bg pointer-events-none">
                                                            <ChartRenderer
                                                                spec={spec}
                                                                compact={true}
                                                            />
                                                        </div>
                                                    );
                                                }
                                            } catch {
                                                /* ignore parse errors */
                                            }
                                            return null;
                                        })()}
                                </button>
                            );
                        })}
                    </div>
                )}
            </div>

            {/* Detail slide-over */}
            {selectedArtifactDetail && (
                <ArtifactDetailPanel
                    artifact={selectedArtifactDetail}
                    onClose={handleClose}
                    onApprove={() =>
                        updateArtifactStatus(selectedArtifactDetail.id, "approved")
                    }
                    onReject={() =>
                        updateArtifactStatus(selectedArtifactDetail.id, "rejected")
                    }
                />
            )}
        </div>
    );
}

// ── Detail Slide-Over Panel ──────────────────────────────────

function ArtifactDetailPanel({
    artifact,
    onClose,
    onApprove,
    onReject,
}: {
    artifact: Artifact;
    onClose: () => void;
    onApprove: () => void;
    onReject: () => void;
}) {
    const Icon = artifactIcon(artifact.artifact_type);

    const formatDate = (ts: string) => {
        try {
            return new Date(ts).toLocaleString([], {
                dateStyle: "medium",
                timeStyle: "short",
            });
        } catch {
            return ts;
        }
    };

    const renderContent = () => {
        const ct = artifact.content_type?.toLowerCase() ?? "";
        const content = artifact.content;

        if (!content) {
            return (
                <div className="text-xs text-cortex-text-muted/60 font-mono italic p-3">
                    No content available
                </div>
            );
        }

        // Code
        if (
            artifact.artifact_type === "code" ||
            ct.includes("code") ||
            ct.includes("javascript") ||
            ct.includes("typescript") ||
            ct.includes("python") ||
            ct.includes("go")
        ) {
            return (
                <pre className="font-mono text-xs bg-cortex-bg p-3 rounded overflow-auto text-cortex-text-main max-h-64">
                    {content}
                </pre>
            );
        }

        // Document / text
        if (
            artifact.artifact_type === "document" ||
            ct.includes("text") ||
            ct.includes("markdown")
        ) {
            return (
                <div className="text-xs text-cortex-text-main whitespace-pre-wrap p-3 leading-relaxed">
                    {content}
                </div>
            );
        }

        // Image
        if (artifact.artifact_type === "image" || ct.includes("image")) {
            return (
                <div className="flex items-center justify-center p-6 text-cortex-text-muted/60">
                    <Image className="w-6 h-6 mr-2 opacity-40" />
                    <span className="font-mono text-xs">Image preview not available</span>
                </div>
            );
        }

        // Chart — interactive visualization
        if (
            artifact.artifact_type === "chart" ||
            ct.includes("vnd.mycelis.chart")
        ) {
            try {
                const spec = JSON.parse(content) as MycelisChartSpec;
                if (spec.chart_type && spec.data) {
                    return (
                        <div className="p-3">
                            <ChartRenderer spec={spec} />
                        </div>
                    );
                }
            } catch {
                // Fall through to raw JSON display
            }
        }

        // Data / JSON
        if (
            artifact.artifact_type === "data" ||
            ct.includes("json")
        ) {
            let formatted = content;
            try {
                formatted = JSON.stringify(JSON.parse(content), null, 2);
            } catch {
                // keep raw
            }
            return (
                <pre className="font-mono text-xs bg-cortex-bg p-3 rounded overflow-auto text-cortex-text-main max-h-64">
                    {formatted}
                </pre>
            );
        }

        // Fallback
        return (
            <div className="text-xs text-cortex-text-muted font-mono p-3">
                Content type: {artifact.content_type}
            </div>
        );
    };

    return (
        <div className={`absolute right-0 top-0 bottom-0 z-30 bg-cortex-surface border-l border-cortex-border shadow-2xl flex flex-col ${artifact.artifact_type === "chart" ? "w-[640px]" : "w-96"}`}>
            {/* Detail header */}
            <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-4 flex-shrink-0">
                <div className="flex items-center gap-2">
                    <Icon className="w-4 h-4 text-cortex-primary" />
                    <span className="text-xs font-mono font-bold text-cortex-text-main truncate max-w-[200px]">
                        {artifact.title}
                    </span>
                </div>
                <button
                    onClick={onClose}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Detail body */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {/* Metadata grid */}
                <div className="grid grid-cols-2 gap-3">
                    <MetaField label="Type" value={artifact.artifact_type} />
                    <MetaField label="Agent" value={artifact.agent_id} />
                    <MetaField label="Content Type" value={artifact.content_type} />
                    <MetaField
                        label="Trust Score"
                        value={
                            artifact.trust_score != null
                                ? `${(artifact.trust_score * 100).toFixed(0)}%`
                                : "N/A"
                        }
                    />
                    <MetaField label="Status" value={artifact.status} />
                    <MetaField
                        label="Created"
                        value={formatDate(artifact.created_at)}
                    />
                </div>

                {artifact.trace_id && (
                    <div className="text-[10px] font-mono text-cortex-text-muted/50 break-all">
                        Trace: {artifact.trace_id}
                    </div>
                )}

                {/* Content */}
                <div>
                    <h4 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-widest mb-2">
                        Content
                    </h4>
                    <div className="bg-cortex-bg border border-cortex-border rounded-lg overflow-hidden">
                        {renderContent()}
                    </div>
                </div>
            </div>

            {/* Governance buttons (for pending artifacts) */}
            {artifact.status === "pending" && (
                <div className="p-4 border-t border-cortex-border flex items-center gap-2 flex-shrink-0">
                    <button
                        onClick={onApprove}
                        className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-cortex-success/15 border border-cortex-success/30 text-cortex-success text-xs font-mono hover:bg-cortex-success/25 transition-all"
                    >
                        <CheckCircle className="w-3.5 h-3.5" />
                        Approve
                    </button>
                    <button
                        onClick={onReject}
                        className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-cortex-danger/15 border border-cortex-danger/30 text-cortex-danger text-xs font-mono hover:bg-cortex-danger/25 transition-all"
                    >
                        <XCircle className="w-3.5 h-3.5" />
                        Reject
                    </button>
                </div>
            )}
        </div>
    );
}

// ── Helper: Metadata field ───────────────────────────────────

function MetaField({ label, value }: { label: string; value: string }) {
    return (
        <div>
            <p className="text-[9px] font-mono text-cortex-text-muted/60 uppercase tracking-wider mb-0.5">
                {label}
            </p>
            <p className="text-xs font-mono text-cortex-text-main truncate">
                {value}
            </p>
        </div>
    );
}
