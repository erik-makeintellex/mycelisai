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
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { ProposalData } from '@/store/cortexStoreTypesChat';
import type { ExecutionSummaryData } from '@/store/cortexStoreTypesExecutionSummary';

type ConfirmFailureBody = {
    error?: string;
    data?: {
        run_id?: string;
        execution_summary?: ExecutionSummaryData;
    };
};

function recoveryTextFromExecutionSummary(summary: ExecutionSummaryData | undefined) {
    const auditRecovery = summary?.audit_recovery && typeof summary.audit_recovery === 'object'
        ? summary.audit_recovery
        : undefined;
    const degradation = auditRecovery?.degradation;
    const code = trimToNonEmpty(degradation?.code);
    const whatFailed = trimToNonEmpty(degradation?.what_failed)
        ?? trimToNonEmpty(auditRecovery?.blocker);
    const safeContinuation = trimToNonEmpty(degradation?.safe_continuation);
    const diagnostics = [
        code,
        whatFailed,
        trimToNonEmpty(degradation?.trusted_state),
        trimToNonEmpty(degradation?.invalidated_proof),
        safeContinuation,
    ].filter(Boolean).join(' | ');
    return {
        code,
        whatFailed,
        safeContinuation,
        diagnostics: trimToNonEmpty(diagnostics),
    };
}

function failureThreadEvent({
    runId,
    recovery,
}: {
    runId?: string | null;
    recovery: ReturnType<typeof recoveryTextFromExecutionSummary>;
}): NonNullable<ChatMessage['thread_events']>[number] {
    const code = recovery.code ?? 'approved_execution_failed';
    const contractUnsatisfied = code === 'result_contract_unsatisfied';
    return {
        kind: 'attention_required',
        label: contractUnsatisfied ? 'Output is not playable yet' : 'Soma needs your direction',
        detail: contractUnsatisfied
            ? 'The team stopped because the required runnable output was not validated. No playable output should be trusted for this attempt.'
            : recovery.whatFailed ?? 'This work stopped before a usable result was produced.',
        tone: 'warning',
        status: code,
        run_id: runId ?? undefined,
        source_kind: 'web_api',
        source_channel: 'api.intent.confirm-action',
        payload_kind: 'soma_thread_event',
        target_reference: code,
        timestamp: new Date().toISOString(),
    };
}

function isMediaDependencyFailure(message?: string | null) {
    const lower = (message ?? '').toLowerCase();
    return lower.includes('comfyui')
        || lower.includes('forge')
        || lower.includes('media engine')
        || lower.includes('media capability')
        || lower.includes('local/private');
}

function mediaDependencyRecoveryCopy(diagnostics: string) {
    const forgeAPIDisabled = diagnostics.toLowerCase().includes('forge')
        && (diagnostics.toLowerCase().includes('api mode') || diagnostics.toLowerCase().includes('api access is off'));
    return {
        summary: forgeAPIDisabled
            ? 'Forge is open, but Soma cannot use image generation until its API mode is enabled.'
            : 'The configured image generator is not ready, so Soma could not create the requested image output.',
        recommendedAction: forgeAPIDisabled
            ? "Enable API mode in Forge's Pinokio launch settings, restart Forge, then tell Soma to try the image again."
            : 'Start or reconnect the configured image generator, then tell Soma to try the image again.',
        diagnostics,
    };
}

function confirmedRunMessage(runId: string | null, summary?: string | null, teamWorkRefs: TeamWorkConfirmationRef[] = []) {
    const state = runId ? `Run ${runId.slice(0, 8)} started.` : 'Proposal approved.';
    const next = runId
        ? 'Soma handed this to the work bus. This is running, not a completed result or proof.'
        : proposalStartedState().detail;
    return [state, next, teamWorkMessage(teamWorkRefs), summary].filter(Boolean).join(' ');
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
                        ui_response_state: configAction ? configurationPendingState(configAction) : proposalStartedState(),
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
                    const completedConfigAction = configAction && confirmationIsCompleted(body, res.status)
                        ? configAction
                        : null;
                    const lifecycle = completedConfigAction || runId ? 'executed' : 'confirmed_pending_execution';
                    const systemMsg: ChatMessage = {
                        role: 'system',
                        content: completedConfigAction
                            ? configurationCompletedMessage(completedConfigAction, proofSummary)
                            : confirmedRunMessage(runId, proofSummary, teamWorkRefs),
                        mode: completedConfigAction || runId ? 'execution_result' : 'proposal',
                        ui_response_state: completedConfigAction
                            ? configurationCompletedState(completedConfigAction)
                            : runId ? undefined : proposalStartedState(),
                        run_id: runId ?? undefined,
                        thread_events: [completedConfigAction
                            ? configurationCompletedEvent(completedConfigAction)
                            : executionStartedEvent(runId, teamWorkRefs)],
                        execution_summary: body?.data?.execution_summary,
                        timestamp: new Date().toISOString(),
                    };
                    set((s) => ({
                        activeRunId: completedConfigAction ? null : runId,
                        activeMode: completedConfigAction || runId ? 'execution_result' : 'proposal',
                        missionChatError: null,
                        missionChatFailure: null,
                        durableWorkRefreshVersion: s.durableWorkRefreshVersion + 1,
                        missionChat: [
                            ...updateProposalLifecycle(s.missionChat, intentProofId, lifecycle, {
                                mode: completedConfigAction || runId ? 'execution_result' : 'proposal',
                                ui_response_state: completedConfigAction
                                    ? configurationCompletedState(completedConfigAction)
                                    : runId ? undefined : proposalStartedState(),
                                run_id: runId ?? undefined,
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
