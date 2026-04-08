import type {
    ChatMessage,
    CTSChatEnvelope,
    ExecutionMode,
    ProposalData,
    ProposalLifecycleStatus,
} from '@/store/cortexStoreTypes';

function extractReadableStructuredText(text: string): string | null {
    const trimmed = text.trim();
    if (!trimmed || !/^[{\[]/.test(trimmed)) {
        return trimmed || null;
    }
    try {
        const parsed = JSON.parse(trimmed);
        if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
            return null;
        }
        const candidate = [parsed.text, parsed.message, parsed.summary].find((value) => typeof value === 'string' && value.trim());
        return typeof candidate === 'string' ? candidate.trim() : null;
    } catch {
        return trimmed;
    }
}

export function fallbackChatContent(payload: CTSChatEnvelope['payload'], assistantName: string): string {
    const text = payload?.text?.trim();
    if (text) {
        const readable = extractReadableStructuredText(text);
        if (readable) {
            return readable;
        }
    }

    if (payload?.artifacts && payload.artifacts.length > 0) {
        return `${assistantName} prepared output for review below.`;
    }

    if (payload?.proposal) {
        return `${assistantName} prepared a proposal for review.`;
    }

    return `${assistantName} could not produce a readable reply for that request. Retry or ask ${assistantName} to summarize the result directly.`;
}

export function trimToNonEmpty(value: unknown): string | null {
    return typeof value === 'string' && value.trim() ? value.trim() : null;
}

export function extractRunIdFromResponse(raw: unknown): string | null {
    if (!raw || typeof raw !== 'object') {
        return null;
    }

    const record = raw as Record<string, unknown>;
    const directRunId = trimToNonEmpty(record.run_id);
    if (directRunId) {
        return directRunId;
    }

    const nestedCandidates = [
        record.data,
        record.execution,
        record.result,
        record.proof,
    ];

    for (const candidate of nestedCandidates) {
        if (!candidate || typeof candidate !== 'object') {
            continue;
        }
        const nestedRunId = trimToNonEmpty((candidate as Record<string, unknown>).run_id);
        if (nestedRunId) {
            return nestedRunId;
        }
    }

    return null;
}

export function isRetryableWorkspaceChatFailure(message: string, statusCode?: number): boolean {
    if (statusCode != null && [500, 502, 503, 504].includes(statusCode)) {
        return true;
    }

    const lower = message.toLowerCase();
    return (
        lower.includes('failed to fetch')
        || lower.includes('deadline exceeded')
        || lower.includes('timeout')
        || lower.includes('unreachable')
        || lower.includes('connection refused')
        || lower.includes('bad gateway')
    );
}

export function delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export function updateProposalLifecycle(
    messages: ChatMessage[],
    intentProofId: string,
    lifecycle: ProposalLifecycleStatus,
    updates?: Partial<ChatMessage>,
): ChatMessage[] {
    if (!intentProofId.trim()) {
        return messages;
    }

    let updated = false;
    const nextMessages = messages.map((message) => {
        if (message.proposal?.intent_proof_id !== intentProofId) {
            return message;
        }

        updated = true;
        return {
            ...message,
            ...updates,
            proposal_status: lifecycle,
        };
    });

    return updated ? nextMessages : messages;
}

export type DerivedMissionChatState = {
    activeMode: ExecutionMode;
    activeRunId: string | null;
    pendingProposal: ProposalData | null;
    activeConfirmToken: string | null;
};

export function deriveMissionChatState(messages: ChatMessage[]): DerivedMissionChatState {
    for (let index = messages.length - 1; index >= 0; index -= 1) {
        const message = messages[index];
        const proposalLifecycle = message.proposal ? message.proposal_status ?? 'active' : null;
        const runId = trimToNonEmpty(message.run_id);

        if (message.proposal && proposalLifecycle === 'active') {
            return {
                activeMode: 'proposal',
                activeRunId: null,
                pendingProposal: message.proposal,
                activeConfirmToken: trimToNonEmpty(message.proposal.confirm_token),
            };
        }

        if (proposalLifecycle === 'confirmed_pending_execution') {
            return {
                activeMode: 'proposal',
                activeRunId: null,
                pendingProposal: null,
                activeConfirmToken: null,
            };
        }

        if (proposalLifecycle === 'executed' || message.mode === 'execution_result' || runId) {
            if (!runId && message.proposal) {
                return {
                    activeMode: 'proposal',
                    activeRunId: null,
                    pendingProposal: null,
                    activeConfirmToken: null,
                };
            }

            if (!runId && message.mode === 'execution_result') {
                return {
                    activeMode: 'answer',
                    activeRunId: null,
                    pendingProposal: null,
                    activeConfirmToken: null,
                };
            }

            return {
                activeMode: 'execution_result',
                activeRunId: runId,
                pendingProposal: null,
                activeConfirmToken: null,
            };
        }

        if (message.mode === 'blocker') {
            return {
                activeMode: 'blocker',
                activeRunId: null,
                pendingProposal: null,
                activeConfirmToken: null,
            };
        }

        if (proposalLifecycle === 'failed' || proposalLifecycle === 'cancelled') {
            return {
                activeMode: 'answer',
                activeRunId: null,
                pendingProposal: null,
                activeConfirmToken: null,
            };
        }

        if (message.mode === 'proposal') {
            return {
                activeMode: 'proposal',
                activeRunId: null,
                pendingProposal: null,
                activeConfirmToken: null,
            };
        }

        if (message.mode === 'answer') {
            return {
                activeMode: 'answer',
                activeRunId: null,
                pendingProposal: null,
                activeConfirmToken: null,
            };
        }
    }

    return {
        activeMode: 'answer',
        activeRunId: null,
        pendingProposal: null,
        activeConfirmToken: null,
    };
}
