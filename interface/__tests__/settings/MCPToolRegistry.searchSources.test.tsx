import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { mockFetch } from '../setup';
import MCPToolRegistry from '@/components/settings/MCPToolRegistry';
import { useCortexStore } from '@/store/useCortexStore';

vi.mock('@/components/settings/MCPServerCard', () => ({
    __esModule: true,
    default: () => <div data-testid="server-card" />,
}));

vi.mock('@/components/settings/MCPLibraryBrowser', () => ({
    __esModule: true,
    default: () => <div data-testid="library-browser">Library Browser</div>,
    MCPLibraryBrowserBody: () => <div data-testid="library-browser">Library Browser</div>,
}));

describe('MCPToolRegistry search sources', () => {
    beforeEach(() => {
        useCortexStore.setState({
            mcpServers: [],
            isFetchingMCPServers: false,
            mcpServersError: null,
            mcpActivity: [],
            isFetchingMCPActivity: false,
            mcpToolSets: [],
            isFetchingMCPToolSets: false,
            mcpToolSetsError: null,
            fetchMCPServers: vi.fn(),
            fetchMCPActivity: vi.fn(),
            fetchMCPToolSets: vi.fn(),
            createMCPToolSet: vi.fn().mockResolvedValue(true),
            fetchSearchCapability: vi.fn(),
            fetchCapabilities: vi.fn(),
            deleteMCPServer: vi.fn(),
            streamLogs: [],
            isStreamConnected: false,
            initializeStream: vi.fn(),
            searchCapability: null,
            isFetchingSearchCapability: false,
            searchCapabilityError: null,
            capabilities: [],
            isFetchingCapabilities: false,
            capabilitiesError: null,
        });
    });

    it('shows the optional registry and adds safe source metadata', async () => {
        mockFetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ ok: true, data: { sources: [{
                    id: 'approved-docs',
                    name: 'Approved docs',
                    source_type: 'knowledge_collection',
                    scope_kind: 'all',
                    boundary: 'Approved company knowledge index',
                    auth_scheme: 'none',
                    mode: 'live',
                    sensitivity_class: 'governed',
                    trust_class: 'trusted_internal',
                    status: 'available',
                }] } }),
            })
            .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: { id: 'team-api' } }) })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ ok: true, data: { sources: [{
                    id: 'team-api',
                    name: 'Team research API',
                    source_type: 'local_api',
                    scope_kind: 'group',
                    scope_ref: 'research',
                    boundary: 'Approved research API',
                    auth_scheme: 'api_token',
                    secret_ref: 'SEARCH_API_KEY',
                    mode: 'live',
                    sensitivity_class: 'governed',
                    trust_class: 'bounded_internal',
                    status: 'available',
                }] } }),
            });

        render(<MCPToolRegistry />);

        await waitFor(() => expect(screen.getByText('Approved docs')).toBeDefined());
        expect(screen.getByText(/Knowledge collection/i)).toBeDefined();
        expect(screen.getByText(/Visible to everyone/i)).toBeDefined();
        expect(screen.getByText(/No secret needed/i)).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /Add search source/i }));
        fireEvent.change(screen.getByLabelText('Source name'), { target: { value: 'Team research API' } });
        fireEvent.change(screen.getByLabelText('Kind'), { target: { value: 'local_api' } });
        fireEvent.change(screen.getByLabelText('Endpoint for web/API'), { target: { value: 'https://search.example.test/api' } });
        fireEvent.change(screen.getByLabelText('Boundary'), { target: { value: 'Approved research API' } });
        fireEvent.change(screen.getByLabelText('Visible to'), { target: { value: 'group' } });
        fireEvent.change(screen.getByLabelText('Scope reference'), { target: { value: 'research' } });
        fireEvent.change(screen.getByLabelText('Authentication'), { target: { value: 'secret_ref' } });
        fireEvent.change(screen.getByLabelText(/Secret reference/i), { target: { value: 'SEARCH_API_KEY' } });
        fireEvent.click(screen.getAllByRole('button', { name: /^Add search source$/i }).at(-1)!);

        await waitFor(() => expect(screen.getByText(/Added Team research API/i)).toBeDefined());
        await waitFor(() => expect(screen.getByText('Team research API')).toBeDefined());

        const postCall = mockFetch.mock.calls.find(([url, init]) => (
            url === '/api/v1/search/sources' && (init as RequestInit | undefined)?.method === 'POST'
        ));
        expect(postCall).toBeDefined();
        const body = JSON.parse(((postCall?.[1] as RequestInit).body ?? '{}') as string);
        expect(body).toMatchObject({
            name: 'Team research API',
            source_type: 'local_api',
            endpoint: 'https://search.example.test/api',
            scope_kind: 'group',
            scope_ref: 'research',
            auth_scheme: 'api_token',
            secret_ref: 'SEARCH_API_KEY',
        });
        expect(JSON.stringify(body)).not.toContain('sk-');
        expect(JSON.stringify(body)).not.toContain('=secret');
    });

    it('updates and removes operator-managed search sources', async () => {
        mockFetch
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ ok: true, data: [{
                    id: 'team-api',
                    name: 'Team research API',
                    managed: true,
                    source_type: 'local_api',
                    endpoint: 'https://search.example.test/api',
                    scope_kind: 'group',
                    scope_ref: 'research',
                    boundary: 'Approved research API',
                    auth_scheme: 'none',
                    mode: 'live',
                    sensitivity_class: 'governed',
                    trust_class: 'bounded_internal',
                    status: 'available',
                }] }),
            })
            .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: { id: 'team-api' } }) })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ ok: true, data: [{
                    id: 'team-api',
                    name: 'Team research API v2',
                    managed: true,
                    source_type: 'local_api',
                    endpoint: 'https://search.example.test/v2',
                    scope_kind: 'group',
                    scope_ref: 'research',
                    boundary: 'Approved research API v2',
                    auth_scheme: 'none',
                    mode: 'live',
                    sensitivity_class: 'governed',
                    trust_class: 'bounded_internal',
                    status: 'available',
                }] }),
            })
            .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: { deleted: true } }) })
            .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: [] }) });

        render(<MCPToolRegistry />);

        await waitFor(() => expect(screen.getByText('Team research API')).toBeDefined());
        fireEvent.click(screen.getByRole('button', { name: /Edit/i }));
        fireEvent.change(screen.getByLabelText('Source name'), { target: { value: 'Team research API v2' } });
        fireEvent.change(screen.getByLabelText('Endpoint for web/API'), { target: { value: 'https://search.example.test/v2' } });
        fireEvent.change(screen.getByLabelText('Boundary'), { target: { value: 'Approved research API v2' } });
        fireEvent.click(screen.getByRole('button', { name: /^Update search source$/i }));

        await waitFor(() => expect(screen.getByText(/Updated Team research API v2/i)).toBeDefined());
        fireEvent.click(screen.getByRole('button', { name: /Remove/i }));

        await waitFor(() => expect(screen.getByText(/Removed Team research API v2/i)).toBeDefined());
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/search/sources/team-api', expect.objectContaining({ method: 'PATCH' }));
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/search/sources/team-api', expect.objectContaining({ method: 'DELETE' }));
    });
});
