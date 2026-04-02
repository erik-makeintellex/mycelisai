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
                description: 'Read and write workspace files',
                transport: 'stdio',
                command: 'npx',
                args: ['-y', '@modelcontextprotocol/server-filesystem', './workspace'],
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

    it('explains the current-group auto-install posture', () => {
        render(<MCPLibraryBrowser />);

        expect(screen.getByText(/Current Group MCP Config/i)).toBeDefined();
        expect(screen.getByText(/Local-first curated entries install directly without another approval step/i)).toBeDefined();
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
});
