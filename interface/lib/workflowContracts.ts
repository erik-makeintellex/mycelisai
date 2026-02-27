export type WorkflowInputChannel = "workspace" | "trigger" | "schedule" | "api" | "sensor";
export type WorkflowOutputChannel = "chat" | "proposal" | "timeline" | "artifact" | "approval";

export type WorkflowGovernanceMode = "passive" | "approval_required" | "halted";

export interface TeamProfileTemplate {
    id: string;
    name: string;
    objectiveKinds: string[];
    description: string;
    defaultGovernanceMode: WorkflowGovernanceMode;
    requiredCapabilities: string[];
    suggestedRoutes: string[];
    inputChannels: WorkflowInputChannel[];
    outputChannels: WorkflowOutputChannel[];
}

export interface ReadinessSnapshot {
    providerReady: boolean;
    mcpReady: boolean;
    governanceReady: boolean;
    natsReady: boolean;
    sseReady: boolean;
    dbReady: boolean;
    blockers: string[];
}

export interface WorkflowIOEnvelope<T = unknown> {
    ok: boolean;
    data: T;
    error: string;
    meta: {
        channel: WorkflowInputChannel;
        run_id: string;
        team_id: string;
        timestamp: string;
    };
}

export const TEAM_PROFILE_TEMPLATES: TeamProfileTemplate[] = [
    {
        id: "research-team",
        name: "Research Team",
        objectiveKinds: ["analysis", "discovery", "summarization"],
        description: "Optimize for source gathering, synthesis, and decision support.",
        defaultGovernanceMode: "approval_required",
        requiredCapabilities: ["provider", "mcp-fetch", "memory"],
        suggestedRoutes: ["swarm.mission.events.research.*", "swarm.team.research.*"],
        inputChannels: ["workspace", "trigger", "api"],
        outputChannels: ["chat", "proposal", "timeline", "artifact"],
    },
    {
        id: "incident-team",
        name: "Incident Team",
        objectiveKinds: ["triage", "stability", "degraded_recovery"],
        description: "Optimize for runtime diagnostics, fallback handling, and stabilization.",
        defaultGovernanceMode: "approval_required",
        requiredCapabilities: ["provider", "nats", "system-observability"],
        suggestedRoutes: ["swarm.alerts.*", "swarm.team.incident.*"],
        inputChannels: ["workspace", "trigger", "schedule"],
        outputChannels: ["chat", "timeline", "approval"],
    },
    {
        id: "delivery-team",
        name: "Delivery Team",
        objectiveKinds: ["build", "execution", "verification"],
        description: "Optimize for implementation throughput with governed execution.",
        defaultGovernanceMode: "approval_required",
        requiredCapabilities: ["provider", "mcp-filesystem", "runs"],
        suggestedRoutes: ["swarm.team.delivery.*", "swarm.runs.*"],
        inputChannels: ["workspace", "api", "schedule"],
        outputChannels: ["proposal", "timeline", "artifact", "approval"],
    },
    {
        id: "governance-first-team",
        name: "Governance-First Team",
        objectiveKinds: ["policy", "high_risk_mutation", "review"],
        description: "Optimize for strict review posture and traceable approvals.",
        defaultGovernanceMode: "approval_required",
        requiredCapabilities: ["provider", "governance", "runs"],
        suggestedRoutes: ["swarm.governance.*", "swarm.approvals.*"],
        inputChannels: ["workspace", "trigger", "api"],
        outputChannels: ["proposal", "approval", "timeline"],
    },
];
