import {
    extractRunIdFromResponse,
    trimToNonEmpty,
    updateProposalLifecycle,
} from '@/store/cortexStoreChatWorkflow';
import type { CortexState } from '@/store/cortexStoreState';
import type { ChatMessage, ConfirmProposalResult } from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';

export function createCortexProposalExecutionSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<CortexState, 'confirmProposal' | 'cancelProposal'> {
    return {
        confirmProposal: async (): Promise<ConfirmProposalResult> => {
            const { activeConfirmToken, pendingProposal } = get();
            if (!activeConfirmToken || !pendingProposal) {
                return {
                    ok: false,
                    runId: null,
                    error: 'No pending proposal to confirm',
                };
            }
            try {
                const res = await fetch('/api/v1/intent/confirm-action', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ confirm_token: activeConfirmToken }),
                });
                if (res.ok) {
                    const body = await res.json();
                    const runId = extractRunIdFromResponse(body);
                    const proofSummary = trimToNonEmpty(body?.data?.message)
                        ?? trimToNonEmpty(body?.message)
                        ?? trimToNonEmpty(body?.data?.summary)
                        ?? trimToNonEmpty(body?.summary);
                    const lifecycle = runId ? 'executed' : 'confirmed_pending_execution';
                    const systemMsg: ChatMessage = {
                        role: 'system',
                        content: proofSummary ?? (runId ? 'Mission activated' : 'Proposal confirmed. Waiting for execution proof.'),
                        mode: runId ? 'execution_result' : 'proposal',
                        run_id: runId ?? undefined,
                        timestamp: new Date().toISOString(),
                    };
                    set((s) => ({
                        activeRunId: runId,
                        activeMode: runId ? 'execution_result' : 'proposal',
                        missionChatError: null,
                        missionChatFailure: null,
                        missionChat: [
                            ...updateProposalLifecycle(s.missionChat, pendingProposal.intent_proof_id, lifecycle, {
                                mode: runId ? 'execution_result' : 'proposal',
                                run_id: runId ?? undefined,
                            }),
                            systemMsg,
                        ],
                        pendingProposal: null,
                        activeConfirmToken: null,
                    }));
                    return { ok: true, runId };
                }

                const text = await res.text();
                let errMsg = 'Confirm action failed';
                try {
                    const parsed = JSON.parse(text);
                    errMsg = parsed.error || errMsg;
                } catch {
                    errMsg = text || errMsg;
                }
                console.error('[CE-1] Confirm action failed:', errMsg);
                set((s) => ({
                    missionChatError: errMsg,
                    missionChatFailure: null,
                    activeMode: 'blocker',
                    activeRunId: null,
                    missionChat: [
                        ...updateProposalLifecycle(s.missionChat, pendingProposal.intent_proof_id, 'failed', {
                            mode: 'blocker',
                        }),
                        { role: 'council', content: errMsg, source_node: 'admin', mode: 'blocker' },
                    ],
                    pendingProposal: null,
                    activeConfirmToken: null,
                }));
                return { ok: false, runId: null, error: errMsg };
            } catch (err) {
                const errMsg = err instanceof Error ? err.message : 'Confirm action failed';
                console.error('[CE-1] confirmProposal error:', err);
                set((s) => ({
                    missionChatError: errMsg,
                    missionChatFailure: null,
                    activeMode: 'blocker',
                    activeRunId: null,
                    missionChat: [
                        ...updateProposalLifecycle(s.missionChat, pendingProposal.intent_proof_id, 'failed', {
                            mode: 'blocker',
                        }),
                        { role: 'council', content: errMsg, source_node: 'admin', mode: 'blocker' },
                    ],
                    pendingProposal: null,
                    activeConfirmToken: null,
                }));
                return { ok: false, runId: null, error: errMsg };
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
