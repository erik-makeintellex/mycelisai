import {
    BarChart3,
    Code,
    Database,
    File,
    FileText,
    Image as ImageIcon,
    Music,
    Video,
} from "lucide-react";
import type {
    ChatArtifactRef,
    ChatConsultation,
    ChatMessage,
} from "@/store/useCortexStore";

export function trustColor(score?: number): string {
    if (score == null) return "";
    if (score >= 0.8) return "text-cortex-success";
    if (score >= 0.5) return "text-cortex-warning";
    return "text-cortex-danger";
}

export const COUNCIL_META: Record<string, { label: string; color: string }> = {
    "council-architect": { label: "Architect", color: "text-cortex-info" },
    "council-coder": { label: "Coder", color: "text-cortex-success" },
    "council-creative": { label: "Creative", color: "text-cortex-warning" },
    "council-sentry": { label: "Sentry", color: "text-cortex-danger" },
};

export const STARTER_PROMPTS = [
    "Search the web for current product news and cite sources",
    "Create or review media output and show it in chat",
    "Schedule this as a recurring review after I approve",
    "Keep a monitor running and summarize changes here",
    "Connect this task to the current team's NATS lane if configured",
    "Propose a small temporary team and ask me to approve",
    "Review private/service boundaries before taking action",
    "Ask the active teams for blockers and summarize",
    "Use host data from workspace/shared-sources",
    "Review my request, match prior context, then ask me to confirm",
] as const;

const ASK_CLASS_BADGES: Record<string, { label: string; tone: string }> = {
    governed_artifact: {
        label: "Artifact result",
        tone: "bg-cortex-primary/10 text-cortex-primary border-cortex-primary/20",
    },
    specialist_consultation: {
        label: "Specialist support",
        tone: "bg-cortex-warning/10 text-cortex-warning border-cortex-warning/20",
    },
};

export function artifactIcon(type: string) {
    switch (type) {
        case "code":
            return Code;
        case "document":
            return FileText;
        case "image":
            return ImageIcon;
        case "audio":
            return Music;
        case "video":
            return Video;
        case "data":
            return Database;
        case "chart":
            return BarChart3;
        default:
            return File;
    }
}

export function askClassBadge(askClass?: ChatMessage["ask_class"]) {
    if (!askClass) return null;
    return ASK_CLASS_BADGES[askClass] ?? null;
}

export function artifactDisplayLabel(artifact: ChatArtifactRef): string {
    const title = artifact.title?.trim();
    if (title) return title;
    return artifact.type ? `${artifact.type} artifact` : "artifact";
}

export function artifactDownloadHref(artifact: ChatArtifactRef): string | null {
    const artifactId = artifact.id?.trim();
    if (!artifactId) {
        return artifact.url?.trim() || null;
    }
    return `/api/v1/artifacts/${encodeURIComponent(artifactId)}/download`;
}

export function binaryArtifactLabel(artifact: ChatArtifactRef): string {
    if (artifact.saved_path?.trim()) {
        return artifact.saved_path.trim();
    }
    return artifactDisplayLabel(artifact);
}

export function artifactResultSummary(artifacts?: ChatArtifactRef[]): string | null {
    if (!artifacts?.length) return null;

    const labels = artifacts.map(artifactDisplayLabel);
    if (labels.length === 1) {
        return `Soma prepared 1 artifact for review: ${labels[0]}.`;
    }
    if (labels.length === 2) {
        return `Soma prepared 2 artifacts for review: ${labels[0]} and ${labels[1]}.`;
    }
    return `Soma prepared ${labels.length} artifacts for review: ${labels[0]}, ${labels[1]}, and ${labels.length - 2} more.`;
}

export function consultationResultSummary(consultations?: ChatConsultation[]): string | null {
    if (!consultations?.length) return null;

    if (consultations.length === 1) {
        const consultation = consultations[0];
        const memberLabel = COUNCIL_META[consultation.member]?.label ?? consultation.member;
        return `Soma checked with ${memberLabel} while shaping this answer: ${ensureSentence(consultation.summary)}`;
    }

    const memberLabels = consultations.map((consultation) => COUNCIL_META[consultation.member]?.label ?? consultation.member);
    if (memberLabels.length === 2) {
        return `Soma checked with ${memberLabels[0]} and ${memberLabels[1]} while shaping this answer.`;
    }
    return `Soma checked with ${memberLabels[0]}, ${memberLabels[1]}, and ${memberLabels.length - 2} more specialists while shaping this answer.`;
}

export function toolToActivity(log: { type?: string; payload?: Record<string, string>; message?: string }): string {
    const payload = (log.payload ?? {}) as Record<string, string>;
    const tool = payload.tool ?? payload.tool_name ?? "";
    switch (tool) {
        case "consult_council":
            return `Consulting ${payload.member ?? "council"}...`;
        case "generate_blueprint":
            return "Generating mission blueprint...";
        case "research_for_blueprint":
            return "Researching past missions...";
        case "write_file":
            return `Writing ${payload.path ?? "file"}...`;
        case "read_file":
            return `Reading ${payload.path ?? "file"}...`;
        case "search_memory":
            return "Searching memory...";
        case "recall":
            return "Recalling context...";
        case "store_artifact":
            return "Storing artifact...";
        case "list_teams":
            return "Checking active teams...";
        case "get_system_status":
            return "Reading system status...";
        default:
            return tool ? `${tool.replace(/_/g, " ")}...` : (log.message ?? "Working...");
    }
}

function ensureSentence(text: string): string {
    const trimmed = text.trim();
    if (!trimmed) return "";
    return /[.!?]$/.test(trimmed) ? trimmed : `${trimmed}.`;
}
