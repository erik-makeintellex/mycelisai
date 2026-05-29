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
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ data: { run_id: 'run-123' } }),
        });

        await useCortexStore.getState().confirmProposal();

        expect(useCortexStore.getState().activeMode).toBe('execution_result');
        expect(useCortexStore.getState().activeRunId).toBe('run-123');
    });

    it('adds returned package artifacts to the mission chat after confirmation', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Create Codex package',
                teams: 1,
                agents: 1,
                tools: ['write_file'],
                risk_level: 'medium',
                confirm_token: 'ct-456',
                intent_proof_id: 'ip-456',
            },
            activeConfirmToken: 'ct-456',
            missionChat: [{
                role: 'council',
                content: 'Create a Codex task bundle under "codex-task-packages/myc-rc-001".',
                mode: 'proposal',
                proposal: {
                    intent: 'Create Codex package',
                    teams: 1,
                    agents: 1,
                    tools: ['write_file'],
                    risk_level: 'medium',
                    confirm_token: 'ct-456',
                    intent_proof_id: 'ip-456',
                },
                proposal_status: 'active',
            }],
            missionChatError: null,
            activeMode: 'proposal',
            activeRunId: null,
        });
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                data: {
                    run_id: 'run-456',
                    message: 'Codex task bundle created under codex-task-packages/myc-rc-001 with proof at codex-task-packages/myc-rc-001/proof.md.',
                    delivery_package: {
                        package_id: 'codex_task_bundle:myc-rc-001',
                        package_kind: 'codex_task_bundle',
                        title: 'Codex task bundle',
                        status: 'verified',
                        run_id: 'run-456',
                        items: [
                            {
                                id: 'codex_task_bundle:myc-rc-001:implementation_task.md',
                                kind: 'document',
                                label: 'Implementation Task',
                                storage_ref: 'codex-task-packages/myc-rc-001/implementation_task.md',
                                source_run_id: 'run-456',
                                source_contract_id: 'ip-456',
                                package_id: 'codex_task_bundle:myc-rc-001',
                                role: 'primary',
                                status: 'final',
                            },
                            {
                                id: 'codex_task_bundle:myc-rc-001:proof.md',
                                kind: 'document',
                                label: 'Package Proof',
                                storage_ref: 'codex-task-packages/myc-rc-001/proof.md',
                                source_run_id: 'run-456',
                                source_contract_id: 'ip-456',
                                package_id: 'codex_task_bundle:myc-rc-001',
                                role: 'proof',
                                status: 'final',
                            },
                        ],
                    },
                    execution_summary: {
                        title: 'Codex task bundle',
                        status: 'verified',
                        run_id: 'run-456',
                    },
                    artifacts: [
                        {
                            type: 'document',
                            title: 'Implementation Task',
                            content_type: 'text/markdown',
                            content: '# Implementation Task',
                            saved_path: 'codex-task-packages/myc-rc-001/implementation_task.md',
                        },
                        {
                            type: 'document',
                            title: 'Package Proof',
                            content_type: 'text/markdown',
                            content: '# Codex Task Bundle Proof',
                            saved_path: 'codex-task-packages/myc-rc-001/proof.md',
                        },
                    ],
                },
            }),
        });

        await useCortexStore.getState().confirmProposal();

        const packageMessage = useCortexStore.getState().missionChat.find((message) => (
            message.mode === 'execution_result' && message.artifacts?.some((artifact) => artifact.title === 'Package Proof')
        ));
        expect(packageMessage).toBeDefined();
        expect(packageMessage?.run_id).toBe('run-456');
        expect(packageMessage?.artifacts).toHaveLength(2);
        expect(packageMessage?.delivery_package?.package_id).toBe('codex_task_bundle:myc-rc-001');
        expect(packageMessage?.execution_summary?.status).toBe('verified');
        expect(packageMessage?.output_refs).toEqual(expect.arrayContaining([
            expect.objectContaining({
                role: 'primary',
                storage_ref: 'codex-task-packages/myc-rc-001/implementation_task.md',
            }),
            expect.objectContaining({
                role: 'proof',
                storage_ref: 'codex-task-packages/myc-rc-001/proof.md',
            }),
        ]));
    });
});
