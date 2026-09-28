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
    capability_ids: ["mcp.filesystem.read_file"],
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
        expect(screen.getByText("Capability bindings")).toBeDefined();
        expect(screen.getByText("mcp.filesystem.read_file")).toBeDefined();
        expect(screen.getByText(/Secrets are shown only as references/i)).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Edit in Library/i }));
        expect(onEdit).toHaveBeenCalledTimes(1);
    });

    // MCPA D10: delete is an admin-only control. isAdmin is a UI hint only
    // (Core still enforces mcp_config:write); it defaults to false so a
    // caller that has not resolved the viewer role yet shows no control it
    // cannot back.
    it("hides the delete control for a non-admin viewer", () => {
        render(<MCPServerCard server={server} onDelete={vi.fn()} recentActivity={[]} />);

        expect(screen.queryByTitle("Delete server")).toBeNull();
    });

    it("lets an admin delete after a confirm click", async () => {
        const onDelete = vi.fn().mockResolvedValue({ ok: true });
        render(<MCPServerCard server={server} onDelete={onDelete} recentActivity={[]} isAdmin />);

        fireEvent.click(screen.getByTitle("Delete server"));
        fireEvent.click(screen.getByTitle("Click again to confirm"));

        expect(onDelete).toHaveBeenCalledWith("srv-001");
    });

    // MCPA D10: honest failure -- a rejected delete renders blocker copy
    // instead of a raw error and never a success toast.
    it("renders blocker copy on a rejected delete instead of a raw error", async () => {
        const onDelete = vi.fn().mockResolvedValue({ ok: false, code: "admin_required", httpStatus: 403 });
        render(<MCPServerCard server={server} onDelete={onDelete} recentActivity={[]} isAdmin />);

        fireEvent.click(screen.getByTitle("Delete server"));
        fireEvent.click(screen.getByTitle("Click again to confirm"));

        expect(await screen.findByRole("alert")).toBeDefined();
        expect(screen.queryByText(/"code":"admin_required"/)).toBeNull();
    });
});
