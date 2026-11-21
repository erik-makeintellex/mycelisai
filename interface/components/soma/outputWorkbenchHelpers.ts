import { ShieldAlert, ShieldCheck, ShieldQuestion, ShieldX } from "lucide-react";
import type {
    ChatArtifactRef,
    DeliveryPackageFile,
    DeliveryPackageSummary,
    OutputProofEnvelope,
    OutputRef,
} from "@/store/useCortexStore";

export function isProofFile(path?: string, title?: string): boolean {
    const lowerPath = path?.toLowerCase() ?? "";
    const lowerTitle = title?.toLowerCase() ?? "";
    return lowerPath.endsWith("proof.md") || lowerTitle.includes("proof");
}

export const READBACK_STYLES: Record<string, { icon: typeof ShieldCheck; tone: string; label: string }> = {
    verified: { icon: ShieldCheck, tone: "border-cortex-success/30 bg-cortex-success/10 text-cortex-success", label: "verified" },
    degraded: { icon: ShieldAlert, tone: "border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning", label: "degraded" },
    missing: { icon: ShieldX, tone: "border-red-500/30 bg-red-500/10 text-red-400", label: "missing" },
    blocked: { icon: ShieldX, tone: "border-red-500/30 bg-red-500/10 text-red-400", label: "blocked" },
};

export function readbackStyle(status?: string) {
    return READBACK_STYLES[status ?? ""] ?? { icon: ShieldQuestion, tone: "border-cortex-border bg-cortex-bg text-cortex-text-muted", label: status || "unknown" };
}

export function packageTitle(packageKind?: string): string {
    switch (packageKind) {
        case "codex_task_bundle":
            return "Codex task bundle";
        case "dev_project_bundle":
            return "Project files";
        case "interactive_game_bundle":
            return "Playable game files";
        case "product_wireframe_bundle":
            return "Interactive app prototype";
        case "team_execution_bundle":
            return "Team execution plan";
        default:
            return packageKind?.replace(/_/g, " ") ?? "Delivery package";
    }
}

export function proofStageState(status?: string): "completed" | "active" | "failed" {
    const normalized = status?.toLowerCase().trim() ?? "";
    if (normalized === "pending_review") return "active";
    if (normalized === "failed" || normalized === "blocked" || normalized === "missing") return "failed";
    return "completed";
}

export const STAGE_STYLES: Record<"completed" | "active" | "failed", string> = {
    completed: "border-cortex-success/30 bg-cortex-success/10 text-cortex-success",
    active: "border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning",
    failed: "border-red-500/30 bg-red-500/10 text-red-400",
};

export function artifactForPath(artifacts: ChatArtifactRef[], path?: string, title?: string): ChatArtifactRef | undefined {
    if (!path && !title) return undefined;
    return artifacts.find((artifact) => (
        (path && artifact.saved_path === path)
        || (title && artifact.title === title)
    ));
}

export function isPlayablePackageFile(file: DeliveryPackageFile, artifact?: ChatArtifactRef): boolean {
    return (
        (file.content_type || artifact?.content_type || "").toLowerCase().includes("html")
        || Boolean((file.path || artifact?.saved_path || "").toLowerCase().endsWith(".html"))
    );
}

export function resolvePackageFiles(
    deliveryPackage?: DeliveryPackageSummary,
    refs: OutputRef[] = [],
    artifacts: ChatArtifactRef[] = [],
): DeliveryPackageFile[] {
    if (deliveryPackage?.files) return deliveryPackage.files;
    if (refs.length > 0) {
        return refs.map((ref) => ({
            path: ref.storage_ref || ref.label,
            title: ref.label,
            content_type: ref.content_type,
            bytes: ref.bytes,
            sha256: ref.sha256,
        }));
    }
    return artifacts.map((artifact) => ({
        path: artifact.saved_path || artifact.title,
        title: artifact.title,
        content_type: artifact.content_type,
    }));
}

export function proofDetailRows(proof: OutputProofEnvelope): [string, string][] {
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
    return rows.filter((entry): entry is [string, string] => Boolean(entry[1]));
}
