import type { ChatMessage } from '@/store/cortexStoreTypes';
import type { TeamWorkConfirmationRef } from '@/store/cortexStoreProposalTeamWorkRefs';

const proposalStartedDetail = 'Soma handed this to the work bus. You can keep talking here while updates arrive.';

export function executionStartedEvent(
    runId: string | null,
    _teamWorkRefs: TeamWorkConfirmationRef[],
): NonNullable<ChatMessage['thread_events']>[number] {
    return {
        kind: 'execution_started',
        label: 'Work started',
        detail: runId
            ? 'Soma handed this to the work bus. It is running, not complete, and you can keep talking here.'
            : proposalStartedDetail,
        tone: 'info',
        status: 'running',
        run_id: runId ?? undefined,
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        target_reference: runId ? `run:${runId}` : undefined,
        timestamp: new Date().toISOString(),
    };
}

export function approvalSentEvent(): NonNullable<ChatMessage['thread_events']>[number] {
    return {
        kind: 'execution_update',
        label: 'Approval sent',
        detail: 'Soma is starting the handoff.',
        tone: 'info',
        status: 'confirming',
        source_kind: 'workspace_ui',
        source_channel: 'soma.proposal.confirm',
        payload_kind: 'soma_thread_event',
        timestamp: new Date().toISOString(),
    };
}
