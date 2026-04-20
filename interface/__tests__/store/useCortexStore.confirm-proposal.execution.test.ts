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
});
