import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import type { MCPServerWithTools } from '@/store/useCortexStore';
import { mockServers, readySearchCapability, webResearchCapability } from './MCPToolRegistry.testData';

type MockMCPServerCardProps = {
    server: MCPServerWithTools;
    onDelete: (id: string) => void;
};

type MockMCPLibraryBrowserProps = {
    onInstalled?: (name: string) => void;
};

// Mock child components to isolate MCPToolRegistry
vi.mock('@/components/settings/MCPServerCard', () => ({
    __esModule: true,
    default: ({ server, onDelete }: MockMCPServerCardProps) => (
        <div data-testid={`server-card-${server.id}`}>
            <span>{server.name}</span>
            <button data-testid={`delete-${server.id}`} onClick={() => onDelete(server.id)}>
                Delete
            </button>
        </div>
    ),
}));

vi.mock('@/components/settings/MCPLibraryBrowser', () => ({
    __esModule: true,
    default: () => <div data-testid="library-browser">Library Browser</div>,
    MCPLibraryBrowserBody: ({ onInstalled }: MockMCPLibraryBrowserProps) => (
        <div data-testid="library-browser">
            <button onClick={() => onInstalled?.('filesystem')}>Mock Install</button>
            Library Browser
        </div>
    ),
}));

import MCPToolRegistry from '@/components/settings/MCPToolRegistry';
import { useCortexStore } from '@/store/useCortexStore';

describe('MCPToolRegistry', () => {
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

    it('renders the installed server list', () => {
        const initializeStream = vi.fn();
        const fetchMCPActivity = vi.fn();
        const fetchMCPToolSets = vi.fn();
        const fetchSearchCapability = vi.fn();
        const fetchCapabilities = vi.fn();
        useCortexStore.setState({
            mcpServers: mockServers,
            initializeStream,
            fetchMCPActivity,
            fetchMCPToolSets,
            fetchSearchCapability,
            fetchCapabilities,
            capabilities: [webResearchCapability],
            searchCapability: readySearchCapability,
        });

        render(<MCPToolRegistry />);

        expect(screen.getByText('Capabilities')).toBeDefined();

        expect(screen.queryByTestId('server-card-srv-001')).toBeNull();
        expect(screen.queryByTestId('server-card-srv-002')).toBeNull();

        expect(screen.getByText(/Capability Workflow/i)).toBeDefined();
        expect(screen.getByText(/Concrete Soma Commands/i)).toBeDefined();
        expect(screen.getByText(/Search web/i)).toBeDefined();
        expect(screen.getAllByText(/Create team/i).length).toBeGreaterThan(0);
        expect(screen.getAllByText(/Read host data/i).length).toBeGreaterThan(0);
        expect(initializeStream).toHaveBeenCalledTimes(1);
        expect(fetchMCPActivity).toHaveBeenCalledTimes(1);
        expect(fetchMCPToolSets).toHaveBeenCalledTimes(1);
        expect(fetchSearchCapability).toHaveBeenCalledTimes(1);
        expect(fetchCapabilities).toHaveBeenCalledTimes(1);
        expect(screen.getByText('What Soma Can Use')).toBeDefined();
        expect(screen.getByText('MCP access layers')).toBeDefined();
        expect(screen.getByText('Can use now')).toBeDefined();
        expect(screen.getByText('Needs repair')).toBeDefined();
        expect(screen.getByText('Can request/add')).toBeDefined();
        expect(screen.getByText('Web Research')).toBeDefined();
        expect(screen.getByText(/risk medium/i)).toBeDefined();
        expect(screen.getByText(/approval optional/i)).toBeDefined();
        expect(screen.getByText(/exchange.browser.research.results/i)).toBeDefined();
        expect(screen.getAllByText(/Inspect capability binding/i).length).toBeGreaterThan(0);
        expect(screen.getByText('Mycelis Search Capability')).toBeDefined();
        expect(screen.getByText('Soma search is ready')).toBeDefined();
        expect(screen.getByText(/No hosted Brave token required/i)).toBeDefined();
        expect(screen.getByText(/local_api/i)).toBeDefined();

        // "Installed" tab should be active by default, showing server count badge
        expect(screen.getByText('2')).toBeDefined();
        fireEvent.click(screen.getByText('Inspect MCP topology'));
        expect(screen.getByTestId('server-card-srv-001')).toBeDefined();
        expect(screen.getByTestId('server-card-srv-002')).toBeDefined();
        expect(screen.getByText('filesystem-server')).toBeDefined();
        expect(screen.getByText('web-scraper')).toBeDefined();
    });

    it('browse library button switches to the library tab', () => {
        useCortexStore.setState({
            mcpServers: [],
        });

        render(<MCPToolRegistry />);

        fireEvent.click(screen.getByText('Request capability'));

        expect(screen.getByTestId('library-browser')).toBeDefined();
    });

    it('separates registry failures from a true empty installed state', () => {
        useCortexStore.setState({
            mcpServers: [],
            mcpServersError: 'MCP registry unreachable (HTTP 500)',
        });

        render(<MCPToolRegistry />);

        expect(screen.getByText('MCP registry unreachable')).toBeDefined();
        expect(screen.queryByText('No MCP servers installed.')).toBeNull();
    });

    it('surfaces recent MCP activity from persisted history', () => {
        useCortexStore.setState({
            mcpServers: mockServers,
            mcpActivity: [
                {
                    id: 'activity-1',
                    server_id: 'srv-001',
                    server_name: 'filesystem-server',
                    tool_name: 'read_file',
                    state: 'completed',
                    summary: 'Read workspace brief successfully.',
                    message: 'Read workspace brief successfully.',
                    channel_name: 'browser.research.results',
                    run_id: 'run-1',
                    team_id: 'alpha',
                    agent_id: 'soma-admin',
                    timestamp: '2026-04-06T12:00:00Z',
                },
            ],
        });

        render(<MCPToolRegistry />);

        expect(screen.getByText(/Recent MCP Activity/i)).toBeDefined();
        expect(screen.getByText(/filesystem-server · read_file/i)).toBeDefined();
        expect(screen.getByText(/Read workspace brief successfully/i)).toBeDefined();
        expect(screen.getByText(/Team alpha · Agent soma-admin · Run run-1/i)).toBeDefined();
    });

    it('falls back to MCP and search data when capability manifests are unavailable', () => {
        useCortexStore.setState({
            mcpServers: mockServers,
            capabilities: [],
            capabilitiesError: 'Capability registry unreachable (HTTP 503)',
            searchCapability: {
                provider: 'local_sources',
                enabled: true,
                configured: true,
                supports_local_sources: true,
                supports_public_web: false,
                soma_tool_name: 'web_search',
                direct_soma_interaction: true,
                requires_hosted_api_token: false,
                max_results: 8,
            },
        });

        render(<MCPToolRegistry />);

        expect(screen.getByText(/Capability registry unreachable/i)).toBeDefined();
        expect(screen.getByText(/derived from MCP servers and search status/i)).toBeDefined();
        expect(screen.getByText('Mycelis Search')).toBeDefined();
        expect(screen.getByText('filesystem-server: read_file')).toBeDefined();
        expect(screen.getByText('filesystem-server: write_file')).toBeDefined();
        expect(screen.getByText(/approval required/i)).toBeDefined();
        expect(screen.getByText(/workspace files/i)).toBeDefined();
        expect(screen.getAllByText(/Inspect capability binding/i).length).toBeGreaterThan(0);
        expect(screen.queryByText('MCP Server Drill-Down')).toBeNull();
        fireEvent.click(screen.getByText('Inspect MCP topology'));
        expect(screen.getByText('MCP Server Drill-Down')).toBeDefined();
    });

    it('keeps degraded capabilities in the repair queue with recovery guidance', () => {
        useCortexStore.setState({
            capabilities: [{
                ...webResearchCapability,
                id: 'media.generate',
                name: 'Local Media Generation',
                category: 'media',
                availability_status: 'degraded',
                fallback_behavior: 'Keep the failed run recoverable, preserve request proof, and retry after the local media provider is reconnected.',
                outputs: ['Image output'],
                writes: ['workspace.saved_media'],
            }],
        });

        render(<MCPToolRegistry />);

        expect(screen.getByText('Needs repair')).toBeDefined();
        expect(screen.getByText('Local Media Generation')).toBeDefined();
        expect(screen.getByText('degraded')).toBeDefined();
        expect(screen.getByText(/Keep the failed run recoverable/i)).toBeDefined();
        expect(screen.getByText(/retry after the local media provider is reconnected/i)).toBeDefined();
        expect(screen.getByText('workspace.saved_media')).toBeDefined();
        expect(screen.queryByText(/derived from MCP servers and search status/i)).toBeNull();
    });

    it('shows search capability blockers as operator guidance', () => {
        useCortexStore.setState({
            searchCapability: {
                provider: 'disabled',
                enabled: false,
                configured: false,
                supports_local_sources: false,
                supports_public_web: false,
                soma_tool_name: 'web_search',
                direct_soma_interaction: true,
                requires_hosted_api_token: false,
                max_results: 8,
                blocker: {
                    code: 'search_provider_disabled',
                    message: 'Mycelis Search is disabled.',
                    next_action: 'Set MYCELIS_SEARCH_PROVIDER=local_sources for governed local-source search or searxng for self-hosted web search.',
                },
            },
        });

        render(<MCPToolRegistry />);

        expect(screen.getByText('Soma search needs configuration')).toBeDefined();
        expect(screen.getAllByText('Mycelis Search is disabled.').length).toBeGreaterThan(0);
        expect(screen.getByText(/Soma direct: web_search/i)).toBeDefined();
    });

    it('returns to installed view with guidance after library install', () => {
        render(<MCPToolRegistry />);

        fireEvent.click(screen.getByText('Request capability'));
        fireEvent.click(screen.getByText('Mock Install'));

        expect(screen.getByText(/Installed filesystem/i)).toBeDefined();
    });

    it('delete action on a server card calls the store delete action', () => {
        const deleteFn = vi.fn();
        useCortexStore.setState({
            mcpServers: mockServers,
            deleteMCPServer: deleteFn,
        });

        render(<MCPToolRegistry />);

        fireEvent.click(screen.getByText('Inspect MCP topology'));
        fireEvent.click(screen.getByTestId('delete-srv-001'));

        expect(deleteFn).toHaveBeenCalledTimes(1);
        expect(deleteFn).toHaveBeenCalledWith('srv-001');
    });
});
