import { beforeEach, describe, expect, it, vi } from 'vitest';
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

    it('returns contract-unsatisfied failures as an attention thread event', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Create a playable voxel browser game',
                teams: 1,
                agents: 3,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-game',
                intent_proof_id: 'ip-game',
            },
            activeConfirmToken: 'ct-game',
            missionChat: [{
                role: 'council',
                content: 'Proposed execution path',
                mode: 'proposal',
                proposal: {
                    intent: 'Create a playable voxel browser game',
                    teams: 1,
                    agents: 3,
                    tools: ['delegate_task'],
                    risk_level: 'medium',
                    confirm_token: 'ct-game',
                    intent_proof_id: 'ip-game',
                },
                proposal_status: 'active',
            }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 500,
            text: async () => JSON.stringify({
                error: 'result contract unsatisfied',
                data: {
                    run_id: 'run-game-failed',
                    execution_summary: {
                        execution: { shape: 'delegated_work', status: 'failed' },
                        audit_recovery: {
                            recovery_state: 'failed',
                            degradation: {
                                code: 'result_contract_unsatisfied',
                                what_failed: 'The team did not retain a runnable browser game package.',
                                invalidated_proof: 'No playable output should be trusted.',
                                safe_continuation: 'Retry with the same app package contract.',
                                requires_attention: true,
                            },
                        },
                    },
                },
            }),
        });

        await useCortexStore.getState().confirmProposal();

        const lastMessage = useCortexStore.getState().missionChat.at(-1);
        expect(lastMessage?.thread_events?.[0]).toMatchObject({
            kind: 'attention_required',
            label: 'Output is not playable yet',
            status: 'result_contract_unsatisfied',
            target_reference: 'result_contract_unsatisfied',
        });
        expect(lastMessage?.content).toBe('The team did not retain a runnable browser game package.');
    });

    it('turns local media engine outages into actionable recovery copy', async () => {
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Generate a comic page',
                teams: 1,
                agents: 3,
                tools: ['generate_image', 'save_cached_image'],
                risk_level: 'medium',
                confirm_token: 'ct-media',
                intent_proof_id: 'ip-media',
            },
            activeConfirmToken: 'ct-media',
            missionChat: [{
                role: 'council',
                content: 'Proposed media execution path',
                mode: 'proposal',
                proposal: {
                    intent: 'Generate a comic page',
                    teams: 1,
                    agents: 3,
                    tools: ['generate_image', 'save_cached_image'],
                    risk_level: 'medium',
                    confirm_token: 'ct-media',
                    intent_proof_id: 'ip-media',
                },
                proposal_status: 'active',
            }],
            missionChatError: null,
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 503,
            text: async () => JSON.stringify({
                error: 'approved execution failed: media engine error (HTTP 503): {"detail":"local/private ComfyUI engine unreachable at configured upstream"}',
                data: {
                    run_id: 'run-media-failed',
                    execution_summary: {
                        execution: {
                            shape: 'guided_proposal',
                            status: 'failed',
                            summary: 'Soma could not complete the approved proposal.',
                        },
                        audit_recovery: {
                            recovery_state: 'failed',
                            degradation: {
                                code: 'approved_execution_failed',
                                what_failed: 'media engine error (HTTP 503): {"detail":"local/private ComfyUI engine unreachable at configured upstream"}',
                                trusted_state: 'The failed run record remains trusted.',
                                safe_continuation: 'Review media provider configuration and retry.',
                                requires_attention: true,
                            },
                        },
                    },
                },
            }),
        });

        const result = await useCortexStore.getState().confirmProposal();

        expect(result).toEqual({
            ok: false,
            runId: 'run-media-failed',
            error: 'The configured image generator is not ready, so Soma could not create the requested image output.',
        });
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'workspace',
            type: 'unreachable',
            summary: 'The configured image generator is not ready, so Soma could not create the requested image output.',
            recommendedAction: expect.stringContaining('Start or reconnect the configured image generator'),
            diagnostics: expect.stringContaining('ComfyUI engine unreachable'),
        });
        expect(warnSpy).toHaveBeenCalledWith(
            '[CE-1] Confirm action blocked by runtime dependency:',
            expect.stringContaining('media engine error'),
        );
        expect(errorSpy).not.toHaveBeenCalled();

        warnSpy.mockRestore();
        errorSpy.mockRestore();
    });
});
