// MCPA D10: blocker copy for the two new MCP configuration codes
// (mcp_server_not_found, mcp_connect_failed). Kept out of blockerCopy.ts
// because that file sits at its 385-line cap; blockerCopy.ts imports and
// merges these into its own CODE_COPY map, so it stays the single lookup
// callers use.
import type { CopyTemplate } from './blockerCopy';

export const MCP_CODE_COPY: Record<string, CopyTemplate> = {
    // MCPA D9: delete resolves the server before auditing; an unknown id
    // answers 404 with no audit and no disconnect (mirrors
    // codeMCPServerNotFound, mcp_config_authority.go).
    mcp_server_not_found: {
        title: "That connector is already gone",
        whatHappened: 'It may already have been removed, maybe in another tab. Nothing was changed.',
        nextAction: { label: 'Refresh', intent: 'retry' },
    },
    // MCPA: install/apply's Connect failed after the server row was
    // created; it is never reported "installed" (mirrors
    // codeMCPConnectFailed, mcp_library.go).
    mcp_connect_failed: {
        title: "That connector couldn't start",
        whatHappened: 'It was registered but could not connect. Nothing is installed yet.',
        nextAction: { label: 'Try again', intent: 'retry' },
        whoCanHelp: 'An admin can check the connection details.',
        adminVariant: {
            title: 'MCP server failed to connect',
            whatHappened: 'The server row was created but Connect failed after install. Check the Core log for the redacted reason, then retry or remove the entry.',
            nextAction: { label: 'Open Servers', href: '/resources?tab=tools', intent: 'open' },
        },
    },
};
