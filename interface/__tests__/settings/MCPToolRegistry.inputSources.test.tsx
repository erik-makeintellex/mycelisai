import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import type { MCPServerWithTools } from '@/store/useCortexStore';
import { readySearchCapability, webResearchCapability } from './MCPToolRegistry.testData';

type MockMCPServerCardProps = {
    server: MCPServerWithTools;
    onDelete: (id: string) => void;
};

vi.mock('@/components/settings/MCPServerCard', () => ({
    __esModule: true,
    default: ({ server, onDelete }: MockMCPServerCardProps) => (
        <div data-testid={`server-card-${server.id}`}>
            <span>{server.name}</span>
            <button onClick={() => onDelete(server.id)}>Delete</button>
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

describe('MCPToolRegistry live inputs', () => {
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
            capabilities: [webResearchCapability],
            isFetchingCapabilities: false,
            capabilitiesError: null,
            searchCapability: readySearchCapability,
            isFetchingSearchCapability: false,
            searchCapabilityError: null,
        });
    });

    afterEach(() => {
        vi.unstubAllGlobals();
    });

    it('shows registered live inputs as buffered context feeds in Access', async () => {
        const fetchMock = vi.fn(async (url: RequestInfo | URL) => {
            const rawUrl = String(url);
            if (rawUrl === '/api/v1/input-sources') {
                return mockResponse(200, [{
                    id: 'warehouse-sensor',
                    name: 'Warehouse sensor feed',
                    source_type: 'sensor',
                    adapter_kind: 'sensor',
                    scope_kind: 'group',
                    scope_ref: 'ops',
                    auth_scheme: 'secret_ref',
                    secret_ref: 'secret://warehouse',
                    allowed_ingress_subject: 'swarm.global.input.warehouse-sensor',
                    buffer_mode: 'latest_state',
                    sensitivity_class: 'internal',
                    trust_class: 'bounded_input',
                    status: 'available',
                }]);
            }
            if (rawUrl.startsWith('/api/v1/input-sources/warehouse-sensor/buffer')) {
                return mockResponse(200, {
                    mode: 'latest_state',
                    latest: [{
                        event_id: 'evt-1',
                        channel_key: 'temperature',
                        payload: { temperature: 72, unit: 'f' },
                    }],
                });
            }
            return mockResponse(404, {});
        });
        vi.stubGlobal('fetch', fetchMock);

        render(<MCPToolRegistry />);

        fireEvent.click(screen.getAllByRole('button', { name: /Access/i })[0]);
        fireEvent.click(screen.getByRole('button', { name: /Live inputs/i }));

        await waitFor(() => expect(screen.getByText('Warehouse sensor feed')).toBeDefined());
        expect(screen.getByText(/event feed Soma can reference/i)).toBeDefined();
        expect(screen.getByText(/Sensor · Latest state · Group scoped/i)).toBeDefined();
        fireEvent.click(screen.getByRole('button', { name: /Warehouse sensor feed/i }));
        await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/api/v1/input-sources/warehouse-sensor/buffer')));
        expect(screen.getByText('temperature')).toBeDefined();
        expect(screen.getByText(/"temperature":72/i)).toBeDefined();
        fireEvent.click(screen.getByText('Inspect source'));
        expect(screen.getByText('swarm.global.input.warehouse-sensor')).toBeDefined();
        expect(screen.getByText(/secret_ref/i)).toBeDefined();
    });
});

function mockResponse(status: number, body: unknown) {
    return {
        ok: status >= 200 && status < 300,
        status,
        json: async () => ({ data: body }),
        text: async () => JSON.stringify(body),
    } as Response;
}
