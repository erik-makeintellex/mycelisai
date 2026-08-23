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
        mockSearchSourceRegistry([{
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
        }, {
            id: 'workspace-code',
            name: 'Workspace code map',
            provider: 'code_context',
            source_type: 'code_context',
            scope_kind: 'host',
            scope_ref: 'dev-workstation',
            boundary: 'Approved workspace repository source',
            auth_scheme: 'none',
            mode: 'snapshot',
            sensitivity_class: 'governed',
            trust_class: 'trusted_internal',
            status: 'available',
            code_context: {
                scope: 'workspace repository',
                snapshot_status: 'ready',
                last_snapshot_at: '2026-08-22T10:00:00Z',
                snapshot_ref: 'snapshot:workspace-code:abc123',
                snapshot_digest: 'sha256:abc123',
                index_status: 'stale',
                last_indexed_at: '2026-08-21T09:30:00Z',
                index_ref: 'index:workspace-code:def456',
                index_digest: 'sha256:def456',
                refresh_action: 'Refresh the repository map after repository changes.',
            },
        }]);

        render(<MCPToolRegistry />);
        openAccessFocus();

        await waitFor(() => expect(screen.getByText('Approved docs')).toBeDefined());
        expect(screen.getByText(/Approved places Soma may search: public web, approved local or mounted data, repository\/code folder context, and private APIs/i)).toBeDefined();
        expect(screen.getByText(/Approved knowledge collection/i)).toBeDefined();
        expect(screen.getByText('Public web research')).toBeDefined();
        expect(screen.getAllByText(/Public web/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Visible to everyone/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/No secret needed/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Ready for Soma to use when this scope is allowed/i).length).toBeGreaterThan(0);
        expect(screen.getByText('Workspace code map')).toBeDefined();
        expect(screen.getAllByText(/Repository or code folder/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/Repository map for Soma/i)).toBeDefined();
        expect(screen.getByText(/Snapshot ready/i)).toBeDefined();
        expect(screen.getByText(/Index stale/i)).toBeDefined();
        expect(screen.getAllByText('workspace repository').length).toBeGreaterThan(0);
        expect(screen.getAllByText('2026-08-22T10:00:00Z').length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Refresh the repository map after repository changes/i).length).toBeGreaterThan(0);

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
        mockSearchSourceRegistry([{
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
        }]);

        render(<MCPToolRegistry />);
        openAccessFocus();

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
        mockSearchSourceRegistry([]);

        render(<MCPToolRegistry />);
        openAccessFocus();

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

    it('adds repository code-folder sources as Soma-searchable context', async () => {
        mockSearchSourceRegistry([]);

        render(<MCPToolRegistry />);
        openAccessFocus();

        await waitFor(() => expect(screen.getByText(/No configured sources reported/i)).toBeDefined());
        fireEvent.click(screen.getByRole('button', { name: /Add search source/i }));
        fireEvent.change(screen.getByLabelText('Source name'), { target: { value: 'Workspace repository map' } });
        fireEvent.change(screen.getByLabelText('Source kind'), { target: { value: 'code_context' } });
        fireEvent.change(screen.getByLabelText('Approved repository or code folder path'), { target: { value: 'core' } });
        fireEvent.change(screen.getByLabelText('Search boundary'), { target: { value: 'Approved Mycelis runtime source tree' } });
        fireEvent.change(screen.getByLabelText('Visible to'), { target: { value: 'group' } });
        fireEvent.change(screen.getByLabelText('Group name'), { target: { value: 'runtime-review' } });
        fireEvent.click(screen.getAllByRole('button', { name: /^Add search source$/i }).at(-1)!);

        await waitFor(() => expect(screen.getByText(/Added Workspace repository map/i)).toBeDefined());
        await waitFor(() => expect(screen.getByText('Workspace repository map')).toBeDefined());
        expect(screen.getAllByText(/Repository or code folder/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Visible to one group/i).length).toBeGreaterThan(0);
        const postCall = mockFetch.mock.calls.find(([url, init]) => (
            url === '/api/v1/search/sources' && (init as RequestInit | undefined)?.method === 'POST'
        ));
        const body = JSON.parse(((postCall?.[1] as RequestInit).body ?? '{}') as string);
        expect(body).toMatchObject({
            name: 'Workspace repository map',
            source_type: 'code_context',
            endpoint: 'core',
            boundary: 'Approved Mycelis runtime source tree',
            scope_kind: 'group',
            scope_ref: 'runtime-review',
            auth_scheme: 'none',
        });
    });
});

function openAccessFocus() {
    fireEvent.click(screen.getAllByRole('button', { name: /Access/i })[0]);
}

function mockSearchSourceRegistry(initialSources: Array<Record<string, unknown>>) {
    let sources = [...initialSources];
    mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = init?.method ?? 'GET';
        if (url === '/api/v1/input-sources') return jsonResponse({ ok: true, data: [] });
        if (!url.startsWith('/api/v1/search/sources')) return jsonResponse({ ok: true, data: [] });

        if (method === 'POST') {
            const draft = JSON.parse(String(init?.body ?? '{}')) as Record<string, unknown>;
            const id = draft.name === 'Client docs mount' ? 'client-docs' : 'team-api';
            const created = { ...draft, id, managed: true, status: 'available' };
            sources = [...sources, created];
            return jsonResponse({ ok: true, data: created });
        }
        if (method === 'PATCH') {
            const id = decodeURIComponent(url.split('/').at(-1) ?? '');
            const draft = JSON.parse(String(init?.body ?? '{}')) as Record<string, unknown>;
            sources = sources.map((source) => source.id === id ? { ...source, ...draft } : source);
            return jsonResponse({ ok: true, data: sources.find((source) => source.id === id) });
        }
        if (method === 'DELETE') {
            const id = decodeURIComponent(url.split('/').at(-1) ?? '');
            sources = sources.filter((source) => source.id !== id);
            return jsonResponse({ ok: true, data: { deleted: true } });
        }
        return jsonResponse({ ok: true, data: { sources } });
    });
}

function jsonResponse(payload: unknown): Response {
    return {
        ok: true,
        status: 200,
        json: async () => payload,
        text: async () => JSON.stringify(payload),
    } as Response;
}
