import type { CortexState } from '@/store/cortexStoreState';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import type { CTSEnvelope } from '@/store/cortexStoreTypes';
import { dispatchSignalToNodes } from '@/store/cortexStoreUtils';
import { normalizeIncomingSignal } from '@/lib/signalNormalize';

let eventSourceRef: EventSource | null = null;

export function createCortexStreamSlice(
    set: CortexSet,
    get: CortexGet,
): Pick<CortexState, 'initializeStream' | 'disconnectStream'> {
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
                set({ isStreamConnected: true, streamConnectionState: 'online' });
            };

            source.onmessage = (event) => {
                try {
                    const signal = normalizeIncomingSignal(JSON.parse(event.data));
                    const { nodes } = get();
                    const nextLogs = [signal, ...get().streamLogs].slice(0, 100);
                    const updatedNodes = dispatchSignalToNodes(signal, nodes);

                    const patch: Partial<CortexState> = updatedNodes
                        ? { streamLogs: nextLogs, nodes: updatedNodes }
                        : { streamLogs: nextLogs };

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
                } catch (error) {
                    console.error('Stream parse error', error);
                }
            };

            source.onerror = () => {
                set({ isStreamConnected: false, streamConnectionState: 'offline' });
                eventSourceRef = null;
                source.close();
            };

            eventSourceRef = source;
        },

        disconnectStream: () => {
            if (eventSourceRef) {
                eventSourceRef.close();
                eventSourceRef = null;
            }
            set({ isStreamConnected: false, streamConnectionState: 'idle' });
        },
    };
}
