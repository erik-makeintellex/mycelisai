import { buildMissionChatFailure, type MissionChatAvailability } from "@/lib/missionChatFailure";
import type { ChatMessage } from "@/store/cortexStoreTypes";
import type { CortexGet, CortexSet } from "@/store/cortexStoreSliceTypes";
import type { CortexState } from "@/store/cortexStoreState";

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

export function buildChatRouteConfig(councilTarget: string, workspaceChatPrimed: boolean): ChatRouteConfig {
    const isSomaRoute = councilTarget === "admin";
    return {
        isSomaRoute,
        chatRoute: isSomaRoute ? "/api/v1/chat" : `/api/v1/council/${councilTarget}/chat`,
        routeLabel: isSomaRoute ? "Soma chat" : "Council agent",
        allowSilentColdStartRetry: isSomaRoute && !workspaceChatPrimed,
    };
}

export function buildRecentMissionMessages(messages: ChatMessage[]) {
    return [...messages]
        .slice(-20)
        .map((entry) => ({
            role: entry.role === "user" ? "user" : "assistant",
            content: entry.content,
        }));
}

export function resolveSelectedTeamContext(get: CortexGet): { id: string; name: string } | null {
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

export function setMissionChatBlocker({
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
            role: "council",
            content: content ?? failure.summary,
            source_node: targetId,
            mode: "blocker",
        }],
        activeMode: "blocker",
        activeRole: targetId || routeLabel,
    }));
}

export function setMissionChatSuccess(set: CortexSet, get: CortexGet, chatMsg: ChatMessage, isSomaRoute: boolean) {
    const govMode = resolveGovernanceMode(get().trustThreshold);
    set((s) => ({
        isMissionChatting: false,
        missionChat: [...s.missionChat, chatMsg],
        missionChatFailure: null,
        missionChatError: null,
        workspaceChatPrimed: s.workspaceChatPrimed || isSomaRoute,
        activeBrain: chatMsg.brain ?? null,
        activeMode: chatMsg.mode || "answer",
        activeRole: chatMsg.source_node || "",
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

function resolveGovernanceMode(trustThreshold: number): CortexState["governanceMode"] {
    if (trustThreshold >= 0.8) return "strict";
    if (trustThreshold >= 0.5) return "active";
    return "passive";
}
