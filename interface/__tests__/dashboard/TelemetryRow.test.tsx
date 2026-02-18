import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import TelemetryRow from '@/components/dashboard/TelemetryRow';
import { mockFetch } from '../setup';

const TELEMETRY_RESPONSE = {
    goroutines: 42,
    heap_alloc_mb: 12.5,
    sys_mem_mb: 64,
    llm_tokens_sec: 3.7,
    timestamp: new Date().toISOString(),
};

describe('TelemetryRow', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('renders loading skeleton initially', () => {
        mockFetch.mockReturnValue(new Promise(() => {})); // never resolves
        render(<TelemetryRow />);
        const row = screen.getByTestId('telemetry-row');
        expect(row).toBeDefined();
        // Should have pulsing skeleton divs
        expect(row.querySelectorAll('.animate-pulse').length).toBe(4);
    });

    it('renders 4 metric cards on successful fetch', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => TELEMETRY_RESPONSE,
        });

        render(<TelemetryRow />);

        await waitFor(() => {
            expect(screen.getAllByTestId('telemetry-card')).toHaveLength(4);
        });

        expect(screen.getByText('42')).toBeDefined();
        expect(screen.getByText('12.5')).toBeDefined();
        expect(screen.getByText('64')).toBeDefined();
        expect(screen.getByText('3.7')).toBeDefined();
    });

    it('shows TELEMETRY OFFLINE after 3 failed fetches', async () => {
        mockFetch.mockRejectedValue(new Error('Connection refused'));

        render(<TelemetryRow />);

        // First call happens on mount
        await act(async () => {
            await vi.advanceTimersByTimeAsync(100);
        });

        // Advance through 2 more poll intervals (5s each)
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        await waitFor(() => {
            expect(screen.getByText(/TELEMETRY OFFLINE/)).toBeDefined();
        });
    });

    it('calls fetch endpoint with correct URL', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => TELEMETRY_RESPONSE,
        });

        render(<TelemetryRow />);

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/telemetry/compute');
        });
    });

    it('polls every 5 seconds', async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => TELEMETRY_RESPONSE,
        });

        render(<TelemetryRow />);

        // Initial call
        await act(async () => {
            await vi.advanceTimersByTimeAsync(100);
        });
        expect(mockFetch).toHaveBeenCalledTimes(1);

        // After 5s
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });
        expect(mockFetch).toHaveBeenCalledTimes(2);

        // After another 5s
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });
        expect(mockFetch).toHaveBeenCalledTimes(3);
    });

    it('recovers from degraded state when backend comes back', async () => {
        // Start with failures
        mockFetch.mockRejectedValue(new Error('down'));

        render(<TelemetryRow />);

        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        await act(async () => { await vi.advanceTimersByTimeAsync(5000); });
        await act(async () => { await vi.advanceTimersByTimeAsync(5000); });

        await waitFor(() => {
            expect(screen.getByText(/TELEMETRY OFFLINE/)).toBeDefined();
        });

        // Backend recovers
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => TELEMETRY_RESPONSE,
        });

        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        await waitFor(() => {
            expect(screen.getAllByTestId('telemetry-card')).toHaveLength(4);
        });
    });

    it('renders sparklines after multiple data points', async () => {
        let callCount = 0;
        mockFetch.mockImplementation(async () => ({
            ok: true,
            json: async () => ({
                ...TELEMETRY_RESPONSE,
                goroutines: 40 + callCount++,
            }),
        }));

        const { container } = render(<TelemetryRow />);

        // Need at least 2 data points for sparklines
        await act(async () => { await vi.advanceTimersByTimeAsync(100); });
        await act(async () => { await vi.advanceTimersByTimeAsync(5000); });

        await waitFor(() => {
            const svgs = container.querySelectorAll('svg');
            expect(svgs.length).toBeGreaterThanOrEqual(1);
        });
    });
});
