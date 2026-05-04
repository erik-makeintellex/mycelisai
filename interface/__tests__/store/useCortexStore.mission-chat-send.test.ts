import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore mission chat send', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('silently retries the first transient Soma failure and recovers on the second attempt', async () => {
        vi.useFakeTimers();
        try {
            mockFetch
                .mockResolvedValueOnce({
                    ok: false,
                    status: 500,
                    text: async () => '{"error":"Soma chat blocked (500)"}',
                })
                .mockResolvedValueOnce({
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
                                text: 'Recovered answer.',
                                tools_used: [],
                            },
                        },
                    }),
                });

            const sendPromise = useCortexStore.getState().sendMissionChat('hello');
            await vi.advanceTimersByTimeAsync(400);
            await sendPromise;

            expect(mockFetch).toHaveBeenCalledTimes(2);
            expect(useCortexStore.getState().missionChatFailure).toBeNull();
            expect(useCortexStore.getState().missionChatError).toBeNull();
            expect(useCortexStore.getState().workspaceChatPrimed).toBe(true);
            expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe('Recovered answer.');
        } finally {
            vi.useRealTimers();
        }
    });

    it('retries a cold-start Soma failure after a scope change even when a prior session was primed', async () => {
        vi.useFakeTimers();
        try {
            useCortexStore.setState({
                workspaceChatPrimed: true,
                missionChat: [],
            });
            useCortexStore.getState().setMissionChatScope('org-123');
            mockFetch
                .mockResolvedValueOnce({
                    ok: false,
                    status: 500,
                    text: async () => '{"error":"Soma chat blocked (500)"}',
                })
                .mockResolvedValueOnce({
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
                                text: 'Recovered answer after scoped reset.',
                                tools_used: [],
                            },
                        },
                    }),
                });

            const sendPromise = useCortexStore.getState().sendMissionChat('hello');
            await vi.advanceTimersByTimeAsync(400);
            await sendPromise;

            expect(mockFetch).toHaveBeenCalledTimes(2);
            expect(useCortexStore.getState().missionChatFailure).toBeNull();
            expect(useCortexStore.getState().workspaceChatPrimed).toBe(true);
            expect(useCortexStore.getState().missionChat.at(-1)?.content).toBe('Recovered answer after scoped reset.');
        } finally {
            vi.useRealTimers();
        }
    });

    it('normalizes team expressions and module bindings from proposal payload', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                    signal_type: 'chat_response',
                    trust_score: 0.5,
                    template_id: 'chat-to-proposal',
                    mode: 'proposal',
                    payload: {
                        text: 'I prepared a governed execution plan.',
                        tools_used: ['delegate'],
                        proposal: {
                            intent: 'chat-action',
                            tools: ['delegate'],
                            risk_level: 'medium',
                            confirm_token: 'ct-123',
                            intent_proof_id: 'ip-123',
                            team_expressions: [
                                {
                                    expression_id: 'expr-1',
                                    team_id: 'admin-core',
                                    objective: 'Execute delegate through governed module binding',
                                    role_plan: ['admin'],
                                    module_bindings: [
                                        {
                                            binding_id: 'binding-1-delegate',
                                            module_id: 'delegate',
                                            adapter_kind: 'internal',
                                            operation: 'delegate',
                                        },
                                    ],
                                },
                            ],
                        },
                    },
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('launch a team');

        expect(useCortexStore.getState().activeMode).toBe('proposal');
        expect(useCortexStore.getState().activeConfirmToken).toBe('ct-123');
        expect(useCortexStore.getState().pendingProposal).toMatchObject({
            intent: 'chat-action',
            teams: 1,
            agents: 1,
            tools: ['delegate'],
            confirm_token: 'ct-123',
            intent_proof_id: 'ip-123',
        });
        expect(useCortexStore.getState().pendingProposal?.team_expressions?.[0]).toMatchObject({
            expression_id: 'expr-1',
            team_id: 'admin-core',
            objective: 'Execute delegate through governed module binding',
        });
        expect(useCortexStore.getState().pendingProposal?.team_expressions?.[0].module_bindings?.[0]).toMatchObject({
            module_id: 'delegate',
            adapter_kind: 'internal',
        });
        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            proposal_status: 'active',
        });
    });

    it('stores governed artifact ask-class metadata for artifact-bearing Soma answers', async () => {
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
                        text: 'I prepared a brief for review.',
                        ask_class: 'governed_artifact',
                        artifacts: [
                            { id: 'doc-1', type: 'document', title: 'Creative Brief', content_type: 'text/markdown', content: '# Brief' },
                        ],
                    },
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('create a brief');

        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            ask_class: 'governed_artifact',
        });
    });

    it('stores specialist consultation ask-class metadata for consulted answers', async () => {
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
                        text: 'The architect reviewed the tradeoffs.',
                        ask_class: 'specialist_consultation',
                        consultations: [
                            { member: 'council-architect', summary: 'Recommend the safer option.' },
                        ],
                    },
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('review the architecture');

        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            ask_class: 'specialist_consultation',
        });
    });
});
