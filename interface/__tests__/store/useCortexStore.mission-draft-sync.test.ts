import { beforeEach, describe, expect, it } from 'vitest';
import { blueprintToGraph } from '@/store/cortexStoreUtils';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';
import { baseBlueprint, resetCortexStore } from './useCortexStoreTestSupport';

describe('useCortexStore mission draft sync', () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it('updateAgentInDraft updates the local blueprint and closes the editor', () => {
        useCortexStore.setState({
            blueprint: structuredClone(baseBlueprint),
            selectedAgentNodeId: 'agent-0-0',
            isAgentEditorOpen: true,
        });

        useCortexStore.getState().updateAgentInDraft(0, 0, { model: 'model-updated' });

        expect(useCortexStore.getState().blueprint?.teams[0].agents[0].model).toBe('model-updated');
        expect(useCortexStore.getState().selectedAgentNodeId).toBeNull();
        expect(useCortexStore.getState().isAgentEditorOpen).toBe(false);
    });

    it('deleteAgentFromDraft removes the last agent team and rebuilds the graph', () => {
        useCortexStore.setState({
            blueprint: {
                mission_id: 'mission-2',
                intent: 'Single agent mission',
                teams: [
                    {
                        name: 'Solo',
                        role: 'operators',
                        agents: [
                            {
                                id: 'solo-agent',
                                role: 'cognitive',
                                outputs: ['solo.out'],
                            },
                        ],
                    },
                ],
            },
            selectedAgentNodeId: 'agent-0-0',
            isAgentEditorOpen: true,
        });

        useCortexStore.getState().deleteAgentFromDraft(0, 0);

        expect(useCortexStore.getState().blueprint?.teams).toEqual([]);
        expect(useCortexStore.getState().nodes).toEqual([]);
        expect(useCortexStore.getState().edges).toEqual([]);
        expect(useCortexStore.getState().isAgentEditorOpen).toBe(false);
    });

    it('updateAgentInMission updates the active blueprint and solidifies draft nodes', async () => {
        const { nodes, edges } = blueprintToGraph(baseBlueprint);
        useCortexStore.setState({
            blueprint: structuredClone(baseBlueprint),
            nodes,
            edges,
            activeMissionId: 'mission-1',
            selectedAgentNodeId: 'agent-0-0',
            isAgentEditorOpen: true,
        });
        mockFetch.mockResolvedValue({ ok: true });

        await useCortexStore.getState().updateAgentInMission('alpha', { model: 'model-live' });

        expect(useCortexStore.getState().blueprint?.teams[0].agents[0].model).toBe('model-live');
        expect(useCortexStore.getState().nodes.some((node) => node.type === 'agentNode' && node.className === '')).toBe(true);
        expect(useCortexStore.getState().selectedAgentNodeId).toBeNull();
        expect(useCortexStore.getState().isAgentEditorOpen).toBe(false);
    });

    it('discardDraft clears the draft workspace state', () => {
        useCortexStore.setState({
            blueprint: structuredClone(baseBlueprint),
            missionStatus: 'active',
            activeMissionId: 'mission-1',
            selectedAgentNodeId: 'agent-0-0',
            isAgentEditorOpen: true,
        });

        useCortexStore.getState().discardDraft();

        expect(useCortexStore.getState().blueprint).toBeNull();
        expect(useCortexStore.getState().missionStatus).toBe('idle');
        expect(useCortexStore.getState().activeMissionId).toBeNull();
        expect(useCortexStore.getState().nodes).toEqual([]);
        expect(useCortexStore.getState().edges).toEqual([]);
        expect(useCortexStore.getState().isAgentEditorOpen).toBe(false);
    });

    it('deleteMission clears the draft workspace after a successful delete', async () => {
        useCortexStore.setState({
            blueprint: structuredClone(baseBlueprint),
            missionStatus: 'active',
            activeMissionId: 'mission-1',
            selectedAgentNodeId: 'agent-0-0',
            isAgentEditorOpen: true,
        });
        mockFetch.mockResolvedValue({ ok: true });

        await useCortexStore.getState().deleteMission('mission-1');

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/missions/mission-1', { method: 'DELETE' });
        expect(useCortexStore.getState().blueprint).toBeNull();
        expect(useCortexStore.getState().missionStatus).toBe('idle');
        expect(useCortexStore.getState().activeMissionId).toBeNull();
        expect(useCortexStore.getState().isAgentEditorOpen).toBe(false);
    });
});
