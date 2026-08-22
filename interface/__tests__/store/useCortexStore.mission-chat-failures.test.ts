import { beforeEach, describe, expect, it, vi } from 'vitest';
import { buildMissionChatBlockerContent } from '@/store/cortexStoreMissionChatHelpers';
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
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe(
            'AI engine setup required. Soma is routed to an AI Engine that is configured but disabled. Next: Open Settings and enable a reachable AI Engine for Soma.',
        );
    });

    it('shows the image-generator setup reason and recovery step in the Soma thread', async () => {
        mockFetch.mockResolvedValue({
            ok: false,
            status: 503,
            text: async () => JSON.stringify({
                ok: false,
                error: 'Forge is open, but image generation is not enabled.',
                data: {
                    code: 'media_provider_not_ready',
                    summary: 'Forge is open, but image generation is not enabled.',
                    recommended_action: 'Enable API mode in Forge, restart Forge, then ask Soma to try again.',
                    setup_required: true,
                    setup_path: '/resources?section=capabilities',
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('create an image generation agent');

        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'workspace',
            type: 'setup_required',
            bannerLabel: 'Image generator setup required',
            setupPath: '/resources?section=capabilities',
        });
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe(
            'Image generator setup required. Forge is open, but image generation is not enabled. Next: Enable API mode in Forge, restart Forge, then ask Soma to try again.',
        );
    });

    it('cleans escaped whitespace from supplied blocker content', () => {
        const content = buildMissionChatBlockerContent({
            routeKind: 'workspace',
            targetId: 'admin',
            targetLabel: 'Soma',
            type: 'setup_required',
            title: 'Image Generator Setup Required',
            bannerLabel: 'Image generator setup required',
            summary: 'Forge is open, but image generation is not enabled.',
            recommendedAction: 'Enable API mode in Forge, restart Forge, then ask Soma to try again.',
            diagnostics: 'Forge is open, but image generation is not enabled.',
            setupPath: '/resources?section=capabilities',
        }, 'Image generator setup required. Review setup.\\\n&#x20;&#x20;');

        expect(content).toBe('Image generator setup required. Review setup.');
    });

    it('replaces obsolete generic operational-alert redirects with actionable workspace guidance', () => {
        const content = buildMissionChatBlockerContent({
            routeKind: 'workspace',
            targetId: 'admin',
            targetLabel: 'Soma',
            type: 'unreachable',
            title: 'Soma Chat Blocked',
            bannerLabel: 'Workspace chat unreachable',
            summary: 'Soma or the local API proxy is currently unreachable from this client.',
            recommendedAction: 'Open System Status and verify Core, NATS, and the local proxy are online.',
            diagnostics: 'Soma chat unreachable',
        }, 'Workspace chat unreachable. Review the operational alert for the safe next step.\\\n&#x20;&#x20;');

        expect(content).toBe(
            'Workspace chat unreachable. Soma or the local API proxy is currently unreachable from this client. Next: Open System Status and verify Core, NATS, and the local proxy are online.',
        );
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
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe(
            'Workspace chat server error. Soma hit a server-side failure while handling the request. Next: Retry once. If the failure persists, inspect System Status and recent startup logs.',
        );
        expect(useCortexStore.getState().missionChat.at(-1)?.content).not.toContain('Internal Server Error');
    });
});
