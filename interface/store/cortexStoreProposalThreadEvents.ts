import type { ChatMessage } from '@/store/cortexStoreTypes';
import type { TeamWorkConfirmationRef } from '@/store/cortexStoreProposalTeamWorkRefs';

export type SynchronousConfigAction = 'store' | 'activate';

const proposalStartedDetail = 'Soma handed this to the work bus. You can keep talking here while updates arrive.';

export function proposalStartedState(): NonNullable<ChatMessage['ui_response_state']> {
    return { kind: 'running', label: 'Started', detail: proposalStartedDetail, tone: 'info' };
}

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

export function synchronousConfigAction(tools: string[] | undefined): SynchronousConfigAction | null {
    if (!tools?.length) return null;
    const configTools = new Set(['store_config_document', 'activate_config_document']);
    if (!tools.every((tool) => configTools.has(tool))) return null;
    if (tools.includes('activate_config_document')) return 'activate';
    if (tools.includes('store_config_document')) return 'store';
    return null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
}

export function confirmationIsCompleted(body: unknown, responseStatus: number) {
    if (responseStatus === 202) return false;
    const root = isRecord(body) ? body : {};
    const data = isRecord(root.data) ? root.data : root;
    const summary = isRecord(data.execution_summary) ? data.execution_summary : {};
    const execution = isRecord(summary.execution) ? summary.execution : {};
    const statuses = [
        data.run_status, data.execution_status,
        execution.status, summary.execution_status,
    ];
    if (data.verified === false || data.execution_state === 'running' || statuses.includes('running')) return false;
    return data.verified === true || statuses.includes('completed');
}

export function configurationPendingEvent(
    action: SynchronousConfigAction,
): NonNullable<ChatMessage['thread_events']>[number] {
    const activating = action === 'activate';
    return {
        kind: 'execution_update',
        label: activating ? 'Activating template' : 'Saving template',
        detail: activating
            ? 'Soma is activating the approved version for this workspace.'
            : 'Soma is saving the approved version for reuse.',
        tone: 'info',
        status: 'confirming',
        source_kind: 'workspace_ui',
        source_channel: 'soma.proposal.confirm',
        payload_kind: 'soma_thread_event',
        timestamp: new Date().toISOString(),
    };
}

export function configurationCompletedEvent(
    action: SynchronousConfigAction,
): NonNullable<ChatMessage['thread_events']>[number] {
    const activating = action === 'activate';
    return {
        kind: 'result_ready',
        label: activating ? 'Template active' : 'Template saved',
        detail: activating
            ? 'New matching work will use this saved version.'
            : 'It is saved but remains inactive until you activate it.',
        tone: 'success',
        status: 'completed',
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        timestamp: new Date().toISOString(),
    };
}

export function configurationPendingState(
    action: SynchronousConfigAction,
): NonNullable<ChatMessage['ui_response_state']> {
    const event = configurationPendingEvent(action);
    return { kind: 'running', label: event.label, detail: event.detail, tone: event.tone };
}

export function configurationCompletedState(
    action: SynchronousConfigAction,
): NonNullable<ChatMessage['ui_response_state']> {
    const event = configurationCompletedEvent(action);
    return { kind: 'execution_result', label: event.label, detail: event.detail, tone: event.tone };
}

export function configurationCompletedMessage(action: SynchronousConfigAction, proofSummary?: string | null) {
    const exactSummary = proofSummary?.trim();
    if (exactSummary) return exactSummary;
    return action === 'activate'
        ? 'Outcome Template active. New matching work will use this saved version.'
        : 'Outcome Template saved, but not active. Ask Soma to activate this template when you are ready.';
}
