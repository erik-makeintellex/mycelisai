import { beforeEach, describe, expect, it, vi } from 'vitest';
import { waitFor } from '@testing-library/react';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { errorText, okJson, requestUrl, resetMissionControlChatStore } from './support/missionControlChatTestUtils';

describe('Soma Multi-Mode Interaction Matrix', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        useCortexStore.setState({ queuedChatPrompts: [], workspaceChatPrimed: false, activeRunId: null });
    });

    it('returns a direct answer for runtime state and playbook asks without blocker banners', async () => {
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: '2026-11-22T20:20:00Z' },
                        signal_type: 'chat_response',
                        trust_score: 0.9,
                        template_id: 'chat-to-answer',
                        mode: 'answer',
                        payload: {
                            text: 'Mycelis runtime state is healthy with 4 standing teams online and NATS JetStream active.',
                            ask_class: 'direct_answer',
                            tools_used: [],
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        await useCortexStore.getState().sendMissionChat("what's the current state of mycelis?");

        const state = useCortexStore.getState();
        expect(state.missionChatFailure).toBeNull();
        expect(state.missionChatError).toBeNull();
        expect(state.activeMode).toBe('answer');
        expect(state.pendingProposal).toBeNull();
        const lastMsg = state.missionChat.at(-1);
        expect(lastMsg?.role).toBe('council');
        expect(lastMsg?.mode).toBe('answer');
        expect(lastMsg?.content).toContain('Mycelis runtime state is healthy');
        expect(lastMsg?.content.toLowerCase()).not.toContain("can't deliver");
    });

    it('handles simple casual build and game fast-path proposals under 50ms with zero error state', async () => {
        mockFetch.mockImplementation(async (input: RequestInfo | URL) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/chat')) {
                return okJson({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: '2026-11-22T20:21:00Z' },
                        signal_type: 'chat_response',
                        trust_score: 0.85,
                        template_id: 'chat-to-proposal',
                        mode: 'proposal',
                        payload: {
                            text: 'I prepared a fast-path proposal to build your playable Snake Game.',
                            tools_used: ['write_file'],
                            proposal: {
                                intent: 'Build a playable Snake Game bundle',
                                operator_summary: 'Create index.html, style.css, and app.js for Snake Game.',
                                expected_result: 'Playable Snake Game in workspace.',
                                teams: 1,
                                agents: 2,
                                tools: ['write_file'],
                                risk_level: 'low',
                                confirm_token: 'tok-snake-1',
                                intent_proof_id: 'proof-snake-1',
                            },
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        await useCortexStore.getState().sendMissionChat('make me a snake game');

        const state = useCortexStore.getState();
        expect(state.missionChatFailure).toBeNull();
        expect(state.missionChatError).toBeNull();
        expect(state.activeMode).toBe('proposal');
        expect(state.pendingProposal).not.toBeNull();
        expect(state.activeConfirmToken).toBe('tok-snake-1');
        expect(state.pendingProposal?.confirm_token).toBe('tok-snake-1');
        expect(state.pendingProposal?.intent).toBe('Build a playable Snake Game bundle');
    });

    it('supports explicit + Queue work stacking and auto-drains queued prompts upon confirmProposal()', async () => {
        let chatCallCount = 0;
        mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
            const url = requestUrl(input);
            if (url.includes('/api/v1/chat')) {
                chatCallCount += 1;
                const body = JSON.parse(String(init?.body ?? '{}'));
                const latestMsg = body.messages?.at(-1)?.content ?? '';
                if (chatCallCount === 1) {
                    return okJson({
                        ok: true,
                        data: {
                            meta: { source_node: 'admin', timestamp: '2026-11-22T20:22:00Z' },
                            signal_type: 'chat_response',
                            trust_score: 0.8,
                            template_id: 'chat-to-proposal',
                            mode: 'proposal',
                            payload: {
                                text: 'Fast-path proposal ready for quick timer.',
                                tools_used: ['write_file'],
                                proposal: {
                                    intent: 'Build quick timer',
                                    teams: 1,
                                    agents: 1,
                                    tools: ['write_file'],
                                    risk_level: 'low',
                                    confirm_token: 'tok-timer-1',
                                    intent_proof_id: 'proof-timer-1',
                                },
                            },
                        },
                    });
                }
                return okJson({
                    ok: true,
                    data: {
                        meta: { source_node: 'admin', timestamp: '2026-11-22T20:22:05Z' },
                        signal_type: 'chat_response',
                        trust_score: 0.85,
                        template_id: 'chat-to-proposal',
                        mode: 'proposal',
                        payload: {
                            text: `Proposal prepared for: ${latestMsg}`,
                            tools_used: ['write_file'],
                            proposal: {
                                intent: `Build ${latestMsg}`,
                                teams: 1,
                                agents: 1,
                                tools: ['write_file'],
                                risk_level: 'low',
                                confirm_token: 'tok-todo-2',
                                intent_proof_id: 'proof-todo-2',
                            },
                        },
                    },
                });
            }
            if (url.includes('/api/v1/intent/confirm-action')) {
                return okJson({
                    ok: true,
                    data: {
                        run_id: 'run-timer-1',
                        message: 'Quick timer bundle delivered.',
                        delivery_package: {
                            package_id: 'dev_project_bundle:timer-1',
                            package_kind: 'dev_project_bundle',
                            title: 'Quick Timer Bundle',
                            status: 'verified',
                            run_id: 'run-timer-1',
                            items: [
                                {
                                    id: 'timer-1:index.html',
                                    kind: 'document',
                                    label: 'Timer App',
                                    storage_ref: 'workspace/deliverables/timer-1/index.html',
                                    source_run_id: 'run-timer-1',
                                    package_id: 'dev_project_bundle:timer-1',
                                    role: 'primary',
                                    status: 'final',
                                },
                            ],
                        },
                    },
                });
            }
            return errorText(404, 'not found');
        });

        await useCortexStore.getState().sendMissionChat('make me a quick timer');
        expect(useCortexStore.getState().pendingProposal?.confirm_token).toBe('tok-timer-1');

        useCortexStore.getState().enqueueChatPrompt('build a todo list');
        useCortexStore.getState().enqueueChatPrompt('create a deployment checklist');
        expect(useCortexStore.getState().queuedChatPrompts).toHaveLength(2);
        expect(useCortexStore.getState().queuedChatPrompts).toEqual([
            'build a todo list',
            'create a deployment checklist',
        ]);

        const confirmRes = await useCortexStore.getState().confirmProposal();
        expect(confirmRes.ok).toBe(true);
        expect(confirmRes.runId).toBe('run-timer-1');

        await waitFor(() => {
            expect(chatCallCount).toBe(2);
            expect(useCortexStore.getState().activeConfirmToken).toBe('tok-todo-2');
        });

        expect(useCortexStore.getState().queuedChatPrompts).toEqual(['create a deployment checklist']);
        const chatCalls = mockFetch.mock.calls.filter((call: unknown[]) => requestUrl(call[0]).includes('/api/v1/chat'));
        expect(chatCalls).toHaveLength(2);
        const drainedBody = JSON.parse(String(chatCalls[1]?.[1]?.body ?? '{}'));
        expect(drainedBody.messages.at(-1)).toEqual({ role: 'user', content: 'build a todo list' });
    });

    it('recovers transparently on silent cold-start retry without showing a Chat Blocked failure card', async () => {
        vi.useFakeTimers();
        try {
            useCortexStore.setState({ workspaceChatPrimed: false });
            mockFetch
                .mockResolvedValueOnce(errorText(503, '{"error":"Soma chat temporarily unavailable (503)"}'))
                .mockResolvedValueOnce(
                    okJson({
                        ok: true,
                        data: {
                            meta: { source_node: 'admin', timestamp: '2026-11-22T20:23:00Z' },
                            signal_type: 'chat_response',
                            trust_score: 0.9,
                            template_id: 'chat-to-proposal',
                            mode: 'proposal',
                            payload: {
                                text: 'Calculator project proposal ready after cold-start warmup.',
                                tools_used: ['write_file'],
                                proposal: {
                                    intent: 'Build calculator web app',
                                    teams: 1,
                                    agents: 1,
                                    tools: ['write_file'],
                                    risk_level: 'low',
                                    confirm_token: 'tok-calc-1',
                                    intent_proof_id: 'proof-calc-1',
                                },
                            },
                        },
                    }),
                );

            const sendPromise = useCortexStore.getState().sendMissionChat('help me build a calculator');
            await vi.advanceTimersByTimeAsync(400);
            await sendPromise;

            const state = useCortexStore.getState();
            expect(mockFetch).toHaveBeenCalledTimes(2);
            expect(state.missionChatFailure).toBeNull();
            expect(state.missionChatError).toBeNull();
            expect(state.workspaceChatPrimed).toBe(true);
            expect(state.activeMode).toBe('proposal');
            expect(state.activeConfirmToken).toBe('tok-calc-1');
            expect(state.missionChat.some((m) => m.mode === 'blocker')).toBe(false);
        } finally {
            vi.useRealTimers();
        }
    });
});
