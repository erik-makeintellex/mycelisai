import type { CapabilityManifest, MCPServerWithTools, SearchCapabilityStatus } from '@/store/useCortexStore';

export const mockServers: MCPServerWithTools[] = [
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

export const webResearchCapability: CapabilityManifest = {
    id: 'browser.search',
    name: 'Web Research',
    description: 'Search public web sources through a governed provider.',
    source: 'builtin',
    category: 'research',
    risk: 'medium',
    approval: 'optional',
    outputs: ['SearchResult', 'ResearchSummary'],
    writes: ['exchange.browser.research.results', 'artifacts.research'],
    allowed_roles: ['soma', 'research_lead'],
    audit: 'required',
    health_check: true,
    availability_status: 'available',
    fallback_behavior: 'Return a search capability blocker.',
    provider: 'searxng',
    bound_tool_name: 'web_search',
};

export const readySearchCapability: SearchCapabilityStatus = {
    provider: 'searxng',
    enabled: true,
    configured: true,
    supports_local_sources: false,
    supports_public_web: true,
    soma_tool_name: 'web_search',
    direct_soma_interaction: true,
    requires_hosted_api_token: false,
    max_results: 8,
    next_actions: ['Ask Soma to search the public web through the self-hosted SearXNG provider.'],
};
