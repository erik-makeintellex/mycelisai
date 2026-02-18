import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { SignalProvider, useSignalStream } from '@/components/dashboard/SignalContext';
import { MockEventSource } from '../setup';

function SignalConsumer() {
    const { isConnected, signals, clearSignals } = useSignalStream();
    return (
        <div>
            <span data-testid="connected">{isConnected ? 'LIVE' : 'OFFLINE'}</span>
            <span data-testid="count">{signals.length}</span>
            <ul data-testid="signals">
                {signals.map((s, i) => (
                    <li key={i} data-testid="signal">{s.type}: {s.message}</li>
                ))}
            </ul>
            <button data-testid="clear" onClick={clearSignals}>Clear</button>
        </div>
    );
}

describe('SignalContext', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        MockEventSource.reset();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('creates EventSource to /api/v1/stream', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        expect(MockEventSource.instances.length).toBe(1);
        expect(MockEventSource.latest()!.url).toBe('/api/v1/stream');
    });

    it('reports connected after EventSource opens', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        // Initially OFFLINE
        expect(screen.getByTestId('connected').textContent).toBe('OFFLINE');

        // Auto-connect fires on next tick
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        expect(screen.getByTestId('connected').textContent).toBe('LIVE');
    });

    it('receives and stores signals', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;

        act(() => {
            es.simulateMessage({ type: 'thought', message: 'Analyzing data', source: 'scanner-1' });
        });

        expect(screen.getByTestId('count').textContent).toBe('1');
        expect(screen.getByText('thought: Analyzing data')).toBeDefined();
    });

    it('caps signals at 100', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;

        act(() => {
            for (let i = 0; i < 150; i++) {
                es.simulateMessage({ type: 'info', message: `Signal ${i}` });
            }
        });

        expect(screen.getByTestId('count').textContent).toBe('100');
    });

    it('reconnects after error and creates new EventSource', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        await act(async () => { await vi.advanceTimersByTimeAsync(10); });
        expect(MockEventSource.instances.length).toBe(1);

        // First error → goes OFFLINE
        act(() => { MockEventSource.latest()!.simulateError(); });
        expect(screen.getByTestId('connected').textContent).toBe('OFFLINE');

        // After base reconnect delay, creates new EventSource
        await act(async () => { await vi.advanceTimersByTimeAsync(2100); });
        expect(MockEventSource.instances.length).toBe(2);

        // New connection auto-opens → goes LIVE again
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });
        expect(screen.getByTestId('connected').textContent).toBe('LIVE');

        // Second error → disconnects again, but will reconnect
        act(() => { MockEventSource.latest()!.simulateError(); });
        expect(screen.getByTestId('connected').textContent).toBe('OFFLINE');

        // After another delay, a third EventSource is created
        await act(async () => { await vi.advanceTimersByTimeAsync(2200); });
        expect(MockEventSource.instances.length).toBe(3);
    });

    it('resets retry counter on successful reconnect', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        // Error → reconnect at 2s
        act(() => { MockEventSource.latest()!.simulateError(); });
        await act(async () => { await vi.advanceTimersByTimeAsync(2100); });
        expect(MockEventSource.instances.length).toBe(2);

        // Successful open → retry counter resets
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });
        expect(screen.getByTestId('connected').textContent).toBe('LIVE');

        // Another error → should reconnect at 2s again (not 4s)
        act(() => { MockEventSource.latest()!.simulateError(); });
        await act(async () => { await vi.advanceTimersByTimeAsync(2100); });
        expect(MockEventSource.instances.length).toBe(3);
    });

    it('clears signals when clearSignals is called', async () => {
        render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        act(() => {
            MockEventSource.latest()!.simulateMessage({ type: 'info', message: 'test' });
        });

        expect(screen.getByTestId('count').textContent).toBe('1');

        act(() => {
            screen.getByTestId('clear').click();
        });

        expect(screen.getByTestId('count').textContent).toBe('0');
    });

    it('closes EventSource on unmount', async () => {
        const { unmount } = render(
            <SignalProvider>
                <SignalConsumer />
            </SignalProvider>
        );

        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        expect(es.readyState).toBe(1);

        unmount();

        expect(es.readyState).toBe(2); // closed
    });
});
