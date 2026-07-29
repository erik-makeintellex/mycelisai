import { afterEach, describe, expect, it } from 'vitest';
import { createCortexStreamSlice } from '@/store/cortexStoreStreamSlice';
import type { CortexSet } from '@/store/cortexStoreSliceTypes';
import type { CortexState } from '@/store/cortexStoreState';
import { MockEventSource } from '../setup';

describe('cortexStoreStreamSlice', () => {
    afterEach(() => {
        MockEventSource.reset();
    });

    it('lets EventSource reconnect after a transient stream error', () => {
        let state = {
            isStreamConnected: false,
            streamConnectionState: 'idle',
            nodes: [],
            streamLogs: [],
            missionChat: [],
            pendingArtifacts: [],
        } as unknown as CortexState;
        const set: CortexSet = (partial) => {
            const patch = typeof partial === 'function' ? partial(state) : partial;
            state = { ...state, ...patch };
        };
        const slice = createCortexStreamSlice(set, () => state);

        slice.initializeStream();
        const source = MockEventSource.latest();
        expect(source).toBeDefined();
        source!.readyState = EventSource.CONNECTING;
        source!.simulateError();

        expect(source!.readyState).toBe(EventSource.CONNECTING);
        expect(state.streamConnectionState).toBe('connecting');
        slice.disconnectStream();
    });
});
