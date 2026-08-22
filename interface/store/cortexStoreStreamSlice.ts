import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { CTSEnvelope } from '@/store/cortexStoreTypes';
import { dispatchSignalToNodes } from '@/store/cortexStoreUtils';
import { normalizeIncomingSignal } from '@/lib/signalNormalize';
import { chatMessageFromThreadSignal } from '@/store/cortexStoreThreadEvents';

let eventSourceRef: EventSource | null = null;
const RECENT_EVENT_ID_LIMIT = 256;
const recentEventIds = new Set<string>();

function hasRecentEventId(eventId: string): boolean {
    return eventId.length > 0 && recentEventIds.has(eventId);
}

function rememberEventId(eventId: string): void {
    if (eventId.length === 0 || recentEventIds.has(eventId)) return;
    recentEventIds.add(eventId);
    if (recentEventIds.size <= RECENT_EVENT_ID_LIMIT) return;
    const oldestEventId = recentEventIds.values().next().value;
    if (oldestEventId) recentEventIds.delete(oldestEventId);
}

export function createCortexStreamSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<'initializeStream' | 'disconnectStream'> {
    return {
        initializeStream: (force = false) => {
            if (eventSourceRef) {
                if (!force) return;
                eventSourceRef.close();
                eventSourceRef = null;
            }

            set({ isStreamConnected: false, streamConnectionState: 'connecting' });
            const source = new EventSource('/api/v1/stream');

            source.onopen = () => {
                if (eventSourceRef !== source) return;
                set({ isStreamConnected: true, streamConnectionState: 'online' });
            };

            source.onmessage = (event) => {
                if (eventSourceRef !== source || hasRecentEventId(event.lastEventId)) return;
                try {
                    const signal = normalizeIncomingSignal(JSON.parse(event.data));
                    const { nodes } = get();
                    const nextLogs = [signal, ...get().streamLogs].slice(0, 100);
                    const updatedNodes = dispatchSignalToNodes(signal, nodes);
                    const threadMessage = chatMessageFromThreadSignal(signal);

                    const patch: Partial<CortexState> = updatedNodes
                        ? { streamLogs: nextLogs, nodes: updatedNodes }
                        : { streamLogs: nextLogs };

                    if (threadMessage) {
                        patch.missionChat = [...get().missionChat, threadMessage];
                    }

                    if (signal.type === 'artifact' && signal.source) {
                        const envelope: CTSEnvelope = {
                            id: `${signal.source}-${signal.timestamp ?? Date.now()}`,
                            source: signal.source,
                            signal: 'artifact',
                            timestamp: signal.timestamp ?? new Date().toISOString(),
                            trust_score: signal.payload?.trust_score,
                            payload: {
                                content: signal.message ?? JSON.stringify(signal.payload ?? {}),
                                content_type: signal.payload?.content_type ?? 'text',
                                title: signal.payload?.title,
                            },
                            proof: signal.payload?.proof,
                        };
                        patch.pendingArtifacts = [envelope, ...get().pendingArtifacts];
                    }

                    if (signal.type === 'governance_halt' && signal.source) {
                        const envelope: CTSEnvelope = {
                            id: `gov-${signal.source}-${signal.timestamp ?? Date.now()}`,
                            source: signal.source,
                            signal: 'governance_halt',
                            timestamp: signal.timestamp ?? new Date().toISOString(),
                            trust_score: signal.payload?.trust_score ?? signal.trust_score,
                            payload: {
                                content: 'Trust score below threshold. Awaiting human approval.',
                                content_type: 'text',
                                title: `Governance Halt: ${signal.source}`,
                            },
                        };
                        patch.pendingArtifacts = [envelope, ...get().pendingArtifacts];
                    }

                    set(patch);
                    rememberEventId(event.lastEventId);
                } catch (error) {
                    console.error('Stream parse error', error);
                }
            };

            source.onerror = () => {
                if (eventSourceRef !== source) return;
                const closed = source.readyState === EventSource.CLOSED;
                set({
                    isStreamConnected: false,
                    streamConnectionState: closed ? 'offline' : 'connecting',
                });
            };

            eventSourceRef = source;
        },

        disconnectStream: () => {
            if (eventSourceRef) {
                eventSourceRef.close();
                eventSourceRef = null;
            }
            recentEventIds.clear();
            set({ isStreamConnected: false, streamConnectionState: 'idle' });
        },
    };
}
