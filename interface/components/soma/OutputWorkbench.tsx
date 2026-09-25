"use client";

import {
    CheckCircle2,
    ClipboardList,
    Download,
    ExternalLink,
    FileText,
    ShieldAlert,
    ShieldCheck,
    ShieldQuestion,
    ShieldX,
} from "lucide-react";
import type {
    ChatArtifactRef,
    DeliveryPackageFile,
    DeliveryPackageSummary,
    OutputProofEnvelope,
    OutputRef,
} from "@/store/useCortexStore";

function isProofFile(path?: string, title?: string): boolean {
    const lowerPath = path?.toLowerCase() ?? "";
    const lowerTitle = title?.toLowerCase() ?? "";
    return lowerPath.endsWith("proof.md") || lowerTitle.includes("proof");
}

const READBACK_STYLES: Record<string, { icon: typeof ShieldCheck; tone: string; label: string }> = {
    verified: { icon: ShieldCheck, tone: "border-cortex-success/30 bg-cortex-success/10 text-cortex-success", label: "verified" },
    degraded: { icon: ShieldAlert, tone: "border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning", label: "degraded" },
    missing: { icon: ShieldX, tone: "border-red-500/30 bg-red-500/10 text-red-400", label: "missing" },
    blocked: { icon: ShieldX, tone: "border-red-500/30 bg-red-500/10 text-red-400", label: "blocked" },
};

function readbackStyle(status?: string) {
    return READBACK_STYLES[status ?? ""] ?? { icon: ShieldQuestion, tone: "border-cortex-border bg-cortex-bg text-cortex-text-muted", label: status || "unknown" };
}

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
    const rows: [string, string | undefined][] = [
        ["Storage ref", proof.storage_ref],
        ["Checksum", proof.checksum ? `${proof.checksum_algorithm ?? "sha256"}:${proof.checksum.slice(0, 16)}...` : undefined],
        ["Bytes", proof.bytes != null ? String(proof.bytes) : undefined],
        ["Source run", proof.source_run_id],
        ["Source contract", proof.source_contract_id],
        ["Proof artifact", proof.artifact_id],
        ["Path boundary", proof.path_boundary_status],
        ["Readback status", proof.readback_status],
        ["Created at", proof.created_at],
    ];
    return (
        <div className="mt-1.5 space-y-1 rounded border border-cortex-border/60 bg-cortex-bg px-2 py-1.5">
            {rows.filter(([, value]) => value).map(([key, value]) => (
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
    outputRefs = [],
    runId,
}: {
    deliveryPackage?: DeliveryPackageSummary;
    artifacts?: ChatArtifactRef[];
    outputRefs?: OutputRef[];
    runId?: string;
}) {
    if (!deliveryPackage && artifacts.length === 0 && outputRefs.length === 0) return null;

    const refs = outputRefs.length > 0 ? outputRefs : deliveryPackage?.items ?? [];
    const files: DeliveryPackageFile[] = deliveryPackage?.files
        ?? (refs.length > 0
            ? refs.map((ref) => ({
                path: ref.storage_ref || ref.label,
                title: ref.label,
                content_type: ref.content_type,
                bytes: ref.bytes,
                sha256: ref.sha256,
            }))
            : artifacts.map((artifact) => ({
                path: artifact.saved_path || artifact.title,
                title: artifact.title,
                content_type: artifact.content_type,
            })));
    const proof = files.find((file) => isProofFile(file.path, file.title));
    const packageRunId = runId || deliveryPackage?.run_id;
    const packageIdentity = deliveryPackage?.base_path || deliveryPackage?.package_id;
    const packageStatus = deliveryPackage?.status || "verified";
    const contentArtifacts = artifacts.filter((artifact) => artifact.content?.trim());

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

            <div className="divide-y divide-cortex-border/60">
                {files.map((file) => {
                    const artifact = artifactForPath(artifacts, file.path, file.title);
                    const ref = refs.find((candidate) => candidate.storage_ref === file.path || candidate.label === file.title);
                    const proofFile = isProofFile(file.path, file.title);
                    const role = ref?.role || (proofFile ? "proof" : "");
                    const openUrl = ref?.open_url ?? artifact?.url;
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
