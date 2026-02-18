import type { StreamSignal, SignalDetail, LogEntry } from '@/store/useCortexStore';

/** Convert an SSE StreamSignal (or SignalContext Signal) to a normalized SignalDetail */
export function streamSignalToDetail(signal: StreamSignal): SignalDetail {
    return {
        type: signal.type ?? 'unknown',
        source: signal.source ?? 'system',
        level: signal.level,
        message: signal.message ?? JSON.stringify(signal.payload ?? {}),
        timestamp: signal.timestamp ?? new Date().toISOString(),
        topic: signal.topic,
        payload: signal.payload,
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
