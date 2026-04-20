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
            text: async () => JSON.stringify({ error: 'confirmation denied' }),
        });

        const result = await useCortexStore.getState().confirmProposal();

        expect(result).toEqual({
            ok: false,
            runId: null,
            error: 'Soma hit a server-side failure while handling the request.',
        });
        expect(useCortexStore.getState().activeMode).toBe('blocker');
        expect(useCortexStore.getState().missionChatError).toBe('Soma hit a server-side failure while handling the request.');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            routeKind: 'workspace',
            type: 'server_error',
        });
        expect(useCortexStore.getState().pendingProposal).toBeNull();
    });
});
