import { extractRunIdFromResponse, trimToNonEmpty, updateProposalLifecycle } from '@/store/cortexStoreChatWorkflow';
import { buildMissionChatFailure } from '@/lib/missionChatFailure';
import type { ChatMessage, ConfirmProposalResult } from '@/store/cortexStoreTypes';
import {
    approvalSentEvent,
    confirmationIsCompleted,
    configurationCompletedEvent,
    configurationCompletedMessage,
    configurationCompletedState,
    configurationPendingEvent,
    configurationPendingState,
    executionStartedEvent,
    proposalStartedState,
    synchronousConfigAction,
} from '@/store/cortexStoreProposalThreadEvents';
import { extractTeamWorkRefs, teamWorkMessage, type TeamWorkConfirmationRef } from '@/store/cortexStoreProposalTeamWorkRefs';
import {
    failureThreadEvent,
    isMediaDependencyFailure,
    mediaDependencyRecoveryCopy,
    recoveryTextFromExecutionSummary,
    type ConfirmFailureBody,
} from '@/store/cortexStoreProposalExecutionRecovery';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { ProposalData } from '@/store/cortexStoreTypesChat';

function confirmedRunMessage(runId: string | null, running: boolean, summary?: string | null, teamWorkRefs: TeamWorkConfirmationRef[] = []) {
    const state = runId ? `Run ${runId.slice(0, 8)} ${running ? 'started' : 'recorded'}.` : 'Proposal approved.';
    const next = running
        ? 'Soma started the work. This is running, not a completed result or proof.'
        : 'Approval was recorded. Completion has not been verified.';
    return [state, next, teamWorkMessage(teamWorkRefs), summary].filter(Boolean).join(' ');
}

const approvalSentState = {
    kind: 'proposal' as const,
    label: 'Approval sent',
    detail: 'Soma is checking the approval.',
    tone: 'info' as const,
};

const approvalRecordedState = {
    kind: 'proposal' as const,
    label: 'Approval recorded',
    detail: 'Completion has not been verified.',
    tone: 'info' as const,
};

function verifiedRunMessage(runId: string, summary?: string | null) {
    return [`Run ${runId.slice(0, 8)} completed. The verified result is available to review.`, summary].filter(Boolean).join(' ');
}

function confirmationPendingEvent(runId: string | null): NonNullable<ChatMessage['thread_events']>[number] {
    return {
        kind: 'execution_update',
        label: 'Approval recorded',
        detail: 'Soma received the approval. Completion has not been verified.',
        tone: 'info',
        status: 'confirmed',
        run_id: runId ?? undefined,
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        target_reference: runId ? `run:${runId}` : undefined,
        timestamp: new Date().toISOString(),
    };
}

export function createCortexProposalExecutionSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<'confirmProposal' | 'cancelProposal'> {
    function latestActiveProposal(): ProposalData | null {
        const messages = get().missionChat;
        for (let index = messages.length - 1; index >= 0; index -= 1) {
            const message = messages[index];
            if (message.proposal && (message.proposal_status ?? 'active') === 'active') {
                return message.proposal;
            }
        }
        return null;
    }

    return {
        confirmProposal: async (proposalOverride?: ProposalData, operatorReply?: string): Promise<ConfirmProposalResult> => {
            const { activeConfirmToken, pendingProposal } = get();
            const proposal = proposalOverride ?? pendingProposal ?? latestActiveProposal();
            const confirmToken = proposalOverride
                ? trimToNonEmpty(proposalOverride.confirm_token)
                : trimToNonEmpty(activeConfirmToken)
                    ?? trimToNonEmpty(pendingProposal?.confirm_token)
                    ?? trimToNonEmpty(proposal?.confirm_token);
            if (!confirmToken || !proposal) {
                return {
                    ok: false,
                    runId: null,
                    error: 'No pending proposal to confirm',
                };
            }
            const intentProofId = trimToNonEmpty(proposal.intent_proof_id);
            if (!intentProofId) {
                return {
                    ok: false,
                    runId: null,
                    error: 'This proposal is missing executable proof. Ask Soma to regenerate it before running.',
                };
            }
            const configAction = synchronousConfigAction(proposal.tools);
            set((s) => {
                const conversationalReply = trimToNonEmpty(operatorReply);
                const missionChat = conversationalReply
                    ? [...s.missionChat, {
                        role: 'user' as const,
                        content: conversationalReply,
                        timestamp: new Date().toISOString(),
                    }]
                    : s.missionChat;
                return {
                    activeMode: 'proposal',
                    missionChatError: null,
                    missionChatFailure: null,
                    missionChat: updateProposalLifecycle(missionChat, intentProofId, 'confirmed_pending_execution', {
                        mode: 'proposal',
                        ui_response_state: configAction ? configurationPendingState(configAction) : approvalSentState,
                        thread_events: [configAction ? configurationPendingEvent(configAction) : approvalSentEvent()],
                    }),
                };
            });
            try {
                const res = await fetch('/api/v1/intent/confirm-action', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ confirm_token: confirmToken }),
                });
                if (res.ok) {
                    const body = await res.json();
                    const runId = extractRunIdFromResponse(body);
                    const teamWorkRefs = extractTeamWorkRefs(body);
                    const proofSummary = trimToNonEmpty(body?.data?.message)
                        ?? trimToNonEmpty(body?.message)
                        ?? trimToNonEmpty(body?.data?.summary)
                        ?? trimToNonEmpty(body?.summary)
                        ?? trimToNonEmpty(body?.data?.execution_summary?.execution?.summary)
                        ?? trimToNonEmpty(body?.data?.execution_summary?.execution_summary);
                    const executionSummary = body?.data?.execution_summary;
                    const summaryStatus = executionSummary?.execution?.status ?? executionSummary?.execution_status;
                    const summaryProof = executionSummary?.proof;
                    const verifiedProof = Array.isArray(summaryProof)
                        ? summaryProof.some((proof) => proof?.verified === true)
                        : summaryProof?.verified === true;
                    const completedRun = Boolean(runId
                        && confirmationIsCompleted(body, res.status)
                        && body?.data?.verified === true
                        && body?.data?.execution_state === 'verified'
                        && body?.data?.run_status === 'completed'
                        && ['completed', 'verified'].includes(summaryStatus)
                        && verifiedProof);
                    const runningRun = body?.data?.verified === false
                        && body?.data?.execution_state === 'running'
                        && body?.data?.run_status === 'running'
                        && summaryStatus === 'running';
                    const completedConfigAction = configAction && confirmationIsCompleted(body, res.status)
                        ? configAction
                        : null;
                    const completed = Boolean(completedConfigAction || completedRun);
                    const lifecycle = completed ? 'executed' : 'confirmed_pending_execution';
                    const completedState = completedConfigAction ? configurationCompletedState(completedConfigAction)
                        : completedRun
                            ? { kind: 'execution_result' as const, label: 'Result verified', detail: 'Soma completed the approved action and saved its proof.', tone: 'success' as const }
                            : runningRun ? proposalStartedState() : approvalRecordedState;
                    const completedEvent = completedConfigAction ? configurationCompletedEvent(completedConfigAction)
                        : completedRun
                            ? {
                                kind: 'result_ready' as const,
                                label: 'Result verified',
                                detail: 'Soma completed the approved action and saved its proof.',
                                tone: 'success' as const,
                                status: 'completed',
                                run_id: runId ?? undefined,
                                source_kind: 'web_api' as const,
                                source_channel: 'api.intent.confirm-action',
                                payload_kind: 'soma_thread_event',
                                timestamp: new Date().toISOString(),
                            }
                            : runningRun ? executionStartedEvent(runId, teamWorkRefs)
                                : confirmationPendingEvent(runId);
                    const systemMsg: ChatMessage = {
                        role: 'system',
                        content: completedConfigAction
                            ? configurationCompletedMessage(completedConfigAction, proofSummary)
                            : completedRun && runId ? verifiedRunMessage(runId, proofSummary)
                                : confirmedRunMessage(runId, runningRun, runningRun ? proofSummary : null, teamWorkRefs),
                        mode: completed || runId ? 'execution_result' : 'proposal',
                        ui_response_state: completedState,
                        run_id: runId ?? undefined,
                        thread_events: [completedEvent],
                        execution_summary: completed || runningRun ? executionSummary : undefined,
                        timestamp: new Date().toISOString(),
                    };
                    set((s) => ({
                        activeRunId: completedConfigAction ? null : runId,
                        activeMode: completed || runId ? 'execution_result' : 'proposal',
                        missionChatError: null,
                        missionChatFailure: null,
                        durableWorkRefreshVersion: s.durableWorkRefreshVersion + 1,
                        missionChat: [
                            ...updateProposalLifecycle(s.missionChat, intentProofId, lifecycle, {
                                mode: completed || runId ? 'execution_result' : 'proposal',
                                ui_response_state: completedState,
                                run_id: runId ?? undefined,
                                execution_summary: completed || runningRun ? executionSummary : undefined,
                            }),
                            systemMsg,
                        ],
                        pendingProposal: null,
                        activeConfirmToken: null,
                    }));
                    void get().fetchTeamsDetail();
                    return { ok: true, runId };
                }

                const text = await res.text();
                let errMsg = 'Confirm action failed';
                let parsedBody: ConfirmFailureBody | null = null;
                try {
                    parsedBody = JSON.parse(text) as ConfirmFailureBody;
                    errMsg = parsedBody.error || errMsg;
                } catch {
                    errMsg = text || errMsg;
                }
                const failureRunId = trimToNonEmpty(parsedBody?.data?.run_id);
                const failureExecutionSummary = parsedBody?.data?.execution_summary;
                const recovery = recoveryTextFromExecutionSummary(failureExecutionSummary);
                const failure = buildMissionChatFailure({
                    assistantName: get().assistantName,
                    targetId: 'admin',
                    message: recovery.whatFailed ?? errMsg,
                    statusCode: res.status,
                });
                const mediaRecovery = isMediaDependencyFailure([
                    errMsg,
                    recovery.whatFailed,
                    recovery.safeContinuation,
                    recovery.diagnostics,
                ].filter(Boolean).join(' '))
                    ? mediaDependencyRecoveryCopy(recovery.diagnostics ?? errMsg)
                    : null;
                const failureWithRecovery = {
                    ...failure,
                    summary: mediaRecovery?.summary ?? recovery.whatFailed ?? failure.summary,
                    recommendedAction: mediaRecovery?.recommendedAction ?? recovery.safeContinuation ?? failure.recommendedAction,
                    diagnostics: mediaRecovery?.diagnostics ?? recovery.diagnostics ?? failure.diagnostics,
                };
                if (res.status === 502 || res.status === 503 || mediaRecovery) {
                    console.warn('[CE-1] Confirm action blocked by runtime dependency:', errMsg);
                } else {
                    console.error('[CE-1] Confirm action failed:', errMsg);
                }
                set((s) => ({
                    missionChatError: failureWithRecovery.summary,
                    missionChatFailure: failureWithRecovery,
                    activeMode: 'blocker',
                    activeRunId: failureRunId ?? null,
                    missionChat: [
                        ...updateProposalLifecycle(s.missionChat, intentProofId, 'failed', {
                            mode: 'blocker',
                            run_id: failureRunId ?? undefined,
                        }),
                        {
                            role: 'council',
                            content: failureWithRecovery.summary,
                            source_node: 'admin',
                            mode: 'blocker',
                            run_id: failureRunId ?? undefined,
                            execution_summary: failureExecutionSummary,
                            thread_events: [failureThreadEvent({ runId: failureRunId, recovery })],
                            timestamp: new Date().toISOString(),
                        },
                    ],
                    pendingProposal: null,
                    activeConfirmToken: null,
                }));
                return { ok: false, runId: failureRunId ?? null, error: failureWithRecovery.summary };
            } catch (err) {
                const errMsg = err instanceof Error ? err.message : 'Confirm action failed';
                const failure = buildMissionChatFailure({
                    assistantName: get().assistantName,
                    targetId: 'admin',
                    message: errMsg,
                });
                console.error('[CE-1] confirmProposal error:', err);
                set((s) => ({
                    missionChatError: failure.summary,
                    missionChatFailure: failure,
                    activeMode: 'blocker',
                    activeRunId: null,
                    missionChat: [
                        ...updateProposalLifecycle(s.missionChat, intentProofId, 'failed', {
                            mode: 'blocker',
                        }),
                        { role: 'council', content: failure.summary, source_node: 'admin', mode: 'blocker' },
                    ],
                    pendingProposal: null,
                    activeConfirmToken: null,
                }));
                return { ok: false, runId: null, error: failure.summary };
            }
        },

        cancelProposal: (operatorReply?: string) => {
            const { pendingProposal } = get();
            if (pendingProposal?.intent_proof_id) {
                void fetch('/api/v1/intent/cancel-action', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ intent_proof_id: pendingProposal.intent_proof_id }),
                });
            }
            set((s) => {
                const conversationalReply = trimToNonEmpty(operatorReply);
                const missionChat = conversationalReply
                    ? [...s.missionChat, {
                        role: 'user' as const,
                        content: conversationalReply,
                        timestamp: new Date().toISOString(),
                    }]
                    : s.missionChat;
                return {
                    missionChat: pendingProposal
                        ? [
                            ...updateProposalLifecycle(missionChat, pendingProposal.intent_proof_id, 'cancelled', {
                                mode: 'proposal',
                            }),
                            {
                                role: 'system',
                                content: 'Proposal cancelled. No action executed.',
                                timestamp: new Date().toISOString(),
                            },
                        ]
                        : missionChat,
                    pendingProposal: null,
                    activeConfirmToken: null,
                    activeRunId: null,
                    activeMode: 'answer',
                    missionChatError: null,
                    missionChatFailure: null,
                };
            });
        },
    };
}
