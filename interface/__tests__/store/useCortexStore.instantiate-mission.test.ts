import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { baseBlueprint, resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore instantiateMission (Launch teams)', () => {
    beforeEach(() => {
        resetCortexStore();
        useCortexStore.setState({
            blueprint: baseBlueprint,
            missionStatus: 'draft',
            activeConfirmToken: 'ct-negotiated-1',
        });
    });

    it('sends the negotiated confirm token and the exact negotiated blueprint to /intent/commit', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            status: 200,
            json: async () => ({ mission_id: 'mission-1', teams: 1, agents: 2 }),
        });

        await useCortexStore.getState().instantiateMission();

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/intent/commit', expect.objectContaining({
            method: 'POST',
        }));
        const call = mockFetch.mock.calls[0];
        const body = JSON.parse(call[1].body);
        expect(body.confirm_token).toBe('ct-negotiated-1');
        expect(body.mission_id).toBe(baseBlueprint.mission_id);
        expect(body.intent).toBe(baseBlueprint.intent);
        expect(body.teams).toEqual(baseBlueprint.teams);
    });

    it('clears the confirm token on a successful launch and reports plain-language success', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            status: 200,
            json: async () => ({ mission_id: 'mission-1', teams: 1, agents: 2 }),
        });

        await useCortexStore.getState().instantiateMission();

        expect(useCortexStore.getState().missionStatus).toBe('active');
        expect(useCortexStore.getState().activeConfirmToken).toBeNull();
        const lastChat = useCortexStore.getState().chatHistory.at(-1);
        expect(lastChat?.content).toContain('Teams launched');
        expect(lastChat?.content).not.toMatch(/instantiated|ACTIVE|swarm/i);
    });

    it('maps a missing-token 403 (invalid_confirm_token) through blockerCopy instead of showing raw backend text', async () => {
        useCortexStore.setState({ activeConfirmToken: null });
        mockFetch.mockResolvedValue({
            ok: false,
            status: 403,
            json: async () => ({
                error: "This team plan needs Soma's proposal before it can launch.",
                data: { code: 'invalid_confirm_token' },
            }),
        });

        await useCortexStore.getState().instantiateMission();

        const lastChat = useCortexStore.getState().chatHistory.at(-1);
        expect(lastChat?.content).toContain('This proposal is no longer valid');
        expect(lastChat?.content).not.toMatch(/confirm_token/);
    });

    it('maps a blueprint_mismatch conflict through blockerCopy', async () => {
        mockFetch.mockResolvedValue({
            ok: false,
            status: 409,
            json: async () => ({
                error: 'invalid confirm_token: blueprint differs from the negotiated proposal',
                data: { code: 'blueprint_mismatch' },
            }),
        });

        await useCortexStore.getState().instantiateMission();

        const lastChat = useCortexStore.getState().chatHistory.at(-1);
        expect(lastChat?.content).toContain('You changed the plan');
        expect(lastChat?.content).not.toMatch(/confirm_token/);
        expect(useCortexStore.getState().missionStatus).toBe('draft');
    });
});
