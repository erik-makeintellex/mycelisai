import type { ChatMessage, SomaThreadEvent, StreamSignal } from '@/store/cortexStoreTypes';

type RawThreadEvent = Partial<SomaThreadEvent> & {
    hrefLabel?: string;
    targetReference?: string;
};

const THREAD_EVENT_KINDS = new Set(['execution_started', 'execution_update', 'result_ready', 'attention_required']);
const THREAD_EVENT_TONES = new Set(['info', 'success', 'warning', 'danger']);

export function chatMessageFromThreadSignal(signal: StreamSignal): ChatMessage | null {
    if (signal.payload_kind !== 'thread_event' && signal.type !== 'thread_event') return null;
    const payload = signal.payload;
    if (!payload || typeof payload !== 'object') return null;
    const raw = payload as RawThreadEvent;
    const kind = typeof raw.kind === 'string' && THREAD_EVENT_KINDS.has(raw.kind)
        ? raw.kind as SomaThreadEvent['kind']
        : 'execution_update';
    const tone = typeof raw.tone === 'string' && THREAD_EVENT_TONES.has(raw.tone)
        ? raw.tone as SomaThreadEvent['tone']
        : toneForKind(kind);
    const label = textValue(raw.label) ?? textValue(raw.title) ?? titleForKind(kind);
    const detail = textValue(raw.detail) ?? signal.message ?? undefined;
    const href = textValue(raw.href);
    const hrefLabel = textValue(raw.href_label) ?? textValue(raw.hrefLabel);
    const targetReference = textValue(raw.target_reference) ?? textValue(raw.targetReference);

    return {
        role: 'system',
        content: [label, detail].filter(Boolean).join(' - '),
        mode: kind === 'attention_required' ? 'blocker' : 'execution_result',
        run_id: signal.run_id,
        timestamp: signal.timestamp,
        thread_events: [{
            kind,
            label,
            title: label,
            detail,
            tone,
            timestamp: signal.timestamp,
            status: textValue(raw.status),
            run_id: signal.run_id,
            team_id: signal.team_id,
            agent_id: signal.agent_id,
            source_kind: signal.source_kind,
            source_channel: signal.source_channel,
            payload_kind: signal.payload_kind,
            href,
            href_label: hrefLabel,
            target_reference: targetReference,
        }],
        thread_event: {
            kind,
            label,
            title: label,
            detail,
            tone,
            timestamp: signal.timestamp,
            status: textValue(raw.status),
            run_id: signal.run_id,
            team_id: signal.team_id,
            agent_id: signal.agent_id,
            source_kind: signal.source_kind,
            source_channel: signal.source_channel,
            payload_kind: signal.payload_kind,
            href,
            href_label: hrefLabel,
            target_reference: targetReference,
        },
    };
}

function textValue(value: unknown) {
    return typeof value === 'string' && value.trim().length > 0 ? value.trim() : undefined;
}

function toneForKind(kind: SomaThreadEvent['kind']): SomaThreadEvent['tone'] {
    if (kind === 'result_ready') return 'success';
    if (kind === 'attention_required') return 'warning';
    return 'info';
}

function titleForKind(kind: SomaThreadEvent['kind']) {
    if (kind === 'execution_started') return 'Execution started';
    if (kind === 'result_ready') return 'Result ready';
    if (kind === 'attention_required') return 'Needs attention';
    return 'Work updated';
}
