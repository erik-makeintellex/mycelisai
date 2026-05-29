"use client";

import { CheckCircle2, ClipboardList, ExternalLink, FileText, ShieldCheck } from "lucide-react";
import type { ChatArtifactRef, DeliveryPackageFile, DeliveryPackageSummary } from "@/store/useCortexStore";

function isProofFile(path?: string, title?: string): boolean {
    const lowerPath = path?.toLowerCase() ?? "";
    const lowerTitle = title?.toLowerCase() ?? "";
    return lowerPath.endsWith("proof.md") || lowerTitle.includes("proof");
}

function packageTitle(packageKind?: string): string {
    switch (packageKind) {
        case "codex_task_bundle":
            return "Codex task bundle";
        default:
            return packageKind?.replace(/_/g, " ") ?? "Delivery package";
    }
}

function artifactForPath(artifacts: ChatArtifactRef[], path?: string, title?: string): ChatArtifactRef | undefined {
    if (!path && !title) return undefined;
    return artifacts.find((artifact) => (
        (path && artifact.saved_path === path)
        || (title && artifact.title === title)
    ));
}

export default function OutputWorkbench({
    deliveryPackage,
    artifacts = [],
    runId,
}: {
    deliveryPackage?: DeliveryPackageSummary;
    artifacts?: ChatArtifactRef[];
    runId?: string;
}) {
    if (!deliveryPackage && artifacts.length === 0) return null;

    const files: DeliveryPackageFile[] = deliveryPackage?.files ?? artifacts.map((artifact) => ({
        path: artifact.saved_path || artifact.title,
        title: artifact.title,
        content_type: artifact.content_type,
    }));
    const proof = files.find((file) => isProofFile(file.path, file.title));
    const packageRunId = runId || deliveryPackage?.run_id;
    const contentArtifacts = artifacts.filter((artifact) => artifact.content?.trim());

    return (
        <div className="mt-2 overflow-hidden rounded-lg border border-cortex-primary/25 bg-cortex-primary/5" data-testid="output-workbench">
            <div className="flex flex-wrap items-center gap-2 border-b border-cortex-primary/20 bg-cortex-bg/70 px-3 py-2">
                <ClipboardList className="h-3.5 w-3.5 text-cortex-primary" />
                <div className="min-w-0 flex-1">
                    <div className="truncate text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">
                        {packageTitle(deliveryPackage?.package_kind)}
                    </div>
                    {deliveryPackage?.base_path ? (
                        <div className="truncate text-[9px] font-mono text-cortex-text-muted">{deliveryPackage.base_path}</div>
                    ) : null}
                </div>
                <span className="inline-flex items-center gap-1 rounded border border-cortex-success/30 bg-cortex-success/10 px-1.5 py-0.5 text-[8px] font-mono font-bold uppercase text-cortex-success">
                    <CheckCircle2 className="h-2.5 w-2.5" />
                    Verified
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

            <div className="divide-y divide-cortex-border/60">
                {files.map((file) => {
                    const artifact = artifactForPath(artifacts, file.path, file.title);
                    const proofFile = isProofFile(file.path, file.title);
                    return (
                        <div key={`${file.path}-${file.title}`} className="grid grid-cols-[auto_1fr_auto] items-center gap-2 px-3 py-2">
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
                            <div className="text-right text-[8px] font-mono text-cortex-text-muted">
                                {file.bytes != null ? `${file.bytes}b` : artifact?.content_type?.replace("text/", "") ?? file.content_type?.replace("text/", "")}
                            </div>
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
