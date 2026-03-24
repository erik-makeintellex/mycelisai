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
                json: async () => [{ id: 'c1', name: 'browser.research.results', type: 'output', schema_id: 'ToolResult', visibility: 'advanced', sensitivity_class: 'team_scoped', owner: 'mcp' }],
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => [{ id: 't1', title: 'Research pass', thread_type: 'review_thread', status: 'active', channel_name: 'browser.research.results', participants: ['soma', 'team_lead'], allowed_reviewers: ['review'] }],
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => [{ id: 'i1', summary: 'Fetch returned market notes.', schema_id: 'ToolResult', channel_name: 'browser.research.results', created_by: 'mcp:fetch', created_at: '2026-03-24T07:00:00Z', sensitivity_class: 'team_scoped', trust_class: 'bounded_external', capability_id: 'browser_research', review_required: true }],
            });

        render(<ExchangeInspector />);

        await waitFor(() => {
            expect(screen.getAllByText('browser.research.results').length).toBeGreaterThan(0);
            expect(screen.getByText('Research pass')).toBeDefined();
            expect(screen.getByText('Fetch returned market notes.')).toBeDefined();
            expect(screen.getAllByText(/team_scoped/i).length).toBeGreaterThan(0);
            expect(screen.getByText(/bounded_external/i)).toBeDefined();
            expect(screen.getByText(/review required/i)).toBeDefined();
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
