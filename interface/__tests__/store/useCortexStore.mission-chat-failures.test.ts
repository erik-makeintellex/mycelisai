import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore mission chat failures', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('stores a structured workspace failure when Soma chat returns 500', async () => {
        vi.useFakeTimers();
        try {
            mockFetch.mockResolvedValue({
                ok: false,
                status: 500,
                text: async () => '{"error":"Soma chat blocked (500)"}',
            });

            const sendPromise = useCortexStore.getState().sendMissionChat('hello');
            await vi.advanceTimersByTimeAsync(400);
            await sendPromise;

            expect(mockFetch).toHaveBeenCalledTimes(2);
            expect(useCortexStore.getState().activeMode).toBe('blocker');
            expect(useCortexStore.getState().missionChatError).toBe('Soma hit a server-side failure while handling the request.');
            expect(useCortexStore.getState().missionChatFailure).toMatchObject({
                routeKind: 'workspace',
                type: 'server_error',
                bannerLabel: 'Workspace chat server error',
            });
        } finally {
            vi.useRealTimers();
        }
    });

    it('stores a setup-required blocker when Soma has no bound AI engine', async () => {
        mockFetch.mockResolvedValue({
            ok: false,
            status: 503,
            text: async () => JSON.stringify({
                ok: false,
                error: 'Soma is routed to an AI Engine that is configured but disabled.',
                data: {
                    code: 'provider_disabled',
                    summary: 'Soma is routed to an AI Engine that is configured but disabled.',
                    recommended_action: 'Open Settings and enable a reachable AI Engine for Soma.',
                    setup_required: true,
                    setup_path: '/settings',
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('hello');

        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'workspace',
            type: 'setup_required',
            bannerLabel: 'AI engine setup required',
            setupPath: '/settings',
        });
    });

    it('routes Soma failures through the workspace contract when no council target is selected', async () => {
        useCortexStore.setState({
            councilTarget: 'admin',
            councilMembers: [],
        });
        mockFetch.mockRejectedValue(new Error('deadline exceeded'));

        await useCortexStore.getState().sendMissionChat('hello');

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/chat', expect.objectContaining({
            method: 'POST',
        }));
        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatError).toBe('Soma did not return a response before the request deadline.');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'workspace',
            targetId: 'admin',
            type: 'timeout',
            title: 'Soma Chat Blocked',
        });
    });

    it('stores a structured council timeout when a direct council request throws', async () => {
        useCortexStore.setState({ councilTarget: 'council-architect' });
        mockFetch.mockRejectedValue(new Error('deadline exceeded'));

        await useCortexStore.getState().sendMissionChat('hello');

        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'council',
            targetId: 'council-architect',
            type: 'timeout',
        });
    });

    it('routes direct council 503 failures through the council unreachable contract', async () => {
        useCortexStore.setState({ councilTarget: 'council-coder' });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 503,
            text: async () => '',
        });

        await useCortexStore.getState().sendMissionChat('hello');

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/council/council-coder/chat', expect.objectContaining({
            method: 'POST',
        }));
        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatError).toBe('The council member service or proxy is currently unreachable from this client.');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'council',
            targetId: 'council-coder',
            type: 'unreachable',
            title: 'Council Call Failed',
        });
    });

    it('does not store raw JSON response text as the final Soma answer', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                    signal_type: 'chat_response',
                    trust_score: 0.5,
                    template_id: 'chat-to-answer',
                    mode: 'answer',
                    payload: {
                        text: '{"tool":"consult_council","status":"ok"}',
                        ask_class: 'direct_answer',
                    },
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('hello');

        expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe(
            'Soma could not produce a readable reply for that request. Retry or ask Soma to summarize the result directly.',
        );
    });

    it('does not store raw council failure JSON as the final Soma answer', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                    signal_type: 'chat_response',
                    trust_score: 0.5,
                    template_id: 'chat-to-answer',
                    mode: 'answer',
                    payload: {
                        text: '{"error":"consult_council requires \\"member\\" and \\"question\\"","tool":"consult_council"}',
                        ask_class: 'direct_answer',
                    },
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('hello');

        expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe(
            'Soma could not produce a readable reply for that request. Retry or ask Soma to summarize the result directly.',
        );
        expect(useCortexStore.getState().missionChat.at(-1)?.content).not.toContain('consult_council requires');
    });

    it('normalizes plain transport failure text into a blocker instead of surfacing it', async () => {
        mockFetch.mockResolvedValue({
            ok: false,
            status: 500,
            text: async () => 'Internal Server Error',
        });

        await useCortexStore.getState().sendMissionChat('hello');

        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatError).toBe('Soma hit a server-side failure while handling the request.');
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe('Workspace chat server error. Review the operational alert for the safe next step.');
        expect(useCortexStore.getState().missionChat.at(-1)?.content).not.toContain('Internal Server Error');
    });
});
