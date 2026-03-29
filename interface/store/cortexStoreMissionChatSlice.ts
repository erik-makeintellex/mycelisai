import { buildMissionChatFailure, type MissionChatAvailability } from '@/lib/missionChatFailure';
import {
    delay,
    deriveMissionChatState,
    fallbackChatContent,
    isRetryableWorkspaceChatFailure,
} from '@/store/cortexStoreChatWorkflow';
import type { CortexState } from '@/store/cortexStoreState';
import type { APIResponse, ChatMessage, CTSChatEnvelope } from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import { clearPersistedChat, loadPersistedChat, normalizeProposalData } from '@/store/cortexStoreUtils';

export function createCortexMissionChatSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<
    CortexState,
    'sendMissionChat' | 'clearMissionChat' | 'setMissionChatScope' | 'broadcastToSwarm'
> {
    return {
        sendMissionChat: async (message: string) => {
            const trimmed = message.trim();
            if (!trimmed) return;

            const { councilTarget, assistantName, workspaceChatPrimed } = get();
            const isSomaRoute = councilTarget === 'admin';
            const chatRoute = isSomaRoute ? '/api/v1/chat' : `/api/v1/council/${councilTarget}/chat`;
            const routeLabel = isSomaRoute ? 'Soma chat' : 'Council agent';
            const allowSilentColdStartRetry = isSomaRoute && !workspaceChatPrimed;

            set((s) => ({
                missionChat: [...s.missionChat, { role: 'user', content: trimmed }],
                isMissionChatting: true,
                missionChatError: null,
                missionChatFailure: null,
            }));

            for (let attempt = 0; attempt < 2; attempt += 1) {
                const isRetryAttempt = attempt > 0;
                try {
                    const messages = [...get().missionChat]
                        .slice(-20)
                        .map((entry) => ({
                            role: entry.role === 'user' ? 'user' : 'assistant',
                            content: entry.content,
                        }));

                    const res = await fetch(chatRoute, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ messages }),
                    });

                    if (!res.ok) {
                        const text = await res.text();
                        let errMsg: string;
                        let availability: MissionChatAvailability | undefined;
                        try {
                            const parsed = JSON.parse(text);
                            errMsg = parsed.error || `${routeLabel} blocked (${res.status})`;
                            if (parsed.data && typeof parsed.data === 'object') {
                                availability = parsed.data as MissionChatAvailability;
                            }
                        } catch {
                            errMsg = `${routeLabel} unreachable (${res.status})`;
                        }

                        if (allowSilentColdStartRetry && !isRetryAttempt && isRetryableWorkspaceChatFailure(errMsg, res.status)) {
                            await delay(350);
                            continue;
                        }

                        set((s) => ({
                            isMissionChatting: false,
                            missionChatError: errMsg,
                            missionChatFailure: buildMissionChatFailure({
                                assistantName,
                                targetId: councilTarget,
                                message: errMsg,
                                statusCode: res.status,
                                availability,
                            }),
                            missionChat: [...s.missionChat, { role: 'council', content: errMsg, source_node: councilTarget, mode: 'blocker' }],
                            activeMode: 'blocker',
                            activeRole: councilTarget,
                        }));
                        return;
                    }

                    const body: APIResponse<CTSChatEnvelope> = await res.json();

                    if (!body.ok || !body.data) {
                        const errText = body.error || `${routeLabel} failed (${res.status})`;
                        const availability = body.data && typeof body.data === 'object'
                            ? body.data as MissionChatAvailability
                            : undefined;
                        set((s) => ({
                            isMissionChatting: false,
                            missionChatError: errText,
                            missionChatFailure: buildMissionChatFailure({
                                assistantName,
                                targetId: councilTarget,
                                message: errText,
                                statusCode: res.status,
                                availability,
                            }),
                            missionChat: [...s.missionChat, { role: 'council', content: errText, source_node: councilTarget, mode: 'blocker' }],
                            activeMode: 'blocker',
                            activeRole: councilTarget,
                        }));
                        return;
                    }

                    const envelope = body.data;
                    const chatMsg: ChatMessage = {
                        role: 'council',
                        content: fallbackChatContent(envelope.payload, assistantName),
                        consultations: envelope.payload.consultations,
                        tools_used: envelope.payload.tools_used,
                        source_node: envelope.meta.source_node,
                        trust_score: envelope.trust_score,
                        timestamp: envelope.meta.timestamp,
                        artifacts: envelope.payload.artifacts,
                        template_id: envelope.template_id || 'chat-to-answer',
                        mode: envelope.mode || 'answer',
                        provenance: envelope.payload?.provenance,
                        brain: envelope.payload?.brain,
                        proposal: normalizeProposalData(envelope.payload?.proposal),
                        proposal_status: envelope.payload?.proposal ? 'active' : undefined,
                    };

                    const trust = get().trustThreshold;
                    const govMode = trust >= 0.8 ? 'strict' as const : trust >= 0.5 ? 'active' as const : 'passive' as const;

                    set((s) => ({
                        isMissionChatting: false,
                        missionChat: [...s.missionChat, chatMsg],
                        missionChatFailure: null,
                        missionChatError: null,
                        workspaceChatPrimed: s.workspaceChatPrimed || isSomaRoute,
                        activeBrain: chatMsg.brain ?? null,
                        activeMode: chatMsg.mode || 'answer',
                        activeRole: chatMsg.source_node || '',
                        governanceMode: govMode,
                        ...(chatMsg.proposal ? {
                            pendingProposal: chatMsg.proposal,
                            activeConfirmToken: chatMsg.proposal.confirm_token,
                        } : {
                            pendingProposal: null,
                            activeConfirmToken: null,
                        }),
                    }));
                    return;
                } catch (err) {
                    const msg = err instanceof Error ? err.message : `${routeLabel} failed`;
                    if (allowSilentColdStartRetry && !isRetryAttempt && isRetryableWorkspaceChatFailure(msg)) {
                        await delay(350);
                        continue;
                    }

                    set((s) => ({
                        isMissionChatting: false,
                        missionChatError: msg,
                        missionChatFailure: buildMissionChatFailure({
                            assistantName,
                            targetId: councilTarget,
                            message: msg,
                        }),
                        missionChat: [...s.missionChat, { role: 'council', content: `Error: ${msg}`, source_node: councilTarget, mode: 'blocker' }],
                        activeMode: 'blocker',
                        activeRole: councilTarget,
                    }));
                    return;
                }
            }
        },

        clearMissionChat: () => {
            clearPersistedChat(get().workspaceChatScope);
            set({
                missionChat: [],
                missionChatError: null,
                missionChatFailure: null,
                workspaceChatPrimed: false,
                pendingProposal: null,
                activeConfirmToken: null,
                activeRunId: null,
                activeMode: 'answer',
            });
        },

        setMissionChatScope: (scope: string | null) => {
            const normalizedScope = typeof scope === 'string' && scope.trim() ? scope.trim() : null;
            if (get().workspaceChatScope === normalizedScope) {
                return;
            }

            const scopedMessages = loadPersistedChat(normalizedScope);
            const derivedState = deriveMissionChatState(scopedMessages);

            set({
                workspaceChatScope: normalizedScope,
                missionChat: scopedMessages,
                missionChatError: null,
                missionChatFailure: null,
                workspaceChatPrimed: false,
                pendingProposal: derivedState.pendingProposal,
                activeConfirmToken: derivedState.activeConfirmToken,
                activeRunId: derivedState.activeRunId,
                activeMode: derivedState.activeMode,
            });
        },

        broadcastToSwarm: async (message: string) => {
            const trimmed = message.trim();
            if (!trimmed) return;

            set((s) => ({
                missionChat: [...s.missionChat, { role: 'user', content: `[BROADCAST] ${trimmed}` }],
                isBroadcasting: true,
                missionChatError: null,
                missionChatFailure: null,
            }));

            try {
                const res = await fetch('/api/v1/swarm/broadcast', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ content: trimmed, source: 'mission-control' }),
                });

                if (!res.ok) {
                    const errText = `Broadcast failed (${res.status})`;
                    set((s) => ({
                        isBroadcasting: false,
                        missionChatError: errText,
                        missionChatFailure: null,
                        missionChat: [...s.missionChat, { role: 'architect', content: errText }],
                    }));
                    return;
                }

                const data = await res.json();
                const replyMessages: ChatMessage[] = [];
                if (Array.isArray(data.replies)) {
                    for (const reply of data.replies) {
                        if (reply.error) {
                            replyMessages.push({
                                role: 'architect',
                                content: `**${reply.team_id}**: _timed out or unavailable_`,
                                source_node: reply.team_id,
                            });
                        } else if (reply.content) {
                            replyMessages.push({
                                role: 'council',
                                content: reply.content,
                                source_node: reply.team_id,
                            });
                        }
                    }
                }
                if (replyMessages.length === 0) {
                    replyMessages.push({
                        role: 'architect',
                        content: `Broadcast sent to ${data.teams_hit} team(s) — no replies received.`,
                    });
                }
                set((s) => ({
                    isBroadcasting: false,
                    lastBroadcastResult: { teams_hit: data.teams_hit },
                    missionChat: [...s.missionChat, ...replyMessages],
                }));
            } catch (err) {
                const msg = err instanceof Error ? err.message : 'Broadcast failed';
                set((s) => ({
                    isBroadcasting: false,
                    missionChatError: msg,
                    missionChatFailure: null,
                    missionChat: [...s.missionChat, { role: 'architect', content: `Error: ${msg}` }],
                }));
            }
        },
    };
}
