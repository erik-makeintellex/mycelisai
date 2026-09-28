import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { SVGProps } from "react";

vi.mock("lucide-react", () => {
    const icon = (props: SVGProps<SVGSVGElement>) => <svg {...props} />;
    return { X: icon, Wrench: icon, Server: icon, Zap: icon, ChevronDown: icon, ChevronRight: icon, Play: icon, Loader2: icon };
});

const mockFetchMCPServers = vi.fn();
const mockFetchMCPTools = vi.fn();
const mockToggle = vi.fn();

const connectedServer = {
    id: "gh-server",
    name: "github",
    status: "connected",
    tools: [{ name: "create_issue", description: "Open an issue" }],
};

type ToolsPaletteStore = {
    isToolsPaletteOpen: boolean;
    toggleToolsPalette: typeof mockToggle;
    mcpServers: Array<Record<string, unknown>>;
    mcpTools: unknown[];
    fetchMCPServers: typeof mockFetchMCPServers;
    fetchMCPTools: typeof mockFetchMCPTools;
};

vi.mock("@/store/useCortexStore", () => ({
    useCortexStore: (selector: (state: ToolsPaletteStore) => unknown) =>
        selector({
            isToolsPaletteOpen: true,
            toggleToolsPalette: mockToggle,
            mcpServers: [connectedServer],
            mcpTools: [],
            fetchMCPServers: mockFetchMCPServers,
            fetchMCPTools: mockFetchMCPTools,
        }),
}));

let mockIsAdmin = false;
vi.mock("@/lib/useIsAdmin", () => ({
    useIsAdmin: () => ({ isAdmin: mockIsAdmin, checked: true }),
}));

import ToolsPalette from "@/components/workspace/ToolsPalette";

describe("ToolsPalette", () => {
    beforeEach(() => {
        vi.restoreAllMocks();
        mockFetchMCPServers.mockReset();
        mockFetchMCPTools.mockReset();
        mockIsAdmin = false;
    });

    it("renders honest blocker copy, not raw backend JSON, when a direct call is forbidden", async () => {
        mockIsAdmin = false;
        vi.stubGlobal(
            "fetch",
            vi.fn(async () =>
                Response.json(
                    { ok: false, error: "Your admin account is missing a permission this needs.", data: { code: "admin_required", recommended_action: "Ask Soma to propose it." } },
                    { status: 403 },
                ),
            ),
        );

        render(<ToolsPalette />);
        fireEvent.click(screen.getByText("github"));
        fireEvent.click(await screen.findByText("create_issue"));
        const executeButton = screen.getByTitle("Execute tool");
        fireEvent.click(executeButton);

        await waitFor(() => {
            expect(screen.getByText(/Only an admin can use this tool directly/i)).toBeDefined();
        });
        expect(screen.queryByText(/"code":"admin_required"/)).toBeNull();
    });

    it("shows the admin variant of the blocker copy for an admin viewer", async () => {
        mockIsAdmin = true;
        vi.stubGlobal(
            "fetch",
            vi.fn(async () =>
                Response.json({ ok: false, error: "service unavailable", data: { code: "service_unavailable" } }, { status: 503 }),
            ),
        );

        render(<ToolsPalette />);
        fireEvent.click(screen.getByText("github"));
        fireEvent.click(await screen.findByText("create_issue"));
        fireEvent.click(screen.getByTitle("Execute tool"));

        await waitFor(() => {
            expect(screen.getByText(/Reconnect the MCP server from Resources/i)).toBeDefined();
        });
    });

    it("shows the raw tool result unchanged on success", async () => {
        vi.stubGlobal("fetch", vi.fn(async () => new Response("ok result text", { status: 200 })));

        render(<ToolsPalette />);
        fireEvent.click(screen.getByText("github"));
        fireEvent.click(await screen.findByText("create_issue"));
        fireEvent.click(screen.getByTitle("Execute tool"));

        await waitFor(() => {
            expect(screen.getByText("ok result text")).toBeDefined();
        });
    });
});
