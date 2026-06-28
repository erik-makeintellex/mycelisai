import {
    extractRunIdFromResponse,
    trimToNonEmpty,
    updateProposalLifecycle,
} from '@/store/cortexStoreChatWorkflow';
import { buildMissionChatFailure } from '@/lib/missionChatFailure';
import type { ChatMessage, ConfirmProposalResult } from '@/store/cortexStoreTypes';
import { approvalSentEvent, executionStartedEvent } from '@/store/cortexStoreProposalThreadEvents';
import { extractTeamWorkRefs, teamWorkMessage, type TeamWorkConfirmationRef } from '@/store/cortexStoreProposalTeamWorkRefs';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { ProposalData } from '@/store/cortexStoreTypesChat';

function recoveryTextFromExecutionSummary(summary: any) {
    const degradation = summary?.audit_recovery?.degradation;
    const whatFailed = trimToNonEmpty(degradation?.what_failed)
        ?? trimToNonEmpty(summary?.audit_recovery?.blocker);
    const safeContinuation = trimToNonEmpty(degradation?.safe_continuation);
    const diagnostics = [
        trimToNonEmpty(degradation?.code),
        whatFailed,
        trimToNonEmpty(degradation?.trusted_state),
        trimToNonEmpty(degradation?.invalidated_proof),
        safeContinuation,
    ].filter(Boolean).join(' | ');
    return {
        whatFailed,
        safeContinuation,
        diagnostics: trimToNonEmpty(diagnostics),
    };
}

function isMediaDependencyFailure(message?: string | null) {
    const lower = (message ?? '').toLowerCase();
    return lower.includes('comfyui')
        || lower.includes('media engine')
        || lower.includes('media capability')
        || lower.includes('local/private');
}

function mediaDependencyRecoveryCopy(diagnostics: string) {
    return {
        summary: 'Local media generation is not reachable, so Soma could not create the requested image output.',
        recommendedAction: 'Start or reconnect the configured ComfyUI upstream, then retry this proposal. If you only need files or text, ask Soma to rerun without image generation.',
        diagnostics,
    };
}

const proposalStartedDetail = 'Soma handed this to the work bus. You can keep talking here while updates arrive.';

function proposalStartedState(): NonNullable<ChatMessage['ui_response_state']> {
    return {
        kind: 'running',
        label: 'Started',
        detail: proposalStartedDetail,
        tone: 'info',
    };
}

function confirmedRunMessage(runId: string | null, summary?: string | null, teamWorkRefs: TeamWorkConfirmationRef[] = []) {
    const state = runId ? `Run ${runId.slice(0, 8)} started.` : 'Proposal approved.';
    const next = runId
        ? 'Soma handed this to the work bus and saved the run receipt.'
        : proposalStartedDetail;
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
        confirmProposal: async (proposalOverride?: ProposalData): Promise<ConfirmProposalResult> => {
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
            set((s) => ({
                activeMode: 'proposal',
                missionChatError: null,
                missionChatFailure: null,
                missionChat: updateProposalLifecycle(s.missionChat, intentProofId, 'confirmed_pending_execution', {
                    mode: 'proposal',
                    ui_response_state: proposalStartedState(),
                    thread_events: [approvalSentEvent()],
                }),
            }));
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
                        ?? trimToNonEmpty(body?.summary);
                    const lifecycle = runId ? 'executed' : 'confirmed_pending_execution';
                    const systemMsg: ChatMessage = {
                        role: 'system',
                        content: confirmedRunMessage(runId, proofSummary, teamWorkRefs),
                        mode: runId ? 'execution_result' : 'proposal',
                        ui_response_state: runId ? undefined : proposalStartedState(),
                        run_id: runId ?? undefined,
                        thread_events: [executionStartedEvent(runId, teamWorkRefs)],
                        execution_summary: body?.data?.execution_summary,
                        timestamp: new Date().toISOString(),
                    };
                    set((s) => ({
                        activeRunId: runId,
                        activeMode: runId ? 'execution_result' : 'proposal',
                        missionChatError: null,
                        missionChatFailure: null,
                        durableWorkRefreshVersion: s.durableWorkRefreshVersion + 1,
                        missionChat: [
                            ...updateProposalLifecycle(s.missionChat, intentProofId, lifecycle, {
                                mode: runId ? 'execution_result' : 'proposal',
                                ui_response_state: runId ? undefined : proposalStartedState(),
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
                let parsedBody: any = null;
                try {
                    parsedBody = JSON.parse(text);
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

        cancelProposal: () => {
            const { pendingProposal } = get();
            if (pendingProposal?.intent_proof_id) {
                void fetch('/api/v1/intent/cancel-action', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ intent_proof_id: pendingProposal.intent_proof_id }),
                });
            }
            set((s) => ({
                missionChat: pendingProposal
                    ? [
                        ...updateProposalLifecycle(s.missionChat, pendingProposal.intent_proof_id, 'cancelled', {
                            mode: 'proposal',
                        }),
                        {
                            role: 'system',
                            content: 'Proposal cancelled. No action executed.',
                            timestamp: new Date().toISOString(),
                        },
                    ]
                    : s.missionChat,
                pendingProposal: null,
                activeConfirmToken: null,
                activeRunId: null,
                activeMode: 'answer',
                missionChatError: null,
                missionChatFailure: null,
            }));
        },
    };
}
