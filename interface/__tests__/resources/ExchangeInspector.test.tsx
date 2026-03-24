import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import ExchangeInspector from '@/components/resources/ExchangeInspector';

describe('ExchangeInspector', () => {
    const fetchMock = vi.fn();

    beforeEach(() => {
        vi.stubGlobal('fetch', fetchMock);
    });

    afterEach(() => {
        vi.unstubAllGlobals();
        fetchMock.mockReset();
    });

    it('renders loaded channels, threads, and recent outputs', async () => {
        fetchMock
            .mockResolvedValueOnce({
                ok: true,
                json: async () => [{ id: 'c1', name: 'browser.research.results', type: 'output', schema_id: 'ToolResult', visibility: 'advanced', owner: 'mcp' }],
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => [{ id: 't1', title: 'Research pass', thread_type: 'review_thread', status: 'active', channel_name: 'browser.research.results', participants: ['soma', 'team_lead'] }],
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => [{ id: 'i1', summary: 'Fetch returned market notes.', schema_id: 'ToolResult', channel_name: 'browser.research.results', created_by: 'mcp:fetch', created_at: '2026-03-24T07:00:00Z' }],
            });

        render(<ExchangeInspector />);

        await waitFor(() => {
            expect(screen.getByText('browser.research.results')).toBeDefined();
            expect(screen.getByText('Research pass')).toBeDefined();
            expect(screen.getByText('Fetch returned market notes.')).toBeDefined();
        });
    });

    it('shows an error state when exchange loading fails', async () => {
        fetchMock.mockResolvedValue({ ok: false, json: async () => ({}) });

        render(<ExchangeInspector />);

        await waitFor(() => {
            expect(screen.getByText(/Managed exchange unavailable/i)).toBeDefined();
        });
    });
});
