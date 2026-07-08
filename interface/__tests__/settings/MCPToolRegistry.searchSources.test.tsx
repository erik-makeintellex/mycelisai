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
                }, {
                    id: 'public-web',
                    name: 'Public web research',
                    source_type: 'public_web',
                    endpoint: 'https://web-search.example.test',
                    scope_kind: 'all',
                    boundary: 'Approved public web search',
                    auth_scheme: 'none',
                    mode: 'live',
                    sensitivity_class: 'public',
                    trust_class: 'bounded_external',
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
        expect(screen.getByText(/Approved places Soma may search: public web, approved local or mounted data, and private APIs/i)).toBeDefined();
        expect(screen.getByText(/Approved knowledge collection/i)).toBeDefined();
        expect(screen.getByText('Public web research')).toBeDefined();
        expect(screen.getAllByText(/Public web/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Visible to everyone/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/No secret needed/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Ready for Soma to use when this scope is allowed/i).length).toBeGreaterThan(0);

        fireEvent.click(screen.getByRole('button', { name: /Add search source/i }));
        fireEvent.change(screen.getByLabelText('Source name'), { target: { value: 'Team research API' } });
        expect(screen.getByText(/Add a place Soma may search after the configured scope allows it/i)).toBeDefined();
        fireEvent.change(screen.getByLabelText('Source kind'), { target: { value: 'local_api' } });
        fireEvent.change(screen.getByLabelText('Private API address'), { target: { value: 'https://search.example.test/api' } });
        fireEvent.change(screen.getByLabelText('Search boundary'), { target: { value: 'Approved research API' } });
        fireEvent.change(screen.getByLabelText('Visible to'), { target: { value: 'group' } });
        fireEvent.change(screen.getByLabelText('Group name'), { target: { value: 'research' } });
        fireEvent.change(screen.getByLabelText('Authentication'), { target: { value: 'secret_ref' } });
        expect(screen.getByPlaceholderText('SOMA_SEARCH_SECRET')).toBeDefined();
        fireEvent.change(screen.getByLabelText(/Secret reference/i), { target: { value: 'SEARCH_API_KEY' } });
        fireEvent.click(screen.getAllByRole('button', { name: /^Add search source$/i }).at(-1)!);

        await waitFor(() => expect(screen.getByText(/Added Team research API/i)).toBeDefined());
        await waitFor(() => expect(screen.getByText('Team research API')).toBeDefined());
        expect(screen.getAllByText(/Private API/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/Visible to one group/i)).toBeDefined();
        expect(screen.getByText(/Uses a saved secret/i)).toBeDefined();
        expect(screen.getByText(/Ready once saved access is available/i)).toBeDefined();

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
        fireEvent.change(screen.getByLabelText('Private API address'), { target: { value: 'https://search.example.test/v2' } });
        fireEvent.change(screen.getByLabelText('Search boundary'), { target: { value: 'Approved research API v2' } });
        fireEvent.click(screen.getByRole('button', { name: /^Update search source$/i }));

        await waitFor(() => expect(screen.getByText(/Updated Team research API v2/i)).toBeDefined());
        fireEvent.click(screen.getByRole('button', { name: /Remove/i }));

        await waitFor(() => expect(screen.getByText(/Removed Team research API v2/i)).toBeDefined());
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/search/sources/team-api', expect.objectContaining({ method: 'PATCH' }));
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/search/sources/team-api', expect.objectContaining({ method: 'DELETE' }));
    });

    it('adds mounted folder sources with local paths instead of HTTP endpoints', async () => {
        mockFetch
            .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: [] }) })
            .mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true, data: { id: 'client-docs' } }) })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ ok: true, data: [{
                    id: 'client-docs',
                    name: 'Client docs mount',
                    managed: true,
                    source_type: 'mounted_folder',
                    endpoint: 'workspace/client-docs',
                    scope_kind: 'host',
                    scope_ref: 'workstation-1',
                    boundary: 'Operator-approved client docs folder',
                    auth_scheme: 'none',
                    mode: 'live',
                    sensitivity_class: 'restricted',
                    trust_class: 'trusted_internal',
                    status: 'available',
                }] }),
            });

        render(<MCPToolRegistry />);

        await waitFor(() => expect(screen.getByText(/No configured sources reported/i)).toBeDefined());
        fireEvent.click(screen.getByRole('button', { name: /Add search source/i }));
        fireEvent.change(screen.getByLabelText('Source name'), { target: { value: 'Client docs mount' } });
        fireEvent.change(screen.getByLabelText('Source kind'), { target: { value: 'mounted_folder' } });
        fireEvent.change(screen.getByLabelText('Approved folder path'), { target: { value: 'workspace/client-docs' } });
        fireEvent.change(screen.getByLabelText('Search boundary'), { target: { value: 'Operator-approved client docs folder' } });
        fireEvent.change(screen.getByLabelText('Visible to'), { target: { value: 'host' } });
        fireEvent.change(screen.getByLabelText('Host name'), { target: { value: 'workstation-1' } });
        fireEvent.click(screen.getAllByRole('button', { name: /^Add search source$/i }).at(-1)!);

        await waitFor(() => expect(screen.getByText(/Added Client docs mount/i)).toBeDefined());
        await waitFor(() => expect(screen.getByText('Client docs mount')).toBeDefined());
        expect(screen.getAllByText(/Approved local or mounted data/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/Visible to one host/i)).toBeDefined();
        const postCall = mockFetch.mock.calls.find(([url, init]) => (
            url === '/api/v1/search/sources' && (init as RequestInit | undefined)?.method === 'POST'
        ));
        const body = JSON.parse(((postCall?.[1] as RequestInit).body ?? '{}') as string);
        expect(body).toMatchObject({
            name: 'Client docs mount',
            source_type: 'mounted_folder',
            endpoint: 'workspace/client-docs',
            scope_kind: 'host',
            scope_ref: 'workstation-1',
            auth_scheme: 'none',
        });
    });
});
