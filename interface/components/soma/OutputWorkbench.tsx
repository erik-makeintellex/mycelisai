"use client";

import { useState } from "react";
import Link from "next/link";
import { CheckCircle2, ClipboardList, Download, ExternalLink, Eye, FileText, Play, ShieldCheck } from "lucide-react";
import type { ChatArtifactRef, DeliveryPackageSummary, OutputProofEnvelope, OutputRef } from "@/store/useCortexStore";
import {
    artifactForPath,
    isPlayablePackageFile,
    isProofFile,
    packageTitle,
    proofDetailRows,
    proofStageState,
    readbackStyle,
    resolvePackageFiles,
    STAGE_STYLES,
} from "./outputWorkbenchHelpers";

function ProofChip({ proof }: { proof?: OutputProofEnvelope }) {
    const { icon: Icon, tone, label } = readbackStyle(proof?.readback_status);
    return (
        <span
            className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[8px] font-mono font-bold uppercase ${tone}`}
            title={proof?.recovery_hint || `readback: ${label}`}
        >
            <Icon className="h-2.5 w-2.5" />
            {label}
        </span>
    );
}

function ProofDetails({ proof, openUrl }: { proof: OutputProofEnvelope; openUrl?: string }) {
    return (
        <div className="mt-1.5 space-y-1 rounded border border-cortex-border/60 bg-cortex-bg px-2 py-1.5">
            {proofDetailRows(proof).map(([key, value]) => (
                <div key={key} className="flex items-baseline justify-between gap-2 text-[9px] font-mono">
                    <span className="text-cortex-text-muted">{key}</span>
                    <span
                        className="max-w-[60%] truncate text-cortex-text-main"
                        title={key === "Checksum" ? proof.checksum : value}
                    >
                        {value}
                    </span>
                </div>
            ))}
            {proof.recovery_hint ? (
                <div className="rounded border border-cortex-warning/30 bg-cortex-warning/10 px-1.5 py-1 text-[9px] font-mono text-cortex-warning">
                    {proof.recovery_hint}
                </div>
            ) : null}
            {openUrl ? (
                <a
                    href={openUrl}
                    className="inline-flex items-center gap-1 text-[9px] font-mono font-bold text-cortex-primary hover:underline"
                >
                    <Download className="h-2.5 w-2.5" />
                    Open / download
                </a>
            ) : null}
        </div>
    );
}

export default function OutputWorkbench({
    deliveryPackage,
    artifacts = [],
    outputRefs = [],
    runId,
}: {
    deliveryPackage?: DeliveryPackageSummary;
    artifacts?: ChatArtifactRef[];
    outputRefs?: OutputRef[];
    runId?: string;
}) {
    const [activePreviewPath, setActivePreviewPath] = useState<string | null>(null);

    if (!deliveryPackage && artifacts.length === 0 && outputRefs.length === 0) return null;

    const refs = outputRefs.length > 0 ? outputRefs : deliveryPackage?.items ?? [];
    const files = resolvePackageFiles(deliveryPackage, refs, artifacts);
    const proof = files.find((file) => isProofFile(file.path, file.title));
    const packageRunId = runId || deliveryPackage?.run_id;
    const packageIdentity = deliveryPackage?.base_path || deliveryPackage?.package_id;
    const packageStatus = deliveryPackage?.status || "verified";
    const contentArtifacts = artifacts.filter((artifact) => artifact.content?.trim());
    const playableEntry = files.find((file) => {
        const artifact = artifactForPath(artifacts, file.path, file.title);
        return Boolean(artifact?.content) && isPlayablePackageFile(file, artifact);
    });
    const activeFile = activePreviewPath
        ? files.find((file) => (file.path || file.title) === activePreviewPath || file.path === activePreviewPath)
        : undefined;
    const activeArtifact = activeFile ? artifactForPath(artifacts, activeFile.path, activeFile.title) : undefined;
    const stages: { label: string; state: "completed" | "active" | "failed" }[] = [
        { label: "Intent", state: "completed" },
        { label: "Approval", state: "completed" },
        { label: "Execution", state: "completed" },
        { label: "Package", state: "completed" },
        { label: "Proof", state: proofStageState(packageStatus) },
    ];

    return (
        <div className="mt-2 overflow-hidden rounded-lg border border-cortex-primary/25 bg-cortex-primary/5" data-testid="output-workbench">
            <div className="flex flex-wrap items-center gap-2 border-b border-cortex-primary/20 bg-cortex-bg/70 px-3 py-2">
                <ClipboardList className="h-3.5 w-3.5 text-cortex-primary" />
                <div className="min-w-0 flex-1">
                    <div className="truncate text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">
                        {packageTitle(deliveryPackage?.package_kind)}
                    </div>
                    {packageIdentity ? (
                        <div className="truncate text-[9px] font-mono text-cortex-text-muted">{packageIdentity}</div>
                    ) : null}
                </div>
                {playableEntry ? (
                    <button
                        type="button"
                        data-testid="output-workbench-play-inline"
                        onClick={() => setActivePreviewPath(activePreviewPath === playableEntry.path ? null : (playableEntry.path || playableEntry.title || null))}
                        className="inline-flex items-center gap-1 rounded border border-cortex-primary/40 bg-cortex-primary/15 px-2 py-0.5 text-[9px] font-mono font-bold text-cortex-primary hover:bg-cortex-primary/25"
                    >
                        <Play className="h-2.5 w-2.5" />
                        {activePreviewPath === playableEntry.path ? "Hide Preview" : "Play / Live Preview"}
                    </button>
                ) : null}
                <Link
                    href="/resources"
                    data-testid="output-workbench-open-resources"
                    className="inline-flex items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[9px] font-mono text-cortex-text-main hover:border-cortex-primary/40 hover:text-cortex-primary"
                >
                    Open in Deliverables
                </Link>
                <span className="inline-flex items-center gap-1 rounded border border-cortex-success/30 bg-cortex-success/10 px-1.5 py-0.5 text-[8px] font-mono font-bold uppercase text-cortex-success">
                    <CheckCircle2 className="h-2.5 w-2.5" />
                    {packageStatus}
                </span>
                {packageRunId ? (
                    <a
                        href={`/runs/${packageRunId}`}
                        className="rounded p-0.5 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-primary"
                        title="Open run"
                    >
                        <ExternalLink className="h-3 w-3" />
                    </a>
                ) : null}
            </div>

            <div
                className="flex flex-wrap items-center gap-1 border-b border-cortex-primary/15 bg-cortex-bg/40 px-3 py-1.5"
                data-testid="output-workbench-stage-strip"
            >
                {stages.map((stage) => (
                    <span
                        key={stage.label}
                        data-stage-state={stage.state}
                        className={`inline-flex items-center rounded border px-1.5 py-0.5 text-[8px] font-mono font-bold uppercase ${STAGE_STYLES[stage.state]}`}
                    >
                        {stage.label}
                    </span>
                ))}
            </div>

            {activeFile && activeArtifact?.content ? (
                <div className="border-b border-cortex-primary/20 bg-cortex-bg p-3" data-testid="output-workbench-inline-preview">
                    {isPlayablePackageFile(activeFile, activeArtifact) ? (
                        <iframe
                            title={activeFile.title || "Live Deliverable Preview"}
                            data-testid="output-workbench-inline-iframe"
                            sandbox="allow-scripts"
                            srcDoc={activeArtifact.content}
                            className="w-full h-[460px] rounded-lg border border-cortex-primary/30 bg-slate-950"
                        />
                    ) : (
                        <pre className="max-h-72 overflow-auto rounded-lg border border-cortex-border bg-cortex-surface p-3 text-xs font-mono text-cortex-text-main whitespace-pre-wrap">
                            {activeArtifact.content}
                        </pre>
                    )}
                </div>
            ) : null}

            <div className="divide-y divide-cortex-border/60">
                {files.map((file) => {
                    const artifact = artifactForPath(artifacts, file.path, file.title);
                    const ref = refs.find((candidate) => candidate.storage_ref === file.path || candidate.label === file.title);
                    const proofFile = isProofFile(file.path, file.title);
                    const role = ref?.role || (proofFile ? "proof" : "");
                    const openUrl = ref?.open_url ?? artifact?.url;
                    const playable = isPlayablePackageFile(file, artifact);
                    return (
                        <div key={`${file.path}-${file.title}`} className="px-3 py-2">
                            <div className="grid grid-cols-[auto_1fr_auto] items-center gap-2">
                                {proofFile ? (
                                    <ShieldCheck className="h-3.5 w-3.5 text-cortex-success" />
                                ) : (
                                    <FileText className="h-3.5 w-3.5 text-cortex-primary" />
                                )}
                                <div className="min-w-0">
                                    <div className="truncate text-[10px] font-mono font-bold text-cortex-text-main">
                                        {file.title || file.path}
                                    </div>
                                    <div className="truncate text-[9px] font-mono text-cortex-text-muted">
                                        {file.path}
                                    </div>
                                </div>
                                <div className="flex items-center gap-1.5">
                                    {role ? (
                                        <span className="text-[8px] font-mono text-cortex-text-muted">{role}</span>
                                    ) : null}
                                    {artifact?.content ? (
                                        <button
                                            type="button"
                                            onClick={() => setActivePreviewPath(activePreviewPath === file.path ? null : (file.path || file.title || null))}
                                            className="inline-flex items-center gap-1 rounded border border-cortex-primary/30 bg-cortex-primary/10 px-1.5 py-0.5 text-[8px] font-mono font-bold text-cortex-primary hover:bg-cortex-primary/20"
                                        >
                                            {playable ? <Play className="h-2.5 w-2.5" /> : <Eye className="h-2.5 w-2.5" />}
                                            {playable ? "Play" : "View"}
                                        </button>
                                    ) : null}
                                    <ProofChip proof={ref?.proof} />
                                    {openUrl ? (
                                        <a
                                            href={openUrl}
                                            className="rounded p-0.5 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-primary"
                                            title="Open / download"
                                        >
                                            <Download className="h-3 w-3" />
                                        </a>
                                    ) : null}
                                </div>
                            </div>
                            {ref?.proof ? (
                                <details className="mt-1">
                                    <summary className="cursor-pointer select-none text-[8px] font-mono font-bold uppercase tracking-wider text-cortex-primary">
                                        Inspect
                                    </summary>
                                    <ProofDetails proof={ref.proof} openUrl={openUrl} />
                                </details>
                            ) : null}
                        </div>
                    );
                })}
            </div>

            {proof?.sha256 ? (
                <div className="border-t border-cortex-primary/20 px-3 py-2 text-[9px] font-mono text-cortex-text-muted">
                    proof sha256 <span className="text-cortex-text-main">{proof.sha256.slice(0, 16)}</span>
                </div>
            ) : null}

            {contentArtifacts.length > 0 ? (
                <details className="border-t border-cortex-primary/20 px-3 py-2">
                    <summary className="cursor-pointer text-[9px] font-mono font-bold uppercase tracking-wider text-cortex-primary">
                        Inspect package files
                    </summary>
                    <div className="mt-2 space-y-2">
                        {contentArtifacts.map((artifact) => (
                            <pre key={artifact.saved_path || artifact.title} className="max-h-40 overflow-auto rounded border border-cortex-border bg-cortex-bg p-2 text-[10px] font-mono text-cortex-text-main">
                                {artifact.content}
                            </pre>
                        ))}
                    </div>
                </details>
            ) : null}
        </div>
    );
}
