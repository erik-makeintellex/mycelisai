import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import MCPServerCard from "@/components/settings/MCPServerCard";
import type { MCPServerWithTools } from "@/store/useCortexStore";

const server: MCPServerWithTools = {
    id: "srv-001",
    name: "filesystem-server",
    transport: "stdio",
    command: "npx",
    args: ["-y", "@modelcontextprotocol/server-filesystem", "workspace/shared-sources"],
    env: { FILESYSTEM_ROOT: "[redacted]" },
    headers: { Authorization: "[redacted]" },
    status: "connected",
    created_at: "2026-04-30T12:00:00Z",
    tools: [
        {
            id: "tool-1",
            server_id: "srv-001",
            name: "read_file",
            description: "Read a host file.",
            input_schema: {},
        },
    ],
};

describe("MCPServerCard", () => {
    it("expands into a readable MCP structure review surface", () => {
        const onEdit = vi.fn();
        render(
            <MCPServerCard
                server={server}
                onDelete={vi.fn()}
                onEdit={onEdit}
                recentActivity={[]}
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /filesystem-server/i }));

        expect(screen.getByText("MCP Structure")).toBeDefined();
        expect(screen.getByText("Command")).toBeDefined();
        expect(screen.getByText("npx")).toBeDefined();
        expect(screen.getByText(/workspace\/shared-sources/i)).toBeDefined();
        expect(screen.getByText("FILESYSTEM_ROOT")).toBeDefined();
        expect(screen.getByText("Authorization")).toBeDefined();
        expect(screen.getByText(/Secrets are shown only as references/i)).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Edit in Library/i }));
        expect(onEdit).toHaveBeenCalledTimes(1);
    });
});
