import type { ChatMessage } from '@/store/cortexStoreTypes';

export const CHAT_STORAGE_KEY = 'mycelis-workspace-chat';
export const CHAT_SESSION_STORAGE_KEY = 'mycelis-workspace-chat-session';
const CHAT_STORAGE_KEY_LEGACY = 'mycelis-mission-chat';
const CHAT_MAX_PERSISTED = 200;

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

export function loadOrCreateChatSessionId(scope?: string | null): string | null {
    if (typeof window === 'undefined') return null;
    try {
        const key = buildChatSessionStorageKey(scope);
        const existing = localStorage.getItem(key);
        if (existing && existing.trim()) {
            return existing.trim();
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
        const msgs: ChatMessage[] = JSON.parse(raw);
        return Array.isArray(msgs) ? msgs.slice(-CHAT_MAX_PERSISTED) : [];
    } catch {
        return [];
    }
}

export function persistChat(messages: ChatMessage[], scope?: string | null) {
    if (typeof window === 'undefined') return;
    try {
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
