import type { ChatMessage } from "@/store/cortexStoreTypes";
import { parseKnownTimestamp, toTitleCase } from "@/components/organizations/organizationSummaryHelpers";

export type ConversationOutcomeSummary = {
    actionLabel: string;
    teamsEngaged: string[];
    outputsGenerated: string[];
    timestamp: number;
};

export function findLatestConversationOutcome(
    messages: ChatMessage[],
    teamLeadName: string,
): ConversationOutcomeSummary | null {
    for (let index = messages.length - 1; index >= 0; index -= 1) {
        const message = messages[index];
        if (message.role === "user") {
            continue;
        }

        const outcome = summarizeConversationOutcome(message, teamLeadName);
        if (outcome) {
            return outcome;
        }
    }

    return null;
}

function summarizeConversationOutcome(
    message: ChatMessage | undefined,
    teamLeadName: string,
): ConversationOutcomeSummary | null {
    if (!message) {
        return null;
    }

    const outputsGenerated = extractOutputsFromConversation(message);
    if (outputsGenerated.length === 0) {
        return null;
    }

    return {
        actionLabel: actionLabelForConversationOutcome(message),
        teamsEngaged: extractTeamsFromConversation(message, teamLeadName),
        outputsGenerated,
        timestamp: parseKnownTimestamp(message.timestamp),
    };
}

function actionLabelForConversationOutcome(message: ChatMessage) {
    const hasRunProof = Boolean(message.run_id?.trim());

    if (message.mode === "blocker") {
        return "Resolve the current workspace blocker";
    }

    if (message.proposal) {
        const lifecycle = message.proposal_status ?? "active";
        if (lifecycle === "cancelled") {
            return "Resume after cancelling the last proposal";
        }
        if (lifecycle === "confirmed_pending_execution") {
            return "Follow the confirmed proposal until proof arrives";
        }
        if (lifecycle === "executed") {
            return hasRunProof ? "Review the verified execution result" : "Follow the confirmed proposal until proof arrives";
        }
        if (lifecycle === "failed") {
            return "Recover from the failed proposal confirmation";
        }
        return "Review the governed proposal";
    }

    if (hasRunProof) {
        return "Inspect the verified run outcome";
    }

    if (message.mode === "execution_result") {
        return "Follow the confirmed proposal until proof arrives";
    }

    return "Continue the current Soma conversation";
}

export function extractTeamsFromConversation(
    message: ChatMessage | undefined,
    teamLeadName: string,
) {
    if (!message) {
        return ["Soma"];
    }

    const consultationLabels = (message.consultations ?? [])
        .map((consultation) => friendlyRoleLabel(consultation.member))
        .filter(Boolean);

    if (consultationLabels.length > 0) {
        return consultationLabels;
    }

    if (message.source_node && message.source_node !== "admin") {
        return [friendlyRoleLabel(message.source_node)];
    }

    return ["Soma", teamLeadName];
}

export function extractOutputsFromConversation(message: ChatMessage | undefined) {
    if (!message) {
        return [];
    }

    if (message.mode === "blocker") {
        return ["Blocked before completion"];
    }

    const hasRunProof = Boolean(message.run_id?.trim());
    const lifecycle = message.proposal?.intent_proof_id ? message.proposal_status ?? "active" : null;
    if (lifecycle === "cancelled") {
        return ["Proposal cancelled"];
    }
    if (lifecycle === "confirmed_pending_execution") {
        return ["Awaiting execution proof"];
    }
    if (lifecycle === "executed") {
        return hasRunProof ? ["Verified run created"] : ["Awaiting execution proof"];
    }
    if (lifecycle === "failed") {
        return ["Proposal confirmation failed"];
    }
    if (lifecycle === "active") {
        return ["Governed proposal ready"];
    }

    if (hasRunProof) {
        return ["Verified run created"];
    }

    if (message.mode === "execution_result") {
        return ["Awaiting execution proof"];
    }

    const artifacts = (message.artifacts ?? []).map((artifact) => artifact.type === "image" ? "Imagery" : artifact.title || `${toTitleCase(artifact.type)} output`);
    if (artifacts.length > 0) {
        return artifacts.slice(0, 3);
    }

    if (message.mode === "answer" && message.content?.trim()) {
        return ["Conversation guidance"];
    }

    return [];
}

function friendlyRoleLabel(value: string) {
    if (value === "admin") {
        return "Soma";
    }
    return value
        .replace(/^council-/, "")
        .replace(/[-_]+/g, " ")
        .replace(/\b\w/g, (segment) => segment.toUpperCase());
}
