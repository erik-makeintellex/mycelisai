import type { ChatMessage } from '@/store/cortexStoreTypes';
import type { TeamWorkConfirmationRef } from '@/store/cortexStoreProposalTeamWorkRefs';

const proposalStartedDetail = 'Soma handed this to the work bus. You can keep talking here while updates arrive.';

export function executionStartedEvent(
    runId: string | null,
    teamWorkRefs: TeamWorkConfirmationRef[],
): NonNullable<ChatMessage['thread_events']>[number] {
    const activeWork = teamWorkRefs.find((ref) => ref.team_id && (ref.work_item_id || ref.work_id || ref.id));
    const workItemId = activeWork?.work_item_id ?? activeWork?.work_id ?? activeWork?.id;
    return {
        kind: 'execution_started',
        label: 'Work started',
        detail: runId
            ? 'Soma handed this to the work bus. It is running, not complete, and you can keep talking here.'
            : proposalStartedDetail,
        tone: 'info',
        status: 'running',
        run_id: runId ?? undefined,
        team_id: activeWork?.team_id,
        work_item_id: workItemId,
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        target_reference: workItemId ? `work:${workItemId}` : runId ? `run:${runId}` : undefined,
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
