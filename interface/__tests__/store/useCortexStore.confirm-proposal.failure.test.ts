import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore confirm proposal failure', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('returns a blocker contract when confirmation fails', async () => {
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
        });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 500,
            text: async () => JSON.stringify({
                error: 'confirmation denied',
                data: {
                    run_id: 'run-failed-1',
                    execution_summary: {
                        execution: {
                            shape: 'guided_proposal',
                            status: 'failed',
                            summary: 'Soma could not complete the approved proposal.',
                        },
                        audit_recovery: {
                            recovery_state: 'failed',
                            blocker: 'tool unavailable',
                            degradation: {
                                code: 'approved_execution_failed',
                                what_failed: 'tool unavailable',
                                trusted_state: 'The failed run record remains trusted.',
                                safe_continuation: 'Review the failed run and retry.',
                                requires_attention: true,
                            },
                        },
                        proof: {
                            run_id: 'run-failed-1',
                            proof_class: 'run_and_audit',
                            verified: false,
                        },
                    },
                },
            }),
        });

        const result = await useCortexStore.getState().confirmProposal();

        expect(result).toEqual({
            ok: false,
            runId: 'run-failed-1',
            error: 'tool unavailable',
        });
        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().activeRunId).toBe('run-failed-1');
        expect(useCortexStore.getState().missionChatError).toBe('tool unavailable');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'workspace',
            type: 'server_error',
            summary: 'tool unavailable',
            recommendedAction: 'Review the failed run and retry.',
            diagnostics: expect.stringContaining('approved_execution_failed'),
        });
        const lastMessage = useCortexStore.getState().missionChat.at(-1);
        expect(lastMessage).toMatchObject({
            mode: 'blocker',
            run_id: 'run-failed-1',
            execution_summary: {
                execution: { status: 'failed' },
                audit_recovery: {
                    degradation: {
                        code: 'approved_execution_failed',
                        requires_attention: true,
                    },
                },
            },
        });
        expect(useCortexStore.getState().pendingProposal).toBeNull();
    });
});
