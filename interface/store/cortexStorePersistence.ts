import type { AskClass, ChatArtifactRef, ChatConsultation, ChatMessage, ExecutionMode, ProposalLifecycleStatus, ResponseDepth, SomaThreadEvent, TemplateID } from '@/store/cortexStoreTypes';
import { normalizeProposalData } from '@/store/cortexStoreProposalData';

export const CHAT_STORAGE_KEY = 'mycelis-workspace-chat';
export const CHAT_SESSION_STORAGE_KEY = 'mycelis-workspace-chat-session';
const CHAT_STORAGE_KEY_LEGACY = 'mycelis-mission-chat';
const CHAT_MAX_PERSISTED = 200;
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const VALID_ROLES = new Set<ChatMessage['role']>(['user', 'architect', 'admin', 'council', 'system']);
const VALID_ASK_CLASSES = new Set<AskClass>(['direct_answer', 'governed_mutation', 'governed_artifact', 'specialist_consultation', 'execution_blocker']);
const VALID_TEMPLATE_IDS = new Set<TemplateID>(['chat-to-answer', 'chat-to-proposal']);
const VALID_MODES = new Set<ExecutionMode>(['answer', 'proposal', 'execution_result', 'blocker']);
const VALID_RESPONSE_DEPTHS = new Set<ResponseDepth>(['quick_box', 'structured_summary', 'decision_brief', 'execution_proposal']);
const VALID_PROPOSAL_STATUSES = new Set<ProposalLifecycleStatus>(['active', 'cancelled', 'confirmed_pending_execution', 'executed', 'failed']);

export function buildChatStorageKey(scope?: string | null): string {
    const normalizedScope = typeof scope === 'string' ? scope.trim() : '';
    return normalizedScope ? `${CHAT_STORAGE_KEY}:${normalizedScope}` : CHAT_STORAGE_KEY;
}

export function buildChatSessionStorageKey(scope?: string | null): string {
    const normalizedScope = typeof scope === 'string' ? scope.trim() : '';
    return normalizedScope ? `${CHAT_SESSION_STORAGE_KEY}:${normalizedScope}` : CHAT_SESSION_STORAGE_KEY;
}

function fallbackUUID(): string {
    const values = new Uint8Array(16);
    if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
        crypto.getRandomValues(values);
    } else {
        for (let i = 0; i < values.length; i += 1) {
            values[i] = Math.floor(Math.random() * 256);
        }
    }
    values[6] = (values[6] & 0x0f) | 0x40;
    values[8] = (values[8] & 0x3f) | 0x80;
    const hex = Array.from(values, (value) => value.toString(16).padStart(2, '0')).join('');
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

function createChatSessionId(): string {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID();
    }
    return fallbackUUID();
}

function isValidChatSessionId(value: string): boolean {
    return UUID_PATTERN.test(value.trim());
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return Boolean(value && typeof value === 'object' && !Array.isArray(value));
}

function boundedString(value: unknown, fallback = ''): string {
    if (typeof value === 'string') return value;
    if (typeof value === 'number' || typeof value === 'boolean') return String(value);
    if (value == null) return fallback;
    try {
        return JSON.stringify(value).slice(0, 1600);
    } catch {
        return fallback;
    }
}

function optionalString(value: unknown): string | undefined {
    const text = boundedString(value).trim();
    return text || undefined;
}

function optionalNumber(value: unknown): number | undefined {
    return typeof value === 'number' && Number.isFinite(value) ? value : undefined;
}

function stringList(value: unknown): string[] | undefined {
    if (!Array.isArray(value)) return undefined;
    const values = value
        .map((item) => optionalString(item))
        .filter((item): item is string => Boolean(item));
    return values.length ? [...new Set(values)] : undefined;
}

function enumValue<T extends string>(value: unknown, allowed: Set<T>): T | undefined {
    return typeof value === 'string' && allowed.has(value as T) ? value as T : undefined;
}

function sanitizeConsultations(value: unknown): ChatConsultation[] | undefined {
    if (!Array.isArray(value)) return undefined;
    const consultations = value.flatMap((item) => {
        if (!isRecord(item)) return [];
        const member = optionalString(item.member);
        const summary = optionalString(item.summary);
        return member && summary ? [{ member, summary }] : [];
    });
    return consultations.length ? consultations : undefined;
}

function sanitizeArtifacts(value: unknown): ChatArtifactRef[] | undefined {
    if (!Array.isArray(value)) return undefined;
    const artifacts = value.flatMap((item) => {
        if (!isRecord(item)) return [];
        const type = optionalString(item.type);
        const title = optionalString(item.title);
        if (!type || !title) return [];
        return [{
            id: optionalString(item.id),
            type,
            output_class: optionalString(item.output_class),
            title,
            content_type: optionalString(item.content_type),
            content: optionalString(item.content),
            url: optionalString(item.url),
            cached: typeof item.cached === 'boolean' ? item.cached : undefined,
            expires_at: optionalString(item.expires_at),
            saved_path: optionalString(item.saved_path),
            entrypoint: optionalString(item.entrypoint),
            folder: optionalString(item.folder),
            files: stringList(item.files),
            validation: optionalString(item.validation),
        }];
    });
    return artifacts.length ? artifacts : undefined;
}

function sanitizeThreadEvent(value: unknown): SomaThreadEvent | undefined {
    if (!isRecord(value)) return undefined;
    const kind = optionalString(value.kind) as SomaThreadEvent['kind'] | undefined;
    const label = optionalString(value.label);
    const tone = optionalString(value.tone) as SomaThreadEvent['tone'] | undefined;
    if (!kind || !label || !tone) return undefined;
    return {
        id: optionalString(value.id),
        kind,
        label,
        title: optionalString(value.title),
        detail: optionalString(value.detail),
        tone,
        timestamp: optionalString(value.timestamp),
        status: optionalString(value.status),
        run_id: optionalString(value.run_id),
        team_id: optionalString(value.team_id),
        agent_id: optionalString(value.agent_id),
        source_kind: optionalString(value.source_kind),
        source_channel: optionalString(value.source_channel),
        payload_kind: optionalString(value.payload_kind),
        href: optionalString(value.href),
        href_label: optionalString(value.href_label),
        target_reference: optionalString(value.target_reference),
    };
}

function sanitizeThreadEvents(value: unknown): SomaThreadEvent[] | undefined {
    if (!Array.isArray(value)) return undefined;
    const events = value
        .map(sanitizeThreadEvent)
        .filter((item): item is SomaThreadEvent => Boolean(item));
    return events.length ? events : undefined;
}

function sanitizePersistedMessage(value: unknown): ChatMessage | null {
    if (!isRecord(value)) return null;
    const role = enumValue(value.role, VALID_ROLES);
    if (!role) return null;
    const content = boundedString(value.content);
    const proposal = normalizeProposalData(value.proposal);
    const threadEvent = sanitizeThreadEvent(value.thread_event);
    const threadEvents = sanitizeThreadEvents(value.thread_events);
    const executionSummary = isRecord(value.execution_summary) ? value.execution_summary as ChatMessage['execution_summary'] : undefined;
    const hasStructuredContent = Boolean(proposal || executionSummary || threadEvent || threadEvents?.length);
    if (!content.trim() && role !== 'system' && !hasStructuredContent) return null;
    return {
        role,
        content,
        consultations: sanitizeConsultations(value.consultations),
        tools_used: stringList(value.tools_used),
        source_node: optionalString(value.source_node),
        trust_score: optionalNumber(value.trust_score),
        timestamp: optionalString(value.timestamp),
        artifacts: sanitizeArtifacts(value.artifacts),
        ask_class: enumValue(value.ask_class, VALID_ASK_CLASSES),
        template_id: enumValue(value.template_id, VALID_TEMPLATE_IDS),
        mode: enumValue(value.mode, VALID_MODES),
        response_depth: enumValue(value.response_depth, VALID_RESPONSE_DEPTHS),
        ui_response_state: isRecord(value.ui_response_state) ? value.ui_response_state as unknown as ChatMessage['ui_response_state'] : undefined,
        provenance: isRecord(value.provenance) ? value.provenance as unknown as ChatMessage['provenance'] : undefined,
        proposal: proposal ?? undefined,
        proposal_status: enumValue(value.proposal_status, VALID_PROPOSAL_STATUSES),
        execution_summary: executionSummary,
        brain: isRecord(value.brain) ? value.brain as unknown as ChatMessage['brain'] : undefined,
        run_id: optionalString(value.run_id),
        thread_event: threadEvent,
        thread_events: threadEvents,
    };
}

function sanitizePersistedMessages(value: unknown): ChatMessage[] {
    if (!Array.isArray(value)) return [];
    return value
        .map(sanitizePersistedMessage)
        .filter((item): item is ChatMessage => Boolean(item))
        .slice(-CHAT_MAX_PERSISTED);
}

export function loadOrCreateChatSessionId(scope?: string | null): string | null {
    if (typeof window === 'undefined') return null;
    try {
        const key = buildChatSessionStorageKey(scope);
        const existing = localStorage.getItem(key);
        const normalizedExisting = existing?.trim();
        if (normalizedExisting && isValidChatSessionId(normalizedExisting)) {
            return normalizedExisting;
        }
        const next = createChatSessionId();
        localStorage.setItem(key, next);
        return next;
    } catch {
        return null;
    }
}

// Soma's memory: chat survives page refreshes. Use clearPersistedChat to reset.
export function loadPersistedChat(scope?: string | null): ChatMessage[] {
    if (typeof window === 'undefined') return [];
    try {
        const scopedKey = buildChatStorageKey(scope);
        const raw = localStorage.getItem(scopedKey) ?? (!scope ? localStorage.getItem(CHAT_STORAGE_KEY_LEGACY) : null);
        if (!raw) return [];
        return sanitizePersistedMessages(JSON.parse(raw));
    } catch {
        return [];
    }
}

export function persistChat(messages: ChatMessage[], scope?: string | null) {
    if (typeof window === 'undefined') return;
    try {
        if (messages.length === 0) {
            localStorage.removeItem(buildChatStorageKey(scope));
            return;
        }
        localStorage.setItem(buildChatStorageKey(scope), JSON.stringify(messages.slice(-CHAT_MAX_PERSISTED)));
    } catch {
        // quota exceeded - silently drop
    }
}

export function clearPersistedChat(scope?: string | null) {
    if (typeof window === 'undefined') return;
    try {
        localStorage.removeItem(buildChatStorageKey(scope));
        localStorage.removeItem(buildChatSessionStorageKey(scope));
        if (!scope) localStorage.removeItem(CHAT_STORAGE_KEY_LEGACY);
    } catch {
        // ignore localStorage failures
    }
}

export function clearAllPersistedChat() {
    if (typeof window === 'undefined') return;
    try {
        const keys: string[] = [];
        for (let index = 0; index < localStorage.length; index += 1) {
            const key = localStorage.key(index);
            if (!key) continue;
            if (
                key === CHAT_STORAGE_KEY_LEGACY
                || key === CHAT_STORAGE_KEY
                || key === CHAT_SESSION_STORAGE_KEY
                || key.startsWith(`${CHAT_STORAGE_KEY}:`)
                || key.startsWith(`${CHAT_SESSION_STORAGE_KEY}:`)
            ) {
                keys.push(key);
            }
        }
        keys.forEach((key) => localStorage.removeItem(key));
    } catch {
        // ignore localStorage failures
    }
}
