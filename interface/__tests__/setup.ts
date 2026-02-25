import { vi, beforeEach, afterEach } from 'vitest';

// ── localStorage polyfill (jsdom sometimes lacks proper methods) ──
if (typeof globalThis.localStorage === 'undefined' || typeof globalThis.localStorage.getItem !== 'function') {
    const store: Record<string, string> = {};
    (globalThis as any).localStorage = {
        getItem: (key: string) => store[key] ?? null,
        setItem: (key: string, value: string) => { store[key] = String(value); },
        removeItem: (key: string) => { delete store[key]; },
        clear: () => { for (const k in store) delete store[k]; },
        get length() { return Object.keys(store).length; },
        key: (i: number) => Object.keys(store)[i] ?? null,
    };
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
    static instances: MockEventSource[] = [];
    url: string;
    onopen: ((ev: any) => void) | null = null;
    onmessage: ((ev: any) => void) | null = null;
    onerror: ((ev: any) => void) | null = null;
    readyState = 0;

    constructor(url: string) {
        this.url = url;
        MockEventSource.instances.push(this);
        // Auto-connect on next tick
        setTimeout(() => {
            this.readyState = 1;
            this.onopen?.({});
        }, 0);
    }

    close() {
        this.readyState = 2;
    }

    // Test helper: simulate a message
    simulateMessage(data: any) {
        this.onmessage?.({ data: JSON.stringify(data) });
    }

    // Test helper: simulate error
    simulateError() {
        this.onerror?.({});
    }

    static reset() {
        MockEventSource.instances = [];
    }

    static latest(): MockEventSource | undefined {
        return MockEventSource.instances[MockEventSource.instances.length - 1];
    }
}

// Install EventSource mock globally
(global as any).EventSource = MockEventSource;

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
