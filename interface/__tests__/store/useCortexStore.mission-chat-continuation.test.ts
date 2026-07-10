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
});
