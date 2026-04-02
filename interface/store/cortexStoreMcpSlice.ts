import { extractApiData } from '@/lib/apiContracts';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type { MCPGovernanceDecision, MCPInstallResult, MCPServerWithTools, MCPLibraryCategory, MCPTool } from '@/store/cortexStoreTypes';

interface MCPLibraryInspectionResponse {
    decision?: string;
    reasons?: string[];
    governance?: MCPGovernanceDecision;
}

const ownedMCPGovernanceContext = {
    source_surface: 'mcp_settings_page',
    config_scope: 'user_group',
} as const;

function inspectionMessage(decision?: string, reasons?: string[], fallback?: string): string {
    if (Array.isArray(reasons) && reasons.length > 0) {
        return reasons.join(' ');
    }
    if (typeof fallback === 'string' && fallback.trim().length > 0) {
        return fallback;
    }
    if (decision === 'require_approval') {
        return 'This MCP entry still needs an explicit approval boundary before it can be installed.';
    }
    return 'MCP configuration could not be completed.';
}

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

        installFromLibrary: async (name: string, env?: Record<string, string>): Promise<MCPInstallResult> => {
            try {
                const inspectRes = await fetch('/api/v1/mcp/library/inspect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, env, governance_context: ownedMCPGovernanceContext }),
                });
                if (!inspectRes.ok) {
                    const message = await inspectRes.text();
                    console.error('[MCP Library] Inspect failed:', message);
                    return { ok: false, message: inspectionMessage(undefined, undefined, message) };
                }

                const inspection = await inspectRes.json() as MCPLibraryInspectionResponse;
                if (inspection.decision !== 'allow') {
                    return {
                        ok: false,
                        message: inspectionMessage(inspection.decision, inspection.reasons),
                        governance: inspection.governance,
                    };
                }

                const res = await fetch('/api/v1/mcp/library/install', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, env, governance_context: ownedMCPGovernanceContext }),
                });
                if (res.ok) {
                    await get().fetchMCPServers();
                    return {
                        ok: true,
                        message: 'Installed into your current MCP group without an extra approval step.',
                        governance: inspection.governance,
                    };
                } else {
                    const message = await res.text();
                    console.error('[MCP Library] Install failed:', message);
                    return { ok: false, message: inspectionMessage(undefined, undefined, message), governance: inspection.governance };
                }
            } catch (err) {
                console.error('[MCP Library] Install failed:', err);
                return { ok: false, message: 'MCP configuration failed before install could complete.' };
            }
        },
    };
}
