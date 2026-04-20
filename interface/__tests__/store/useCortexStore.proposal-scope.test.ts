import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore proposal and governance scope', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    describe('fetchProposals', () => {
        it('stores proposals from API response', async () => {
            const proposals = [
                { id: 'p1', name: 'Squad', role: 'analytics', agents: [], reason: 'Test', status: 'pending', created_at: '' },
            ];
            mockFetch.mockResolvedValue({
                ok: true,
                json: async () => ({ proposals }),
            });

            await useCortexStore.getState().fetchProposals();

            expect(useCortexStore.getState().teamProposals).toEqual(proposals);
        });

        it('approveProposal updates status in store', async () => {
            useCortexStore.setState({
                teamProposals: [
                    { id: 'p1', name: 'Squad', role: 'test', agents: [], reason: 'r', status: 'pending', created_at: '' },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({}) });

            await useCortexStore.getState().approveProposal('p1');

            expect(useCortexStore.getState().teamProposals[0].status).toBe('approved');
        });

        it('rejectProposal updates status in store', async () => {
            useCortexStore.setState({
                teamProposals: [
                    { id: 'p1', name: 'Squad', role: 'test', agents: [], reason: 'r', status: 'pending', created_at: '' },
                ],
            });
            mockFetch.mockResolvedValue({ ok: true, json: async () => ({}) });

            await useCortexStore.getState().rejectProposal('p1');

            expect(useCortexStore.getState().teamProposals[0].status).toBe('rejected');
        });
    });

    describe('governance artifact state', () => {
        it('selectArtifact sets selected artifact', () => {
            const artifact = {
                id: 'a1', source: 'agent-1', signal: 'artifact' as const,
                timestamp: '', trust_score: 0.8,
                payload: { content: 'test', content_type: 'text' as const },
            };

            useCortexStore.getState().selectArtifact(artifact);

            expect(useCortexStore.getState().selectedArtifact).toEqual(artifact);
        });

        it('approveArtifact removes from pending list', () => {
            useCortexStore.setState({
                pendingArtifacts: [
                    { id: 'a1', source: 's', signal: 'artifact' as const, timestamp: '', payload: { content: 'c', content_type: 'text' as const } },
                    { id: 'a2', source: 's', signal: 'artifact' as const, timestamp: '', payload: { content: 'c', content_type: 'text' as const } },
                ],
            });

            useCortexStore.getState().approveArtifact('a1');

            expect(useCortexStore.getState().pendingArtifacts).toHaveLength(1);
            expect(useCortexStore.getState().pendingArtifacts[0].id).toBe('a2');
        });

        it('rejectArtifact removes from pending and clears selection', () => {
            const artifact = { id: 'a1', source: 's', signal: 'artifact' as const, timestamp: '', payload: { content: 'c', content_type: 'text' as const } };
            useCortexStore.setState({
                pendingArtifacts: [artifact],
                selectedArtifact: artifact,
            });

            useCortexStore.getState().rejectArtifact('a1', 'Not accurate');

            expect(useCortexStore.getState().pendingArtifacts).toHaveLength(0);
            expect(useCortexStore.getState().selectedArtifact).toBeNull();
        });
    });
});
