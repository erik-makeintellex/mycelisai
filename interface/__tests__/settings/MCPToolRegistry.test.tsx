import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { mockFetch } from '../setup';

// Mock child components to isolate MCPToolRegistry
vi.mock('@/components/settings/MCPServerCard', () => ({
    __esModule: true,
    default: ({ server, onDelete }: any) => (
        <div data-testid={`server-card-${server.id}`}>
            <span>{server.name}</span>
            <button data-testid={`delete-${server.id}`} onClick={() => onDelete(server.id)}>
                Delete
            </button>
        </div>
    ),
}));

vi.mock('@/components/settings/MCPInstallModal', () => ({
    __esModule: true,
    default: ({ isOpen, onClose, onInstall }: any) =>
        isOpen ? (
            <div data-testid="install-modal">
                <span>Install MCP Server</span>
                <button onClick={onClose}>Close Modal</button>
                <button onClick={() => onInstall({ name: 'test-server', transport: 'stdio' })}>
                    Confirm Install
                </button>
            </div>
        ) : null,
}));

vi.mock('@/components/settings/MCPLibraryBrowser', () => ({
    __esModule: true,
    default: () => <div data-testid="library-browser">Library Browser</div>,
}));

import MCPToolRegistry from '@/components/settings/MCPToolRegistry';
import { useCortexStore } from '@/store/useCortexStore';
import type { MCPServerWithTools } from '@/store/useCortexStore';

const mockServers: MCPServerWithTools[] = [
    {
        id: 'srv-001',
        name: 'filesystem-server',
        transport: 'stdio',
        command: 'npx -y @modelcontextprotocol/server-fs',
        args: ['/home/user/docs'],
        status: 'connected',
        created_at: '2025-01-01T00:00:00Z',
        tools: [
            { id: 'tool-1', server_id: 'srv-001', name: 'read_file', description: 'Read a file', input_schema: {} },
            { id: 'tool-2', server_id: 'srv-001', name: 'write_file', description: 'Write a file', input_schema: {} },
        ],
    },
    {
        id: 'srv-002',
        name: 'web-scraper',
        transport: 'sse',
        url: 'http://localhost:3001/sse',
        status: 'connected',
        created_at: '2025-01-02T00:00:00Z',
        tools: [],
    },
];

describe('MCPToolRegistry', () => {
    beforeEach(() => {
        useCortexStore.setState({
            mcpServers: [],
            isFetchingMCPServers: false,
            fetchMCPServers: vi.fn(),
            installMCPServer: vi.fn(),
            deleteMCPServer: vi.fn(),
        });
    });

    it('renders the installed server list', () => {
        useCortexStore.setState({
            mcpServers: mockServers,
        });

        render(<MCPToolRegistry />);

        // Header should show registry title
        expect(screen.getByText('MCP Tool Registry')).toBeDefined();

        // Both server cards should be rendered
        expect(screen.getByTestId('server-card-srv-001')).toBeDefined();
        expect(screen.getByTestId('server-card-srv-002')).toBeDefined();

        // Server names should be visible
        expect(screen.getByText('filesystem-server')).toBeDefined();
        expect(screen.getByText('web-scraper')).toBeDefined();

        // "Installed" tab should be active by default, showing server count badge
        expect(screen.getByText('2')).toBeDefined();
    });

    it('install button opens the install modal', () => {
        useCortexStore.setState({
            mcpServers: [],
        });

        render(<MCPToolRegistry />);

        // No modal initially
        expect(screen.queryByTestId('install-modal')).toBeNull();

        // Click the INSTALL button
        fireEvent.click(screen.getByText('INSTALL'));

        // Modal should now be visible
        expect(screen.getByTestId('install-modal')).toBeDefined();
        expect(screen.getByText('Install MCP Server')).toBeDefined();
    });

    it('delete action on a server card calls the store delete action', () => {
        const deleteFn = vi.fn();
        useCortexStore.setState({
            mcpServers: mockServers,
            deleteMCPServer: deleteFn,
        });

        render(<MCPToolRegistry />);

        // Click delete on srv-001
        fireEvent.click(screen.getByTestId('delete-srv-001'));

        // Store's deleteMCPServer should have been called
        expect(deleteFn).toHaveBeenCalledTimes(1);
        expect(deleteFn).toHaveBeenCalledWith('srv-001');
    });
});
