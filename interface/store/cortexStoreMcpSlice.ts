import { extractApiData } from '@/lib/apiContracts';
import { normalizeCapabilitiesPayload, normalizeSearchCapabilityStatus } from '@/store/cortexStoreMcpCapabilities';
import type { CortexGet, CortexSet, CortexSlice } from '@/store/cortexStoreSliceTypes';
import type {
    CapabilityManifest,
    MCPActivityEntry,
    MCPGovernanceDecision,
    MCPInstallResult,
    MCPServerWithTools,
    MCPLibraryCategory,
    MCPTool,
    MCPToolSet,
    MCPToolSetCreate,
    MCPWriteResult,
    SearchCapabilityStatus,
} from '@/store/cortexStoreTypes';

interface MCPLibraryInspectionResponse {
    decision?: string;
    reasons?: string[];
    governance?: MCPGovernanceDecision;
}

const ownedMCPGovernanceContext = {
    source_surface: 'mcp_settings_page',
    config_scope: 'user_group',
} as const;

// MCPA D10: every write route answers a non-2xx as the blocker envelope
// `{ok:false, error, data:{code, ...}}` (respondBlocker/blocker_copy.go).
// This is the one place that reads `data.code` so callers never hand-parse
// or display raw backend text; an envelope with no code (a plain
// respondError 400/404/500) falls back to the generic `request_failed`
// copy, which is still honest -- never a success toast on failure.
async function readMCPBlocker(res: Response): Promise<{ code: string; httpStatus: number }> {
    let code = 'request_failed';
    try {
        const payload = await res.json();
        if (typeof payload?.data?.code === 'string') {
            code = payload.data.code;
        }
    } catch {
        // no JSON body: keep the generic fallback code
    }
    return { code, httpStatus: res.status };
}

export function createCortexMcpSlice(
    set: CortexSet,
    get: CortexGet,
): CortexSlice<
    | 'fetchMCPServers'
    | 'fetchMCPActivity'
    | 'deleteMCPServer'
    | 'fetchMCPTools'
    | 'fetchMCPToolSets'
    | 'createMCPToolSet'
    | 'fetchMCPLibrary'
    | 'installFromLibrary'
    | 'fetchSearchCapability'
    | 'fetchCapabilities'
> {
    return {
        fetchMCPServers: async () => {
            set({ isFetchingMCPServers: true, mcpServersError: null });
            try {
                const res = await fetch('/api/v1/mcp/servers');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<MCPServerWithTools[] | unknown>(payload);
                    set({ mcpServers: Array.isArray(data) ? data : [], isFetchingMCPServers: false, mcpServersError: null });
                } else {
                    set({ mcpServers: [], isFetchingMCPServers: false, mcpServersError: `MCP registry unreachable (HTTP ${res.status})` });
                }
            } catch (err) {
                const message = err instanceof Error ? err.message : 'network error';
                set({ mcpServers: [], isFetchingMCPServers: false, mcpServersError: `MCP registry unreachable (${message})` });
            }
        },

        fetchMCPActivity: async () => {
            set({ isFetchingMCPActivity: true });
            try {
                const res = await fetch('/api/v1/mcp/activity?limit=12');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<MCPActivityEntry[] | unknown>(payload);
                    set({ mcpActivity: Array.isArray(data) ? data : [], isFetchingMCPActivity: false });
                } else {
                    set({ mcpActivity: [], isFetchingMCPActivity: false });
                }
            } catch {
                set({ mcpActivity: [], isFetchingMCPActivity: false });
            }
        },

        // MCPA D9/D10: admin-only, fail-closed. A non-2xx (403 admin_required,
        // 404 mcp_server_not_found, 503 service_unavailable, ...) never
        // touches local state, so a rejected delete cannot be mistaken for a
        // removed server.
        deleteMCPServer: async (id: string): Promise<MCPWriteResult> => {
            try {
                const res = await fetch(`/api/v1/mcp/servers/${id}`, { method: 'DELETE' });
                if (!res.ok) {
                    const blocker = await readMCPBlocker(res);
                    console.error('[MCP] Delete failed:', blocker.code, res.status);
                    return { ok: false, code: blocker.code, httpStatus: blocker.httpStatus };
                }
                set((s) => ({
                    mcpServers: s.mcpServers.filter((server) => server.id !== id),
                }));
                return { ok: true };
            } catch (err) {
                console.error('[MCP] Delete failed:', err);
                return { ok: false, code: 'request_failed', httpStatus: 0 };
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

        fetchMCPToolSets: async () => {
            set({ isFetchingMCPToolSets: true, mcpToolSetsError: null });
            try {
                const res = await fetch('/api/v1/mcp/toolsets');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<MCPToolSet[] | unknown>(payload);
                    set({
                        mcpToolSets: Array.isArray(data) ? data : [],
                        isFetchingMCPToolSets: false,
                        mcpToolSetsError: null,
                    });
                } else {
                    set({
                        mcpToolSets: [],
                        isFetchingMCPToolSets: false,
                        mcpToolSetsError: `MCP access layers unreachable (HTTP ${res.status})`,
                    });
                }
            } catch (err) {
                const message = err instanceof Error ? err.message : 'network error';
                set({
                    mcpToolSets: [],
                    isFetchingMCPToolSets: false,
                    mcpToolSetsError: `MCP access layers unreachable (${message})`,
                });
            }
        },

        // MCPA D10: admin-only, fail-closed. A non-2xx surfaces the blocker
        // envelope's code instead of raw backend text; `mcpToolSetsError`
        // stays a short, non-raw summary for the list-level banner.
        createMCPToolSet: async (input: MCPToolSetCreate): Promise<MCPWriteResult> => {
            try {
                const res = await fetch('/api/v1/mcp/toolsets', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(input),
                });
                if (!res.ok) {
                    const blocker = await readMCPBlocker(res);
                    set({ mcpToolSetsError: `MCP access layer was rejected (HTTP ${res.status})` });
                    return { ok: false, code: blocker.code, httpStatus: blocker.httpStatus };
                }
                await get().fetchMCPToolSets();
                return { ok: true };
            } catch (err) {
                const message = err instanceof Error ? err.message : 'network error';
                set({ mcpToolSetsError: `MCP access layer could not be saved (${message})` });
                return { ok: false, code: 'request_failed', httpStatus: 0 };
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

        fetchSearchCapability: async () => {
            set({ isFetchingSearchCapability: true, searchCapabilityError: null });
            try {
                const res = await fetch('/api/v1/search/status');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<SearchCapabilityStatus | unknown>(payload);
                    const status = normalizeSearchCapabilityStatus(data);
                    set({
                        searchCapability: status,
                        isFetchingSearchCapability: false,
                        searchCapabilityError: status ? null : 'Search capability status was not readable.',
                    });
                } else {
                    set({
                        searchCapability: null,
                        isFetchingSearchCapability: false,
                        searchCapabilityError: `Search capability status unreachable (HTTP ${res.status})`,
                    });
                }
            } catch (err) {
                const message = err instanceof Error ? err.message : 'network error';
                set({
                    searchCapability: null,
                    isFetchingSearchCapability: false,
                    searchCapabilityError: `Search capability status unreachable (${message})`,
                });
            }
        },

        fetchCapabilities: async () => {
            set({ isFetchingCapabilities: true, capabilitiesError: null });
            try {
                const res = await fetch('/api/v1/capabilities');
                if (res.ok) {
                    const payload = await res.json();
                    const data = extractApiData<CapabilityManifest[] | { capabilities?: unknown; manifests?: unknown } | unknown>(payload);
                    const capabilities = normalizeCapabilitiesPayload(data);
                    set({
                        capabilities,
                        isFetchingCapabilities: false,
                        capabilitiesError: capabilities.length > 0 ? null : 'Capability registry returned no readable manifests.',
                    });
                } else {
                    set({
                        capabilities: [],
                        isFetchingCapabilities: false,
                        capabilitiesError: `Capability registry unreachable (HTTP ${res.status})`,
                    });
                }
            } catch (err) {
                const message = err instanceof Error ? err.message : 'network error';
                set({
                    capabilities: [],
                    isFetchingCapabilities: false,
                    capabilitiesError: `Capability registry unreachable (${message})`,
                });
            }
        },

        // MCPA D5/D10: `require_approval` is a real tier-2 path now, not a
        // dead end -- install always goes through for an approver
        // (self-approved tier 2), so the client never blocks on the
        // inspect decision itself. The server is the only place that
        // enforces `approvals:decide` (403 admin_required otherwise). No
        // path here expects a 202: install/apply answer 200 or a blocker.
        installFromLibrary: async (name: string, env?: Record<string, string>): Promise<MCPInstallResult> => {
            try {
                const inspectRes = await fetch('/api/v1/mcp/library/inspect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, env, governance_context: ownedMCPGovernanceContext }),
                });
                if (!inspectRes.ok) {
                    const blocker = await readMCPBlocker(inspectRes);
                    console.error('[MCP Library] Inspect failed:', blocker.code, inspectRes.status);
                    return { ok: false, code: blocker.code, httpStatus: blocker.httpStatus };
                }
                const inspection = await inspectRes.json() as MCPLibraryInspectionResponse;

                const res = await fetch('/api/v1/mcp/library/install', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, env, governance_context: ownedMCPGovernanceContext }),
                });
                if (!res.ok) {
                    const blocker = await readMCPBlocker(res);
                    console.error('[MCP Library] Install failed:', blocker.code, res.status);
                    return { ok: false, code: blocker.code, httpStatus: blocker.httpStatus, governance: inspection.governance };
                }
                await get().fetchMCPServers();
                await get().fetchMCPActivity();
                // Servers are global (no owner/group column), never scoped
                // to "your current MCP group".
                return {
                    ok: true,
                    message: 'Installed. It is available across the workspace.',
                    governance: inspection.governance,
                };
            } catch (err) {
                console.error('[MCP Library] Install failed:', err);
                return { ok: false, code: 'request_failed', httpStatus: 0 };
            }
        },
    };
}
