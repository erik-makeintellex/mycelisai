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
            json: async () => ({ mission_id: 'mission-1', teams: 1, agents: 2, status: 'active' }),
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
            json: async () => ({ mission_id: 'mission-1', teams: 1, agents: 2, status: 'active' }),
        });

        await useCortexStore.getState().instantiateMission();

        expect(useCortexStore.getState().missionStatus).toBe('active');
        expect(useCortexStore.getState().activeConfirmToken).toBeNull();
        const lastChat = useCortexStore.getState().chatHistory.at(-1);
        expect(lastChat?.content).toContain('Teams launched');
        expect(lastChat?.content).not.toMatch(/instantiated|ACTIVE|swarm/i);
        const agentNode = useCortexStore.getState().nodes.find((n) => n.type === 'agentNode');
        expect(agentNode?.data?.status).toBe('online');
    });

    it('reports a partial start honestly (never "active") when only some teams activated', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            status: 200,
            json: async () => ({
                mission_id: 'mission-1',
                teams: 3,
                agents: 5,
                status: 'partially_active',
                activation: { teams_spawned: 1, teams_skipped: 1, sensors_spawned: 0, errors: ['team-3: spawn failed'] },
            }),
        });

        await useCortexStore.getState().instantiateMission();

        const lastChat = useCortexStore.getState().chatHistory.at(-1);
        expect(lastChat?.content).toContain('Partly started (2 of 3 team');
        expect(lastChat?.content).not.toMatch(/Teams launched/);
        const agentNode = useCortexStore.getState().nodes.find((n) => n.type === 'agentNode');
        expect(agentNode?.data?.status).not.toBe('online');
    });

    it('reports "Saved, not started yet" when the mission persisted but nothing activated', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            status: 200,
            json: async () => ({
                mission_id: 'mission-1',
                teams: 1,
                agents: 2,
                status: 'persisted_not_activated',
                activation: null,
            }),
        });

        await useCortexStore.getState().instantiateMission();

        const lastChat = useCortexStore.getState().chatHistory.at(-1);
        expect(lastChat?.content).toContain('Saved, not started yet');
        const agentNode = useCortexStore.getState().nodes.find((n) => n.type === 'agentNode');
        expect(agentNode?.data?.status).not.toBe('online');
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
