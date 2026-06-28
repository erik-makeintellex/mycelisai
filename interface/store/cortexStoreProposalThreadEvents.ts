import type { ChatMessage } from '@/store/cortexStoreTypes';
import type { TeamWorkConfirmationRef } from '@/store/cortexStoreProposalTeamWorkRefs';

const proposalStartedDetail = 'Soma handed this to the work bus. You can keep talking here while updates arrive.';

export function executionStartedEvent(
    runId: string | null,
    teamWorkRefs: TeamWorkConfirmationRef[],
): NonNullable<ChatMessage['thread_events']>[number] {
    return {
        kind: 'execution_started',
        label: runId ? 'Execution started' : 'Work approved',
        detail: runId
            ? 'Soma handed this to the work bus and saved the run receipt.'
            : proposalStartedDetail,
        tone: 'info',
        status: teamWorkRefs.length ? 'team handoff' : 'running',
        run_id: runId ?? undefined,
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        href: runId ? `/runs/${runId}` : undefined,
        href_label: runId ? 'Open run receipt' : undefined,
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
