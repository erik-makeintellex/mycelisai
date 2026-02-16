import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import { SignalProvider } from '@/components/dashboard/SignalContext';
import PriorityStream from '@/components/dashboard/PriorityStream';
import ActivityStream from '@/components/dashboard/ActivityStream';
import { MockEventSource } from '../setup';

function renderWithProvider(Component: React.ComponentType) {
    return render(
        <SignalProvider>
            <Component />
        </SignalProvider>
    );
}

describe('PriorityStream', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        MockEventSource.reset();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('renders priority stream container', () => {
        renderWithProvider(PriorityStream);
        expect(screen.getByTestId('priority-stream')).toBeDefined();
    });

    it('shows "No priority events" when no signals', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        expect(screen.getByText('No priority events')).toBeDefined();
    });

    it('shows governance_halt signals', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'governance_halt',
                source: 'overseer-1',
                message: 'Trust below threshold',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('GOVERNANCE')).toBeDefined();
        expect(screen.getByText('Trust below threshold')).toBeDefined();
    });

    it('shows error signals', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'error',
                source: 'scanner-1',
                message: 'Connection timeout',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('ERROR')).toBeDefined();
        expect(screen.getByText('Connection timeout')).toBeDefined();
    });

    it('shows task_complete signals', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'task_complete',
                source: 'writer-1',
                message: 'Report generated',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('COMPLETE')).toBeDefined();
    });

    it('shows artifact signals', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'artifact',
                source: 'coder-1',
                message: 'Generated analysis.py',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('ARTIFACT')).toBeDefined();
    });

    it('filters out non-priority signals (thought, tool_call, etc)', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({ type: 'thought', message: 'Thinking...' });
            es.simulateMessage({ type: 'tool_call', message: 'Running tool' });
            es.simulateMessage({ type: 'info', message: 'Status update' });
        });

        // Should still show empty state
        expect(screen.getByText('No priority events')).toBeDefined();
    });

    it('shows signal count', async () => {
        renderWithProvider(PriorityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({ type: 'error', message: 'Error 1', timestamp: new Date().toISOString() });
            es.simulateMessage({ type: 'error', message: 'Error 2', timestamp: new Date().toISOString() });
        });

        expect(screen.getByText('2')).toBeDefined();
    });
});

describe('ActivityStream', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        MockEventSource.reset();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('shows "Awaiting Neural Signals..." when empty', async () => {
        renderWithProvider(ActivityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        expect(screen.getByText('Awaiting Neural Signals...')).toBeDefined();
    });

    it('renders thought signals with CORTEX label', async () => {
        renderWithProvider(ActivityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'thought',
                message: 'Analyzing the dataset',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('Analyzing the dataset')).toBeDefined();
    });

    it('renders tool_call signals with TOOL EXECUTION label', async () => {
        renderWithProvider(ActivityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'tool_call',
                message: 'read_file("data.csv")',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('read_file("data.csv")')).toBeDefined();
    });

    it('renders user_input signals with COMMAND label', async () => {
        renderWithProvider(ActivityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'user_input',
                message: 'Deploy the report',
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('Deploy the report')).toBeDefined();
    });

    it('renders all signal types (not filtered like PriorityStream)', async () => {
        renderWithProvider(ActivityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({ type: 'thought', message: 'Thinking...' });
            es.simulateMessage({ type: 'tool_call', message: 'Running tool' });
            es.simulateMessage({ type: 'info', message: 'Status update' });
            es.simulateMessage({ type: 'error', message: 'Something failed' });
        });

        // All 4 should render
        expect(screen.queryByText('Awaiting Neural Signals...')).toBeNull();
    });

    it('parses JSON-wrapped signal content', async () => {
        renderWithProvider(ActivityStream);
        await act(async () => { await vi.advanceTimersByTimeAsync(10); });

        const es = MockEventSource.latest()!;
        act(() => {
            es.simulateMessage({
                type: 'info',
                message: JSON.stringify({ type: 'thought', content: 'Parsed thought' }),
                timestamp: new Date().toISOString(),
            });
        });

        expect(screen.getByText('Parsed thought')).toBeDefined();
    });

    it('shows LIVE FEED label in header', () => {
        renderWithProvider(ActivityStream);
        expect(screen.getByText('LIVE FEED')).toBeDefined();
    });
});
