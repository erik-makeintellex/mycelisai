import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

function okChat(text: string) {
    return {
        ok: true,
        json: async () => ({
            ok: true,
            data: {
                meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                signal_type: 'chat_response',
                trust_score: 0.5,
                template_id: 'chat-to-answer',
                mode: 'answer',
                payload: { text, tools_used: [] },
            },
        }),
    };
}

describe('useCortexStore mission chat continuation context', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('includes delivered-output continuation context when supplied', async () => {
        mockFetch.mockResolvedValue(okChat('I can continue from that output.'));

        await useCortexStore.getState().sendMissionChat('revise it', {
            continuation_context: {
                kind: 'output',
                title: 'Trusted Outcome Kit',
                reference: 'workspace/generated/trusted-outcome-kit',
                proof: 'proof-artifact-trusted-outcome',
            },
        });

        const request = mockFetch.mock.calls[0]?.[1] as RequestInit;
        const body = JSON.parse(String(request.body));
        expect(body.continuation_context).toEqual({
            kind: 'output',
            title: 'Trusted Outcome Kit',
            reference: 'workspace/generated/trusted-outcome-kit',
            proof: 'proof-artifact-trusted-outcome',
        });
    });

    it('omits continuation context for ordinary chat', async () => {
        mockFetch.mockResolvedValue(okChat('Ordinary answer.'));

        await useCortexStore.getState().sendMissionChat('hello');

        const request = mockFetch.mock.calls[0]?.[1] as RequestInit;
        const body = JSON.parse(String(request.body));
        expect(body).not.toHaveProperty('continuation_context');
    });

    it('adds an idempotent active-work reference without changing the visible message', async () => {
        mockFetch.mockResolvedValue(okChat('I passed that guidance to the team.'));

        await useCortexStore.getState().sendMissionChat('Also add a launch note.', {
            active_work_context: {
                type: 'team_work',
                id: '22222222-2222-4222-8222-222222222222',
                run_id: '11111111-1111-4111-8111-111111111111',
                team_id: 'launch-team',
                work_item_id: '22222222-2222-4222-8222-222222222222',
            },
        });

        const request = mockFetch.mock.calls[0]?.[1] as RequestInit;
        const body = JSON.parse(String(request.body));
        expect(body.active_work_context).toMatchObject({
            type: 'team_work',
            id: '22222222-2222-4222-8222-222222222222',
            run_id: '11111111-1111-4111-8111-111111111111',
            team_id: 'launch-team',
            work_item_id: '22222222-2222-4222-8222-222222222222',
        });
        expect(body.active_work_context.steering_id).toMatch(/^[0-9a-f-]{36}$/i);
        expect(body.messages.at(-1)).toEqual({ role: 'user', content: 'Also add a launch note.' });
    });
});
