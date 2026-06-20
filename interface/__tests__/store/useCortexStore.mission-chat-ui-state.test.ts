import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore mission chat UI state', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('preserves structured UI response state from chat envelopes', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    meta: { source_node: 'admin', timestamp: new Date().toISOString() },
                    signal_type: 'chat_response',
                    trust_score: 0.5,
                    template_id: 'chat-to-proposal',
                    mode: 'proposal',
                    ui_response_state: {
                        kind: 'awaiting_approval',
                        label: 'Waiting for approval',
                        detail: 'Review the proposed action before Soma starts work.',
                        tone: 'warning',
                    },
                    payload: {
                        text: 'I prepared a governed execution plan.',
                        proposal: {
                            intent: 'chat-action',
                            tools: ['delegate_task'],
                            risk_level: 'medium',
                            confirm_token: 'ct-123',
                            intent_proof_id: 'ip-123',
                        },
                    },
                },
            }),
        });

        await useCortexStore.getState().sendMissionChat('start the work');

        expect(useCortexStore.getState().missionChat.at(-1)?.ui_response_state).toMatchObject({
            kind: 'awaiting_approval',
            label: 'Waiting for approval',
            detail: 'Review the proposed action before Soma starts work.',
            tone: 'warning',
        });
    });
});
