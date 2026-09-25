import { beforeEach, describe, expect, it } from 'vitest';
import { useCortexStore } from '@/store/useCortexStore';

describe('restorePersistedClientState', () => {
    beforeEach(() => {
        localStorage.clear();
    });

    it('starts from server-identical state and restores persisted chat and advanced mode after mount', () => {
        localStorage.setItem('mycelis-advanced-mode', 'true');
        localStorage.setItem('mycelis-workspace-chat', JSON.stringify([
            { role: 'user', content: 'Plan a rollout', timestamp: '2026-09-25T00:00:00Z' },
        ]));
        useCortexStore.setState({ missionChat: [], advancedMode: false, workspaceChatScope: null });

        useCortexStore.getState().restorePersistedClientState();

        const state = useCortexStore.getState();
        expect(state.advancedMode).toBe(true);
        expect(state.missionChat.map((m) => m.content)).toEqual(['Plan a rollout']);
    });

    it('keeps defaults when nothing is persisted', () => {
        useCortexStore.setState({ missionChat: [], advancedMode: false, workspaceChatScope: null });

        useCortexStore.getState().restorePersistedClientState();

        expect(useCortexStore.getState().advancedMode).toBe(false);
        expect(useCortexStore.getState().missionChat).toEqual([]);
    });
});
