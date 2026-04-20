import { beforeEach, describe, expect, it } from 'vitest';
import { buildMissionChatFailure } from '@/lib/missionChatFailure';
import { useCortexStore } from '@/store/useCortexStore';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore mission chat scope', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('marks the proposal cancelled and appends a no-op system message when cancelled', () => {
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
            activeMode: 'proposal',
        });

        useCortexStore.getState().cancelProposal();

        expect(useCortexStore.getState().activeMode).toBe('answer');
        expect(useCortexStore.getState().pendingProposal).toBeNull();
        expect(useCortexStore.getState().activeConfirmToken).toBeNull();
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'cancelled',
        });
        expect(useCortexStore.getState().missionChat.at(-1)).toMatchObject({
            role: 'system',
            content: 'Proposal cancelled. No action executed.',
        });
    });

    it('rehydrates scoped chat history when the organization scope changes', () => {
        localStorage.setItem('mycelis-workspace-chat:org-1', JSON.stringify([{ role: 'user', content: 'org-1 history' }]));
        localStorage.setItem('mycelis-workspace-chat:org-2', JSON.stringify([{ role: 'user', content: 'org-2 history' }]));

        useCortexStore.getState().setMissionChatScope('org-1');
        expect(useCortexStore.getState().workspaceChatScope).toBe('org-1');
        expect(useCortexStore.getState().missionChat).toMatchObject([{ role: 'user', content: 'org-1 history' }]);

        useCortexStore.getState().setMissionChatScope('org-2');
        expect(useCortexStore.getState().workspaceChatScope).toBe('org-2');
        expect(useCortexStore.getState().missionChat).toMatchObject([{ role: 'user', content: 'org-2 history' }]);
    });

    it('clears in-flight workspace chat state when switching organization scope', () => {
        localStorage.setItem('mycelis-workspace-chat:org-2', JSON.stringify([{ role: 'user', content: 'org-2 history' }]));

        useCortexStore.setState({
            workspaceChatScope: 'org-1',
            missionChat: [{ role: 'user', content: 'org-1 history' }],
            isMissionChatting: true,
            isBroadcasting: true,
            missionChatError: 'still loading',
            missionChatFailure: buildMissionChatFailure({
                assistantName: 'Soma',
                targetId: 'admin',
                message: 'still loading',
                statusCode: 503,
            }),
            activeRole: 'admin',
            lastBroadcastResult: { teams_hit: 2 },
        });

        useCortexStore.getState().setMissionChatScope('org-2');

        expect(useCortexStore.getState().workspaceChatScope).toBe('org-2');
        expect(useCortexStore.getState().missionChat).toMatchObject([{ role: 'user', content: 'org-2 history' }]);
        expect(useCortexStore.getState().isMissionChatting).toBe(false);
        expect(useCortexStore.getState().isBroadcasting).toBe(false);
        expect(useCortexStore.getState().missionChatError).toBeNull();
        expect(useCortexStore.getState().missionChatFailure).toBeNull();
        expect(useCortexStore.getState().lastBroadcastResult).toBeNull();
        expect(useCortexStore.getState().activeRole).toBe('');
    });

    it('rehydrates a proof-pending proposal without promoting it to verified execution', () => {
        localStorage.setItem('mycelis-workspace-chat:org-proof', JSON.stringify([
            {
                role: 'council',
                content: 'Proposal confirmed. Waiting for execution proof.',
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
                proposal_status: 'executed',
            },
        ]));

        useCortexStore.getState().setMissionChatScope('org-proof');

        expect(useCortexStore.getState().workspaceChatScope).toBe('org-proof');
        expect(useCortexStore.getState().activeMode).toBe('proposal');
        expect(useCortexStore.getState().activeRunId).toBeNull();
        expect(useCortexStore.getState().pendingProposal).toBeNull();
        expect(useCortexStore.getState().missionChat[0]).toMatchObject({
            proposal_status: 'executed',
        });
    });
});
