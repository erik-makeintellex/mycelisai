import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

// Split out of useCortexStore.confirm-proposal.failure.test.ts to stay
// within its line cap. Covers the "token kept" contract: approver_required
// and confirmer_not_proposer must not fail or clear the proposal, and
// token_already_used must show "Already approved" with no retry.
describe('useCortexStore confirm proposal: token-kept codes', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('keeps the proposal and confirm token when the backend returns approver_required', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Launch a docs crew',
                teams: 1,
                agents: 2,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-approver',
                intent_proof_id: 'ip-approver',
            },
            activeConfirmToken: 'ct-approver',
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
                    confirm_token: 'ct-approver',
                    intent_proof_id: 'ip-approver',
                },
                proposal_status: 'active',
            }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 403,
            text: async () => JSON.stringify({
                error: 'This work needs admin approval.',
                data: { code: 'approver_required', recommended_action: 'Ask an admin to review and confirm this proposal.' },
            }),
        });

        await useCortexStore.getState().confirmProposal();

        // Token-kept codes must not fail the proposal or clear the token.
        expect(useCortexStore.getState().pendingProposal).not.toBeNull();
        expect(useCortexStore.getState().activeConfirmToken).toBe('ct-approver');
        expect(useCortexStore.getState().activeMode).toBe('proposal');
        expect(useCortexStore.getState().missionChat[0].proposal_status).toBe('active');
        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            code: 'approver_required',
            title: 'Waiting for an admin to approve',
        });
    });

    it('keeps the proposal and confirm token when the backend returns confirmer_not_proposer', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Launch a docs crew',
                teams: 1,
                agents: 2,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-proposer',
                intent_proof_id: 'ip-proposer',
            },
            activeConfirmToken: 'ct-proposer',
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
                    confirm_token: 'ct-proposer',
                    intent_proof_id: 'ip-proposer',
                },
                proposal_status: 'active',
            }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 403,
            text: async () => JSON.stringify({
                error: 'invalid confirm_token: only the proposer or an approver may confirm this proposal',
                data: { code: 'confirmer_not_proposer' },
            }),
        });

        await useCortexStore.getState().confirmProposal();

        expect(useCortexStore.getState().pendingProposal).not.toBeNull();
        expect(useCortexStore.getState().activeConfirmToken).toBe('ct-proposer');
        expect(useCortexStore.getState().missionChat[0].proposal_status).toBe('active');
    });

    it('shows "Already approved" with no retry and clears the token for token_already_used', async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: 'Launch a docs crew',
                teams: 1,
                agents: 2,
                tools: ['delegate_task'],
                risk_level: 'medium',
                confirm_token: 'ct-used',
                intent_proof_id: 'ip-used',
            },
            activeConfirmToken: 'ct-used',
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
                    confirm_token: 'ct-used',
                    intent_proof_id: 'ip-used',
                },
                proposal_status: 'active',
            }],
            activeMode: 'proposal',
        });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 409,
            text: async () => JSON.stringify({
                error: 'invalid confirm_token: already used',
                data: { code: 'token_already_used' },
            }),
        });

        await useCortexStore.getState().confirmProposal();

        expect(useCortexStore.getState().missionChatFailure).toMatchObject({
            code: 'token_already_used',
            title: 'Already approved',
            retryable: false,
        });
        expect(useCortexStore.getState().pendingProposal).toBeNull();
        expect(useCortexStore.getState().activeConfirmToken).toBeNull();
    });
});
