import {
    extractRunIdFromResponse,
    trimToNonEmpty,
    updateProposalLifecycle,
} from '@/store/cortexStoreChatWorkflow';
import { buildMissionChatFailure } from '@/lib/missionChatFailure';
import type {
    ChatArtifactRef,
    ChatMessage,
    ConfirmProposalResult,
    DeliveryPackageSummary,
    ExecutionSummary,
    OutputRef,
} from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import { normalizeOutputRefs } from '@/store/cortexStoreOutputRefs';

function extractArtifactsFromConfirmResponse(raw: unknown): ChatArtifactRef[] {
    if (!raw || typeof raw !== 'object') {
        return [];
    }
    const record = raw as Record<string, unknown>;
    const data = record.data && typeof record.data === 'object' ? record.data as Record<string, unknown> : null;
    const artifacts = data?.artifacts;
    return Array.isArray(artifacts) ? artifacts.filter((artifact): artifact is ChatArtifactRef => (
        artifact != null
        && typeof artifact === 'object'
        && typeof (artifact as ChatArtifactRef).type === 'string'
        && typeof (artifact as ChatArtifactRef).title === 'string'
    )) : [];
}

function extractDeliveryPackageFromConfirmResponse(raw: unknown): DeliveryPackageSummary | undefined {
    if (!raw || typeof raw !== 'object') {
        return undefined;
    }
    const record = raw as Record<string, unknown>;
    const data = record.data && typeof record.data === 'object' ? record.data as Record<string, unknown> : null;
    const deliveryPackage = data?.delivery_package ?? data?.package;
    if (!deliveryPackage || typeof deliveryPackage !== 'object') {
        return undefined;
    }
    const candidate = deliveryPackage as Partial<DeliveryPackageSummary>;
    return typeof candidate.package_kind === 'string' ? candidate as DeliveryPackageSummary : undefined;
}

function extractExecutionSummaryFromConfirmResponse(raw: unknown): ExecutionSummary | undefined {
    if (!raw || typeof raw !== 'object') {
        return undefined;
    }
    const record = raw as Record<string, unknown>;
    const data = record.data && typeof record.data === 'object' ? record.data as Record<string, unknown> : null;
    const summary = data?.execution_summary;
    if (!summary || typeof summary !== 'object') {
        return undefined;
    }
    const candidate = summary as Partial<ExecutionSummary>;
    return typeof candidate.title === 'string' ? candidate as ExecutionSummary : undefined;
}

function extractOutputRefsFromConfirmResponse(raw: unknown): OutputRef[] {
    if (!raw || typeof raw !== 'object') {
        return [];
    }
    const record = raw as Record<string, unknown>;
    const data = record.data && typeof record.data === 'object' ? record.data as Record<string, unknown> : null;
    return Array.isArray(data?.output_refs) ? data.output_refs as OutputRef[] : [];
}

export function createCortexProposalExecutionSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<'confirmProposal' | 'cancelProposal'> {
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
                    const artifacts = extractArtifactsFromConfirmResponse(body);
                    const deliveryPackage = extractDeliveryPackageFromConfirmResponse(body);
                    const executionSummary = extractExecutionSummaryFromConfirmResponse(body);
                    const explicitOutputRefs = extractOutputRefsFromConfirmResponse(body);
                    const outputRefs = explicitOutputRefs.length > 0
                        ? explicitOutputRefs
                        : normalizeOutputRefs({ artifacts, deliveryPackage, executionSummary, runId });
                    const lifecycle = runId ? 'executed' : 'confirmed_pending_execution';
                    const systemMsg: ChatMessage = {
                        role: 'system',
                        content: proofSummary ?? (runId ? 'Mission activated' : 'Proposal confirmed. Waiting for execution proof.'),
                        mode: runId ? 'execution_result' : 'proposal',
                        run_id: runId ?? undefined,
                        timestamp: new Date().toISOString(),
                    };
                    const packageMsg: ChatMessage | null = artifacts.length > 0 || deliveryPackage || outputRefs.length > 0
                        ? {
                            role: 'council',
                            content: proofSummary
                                ?? trimToNonEmpty(executionSummary?.summary)
                                ?? 'Codex task bundle created with retained package proof.',
                            mode: 'execution_result',
                            run_id: runId ?? undefined,
                            artifacts,
                            delivery_package: deliveryPackage,
                            output_refs: outputRefs,
                            execution_summary: executionSummary,
                            tools_used: ['write_file'],
                            timestamp: new Date().toISOString(),
                        }
                        : null;
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
                            ...(packageMsg ? [packageMsg] : []),
                        ],
                        pendingProposal: null,
                        activeConfirmToken: null,
                    }));
                    if (!get().isMissionChatting && (get().queuedChatPrompts ?? []).length > 0) {
                        get().dispatchNextQueuedPrompt?.();
                    }
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
                const failure = buildMissionChatFailure({
                    assistantName: get().assistantName,
                    targetId: 'admin',
                    message: errMsg,
                    statusCode: res.status,
                });
                console.error('[CE-1] Confirm action failed:', errMsg);
                set((s) => ({
                    missionChatError: failure.summary,
                    missionChatFailure: failure,
                    activeMode: 'blocker',
                    activeRunId: null,
                    missionChat: [
                        ...updateProposalLifecycle(s.missionChat, pendingProposal.intent_proof_id, 'failed', {
                            mode: 'blocker',
                        }),
                        { role: 'council', content: failure.summary, source_node: 'admin', mode: 'blocker' },
                    ],
                    pendingProposal: null,
                    activeConfirmToken: null,
                }));
                return { ok: false, runId: null, error: failure.summary };
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
                        ...updateProposalLifecycle(s.missionChat, pendingProposal.intent_proof_id, 'failed', {
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
