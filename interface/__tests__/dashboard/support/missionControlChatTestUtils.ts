import { act } from '@testing-library/react';
import { useCortexStore } from '@/store/useCortexStore';

export const COUNCIL_MEMBERS = [
    { id: 'admin', role: 'admin', team: 'admin-core' },
    { id: 'council-architect', role: 'architect', team: 'council-core' },
    { id: 'council-coder', role: 'coder', team: 'council-core' },
    { id: 'council-creative', role: 'creative', team: 'council-core' },
    { id: 'council-sentry', role: 'sentry', team: 'council-core' },
];

export const CTS_CHAT_RESPONSE = {
    ok: true,
    data: {
        meta: { source_node: 'admin', timestamp: '2026-02-16T12:00:00Z' },
        signal_type: 'chat_response',
        trust_score: 0.5,
        payload: { text: 'Hello from admin agent', consultations: null, tools_used: null },
    },
};

export function requestUrl(input: unknown): string {
    if (typeof input === 'string') {
        return input;
    }
    if (typeof Request !== 'undefined' && input instanceof Request) {
        return input.url;
    }
    if (typeof URL !== 'undefined' && input instanceof URL) {
        return input.toString();
    }
    if (input && typeof input === 'object' && 'url' in input && typeof (input as { url?: unknown }).url === 'string') {
        return (input as { url: string }).url;
    }
    return String(input ?? '');
}

export function okJson(body: unknown) {
    return {
        ok: true,
        status: 200,
        json: async () => body,
        text: async () => JSON.stringify(body),
    } as any;
}

export function errorText(status: number, text: string) {
    return {
        ok: false,
        status,
        text: async () => text,
    } as any;
}

export function resetMissionControlChatStore() {
    useCortexStore.setState({
        missionChat: [],
        workspaceChatScope: null,
        isMissionChatting: false,
        missionChatError: null,
        missionChatFailure: null,
        activeMode: 'answer',
        activeRole: '',
        assistantName: 'Soma',
        councilTarget: 'admin',
        councilMembers: [],
        pendingProposal: null,
        activeConfirmToken: null,
        isBroadcasting: false,
        lastBroadcastResult: null,
        streamLogs: [],
        selectedTeamId: null,
        teamsDetail: [],
    });
}

export async function settleMissionControlChat() {
    await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 0));
    });
}
