import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore confirm proposal execution', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('records an execution result and run id on successful confirmation', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Launch a docs crew',
                teams: 1,
                agents: 2,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
            },
            activeConfirmToken: 'ct-123',
            missionChat: [{
                role: 'council',
                content: 'Proposed execution path',
                mode: 'proposal',
                proposal: {
                    intent: 'Launch a docs crew',
                    teams: 1,
                    agents: 2,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                proposal_status: 'active',
            }],
            missionChatError: null,
            activeMode: 'proposal',
            activeRunId: null,
        });
        mockFetch
            .mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                data: {
                    run_id: 'run-123',
                    execution_summary: {
                        outputs: [
                            {
                                id: 'workspace/logs/game.html',
                                kind: 'code',
                                title: 'workspace/logs/game.html',
                                href: '/api/v1/workspace/files/view?path=workspace%2Flogs%2Fgame.html',
                                retained: true,
                            },
                        ],
                    },
                },
            }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ([]),
            });

        await useCortexStore.getState().confirmProposal();

        expect(useCortexStore.getState().activeMode).toBe('execution_result');
        expect(useCortexStore.getState().activeRunId).toBe('run-123');
        expect(useCortexStore.getState().durableWorkRefreshVersion).toBe(1);
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/teams/detail');
        expect(useCortexStore.getState().missionChat.at(-1)?.execution_summary?.outputs).toEqual([
            {
                id: 'workspace/logs/game.html',
                kind: 'code',
                title: 'workspace/logs/game.html',
                href: '/api/v1/workspace/files/view?path=workspace%2Flogs%2Fgame.html',
                retained: true,
            },
        ]);
    });

    it('can confirm from the rendered proposal when scoped store state lost the active token', async () => {
        const renderedProposal = {
            intent: 'Launch a docs crew',
            teams: 1,
            agents: 2,
            tools: ['delegate_task'],
            risk_level: 'medium',
            confirm_token: 'ct-rendered',
            intent_proof_id: 'ip-rendered',
        };
        useCortexStore.setState({
            pendingProposal: null,
            activeConfirmToken: null,
            missionChat: [{
                role: 'council',
                content: 'Proposed execution path',
                mode: 'proposal',
                proposal: renderedProposal,
                proposal_status: 'active',
            }],
            missionChatError: null,
            activeMode: 'proposal',
            activeRunId: null,
        });
        mockFetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    data: {
                        run_id: 'run-rendered',
                        verified: true,
                    },
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ([]),
            });

        const result = await useCortexStore.getState().confirmProposal(renderedProposal);

        expect(result).toEqual({ ok: true, runId: 'run-rendered' });
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/intent/confirm-action', expect.objectContaining({
            body: JSON.stringify({ confirm_token: 'ct-rendered' }),
        }));
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'executed',
            run_id: 'run-rendered',
        });
    });

    it('confirms the clicked rendered proposal instead of a stale pending proposal token', async () => {
        const renderedProposal = {
            intent: 'Create this visible file',
            teams: 1,
            agents: 1,
            tools: ['write_file'],
            risk_level: 'medium',
            confirm_token: 'ct-visible',
            intent_proof_id: 'ip-visible',
        };
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Older pending proposal',
                teams: 1,
                agents: 1,
                tools: ['write_file'],
                risk_level: 'medium',
                confirm_token: 'ct-stale',
                intent_proof_id: 'ip-stale',
            },
            activeConfirmToken: 'ct-stale',
            missionChat: [
                {
                    role: 'council',
                    content: 'Older pending proposal',
                    mode: 'proposal',
                    proposal: {
                        intent: 'Older pending proposal',
                        teams: 1,
                        agents: 1,
                        tools: ['write_file'],
                        risk_level: 'medium',
                        confirm_token: 'ct-stale',
                        intent_proof_id: 'ip-stale',
                    },
                    proposal_status: 'active',
                },
                {
                    role: 'council',
                    content: 'Visible proposal',
                    mode: 'proposal',
                    proposal: renderedProposal,
                    proposal_status: 'active',
                },
            ],
            missionChatError: null,
            activeMode: 'proposal',
            activeRunId: null,
        });
        mockFetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    data: {
                        run_id: 'run-visible',
                        verified: true,
                    },
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ([]),
            });

        const result = await useCortexStore.getState().confirmProposal(renderedProposal);

        expect(result).toEqual({ ok: true, runId: 'run-visible' });
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/intent/confirm-action', expect.objectContaining({
            body: JSON.stringify({ confirm_token: 'ct-visible' }),
        }));
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'active',
        });
        expect(useCortexStore.getState().missionChat[1]).toMatchObject({
            proposal_status: 'executed',
            run_id: 'run-visible',
        });
    });
});
