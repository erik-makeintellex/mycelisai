import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore confirm proposal pending proof', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('keeps the proposal in a pending-proof state when confirmation succeeds without a run id', async () => {
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
            json: async () => ({ data: { confirmed: true, run_id: null } }),
        });

        const result = await useCortexStore.getState().confirmProposal();

        expect(result).toEqual({ ok: true, runId: null });
        expect(useCortexStore.getState().activeMode).toBe('proposal');
        expect(useCortexStore.getState().activeRunId).toBeNull();
        expect(useCortexStore.getState().pendingProposal).toBeNull();
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'confirmed_pending_execution',
            mode: 'proposal',
        });
        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            role: 'system',
            mode: 'proposal',
            content: 'Proposal confirmed. Waiting for execution proof.',
        });
    });
});
