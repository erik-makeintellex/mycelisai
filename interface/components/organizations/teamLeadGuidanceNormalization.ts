import type {
    TeamLeadExecutionContract,
    TeamLeadExecutionWorkstream,
    TeamLeadGuidanceResponse,
    TeamLeadGuidedAction,
    TeamLeadWorkflowGroupDraft,
} from "@/lib/organizations";

export type GuidedExecutionContract = TeamLeadExecutionContract & {
    recommended_team_shape?: string;
    coordination_model?: string;
    recommended_team_count?: number;
    recommended_team_member_limit?: number;
    workstreams?: TeamLeadExecutionWorkstream[];
};

export const GUIDED_ACTIONS: Array<{ action: TeamLeadGuidedAction; label: string; detail: string }> = [
    {
        action: "plan_next_steps",
        label: "Run a quick strategy check",
        detail: "See the next priorities Soma would set so the workspace starts moving in a visible direction.",
    },
    {
        action: "focus_first",
        label: "Choose the first priority",
        detail: "Ask Soma to identify the first move most likely to create visible progress across the workspace.",
    },
    {
        action: "review_setup",
        label: "Review your organization setup",
        detail: "Check whether Advisors, Departments, and Specialists are ready and see what Soma recommends inspecting next.",
    },
    {
        action: "resume_retained_package",
        label: "Resume retained package",
        detail: "Recover the next step from durable outputs after a reboot, reload, or interruption without rebuilding finished work.",
    },
];

export function normalizeGuidanceResponse(
    payload: Partial<TeamLeadGuidanceResponse> | null,
    action: TeamLeadGuidedAction,
    organizationName: string,
    somaName: string,
    teamLeadName: string,
): TeamLeadGuidanceResponse {
    const requestLabel = rewriteGuidanceText(sanitizeGuidanceText(payload?.request_label, defaultRequestLabel(action)), somaName, teamLeadName);
    const headline = rewriteGuidanceText(sanitizeGuidanceText(payload?.headline, `${somaName} guidance for ${organizationName}`), somaName, teamLeadName);
    const summary = rewriteGuidanceText(sanitizeGuidanceText(payload?.summary, `${somaName} has guidance ready for ${organizationName}.`), somaName, teamLeadName);
    const prioritySteps = normalizeGuidanceList(payload?.priority_steps, defaultPrioritySteps(action, organizationName), somaName, teamLeadName);
    const suggestedFollowUps = normalizeGuidanceList(payload?.suggested_follow_ups, defaultFollowUps(action), somaName, teamLeadName);
    const executionContract = normalizeExecutionContract(payload?.execution_contract, somaName, teamLeadName);

    return {
        action,
        request_label: requestLabel,
        headline,
        summary,
        priority_steps: prioritySteps,
        suggested_follow_ups: suggestedFollowUps,
        execution_contract: executionContract,
    };
}

export function defaultRequestLabel(action: TeamLeadGuidedAction) {
    return GUIDED_ACTIONS.find((item) => item.action === action)?.label ?? "Work with Soma";
}

export function resolvePromptAction(prompt: string): TeamLeadGuidedAction {
    const normalized = prompt.trim().toLowerCase();
    if (/(resume|reboot|reload|retained package|retained output|continue from|pick back up)/i.test(normalized)) {
        return "resume_retained_package";
    }
    if (/(review|check|audit|inspect|understand|setup|structure)/i.test(normalized)) {
        return "review_setup";
    }
    if (/(first|priority|focus|start with|begin with|most important)/i.test(normalized)) {
        return "focus_first";
    }
    return "plan_next_steps";
}

export function classifyRequestScope(value: string): "compact" | "broad" {
    const normalized = value.toLowerCase();
    const broadSignals = [
        "company-wide",
        "organization-wide",
        "enterprise",
        "enterprise-wide",
        "cross-functional",
        "multi-team",
        "multi team",
        "multiple teams",
        "several teams",
        "several workstreams",
        "full suite",
        "end-to-end",
        "whole organization",
        "entire organization",
        "across teams",
        "across functions",
    ];
    let score = 0;
    for (const signal of broadSignals) {
        if (normalized.includes(signal)) {
            score += 1;
        }
    }
    if ((normalized.match(/\b(and|plus|with)\b/g) ?? []).length >= 3) {
        score += 1;
    }
    if ((normalized.match(/[,/]/g) ?? []).length >= 3) {
        score += 1;
    }
    if (normalized.split(/\s+/).filter(Boolean).length >= 18) {
        score += 1;
    }
    return score >= 1 ? "broad" : "compact";
}

function defaultPrioritySteps(action: TeamLeadGuidedAction, organizationName: string) {
    switch (action) {
        case "resume_retained_package":
            return [`Resume the retained package for ${organizationName}.`, "Confirm what is done, what remains, and who owns the next step."];
        case "focus_first":
            return [`Confirm the first priority for ${organizationName}.`, "Use that priority to create the first visible movement across the workspace."];
        case "review_setup":
            return ["Check whether Advisors, Departments, and Specialists are ready for the next move.", "Open the parts of the organization that need attention first."];
        case "plan_next_steps":
        default:
            return [`Turn ${organizationName} into a clear next move.`, "Use Soma guidance to choose the next action that will show up across the workspace."];
    }
}

function defaultFollowUps(action: TeamLeadGuidedAction) {
    return [
        "Run a quick strategy check",
        "Choose the first priority",
        "Review your organization setup",
        "Resume retained package",
    ].filter((label) => label !== defaultRequestLabel(action));
}

function normalizeGuidanceList(value: unknown, fallback: string[], somaName: string, teamLeadName: string) {
    if (!Array.isArray(value)) {
        return fallback;
    }
    const normalized = value
        .map((entry) => rewriteGuidanceText(sanitizeGuidanceText(entry, ""), somaName, teamLeadName))
        .filter((entry) => entry.length > 0);
    return normalized.length > 0 ? normalized : fallback;
}

function normalizeExecutionContract(value: unknown, somaName: string, teamLeadName: string): GuidedExecutionContract | undefined {
    if (!value || typeof value !== "object") {
        return undefined;
    }

    const contract = value as Partial<GuidedExecutionContract>;
    const executionMode = contract.execution_mode;
    if (executionMode !== "guided_review" && executionMode !== "native_team" && executionMode !== "external_workflow_contract" && executionMode !== "continuity_resume") {
        return undefined;
    }

    const targetOutputs = Array.isArray(contract.target_outputs)
        ? contract.target_outputs.map((entry) => rewriteGuidanceText(sanitizeExecutionContractText(entry, ""), somaName, teamLeadName)).filter((entry) => entry.length > 0)
        : [];
    const workstreams = Array.isArray(contract.workstreams)
        ? contract.workstreams.map((entry) => normalizeExecutionWorkstream(entry, somaName, teamLeadName)).filter((entry): entry is TeamLeadExecutionWorkstream => Boolean(entry))
        : [];

    return {
        execution_mode: executionMode,
        owner_label: rewriteGuidanceText(sanitizeExecutionContractText(contract.owner_label, "Execution path"), somaName, teamLeadName),
        summary: rewriteGuidanceText(sanitizeExecutionContractText(contract.summary, "Soma has an execution path ready."), somaName, teamLeadName),
        continuity_label: rewriteGuidanceText(sanitizeExecutionContractText(contract.continuity_label, ""), somaName, teamLeadName),
        continuity_summary: rewriteGuidanceText(sanitizeExecutionContractText(contract.continuity_summary, ""), somaName, teamLeadName),
        resume_checkpoint: rewriteGuidanceText(sanitizeExecutionContractText(contract.resume_checkpoint, ""), somaName, teamLeadName),
        team_name: sanitizeExecutionContractText(contract.team_name, ""),
        external_target: sanitizeExecutionContractText(contract.external_target, ""),
        target_outputs: targetOutputs,
        workstreams,
        workflow_group: normalizeWorkflowGroupDraft(contract.workflow_group, somaName, teamLeadName),
        recommended_team_shape: sanitizeExecutionContractText(contract.recommended_team_shape, ""),
        coordination_model: sanitizeExecutionContractText(contract.coordination_model, ""),
        recommended_team_count: typeof contract.recommended_team_count === "number" ? contract.recommended_team_count : undefined,
        recommended_team_member_limit: typeof contract.recommended_team_member_limit === "number" ? contract.recommended_team_member_limit : undefined,
    };
}

function normalizeWorkflowGroupDraft(value: unknown, somaName: string, teamLeadName: string): TeamLeadWorkflowGroupDraft | undefined {
    if (!value || typeof value !== "object") {
        return undefined;
    }

    const draft = value as Partial<TeamLeadWorkflowGroupDraft>;
    const workMode = draft.work_mode;
    if (workMode !== "read_only" && workMode !== "propose_only" && workMode !== "execute_with_approval" && workMode !== "execute_bounded" && workMode !== "resume_continuity") {
        return undefined;
    }

    return {
        group_id: sanitizeExecutionContractText(draft.group_id, ""),
        name: rewriteGuidanceText(sanitizeExecutionContractText(draft.name, "Temporary workflow group"), somaName, teamLeadName),
        goal_statement: rewriteGuidanceText(sanitizeExecutionContractText(draft.goal_statement, "Coordinate a focused workflow."), somaName, teamLeadName),
        work_mode: workMode,
        coordinator_profile: rewriteGuidanceText(sanitizeExecutionContractText(draft.coordinator_profile, "Workflow lead"), somaName, teamLeadName),
        allowed_capabilities: Array.isArray(draft.allowed_capabilities) ? draft.allowed_capabilities.map((entry) => sanitizeExecutionContractText(entry, "")).filter((entry) => entry.length > 0) : [],
        recommended_member_limit: typeof draft.recommended_member_limit === "number" ? draft.recommended_member_limit : undefined,
        expiry_hours: typeof draft.expiry_hours === "number" ? draft.expiry_hours : undefined,
        summary: rewriteGuidanceText(sanitizeExecutionContractText(draft.summary, "Launch a temporary workflow group from this Soma plan."), somaName, teamLeadName),
    };
}

function normalizeExecutionWorkstream(value: unknown, somaName: string, teamLeadName: string): TeamLeadExecutionWorkstream | undefined {
    if (!value || typeof value !== "object") {
        return undefined;
    }

    const workstream = value as Partial<TeamLeadExecutionWorkstream>;
    const label = rewriteGuidanceText(sanitizeExecutionContractText(workstream.label, ""), somaName, teamLeadName);
    const ownerLabel = rewriteGuidanceText(sanitizeExecutionContractText(workstream.owner_label, ""), somaName, teamLeadName);
    const summary = rewriteGuidanceText(sanitizeExecutionContractText(workstream.summary, ""), somaName, teamLeadName);
    const nextStep = rewriteGuidanceText(sanitizeExecutionContractText(workstream.next_step, ""), somaName, teamLeadName);
    const targetOutputs = Array.isArray(workstream.target_outputs)
        ? workstream.target_outputs.map((entry) => rewriteGuidanceText(sanitizeExecutionContractText(entry, ""), somaName, teamLeadName)).filter((entry) => entry.length > 0)
        : [];

    if (!label || !ownerLabel || !summary || !nextStep) {
        return undefined;
    }

    return {
        label,
        owner_label: ownerLabel,
        status: sanitizeExecutionContractText(workstream.status, ""),
        summary,
        next_step: nextStep,
        target_outputs: targetOutputs,
    };
}

export function rewriteGuidanceText(value: string, somaName: string, teamLeadName: string) {
    return value.replaceAll(teamLeadName, somaName).replace(/\bTeam Lead\b/g, "Soma");
}

function sanitizeGuidanceText(value: unknown, fallback: string) {
    const normalized = sanitizeText(value);
    if (!normalized || containsForbiddenGuidanceCopy(normalized)) {
        return fallback;
    }
    return normalized;
}

function sanitizeExecutionContractText(value: unknown, fallback: string) {
    return sanitizeText(value) || fallback;
}

function sanitizeText(value: unknown) {
    if (typeof value !== "string") {
        return "";
    }
    const normalized = value
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter((line) => line.length > 0)
        .filter((line) => !/^(system|debug|trace|tool|agent_id)\s*:/i.test(line))
        .join(" ")
        .replace(/\s+/g, " ")
        .trim();
    return !normalized || normalized.startsWith("{") || normalized.startsWith("[") ? "" : normalized;
}

function containsForbiddenGuidanceCopy(value: string) {
    return [
        /v8 entry flow/i,
        /bounded slice/i,
        /implementation slice/i,
        /context shell/i,
        /raw architecture controls/i,
        /\bcontract\b/i,
        /\bloop\b/i,
        /scheduler/i,
        /inception/i,
        /soma kernel/i,
        /\bcouncil\b/i,
        /provider policy/i,
        /identity\s*\/\s*continuity/i,
        /memory promotion/i,
        /pgvector/i,
    ].some((pattern) => pattern.test(value));
}
