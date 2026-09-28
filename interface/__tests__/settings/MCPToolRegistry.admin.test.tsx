import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import type { MCPServerWithTools } from '@/store/useCortexStore';
import { mockServers } from './MCPToolRegistry.testData';

// MCPA D10: MCPToolRegistry reads useIsAdmin() (a UI hint; Core still
// enforces mcp_config:write) and forwards it to MCPServerCard, which renders
// the delete control only for an admin viewer. Kept in its own file so the
// U1 lane's MCPToolRegistry.test.tsx stays untouched.
let mockIsAdmin = true;
vi.mock('@/lib/useIsAdmin', () => ({
    useIsAdmin: () => ({ isAdmin: mockIsAdmin, checked: true }),
}));

vi.mock('@/components/settings/MCPServerCard', () => ({
    __esModule: true,
    default: ({ server, onDelete, isAdmin }: { server: MCPServerWithTools; onDelete: (id: string) => void; isAdmin?: boolean }) => (
        <div data-testid={`server-card-${server.id}`}>
            <span>{server.name}</span>
            {isAdmin && (
                <button data-testid={`delete-${server.id}`} onClick={() => onDelete(server.id)}>
                    Delete
                </button>
            )}
        </div>
    ),
}));

vi.mock('@/components/settings/MCPLibraryBrowser', () => ({
    __esModule: true,
    default: () => <div data-testid="library-browser">Library Browser</div>,
    MCPLibraryBrowserBody: () => <div data-testid="library-browser">Library Browser</div>,
}));

import MCPToolRegistry from '@/components/settings/MCPToolRegistry';
import { useCortexStore } from '@/store/useCortexStore';

describe('MCPToolRegistry admin gating (MCPA D10)', () => {
    beforeEach(() => {
        mockIsAdmin = true;
        // Same store stubs as MCPToolRegistry.test.tsx: mount-time fetches are
        // no-ops so the preloaded servers are what renders.
        useCortexStore.setState({
            mcpServers: mockServers,
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
            createMCPToolSet: vi.fn().mockResolvedValue({ ok: true }),
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

    it('shows the delete control on server cards for an admin viewer', () => {
        render(<MCPToolRegistry />);
        fireEvent.click(screen.getByRole('button', { name: /Servers\s*2/i }));
        expect(screen.getByTestId('delete-srv-001')).toBeTruthy();
    });

    it('hides the delete control on server cards for a non-admin viewer', () => {
        mockIsAdmin = false;
        render(<MCPToolRegistry />);
        fireEvent.click(screen.getByRole('button', { name: /Servers\s*2/i }));
        expect(screen.queryByTestId('delete-srv-001')).toBeNull();
    });
});
