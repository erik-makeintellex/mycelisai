import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

import MCPLibraryBrowser from '@/components/settings/MCPLibraryBrowser';
import { useCortexStore, type MCPLibraryCategory } from '@/store/useCortexStore';

const mockLibrary: MCPLibraryCategory[] = [
    {
        name: 'Development',
        servers: [
            {
                name: 'filesystem',
                title: 'Filesystem',
                description: 'Read and write workspace files',
                version: 'latest',
                transport: 'stdio',
                command: 'npx',
                args: ['-y', '@modelcontextprotocol/server-filesystem', './workspace'],
                packages: [
                    {
                        registry_type: 'npm',
                        identifier: '@modelcontextprotocol/server-filesystem',
                        version: 'latest',
                        transport: { type: 'stdio' },
                    },
                ],
                repository: 'https://github.com/modelcontextprotocol/servers',
                homepage: 'https://modelcontextprotocol.io',
                tags: ['files', 'local'],
            },
        ],
    },
];

describe('MCPLibraryBrowser', () => {
    beforeEach(() => {
        useCortexStore.setState({
            mcpLibrary: mockLibrary,
            mcpServers: [],
            isFetchingMCPLibrary: false,
            fetchMCPLibrary: vi.fn(),
            installFromLibrary: vi.fn().mockResolvedValue({ ok: true, message: 'Installed into your current MCP group without an extra approval step.' }),
        });
    });

    it('explains connector installation and configuration posture', () => {
        render(<MCPLibraryBrowser />);

        expect(screen.getByText(/Add capability connector/i)).toBeDefined();
        expect(screen.getByText(/connectors add explicit URL reading/i)).toBeDefined();
        expect(screen.getByText(/@modelcontextprotocol\/server-filesystem/i)).toBeDefined();
        expect(screen.getByText(/Version policy: latest \(curated upstream tracking\)/i)).toBeDefined();
        expect(screen.getByText(/Ready to install with default configuration/i)).toBeDefined();
        expect(screen.getByText(/Capability binding: filesystem/i)).toBeDefined();
        expect(screen.getByText(/outputs normalize through Managed Exchange/i)).toBeDefined();
        expect((screen.getByRole('link', { name: 'Repository' }) as HTMLAnchorElement).href).toBe('https://github.com/modelcontextprotocol/servers');
        expect((screen.getByRole('link', { name: 'Homepage' }) as HTMLAnchorElement).href).toBe('https://modelcontextprotocol.io/');
    });

    it('surfaces install status instead of a follow-up approval prompt', async () => {
        const installFromLibrary = vi.fn().mockResolvedValue({
            ok: false,
            message: 'This MCP entry still needs an explicit approval boundary before it can be installed.',
        });

        useCortexStore.setState({ installFromLibrary });

        render(<MCPLibraryBrowser />);

        fireEvent.click(screen.getByRole('button', { name: /install/i }));

        await waitFor(() => {
            expect(installFromLibrary).toHaveBeenCalledWith('filesystem', undefined);
        });
        expect(screen.getByText(/still needs an explicit approval boundary/i)).toBeDefined();
    });

    it('renders typed environment variable guidance when present', async () => {
        useCortexStore.setState({
            mcpLibrary: [
                {
                    name: 'Development',
                    servers: [
                        {
                            name: 'github',
                            title: 'GitHub',
                            description: 'GitHub API integration',
                            version: 'latest',
                            transport: 'stdio',
                            command: 'npx',
                            args: ['-y', '@modelcontextprotocol/server-github'],
                            packages: [
                                {
                                    registry_type: 'npm',
                                    identifier: '@modelcontextprotocol/server-github',
                                    version: 'latest',
                                    transport: { type: 'stdio' },
                                },
                            ],
                            environment_variables: [
                                {
                                    name: 'GITHUB_PERSONAL_ACCESS_TOKEN',
                                    description: 'GitHub personal access token used for repository access.',
                                    required: true,
                                    secret: true,
                                },
                            ],
                            tags: ['git'],
                        },
                    ],
                },
            ],
        });

        render(<MCPLibraryBrowser />);

        fireEvent.click(screen.getByRole('button', { name: /install/i }));

        expect(await screen.findByText(/GitHub personal access token used for repository access/i)).toBeDefined();
        expect((screen.getByLabelText(/GITHUB_PERSONAL_ACCESS_TOKEN \*/i) as HTMLInputElement).type).toBe('password');
    });

    it('shows installed service connectors as configurable connection profiles', () => {
        useCortexStore.setState({
            mcpServers: [{
                id: 'postgres-1',
                name: 'postgres',
                transport: 'stdio',
                command: 'npx',
                args: ['-y', '@modelcontextprotocol/server-postgres'],
                status: 'connected',
                created_at: '2026-07-15T00:00:00Z',
                tools: [],
            }],
            mcpLibrary: [{
                name: 'Data & Search',
                servers: [{
                    name: 'postgres',
                    title: 'PostgreSQL',
                    description: 'Connect named PostgreSQL data sources for governed Soma/team querying',
                    version: 'latest',
                    transport: 'stdio',
                    command: 'npx',
                    args: ['-y', '@modelcontextprotocol/server-postgres'],
                    environment_variables: [{
                        name: 'POSTGRES_URL',
                        description: 'Connection string for one named PostgreSQL data source.',
                        required: true,
                        secret: true,
                    }],
                    tags: ['database', 'sql', 'data-source'],
                    tool_set: 'data_access',
                    configuration_kind: 'connection_profiles',
                    connection_resource: 'postgresql_database',
                    multiple_connections: true,
                    configuration_hint: 'Use named connections for user or customer data sources.',
                }],
            }],
        });

        render(<MCPLibraryBrowser />);

        expect(screen.getByRole('button', { name: /configure/i })).toBeDefined();
        expect(screen.getByText(/Use named data-source connections/i)).toBeDefined();
        expect(screen.getByText(/multiple named connections/i)).toBeDefined();
        expect(screen.getByText(/postgresql_database/i)).toBeDefined();
    });
});
