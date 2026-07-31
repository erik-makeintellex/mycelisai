import { afterEach, describe, expect, it, vi } from 'vitest';
import { createCortexStreamSlice } from '@/store/cortexStoreStreamSlice';
import type { CortexSet } from '@/store/cortexStoreSliceTypes';
import type { CortexState } from '@/store/cortexStoreState';
import { MockEventSource } from '../setup';

describe('cortexStoreStreamSlice', () => {
    let state: CortexState;
    let slice: ReturnType<typeof createCortexStreamSlice>;

    const signal = (message: string) => ({
        type: 'status',
        source: 'team-alpha',
        message,
        timestamp: '2026-07-31T12:00:00Z',
    });

    const send = (source: MockEventSource, data: unknown, lastEventId = '') => {
        source.onmessage?.(new MessageEvent('message', {
            data: typeof data === 'string' ? data : JSON.stringify(data),
            lastEventId,
        }));
    };

    const createSlice = () => {
        state = {
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
        slice = createCortexStreamSlice(set, () => state);
        return slice;
    };

    afterEach(() => {
        slice?.disconnectStream();
        MockEventSource.reset();
    });

    it('lets EventSource reconnect after a transient stream error', () => {
        createSlice();

        slice.initializeStream();
        const source = MockEventSource.latest();
        expect(source).toBeDefined();
        source!.readyState = EventSource.CONNECTING;
        source!.simulateError();

        expect(source!.readyState).toBe(EventSource.CONNECTING);
        expect(state.streamConnectionState).toBe('connecting');
        expect(state.isStreamConnected).toBe(false);

        source!.readyState = EventSource.OPEN;
        source!.onopen?.(new Event('open'));

        expect(state.streamConnectionState).toBe('online');
        expect(state.isStreamConnected).toBe(true);
    });

    it('applies unique replay ids once while preserving ordinary signals', () => {
        createSlice().initializeStream();
        const source = MockEventSource.latest()!;

        send(source, signal('first'), 'event-1');
        send(source, signal('second'), 'event-2');

        expect(state.streamLogs.map((entry) => entry.message)).toEqual(['second', 'first']);
    });

    it('suppresses a duplicate event id without applying its replay payload', () => {
        createSlice().initializeStream();
        const source = MockEventSource.latest()!;

        send(source, signal('original'), 'event-1');
        send(source, signal('duplicate replay'), 'event-1');

        expect(state.streamLogs).toHaveLength(1);
        expect(state.streamLogs[0]?.message).toBe('original');
    });

    it('preserves id-less, artifact, and thread-event behavior', () => {
        createSlice().initializeStream();
        const source = MockEventSource.latest()!;

        send(source, signal('id-less first'));
        send(source, signal('id-less second'));
        send(source, {
            type: 'artifact',
            source: 'team-alpha',
            message: 'retained output',
            timestamp: '2026-07-31T12:00:01Z',
        }, 'artifact-1');
        send(source, {
            type: 'thread_event',
            source: 'team-alpha',
            message: 'work is running',
            timestamp: '2026-07-31T12:00:02Z',
            meta: { payload_kind: 'thread_event' },
            payload: { kind: 'execution_started', label: 'Work started' },
        }, 'thread-1');

        expect(state.streamLogs).toHaveLength(4);
        expect(state.pendingArtifacts).toHaveLength(1);
        expect(state.missionChat).toHaveLength(1);
        expect(state.missionChat[0]?.content).toContain('Work started');
    });

    it('retains replay ids when force recreates EventSource', () => {
        createSlice().initializeStream();
        const firstSource = MockEventSource.latest()!;
        send(firstSource, signal('original'), 'event-1');

        slice.initializeStream(true);
        const replacementSource = MockEventSource.latest()!;
        send(replacementSource, signal('replayed'), 'event-1');
        send(replacementSource, signal('new'), 'event-2');

        expect(firstSource.readyState).toBe(EventSource.CLOSED);
        expect(state.streamConnectionState).toBe('connecting');
        expect(state.streamLogs.map((entry) => entry.message)).toEqual(['new', 'original']);
    });

    it('ignores malformed events without consuming their replay id', () => {
        const error = vi.spyOn(console, 'error').mockImplementation(() => undefined);
        createSlice().initializeStream();
        const source = MockEventSource.latest()!;

        send(source, '{not-json', 'event-1');
        send(source, signal('corrected replay'), 'event-1');

        expect(error).toHaveBeenCalledWith('Stream parse error', expect.any(SyntaxError));
        expect(state.streamLogs.map((entry) => entry.message)).toEqual(['corrected replay']);
    });

    it('clears replay ids on explicit disconnect for a new module session', () => {
        createSlice().initializeStream();
        send(MockEventSource.latest()!, signal('before disconnect'), 'event-1');

        slice.disconnectStream();
        slice.initializeStream();
        send(MockEventSource.latest()!, signal('after disconnect'), 'event-1');

        expect(state.streamLogs.map((entry) => entry.message)).toEqual([
            'after disconnect',
            'before disconnect',
        ]);
    });

    it('bounds retained replay ids and evicts the oldest id', () => {
        createSlice().initializeStream();
        const source = MockEventSource.latest()!;
        for (let index = 0; index <= 256; index += 1) {
            send(source, signal(`event ${index}`), `event-${index}`);
        }

        send(source, signal('oldest replay accepted'), 'event-0');
        send(source, signal('newest replay suppressed'), 'event-256');

        expect(state.streamLogs[0]?.message).toBe('oldest replay accepted');
        expect(state.streamLogs.some((entry) => entry.message === 'newest replay suppressed')).toBe(false);
    });
});
