// activityLabels.ts — Human-readable translation helpers for activity events and proposals

import type { MissionEvent, ProposalData } from "@/store/useCortexStore";

export interface ActivityEventLike {
    event_type?: string;
    type?: string;
    source_agent?: string;
    source_team?: string;
    source?: string;
    message?: string;
    payload?: Record<string, unknown>;
}

const ACTOR_LABELS: Record<string, string> = {
    admin: "Soma",
    architect: "Architect",
    coder: "Coder",
    creative: "Creative",
    sentry: "Sentry",
};

const HUMAN_TOOL_LABELS: Record<string, string> = {
    consult_council: "Consult Council",
    delegate_task: "Assign to Team",
    delegate: "Assign to Team",
    create_team: "Create Team",
    search_memory: "Memory Search",
    load_deployment_context: "Load Knowledge",
    promote_deployment_context: "Promote Knowledge",
    generate_blueprint: "Generate Blueprint",
    write_file: "Save Deliverable",
    publish_signal: "Send Signal",
    recall: "Recall Knowledge",
    research_for_blueprint: "Pre-Blueprint Research",
    list_teams: "List Teams",
    list_missions: "List Missions",
    get_system_status: "System Check",
    list_available_tools: "List Tools",
    list_catalogue: "Browse Catalogue",
    remember: "Save to Memory",
    store_artifact: "Store Artifact",
    broadcast: "Broadcast",
    read_signals: "Listen to Signals",
    read_file: "Read File",
    generate_image: "Generate Image",
    save_cached_image: "Save Image",
    summarize_conversation: "Summarize Conversation",
};

function resolveActorLabel(event: ActivityEventLike, assistantName = "Soma"): string {
    const rawActor = event.source_agent?.trim() || event.source_team?.trim() || event.source?.trim();
    if (!rawActor) return assistantName;
    if (rawActor === "admin" || rawActor === "soma") return assistantName;
    const stripped = rawActor.replace(/^council-/, "");
    return ACTOR_LABELS[stripped] ?? rawActor;
}

function resolveToolLabel(tool: string): string {
    return HUMAN_TOOL_LABELS[tool] ?? tool;
}

function resolveFilename(payload: Record<string, unknown>): string {
    const candidates = [
        payload.filename,
        payload.file_path,
        payload.path,
        payload.saved_path,
        payload.title,
        payload.package_slug,
        payload.base_path,
        payload.resource,
    ];
    for (const candidate of candidates) {
        if (typeof candidate === "string" && candidate.trim().length > 0) {
            return candidate.trim();
        }
    }
    return "deliverable";
}

function resolveToolName(payload: Record<string, unknown>): string | null {
    const candidates = [payload.tool, payload.tool_name, payload.action, payload.capability];
    for (const candidate of candidates) {
        if (typeof candidate === "string" && candidate.trim().length > 0) {
            return candidate.trim();
        }
    }
    return null;
}

/**
 * Converts raw NATS / MissionEventEnvelope items into friendly operator sentences
 * instead of raw JSON or SDK dumps.
 */
export function formatHumanActivityEvent(
    event: MissionEvent | ActivityEventLike,
    assistantName = "Soma",
): string {
    const eventType = (event.event_type || (event as ActivityEventLike).type || "").trim();
    const payload = (event.payload ?? {}) as Record<string, unknown>;
    const actor = resolveActorLabel(event, assistantName);
    const filename = resolveFilename(payload);
    const tool = resolveToolName(payload);

    switch (eventType) {
        case "artifact.created": {
            if (payload.package_kind === "codex_task_bundle") {
                const label = String(payload.package_slug ?? payload.base_path ?? "codex_task_bundle");
                return `${actor} created task package ${label}`;
            }
            return `${actor} created ${filename}`;
        }
        case "tool.invoked":
        case "tool.completed": {
            if (tool === "write_file" || tool === "store_artifact" || tool === "save_cached_image") {
                return `${actor} saved ${filename} to Durable Files`;
            }
            if (tool) {
                return `${actor} ran ${resolveToolLabel(tool)}`;
            }
            return `${actor} completed a workspace tool action`;
        }
        case "approval.granted":
        case "approval.approved": {
            const actionLabel = tool
                ? resolveToolLabel(tool)
                : typeof payload.intent === "string" && payload.intent.trim()
                ? payload.intent.trim()
                : "governed action";
            return `${assistantName} approved ${actionLabel}`;
        }
        case "approval.requested":
        case "approval.pending": {
            const actionLabel = tool ? resolveToolLabel(tool) : "governed action";
            return `${actor} requested approval for ${actionLabel}`;
        }
        case "mission.started": {
            const title = typeof payload.title === "string" && payload.title.trim() ? `: ${payload.title.trim()}` : "";
            return `${actor} started workflow${title}`;
        }
        case "mission.completed":
            return `${actor} completed workflow`;
        case "mission.failed":
        case "tool.failed":
            return `${actor} encountered an issue`;
        case "memory.stored":
            return `${actor} saved context to Memory`;
        case "memory.recalled":
            return `${actor} recalled context from Memory`;
        case "trigger.fired": {
            const rule = typeof payload.rule_name === "string"
                ? payload.rule_name
                : typeof payload.title === "string"
                ? payload.title
                : "";
            return rule ? `Automation trigger fired: ${rule}` : "Automation trigger fired";
        }
        default: {
            if (typeof payload.error === "string" && payload.error.trim()) {
                return `${actor} reported error: ${payload.error.trim().slice(0, 80)}`;
            }
            if (tool) {
                return `${actor} used ${resolveToolLabel(tool)}`;
            }
            if (typeof payload.title === "string" && payload.title.trim()) {
                return `${actor}: ${payload.title.trim()}`;
            }
            const msg = (event as ActivityEventLike).message;
            if (typeof msg === "string" && msg.trim()) {
                return msg.trim();
            }
            const friendlyType = eventType ? eventType.replace(/[._-]+/g, " ") : "activity";
            return `${actor} recorded ${friendlyType}`;
        }
    }
}

/**
 * Generates a plain-English headline for a governed proposal card without exposing raw SDK tool IDs.
 */
export function formatProposalHeadline(proposal: ProposalData, assistantName = "Soma"): string {
    const tools = proposal.tools ?? [];
    const capabilities = proposal.capability_ids ?? [];
    if (tools.includes("write_file") || capabilities.includes("write_file") || tools.includes("store_artifact")) {
        return `${assistantName} proposes saving a deliverable to your workspace`;
    }
    if (tools.includes("create_team") || tools.includes("generate_blueprint")) {
        return `${assistantName} proposes creating a new team workflow`;
    }
    if (tools.includes("delegate") || tools.includes("delegate_task")) {
        return `${assistantName} proposes coordinating a team task`;
    }
    if (tools.includes("publish_signal") || tools.includes("broadcast")) {
        return `${assistantName} proposes sending a team update`;
    }
    return `${assistantName} proposes a governed workspace action`;
}
