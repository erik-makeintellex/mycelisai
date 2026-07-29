import { vi, beforeEach, afterEach } from 'vitest';

// ── localStorage polyfill (jsdom sometimes lacks proper methods) ──
if (typeof globalThis.localStorage === 'undefined' || typeof globalThis.localStorage.getItem !== 'function') {
    const store: Record<string, string> = {};
    const localStorageMock: Storage = {
        getItem: (key: string) => store[key] ?? null,
        setItem: (key: string, value: string) => { store[key] = String(value); },
        removeItem: (key: string) => { delete store[key]; },
        clear: () => { for (const k in store) delete store[k]; },
        get length() { return Object.keys(store).length; },
        key: (i: number) => Object.keys(store)[i] ?? null,
    };
    Object.defineProperty(globalThis, 'localStorage', {
        configurable: true,
        value: localStorageMock,
    });
}

// ── Global fetch mock ────────────────────────────────────────
// Each test file configures its own fetch responses via mockFetch.
export const mockFetch = vi.fn();

beforeEach(() => {
    // Reset fetch mock before each test
    mockFetch.mockReset();
    global.fetch = mockFetch;
});

afterEach(() => {
    vi.restoreAllMocks();
});

// ── EventSource mock ─────────────────────────────────────────
export class MockEventSource {
    static readonly CONNECTING = 0;
    static readonly OPEN = 1;
    static readonly CLOSED = 2;
    static instances: MockEventSource[] = [];
    url: string;
    onopen: ((ev: Event) => void) | null = null;
    onmessage: ((ev: MessageEvent<string>) => void) | null = null;
    onerror: ((ev: Event) => void) | null = null;
    readyState = 0;

    constructor(url: string) {
        this.url = url;
        MockEventSource.instances.push(this);
        // Auto-connect on next tick
        setTimeout(() => {
            this.readyState = 1;
            this.onopen?.(new Event('open'));
        }, 0);
    }

    close() {
        this.readyState = 2;
    }

    // Test helper: simulate a message
    simulateMessage(data: unknown) {
        this.onmessage?.(new MessageEvent('message', { data: JSON.stringify(data) }));
    }

    // Test helper: simulate error
    simulateError() {
        this.onerror?.(new Event('error'));
    }

    static reset() {
        MockEventSource.instances = [];
    }

    static latest(): MockEventSource | undefined {
        return MockEventSource.instances[MockEventSource.instances.length - 1];
    }
}

// Install EventSource mock globally
Object.defineProperty(globalThis, 'EventSource', {
    configurable: true,
    value: MockEventSource,
});

// ── Next.js navigation mock ──────────────────────────────────
vi.mock('next/navigation', () => ({
    useRouter: () => ({
        push: vi.fn(),
        replace: vi.fn(),
        back: vi.fn(),
        prefetch: vi.fn(),
    }),
    usePathname: () => '/dashboard',
}));
