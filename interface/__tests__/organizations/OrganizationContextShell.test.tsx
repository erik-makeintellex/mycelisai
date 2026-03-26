import { describe, expect, it, vi } from 'vitest';

vi.mock('reactflow', async () => {
    const mock = await import('../mocks/reactflow');
    return mock;
});

import {
    extractOutputsFromConversation,
    findLatestConversationOutcome,
} from '@/components/organizations/OrganizationContextShell';

describe('OrganizationContextShell causal helpers', () => {
    it('shows execution proof is still pending when an execution_result lacks durable proof', () => {
        expect(extractOutputsFromConversation({
            role: 'system',
            content: 'Mission activated',
            mode: 'execution_result',
        })).toEqual(['Awaiting execution proof']);
    });

    it('keeps an executed proposal in a pending-proof posture until run proof exists', () => {
        expect(extractOutputsFromConversation({
            role: 'council',
            content: 'Proposal confirmed.',
            mode: 'proposal',
            proposal: {
                intent: 'chat-action',
                teams: 1,
                agents: 1,
                tools: ['delegate'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
            },
            proposal_status: 'executed',
        })).toEqual(['Awaiting execution proof']);
    });

    it('surfaces pending-proof proposals as awaiting execution proof', () => {
        expect(extractOutputsFromConversation({
            role: 'council',
            content: 'Proposal confirmed. Waiting for execution proof.',
            mode: 'proposal',
            proposal: {
                intent: 'chat-action',
                teams: 1,
                agents: 1,
                tools: ['delegate'],
                risk_level: 'medium',
                confirm_token: 'ct-123',
                intent_proof_id: 'ip-123',
            },
            proposal_status: 'confirmed_pending_execution',
        })).toEqual(['Awaiting execution proof']);
    });

    it('returns the latest verified conversation outcome instead of the latest generic message', () => {
        const outcome = findLatestConversationOutcome([
            {
                role: 'council',
                content: 'I prepared a governed plan.',
                mode: 'proposal',
                timestamp: '2026-03-25T10:00:00.000Z',
                proposal: {
                    intent: 'chat-action',
                    teams: 1,
                    agents: 1,
                    tools: ['delegate'],
                    risk_level: 'medium',
                    confirm_token: 'ct-123',
                    intent_proof_id: 'ip-123',
                },
                proposal_status: 'active',
            },
            {
                role: 'system',
                content: 'Mission activated',
                mode: 'execution_result',
                timestamp: '2026-03-25T10:01:00.000Z',
            },
        ], 'Team Lead for QA Org');

        expect(outcome).toMatchObject({
            actionLabel: 'Follow the confirmed proposal until proof arrives',
            outputsGenerated: ['Awaiting execution proof'],
        });
    });

    it('recognizes a verified execution only when run proof exists', () => {
        expect(extractOutputsFromConversation({
            role: 'system',
            content: 'Mission activated',
            mode: 'execution_result',
            run_id: 'run-123',
        })).toEqual(['Verified run created']);
    });
});
