import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { mockFetch } from '../setup';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import MissionControlChat from '@/components/dashboard/MissionControlChat';
import { useCortexStore } from '@/store/useCortexStore';
import {
    COUNCIL_MEMBERS,
    resetMissionControlChatStore,
    settleMissionControlChat,
} from './support/missionControlChatTestUtils';

describe('MissionControlChat package outputs', () => {
    beforeEach(() => {
        localStorage.clear();
        resetMissionControlChatStore();
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: COUNCIL_MEMBERS }),
        });
    });

    it('renders confirmed Codex package artifacts as one output workbench', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'Codex task bundle created under codex-task-packages/myc-rc-001 with proof at codex-task-packages/myc-rc-001/proof.md.',
                    source_node: 'admin',
                    mode: 'execution_result',
                    run_id: 'run-456',
                    delivery_package: {
                        package_kind: 'codex_task_bundle',
                        package_slug: 'myc-rc-001',
                        base_path: 'codex-task-packages/myc-rc-001',
                        proof_path: 'codex-task-packages/myc-rc-001/proof.md',
                        files: [
                            {
                                path: 'codex-task-packages/myc-rc-001/implementation_task.md',
                                title: 'Implementation Task',
                                bytes: 120,
                                sha256: 'abc1230000000000000000000000000000000000000000000000000000000000',
                            },
                            {
                                path: 'codex-task-packages/myc-rc-001/proof.md',
                                title: 'Package Proof',
                                bytes: 80,
                                sha256: 'def4560000000000000000000000000000000000000000000000000000000000',
                            },
                        ],
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
                    output_refs: [
                        {
                            id: 'codex_task_bundle:myc-rc-001:implementation_task.md',
                            kind: 'document',
                            label: 'Implementation Task',
                            storage_ref: 'codex-task-packages/myc-rc-001/implementation_task.md',
                            package_id: 'codex_task_bundle:myc-rc-001',
                            role: 'primary',
                            status: 'final',
                            artifact_id: 'artifact-001',
                            open_url: '/api/v1/artifacts/artifact-001/download',
                            proof: {
                                proof_id: 'proof-001',
                                output_ref_id: 'codex_task_bundle:myc-rc-001:implementation_task.md',
                                artifact_id: 'artifact-001',
                                storage_ref: 'codex-task-packages/myc-rc-001/implementation_task.md',
                                source_run_id: 'run-456',
                                source_contract_id: 'ip-456',
                                execution_status: 'completed',
                                path_boundary_status: 'ok',
                                checksum: 'abc1230000000000000000000000000000000000000000000000000000000000',
                                checksum_algorithm: 'sha256',
                                bytes: 120,
                                content_type: 'text/markdown',
                                created_at: '2026-05-29T00:00:00Z',
                                readback_status: 'verified',
                            },
                        },
                        {
                            id: 'codex_task_bundle:myc-rc-001:proof.md',
                            kind: 'document',
                            label: 'Package Proof',
                            storage_ref: 'codex-task-packages/myc-rc-001/proof.md',
                            package_id: 'codex_task_bundle:myc-rc-001',
                            role: 'proof',
                            status: 'final',
                        },
                    ],
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByTestId('output-workbench')).toBeDefined();
        expect(screen.getByText('Codex task bundle')).toBeDefined();
        expect(screen.getByText('codex-task-packages/myc-rc-001')).toBeDefined();
        expect(screen.getByText('Implementation Task')).toBeDefined();
        expect(screen.getByText('Package Proof')).toBeDefined();
        expect(screen.getByText(/proof sha256/i)).toBeDefined();
        expect(screen.queryByText('Returned output')).toBeNull();

        // Verified item: status chip renders and the checksum is inspectable.
        expect(screen.getAllByText('verified').length).toBeGreaterThan(0);
        expect(screen.getByText(/abc1230000000000\.\.\./)).toBeDefined();
    });

    it('shows a degraded status with a recovery hint when the backend reports an unverified proof', async () => {
        useCortexStore.setState({
            missionChat: [
                {
                    role: 'council',
                    content: 'Codex task bundle created under codex-task-packages/myc-rc-002.',
                    source_node: 'admin',
                    mode: 'execution_result',
                    run_id: 'run-789',
                    delivery_package: {
                        package_kind: 'codex_task_bundle',
                        package_slug: 'myc-rc-002',
                        base_path: 'codex-task-packages/myc-rc-002',
                        files: [
                            {
                                path: 'codex-task-packages/myc-rc-002/implementation_task.md',
                                title: 'Implementation Task',
                                bytes: 120,
                                sha256: 'deadbeef00000000000000000000000000000000000000000000000000000000',
                            },
                        ],
                    },
                    artifacts: [],
                    output_refs: [
                        {
                            id: 'codex_task_bundle:myc-rc-002:implementation_task.md',
                            kind: 'document',
                            label: 'Implementation Task',
                            storage_ref: 'codex-task-packages/myc-rc-002/implementation_task.md',
                            package_id: 'codex_task_bundle:myc-rc-002',
                            role: 'primary',
                            status: 'final',
                            proof: {
                                proof_id: 'proof-002',
                                output_ref_id: 'codex_task_bundle:myc-rc-002:implementation_task.md',
                                storage_ref: 'codex-task-packages/myc-rc-002/implementation_task.md',
                                readback_status: 'degraded',
                                recovery_hint: 'Artifact storage was unavailable; falling back to workspace file readback.',
                            },
                        },
                    ],
                },
            ],
        });

        render(<MissionControlChat />);
        await settleMissionControlChat();

        expect(screen.getByTestId('output-workbench')).toBeDefined();
        expect(screen.getAllByText('degraded').length).toBeGreaterThan(0);
        expect(screen.getByText(/Artifact storage was unavailable/i)).toBeDefined();
    });
});
