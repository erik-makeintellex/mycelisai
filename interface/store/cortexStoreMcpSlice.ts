import { extractApiData } from '@/lib/apiContracts';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { MCPServerWithTools, MCPLibraryCategory, MCPTool } from '@/store/cortexStoreTypes';

export function createCortexMcpSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<
    | 'fetchMCPServers'
    | 'deleteMCPServer'
    | 'fetchMCPTools'
    | 'fetchMCPLibrary'
    | 'installFromLibrary'
> {
    return {
        fetchMCPServers: async () => {
            set({ isFetchingMCPServers: true });
            try {
                const res = await fetch('/api/v1/mcp/servers');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<MCPServerWithTools[] | unknown>(payload);
                    set({ mcpServers: Array.isArray(data) ? data : [], isFetchingMCPServers: false });
                } else {
                    set({ mcpServers: [], isFetchingMCPServers: false });
                }
            } catch {
                set({ mcpServers: [], isFetchingMCPServers: false });
            }
        },

        deleteMCPServer: async (id: string) => {
            try {
                const res = await fetch(`/api/v1/mcp/servers/${id}`, { method: 'DELETE' });
                if (res.ok) {
                    set((s) => ({
                        mcpServers: s.mcpServers.filter((server) => server.id !== id),
                    }));
                }
            } catch (err) {
                console.error('[MCP] Delete failed:', err);
            }
        },

        fetchMCPTools: async () => {
            try {
                const res = await fetch('/api/v1/mcp/tools');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<MCPTool[] | unknown>(payload);
                    set({ mcpTools: Array.isArray(data) ? data : [] });
                }
            } catch {
                set({ mcpTools: [] });
            }
        },

        fetchMCPLibrary: async () => {
            set({ isFetchingMCPLibrary: true });
            try {
                const res = await fetch('/api/v1/mcp/library');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<MCPLibraryCategory[] | unknown>(payload);
                    set({ mcpLibrary: Array.isArray(data) ? data : [], isFetchingMCPLibrary: false });
                } else {
                    set({ mcpLibrary: [], isFetchingMCPLibrary: false });
                }
            } catch {
                set({ mcpLibrary: [], isFetchingMCPLibrary: false });
            }
        },

        installFromLibrary: async (name: string, env?: Record<string, string>) => {
            try {
                const res = await fetch('/api/v1/mcp/library/install', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, env }),
                });
                if (res.ok) {
                    get().fetchMCPServers();
                } else {
                    console.error('[MCP Library] Install failed:', await res.text());
                }
            } catch (err) {
                console.error('[MCP Library] Install failed:', err);
            }
        },
    };
}
