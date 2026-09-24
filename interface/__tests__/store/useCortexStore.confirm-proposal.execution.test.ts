import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore confirm proposal execution', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    function completedResponse(runId: string) {
        return { data: {
            confirmed: true,
            run_id: runId,
            verified: true,
            run_status: 'completed',
            execution_state: 'verified',
            execution_summary: {
                execution: { shape: 'guided_proposal', status: 'completed', summary: 'Approved action completed.' },
                proof: { run_id: runId, verified: true, proof_id: `proof-${runId}` },
            },
        } };
    }

    it.each([
        ['run id only', { run_id: 'run-sparse' }],
        ['verified flag without proof', { run_id: 'run-sparse', verified: true }],
        ['running with contradictory proof', {
            run_id: 'run-sparse', verified: false, run_status: 'running', execution_state: 'running',
            execution_summary: { execution: { status: 'completed' }, proof: { verified: true } },
        }],
        ['failed with contradictory proof', {
            run_id: 'run-sparse', verified: true, run_status: 'failed', execution_state: 'failed',
            execution_summary: { execution: { status: 'completed' }, proof: { verified: true } },
        }],
    ])('does not promote %s to a completed result', async (_caseName, data) => {
        const proposal = {
            intent: 'Write a file', teams: 1, agents: 1, tools: ['write_file'], risk_level: 'medium',
            confirm_token: 'ct-sparse', intent_proof_id: 'ip-sparse',
        };
        useCortexStore.setState({
            pendingProposal: proposal,
            activeConfirmToken: proposal.confirm_token,
            missionChat: [{ role: 'council', content: 'Write a file?', mode: 'proposal', proposal, proposal_status: 'active' }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValueOnce({ ok: true, status: 200, json: async () => ({ data }) })
            .mockResolvedValueOnce({ ok: true, json: async () => ([]) });

        expect(await useCortexStore.getState().confirmProposal()).toEqual({ ok: true, runId: 'run-sparse' });
        const [original, receipt] = useCortexStore.getState().missionChat;
        expect(original).toMatchObject({ proposal_status: 'confirmed_pending_execution', run_id: 'run-sparse' });
        expect(original.execution_summary).toBeUndefined();
        expect(receipt).toMatchObject({
            ui_response_state: { label: 'Approval recorded' },
            thread_events: [{ label: 'Approval recorded', status: 'confirmed' }],
        });
        expect(JSON.stringify([original, receipt])).not.toMatch(/Result verified|Result saved|Proof is available in Trust/);
    });

    it('records an execution result and run id on successful confirmation without team work refs', async () => {
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
                    verified: false,
                    run_status: 'running',
                    execution_state: 'running',
                    execution_summary: {
                        execution: { shape: 'team_execution', status: 'running' },
                        proof: { verified: false },
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
        expect(useCortexStore.getState().missionChat[0].proposal_status).toBe('confirmed_pending_execution');
        expect(useCortexStore.getState().durableWorkRefreshVersion).toBe(1);
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/teams/detail');
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toContain('Run run-123 started.');
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toContain('Soma started the work');
        expect(useCortexStore.getState().missionChat.at(-1)?.content).toContain('not a completed result or proof');
        expect(useCortexStore.getState().missionChat.at(-1)?.content).not.toContain('Active Work');
        expect(useCortexStore.getState().missionChat.at(-1)?.thread_events?.[0]).toMatchObject({
            kind: 'execution_update',
            label: 'Work queued',
            status: 'queued',
            target_reference: 'run:run-123',
            source_kind: 'web_api',
            source_channel: 'api.intent.confirm-action',
            payload_kind: 'soma_thread_event',
        });
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
                json: async () => completedResponse('run-rendered'),
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
            execution_summary: { execution: { status: 'completed' }, proof: { verified: true } },
        });
        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            content: expect.stringContaining('verified result is available to review'),
            thread_events: [{ kind: 'result_ready', status: 'completed' }],
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
                json: async () => completedResponse('run-visible'),
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

    it('does not run proposals that are missing executable proof linkage', async () => {
        const renderedProposal = {
            intent: 'Create this visible file',
            teams: 1,
            agents: 1,
            tools: ['write_file'],
            risk_level: 'medium',
            confirm_token: 'ct-visible',
            intent_proof_id: '',
        };
        useCortexStore.setState({
            pendingProposal: null,
            activeConfirmToken: null,
            missionChat: [{
                role: 'council',
                content: 'Visible proposal',
                mode: 'proposal',
                proposal: renderedProposal,
                proposal_status: 'active',
            }],
            missionChatError: null,
            activeMode: 'proposal',
            activeRunId: null,
        });

        const result = await useCortexStore.getState().confirmProposal(renderedProposal);

        expect(result).toEqual({
            ok: false,
            runId: null,
            error: 'This proposal is missing executable proof. Ask Soma to regenerate it before running.',
        });
        expect(mockFetch).not.toHaveBeenCalled();
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({ proposal_status: 'active' });
    });
});
