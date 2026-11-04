"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Package, Loader2 } from "lucide-react";
import {
    useCortexStore,
    type Artifact,
    type ArtifactType,
    type ArtifactStatus,
} from "@/store/useCortexStore";
import { ChartRenderer, type MycelisChartSpec } from "@/components/charts";
import {
    ArtifactDetailPanel,
    artifactIcon,
    statusBadgeClasses,
} from "./ArtifactDetailPanel";

interface ArtifactViewerProps {
    missionId: string;
}

type FilterTypeOption = "all" | ArtifactType;
type FilterStatusOption = "all" | ArtifactStatus;

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
