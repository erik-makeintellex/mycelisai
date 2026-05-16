import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import WorkspaceExplorer from "@/components/resources/WorkspaceExplorer";

const mockFetchMCPServers = vi.fn();

vi.mock("@/store/useCortexStore", () => ({
    useCortexStore: (selector: any) =>
        selector({
            mcpServers: [
                {
                    id: "filesystem-server",
                    name: "filesystem",
                    status: "connected",
                    tools: [
                        { name: "list_directory" },
                        { name: "read_text_file" },
                        { name: "create_directory" },
                        { name: "write_file" },
                    ],
                },
            ],
            isFetchingMCPServers: false,
            fetchMCPServers: mockFetchMCPServers,
        }),
}));

function mockToolFetch() {
    const calls: Array<{ tool: string; body: any }> = [];
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const match = url.match(/\/tools\/([^/]+)\/call$/);
        const tool = match?.[1] ?? "";
        const body = init?.body ? JSON.parse(String(init.body)) : {};
        calls.push({ tool, body });

        if (tool === "list_directory") {
            return Response.json({ content: [{ type: "text", text: "[FILE] proof.md\n[DIR] outputs" }] });
        }
        if (tool === "read_text_file") {
            return Response.json({ content: [{ type: "text", text: "# Proof\nReadable through filesystem MCP." }] });
        }
        if (tool === "create_directory" || tool === "write_file") {
            return Response.json({ content: [{ type: "text", text: "ok" }] });
        }
        return Response.json({ error: `unexpected tool ${tool}` }, { status: 500 });
    });
    vi.stubGlobal("fetch", fetchMock);
    return { calls, fetchMock };
}

describe("WorkspaceExplorer", () => {
    beforeEach(() => {
        vi.restoreAllMocks();
        mockFetchMCPServers.mockReset();
    });

    it("uses current filesystem MCP tool names with the arguments envelope", async () => {
        const { calls } = mockToolFetch();

        render(<WorkspaceExplorer onOpenToolsTab={vi.fn()} />);

        await waitFor(() => {
            expect(calls.some((call) => call.tool === "list_directory")).toBe(true);
        });
        expect(calls[0].body).toEqual({ arguments: { path: "workspace" } });

        fireEvent.click(await screen.findByText("proof.md"));
        await waitFor(() => {
            expect(calls.some((call) => call.tool === "read_text_file")).toBe(true);
        });
        expect(screen.getByDisplayValue(/Readable through filesystem MCP/i)).toBeDefined();

        fireEvent.change(screen.getByPlaceholderText("new directory name"), { target: { value: "generated" } });
        fireEvent.click(screen.getByRole("button", { name: /Create Dir/i }));
        await waitFor(() => {
            expect(calls.some((call) => call.tool === "create_directory")).toBe(true);
        });

        fireEvent.change(screen.getByPlaceholderText("new file name"), { target: { value: "new-proof.md" } });
        fireEvent.change(screen.getByPlaceholderText("Optional content for new file"), { target: { value: "# New proof" } });
        fireEvent.click(screen.getByRole("button", { name: /Write File/i }));
        await waitFor(() => {
            expect(calls.some((call) => call.tool === "write_file")).toBe(true);
        });
        const writeCall = calls.find((call) => call.tool === "write_file");
        expect(writeCall?.body).toEqual({ arguments: { path: "workspace/new-proof.md", content: "# New proof" } });
    });
});
