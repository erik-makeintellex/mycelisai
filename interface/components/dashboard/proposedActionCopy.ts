import type { ProposalData } from "@/store/useCortexStore";

const plainLabelMap: Record<string, string> = {
    auto_approve: "Low-risk action",
    capability_risk: "Needs approval",
    governed_mutation_intent: "Requested change",
    governed_state: "Current status",
    write_file: "File changes",
};

export function humanizeLabel(value: string): string {
    const mapped = plainLabelMap[value.toLowerCase()];
    if (mapped) return mapped;
    const normalized = value.replace(/[_-]+/g, " ").trim();
    if (!normalized) return "";
    return normalized.charAt(0).toUpperCase() + normalized.slice(1);
}

export function plainExecutionText(value: string): string {
    return value
        .replace(/\bgoverned mutation intent\b/gi, "requested change")
        .replace(/\bgoverned state\b/gi, "current status")
        .replace(/\bcapability[_ -]risk\b/gi, "approval need")
        .replace(/\bgoverned module binding\b/gi, "tool step")
        .replace(/\bmodule binding\b/gi, "tool step")
        .replace(/\bgoverned\b/gi, "reviewed");
}

export function fallbackOperatorSummary(proposal: ProposalData): string {
    if (proposal.tools.includes("write_file")) return "create or update files in your workspace.";
    if (proposal.tools.includes("generate_blueprint")) return "prepare a durable blueprint artifact.";
    if (proposal.tools.includes("publish_signal") || proposal.tools.includes("broadcast")) return "send a platform message.";
    if (proposal.tools.includes("delegate") || proposal.tools.includes("delegate_task")) return "coordinate team work.";
    return "carry out the requested action.";
}

export function fallbackExpectedResult(proposal: ProposalData): string {
    if (proposal.tools.includes("write_file")) return "The requested file change will be created after approval.";
    if (proposal.tools.includes("generate_blueprint")) return "A saved blueprint artifact will be returned in this conversation.";
    if (proposal.tools.includes("publish_signal") || proposal.tools.includes("broadcast")) return "The platform message will be sent and the outcome will be returned here.";
    if (proposal.tools.includes("delegate") || proposal.tools.includes("delegate_task")) return "Soma will return the result in this conversation.";
    return "Soma will return the result in this conversation.";
}

export function fallbackAffectedResources(proposal: ProposalData): string[] {
    if (proposal.tools.includes("write_file")) return ["Workspace files"];
    if (proposal.tools.includes("generate_blueprint")) return ["Blueprint artifact"];
    if (proposal.tools.includes("publish_signal") || proposal.tools.includes("broadcast")) return ["Platform message"];
    if (proposal.tools.includes("delegate") || proposal.tools.includes("delegate_task")) return ["Team workflow"];
    return [];
}

export function explainApprovalPosture(proposal: ProposalData, approvalRequired: boolean, approvalMode: string): string {
    if (approvalRequired) {
        if (proposal.capability_ids?.includes("write_file")) return "This action will change your workspace, so Soma needs your approval before running it.";
        if (proposal.external_data_use) return "This action may use external systems or data, so Soma needs your approval before running it.";
        if (proposal.approval_reason === "cost_threshold") return "This action may incur additional spend, so Soma needs your approval before running it.";
        return "This action needs your approval before Soma runs it.";
    }

    return approvalMode === "optional"
        ? "This action stays within current policy thresholds, but you can still review it before execution."
        : "This action is within current policy thresholds and can run without a mandatory approval.";
}
