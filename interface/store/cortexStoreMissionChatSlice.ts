import { buildMissionChatFailure, type MissionChatAvailability } from '@/lib/missionChatFailure';
import {
    delay,
    deriveMissionChatState,
    fallbackChatContent,
    isRetryableWorkspaceChatFailure,
} from '@/store/cortexStoreChatWorkflow';
import type { CortexState } from '@/store/cortexStoreState';
import type { APIResponse, ChatMessage, CTSChatEnvelope } from '@/store/cortexStoreTypes';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import { clearPersistedChat, loadPersistedChat, normalizeProposalData } from '@/store/cortexStoreUtils';

interface ChatRouteConfig {
    isSomaRoute: boolean;
    chatRoute: string;
    routeLabel: string;
    allowSilentColdStartRetry: boolean;
}

interface MissionChatBlockerOptions {
    assistantName: string;
    targetId: string;
    message: string;
    routeLabel: string;
    set: CortexSet;
    statusCode?: number;
    availability?: MissionChatAvailability;
    content?: string;
}

function buildChatRouteConfig(councilTarget: string, workspaceChatPrimed: boolean): ChatRouteConfig {
    const isSomaRoute = councilTarget === 'admin';
    return {
        isSomaRoute,
        chatRoute: isSomaRoute ? '/api/v1/chat' : `/api/v1/council/${councilTarget}/chat`,
        routeLabel: isSomaRoute ? 'Soma chat' : 'Council agent',
        allowSilentColdStartRetry: isSomaRoute && !workspaceChatPrimed,
    };
}

function buildRecentMissionMessages(messages: ChatMessage[]) {
    return [...messages]
        .slice(-20)
        .map((entry) => ({
            role: entry.role === 'user' ? 'user' : 'assistant',
            content: entry.content,
        }));
}

function resolveSelectedTeamContext(get: CortexGet): { id: string; name: string } | null {
    const { selectedTeamId, teamsDetail } = get();
    if (!selectedTeamId) {
        return null;
    }
    const team = teamsDetail.find((entry) => entry.id === selectedTeamId);
    if (!team) {
        return { id: selectedTeamId, name: selectedTeamId };
    }
    return {
        id: team.id,
        name: team.name || team.id,
    };
}

function resolveGovernanceMode(trustThreshold: number): CortexState['governanceMode'] {
    if (trustThreshold >= 0.8) return 'strict';
    if (trustThreshold >= 0.5) return 'active';
    return 'passive';
}

function setMissionChatBlocker({
    assistantName,
    targetId,
    message,
    routeLabel,
    set,
    statusCode,
    availability,
    content,
}: MissionChatBlockerOptions) {
    const failure = buildMissionChatFailure({
        assistantName,
        targetId,
        message,
        statusCode,
        availability,
    });
    set((s) => ({
        isMissionChatting: false,
        missionChatError: failure.summary,
        missionChatFailure: failure,
        missionChat: [...s.missionChat, {
            role: 'council',
            content: content ?? failure.summary,
            source_node: targetId,
            mode: 'blocker',
        }],
        activeMode: 'blocker',
        activeRole: targetId || routeLabel,
    }));
}

function setMissionChatSuccess(set: CortexSet, get: CortexGet, chatMsg: ChatMessage, isSomaRoute: boolean) {
    const govMode = resolveGovernanceMode(get().trustThreshold);
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
}

export function createCortexMissionChatSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<
    'sendMissionChat' | 'clearMissionChat' | 'setMissionChatScope' | 'broadcastToSwarm'
> {
    return {
        sendMissionChat: async (message: string) => {
            const trimmed = message.trim();
            if (!trimmed) return;

            const { councilTarget, assistantName, workspaceChatPrimed } = get();
            const {
                isSomaRoute,
                chatRoute,
                routeLabel,
                allowSilentColdStartRetry,
            } = buildChatRouteConfig(councilTarget, workspaceChatPrimed);

            set((s) => ({
                missionChat: [...s.missionChat, { role: 'user', content: trimmed }],
                isMissionChatting: true,
                missionChatError: null,
                missionChatFailure: null,
            }));

            for (let attempt = 0; attempt < 2; attempt += 1) {
                const isRetryAttempt = attempt > 0;
                try {
                    const messages = buildRecentMissionMessages(get().missionChat);
                    const teamContext = resolveSelectedTeamContext(get);

                    const res = await fetch(chatRoute, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            messages,
                            organization_id: get().workspaceChatScope ?? undefined,
                            team_id: teamContext?.id,
                            team_name: teamContext?.name,
                        }),
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
                            errMsg = text.trim() || `${routeLabel} unreachable (${res.status})`;
                        }

                        if (allowSilentColdStartRetry && !isRetryAttempt && isRetryableWorkspaceChatFailure(errMsg, res.status)) {
                            await delay(350);
                            continue;
                        }

                        setMissionChatBlocker({
                            assistantName,
                            targetId: councilTarget,
                            message: errMsg,
                            routeLabel,
                            set,
                            statusCode: res.status,
                            availability,
                        });
                        return;
                    }

                    const body: APIResponse<CTSChatEnvelope> = await res.json();

                    if (!body.ok || !body.data) {
                        const errText = body.error || `${routeLabel} failed (${res.status})`;
                        const availability = body.data && typeof body.data === 'object'
                            ? body.data as MissionChatAvailability
                            : undefined;
                        setMissionChatBlocker({
                            assistantName,
                            targetId: councilTarget,
                            message: errText,
                            routeLabel,
                            set,
                            statusCode: res.status,
                            availability,
                        });
                        return;
                    }

                    const envelope = body.data;
                    const chatMsg: ChatMessage = {
                        role: 'council',
                        content: fallbackChatContent(envelope.payload, assistantName),
                        consultations: envelope.payload.consultations,
                        tools_used: envelope.payload.tools_used,
                        ask_class: envelope.payload.ask_class,
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

                    setMissionChatSuccess(set, get, chatMsg, isSomaRoute);
                    return;
                } catch (err) {
                    const msg = err instanceof Error ? err.message : `${routeLabel} failed`;
                    if (allowSilentColdStartRetry && !isRetryAttempt && isRetryableWorkspaceChatFailure(msg)) {
                        await delay(350);
                        continue;
                    }

                    setMissionChatBlocker({
                        assistantName,
                        targetId: councilTarget,
                        message: msg,
                        routeLabel,
                        set,
                        content: `Error: ${msg}`,
                    });
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
                isMissionChatting: false,
                missionChatError: null,
                missionChatFailure: null,
                workspaceChatPrimed: false,
                isBroadcasting: false,
                lastBroadcastResult: null,
                pendingProposal: derivedState.pendingProposal,
                activeConfirmToken: derivedState.activeConfirmToken,
                activeRunId: derivedState.activeRunId,
                activeMode: derivedState.activeMode,
                activeRole: '',
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
