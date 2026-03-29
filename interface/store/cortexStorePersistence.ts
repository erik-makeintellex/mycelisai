import type { ChatMessage } from '@/store/cortexStoreTypes';

export const CHAT_STORAGE_KEY = 'mycelis-workspace-chat';
const CHAT_STORAGE_KEY_LEGACY = 'mycelis-mission-chat';
const CHAT_MAX_PERSISTED = 200;

export function buildChatStorageKey(scope?: string | null): string {
    const normalizedScope = typeof scope === 'string' ? scope.trim() : '';
    return normalizedScope ? `${CHAT_STORAGE_KEY}:${normalizedScope}` : CHAT_STORAGE_KEY;
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
        if (!scope) localStorage.removeItem(CHAT_STORAGE_KEY_LEGACY);
    } catch {
        // ignore localStorage failures
    }
}
