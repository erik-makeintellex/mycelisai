import type { StreamSignal, SignalDetail, LogEntry } from '@/store/useCortexStore';

function stringifySignalPayload(payload: unknown): string {
    if (payload == null) return '';
    if (typeof payload === 'string') return payload;
    try {
        return JSON.stringify(payload);
    } catch {
        return String(payload);
    }
}

type RawSignalEnvelope = {
    meta?: {
        timestamp?: string;
        source_kind?: string;
        source_channel?: string;
        payload_kind?: string;
        team_id?: string;
        agent_id?: string;
        run_id?: string;
    };
    text?: string;
    payload?: unknown;
    type?: string;
    source?: string;
    level?: string;
    message?: string;
    timestamp?: string;
    topic?: string;
};

/** Normalize raw SSE input into the shared StreamSignal shape. */
export function normalizeIncomingSignal(raw: RawSignalEnvelope): StreamSignal {
    const meta = raw.meta ?? {};
    const payloadKind = typeof meta.payload_kind === 'string' ? meta.payload_kind : undefined;
    const sourceChannel =
        typeof meta.source_channel === 'string' && meta.source_channel.trim().length > 0
            ? meta.source_channel
            : raw.topic;
    const source =
        raw.source
        ?? (typeof meta.team_id === 'string' && meta.team_id.trim().length > 0 ? meta.team_id : undefined)
        ?? (typeof meta.agent_id === 'string' && meta.agent_id.trim().length > 0 ? meta.agent_id : undefined)
        ?? (typeof meta.source_kind === 'string' ? meta.source_kind : undefined)
        ?? 'system';
    const payload = raw.payload;
    const message =
        raw.message
        ?? raw.text
        ?? stringifySignalPayload(payload)
        ?? '';

    return {
        type: raw.type ?? payloadKind ?? 'unknown',
        source,
        level: raw.level,
        message,
        timestamp: raw.timestamp ?? meta.timestamp ?? new Date().toISOString(),
        payload,
        topic: sourceChannel,
        source_kind: typeof meta.source_kind === 'string' ? meta.source_kind : undefined,
        source_channel: sourceChannel,
        payload_kind: payloadKind,
        team_id: typeof meta.team_id === 'string' ? meta.team_id : undefined,
        agent_id: typeof meta.agent_id === 'string' ? meta.agent_id : undefined,
        run_id: typeof meta.run_id === 'string' ? meta.run_id : undefined,
    };
}

/** Convert a StreamSignal into a normalized SignalDetail. */
export function streamSignalToDetail(signal: StreamSignal): SignalDetail {
    return {
        type: signal.type ?? 'unknown',
        source: signal.source ?? 'system',
        level: signal.level,
        message: signal.message ?? JSON.stringify(signal.payload ?? {}),
        timestamp: signal.timestamp ?? new Date().toISOString(),
        topic: signal.topic,
        payload: signal.payload,
        source_kind: signal.source_kind,
        source_channel: signal.source_channel,
        payload_kind: signal.payload_kind,
        team_id: signal.team_id,
        agent_id: signal.agent_id,
        run_id: signal.run_id,
        trust_score: signal.payload?.trust_score,
    };
}

/** Convert a LogEntry (from /api/v1/memory/stream) to a normalized SignalDetail */
export function logEntryToDetail(entry: LogEntry): SignalDetail {
    return {
        type: entry.intent || entry.level,
        source: entry.source,
        level: entry.level,
        message: entry.message,
        timestamp: entry.timestamp,
        id: entry.id,
        trace_id: entry.trace_id,
        intent: entry.intent,
        context: entry.context,
    };
}
