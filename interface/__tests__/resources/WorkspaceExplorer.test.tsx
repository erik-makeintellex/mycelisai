import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import WorkspaceExplorer from "@/components/resources/WorkspaceExplorer";

const mockFetchMCPServers = vi.fn();
const connectedFilesystemServer = {
    id: "filesystem-server",
    name: "filesystem",
    status: "connected",
    tools: [
        { name: "list_directory" },
        { name: "read_text_file" },
        { name: "create_directory" },
        { name: "write_file" },
    ],
};

let mockMCPServers: Array<Record<string, unknown>> = [connectedFilesystemServer];

vi.mock("@/store/useCortexStore", () => ({
    useCortexStore: (selector: any) =>
        selector({
            mcpServers: mockMCPServers,
            isFetchingMCPServers: false,
            fetchMCPServers: mockFetchMCPServers,
        }),
}));

function mockToolFetch() {
    const calls: Array<{ tool: string; body: any }> = [];
    const revealCalls: string[] = [];
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        if (url === "/api/v1/groups") {
            return Response.json({
                data: [
                    {
                        group_id: "group-with-output",
                        name: "Game Delivery Group",
                        workspace_folder: "groups/game-delivery",
                    },
                    {
                        group_id: "group-empty",
                        name: "Empty Group",
                        workspace_folder: "groups/empty",
                    },
                ],
            });
        }
        if (url.includes("/api/v1/groups/group-with-output/outputs")) {
            return Response.json({
                data: [
                    {
                        id: "artifact-final",
                        agent_id: "lead",
                        artifact_type: "document",
                        title: "Final Game Brief",
                        content_type: "text/markdown",
                        file_path: "groups/game-delivery/final/game-brief.md",
                        metadata: {},
                        status: "approved",
                        created_at: new Date().toISOString(),
                    },
                    {
                        id: "artifact-code",
                        agent_id: "gameplay-coder",
                        artifact_type: "code",
                        title: "Gameplay Loop",
                        content_type: "text/javascript",
                        file_path: "groups/game-delivery/source/gameplay.js",
                        metadata: { role: "coder" },
                        status: "approved",
                        created_at: new Date().toISOString(),
                    },
                    {
                        id: "artifact-review",
                        agent_id: "qa-reviewer",
                        artifact_type: "document",
                        title: "QA Review Notes",
                        content_type: "text/markdown",
                        file_path: "groups/game-delivery/review/qa.md",
                        metadata: { role: "reviewer" },
                        status: "approved",
                        created_at: new Date().toISOString(),
                    },
                    {
                        id: "artifact-media",
                        agent_id: "asset-artist",
                        artifact_type: "image",
                        title: "Sprite Sheet",
                        content_type: "image/png",
                        file_path: "groups/game-delivery/media/sprites.png",
                        metadata: { role: "media artist" },
                        status: "approved",
                        created_at: new Date().toISOString(),
                    },
                ],
            });
        }
        if (url.includes("/api/v1/groups/group-empty/outputs")) {
            return Response.json({ data: [] });
        }
        if (url.includes("/api/v1/workspace/files/reveal")) {
            revealCalls.push(url);
            return Response.json({ ok: true, data: { workspace_path: ".", folder_path: "workspace" } });
        }
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
    return { calls, revealCalls, fetchMock };
}

describe("WorkspaceExplorer", () => {
    beforeEach(() => {
        vi.restoreAllMocks();
        mockFetchMCPServers.mockReset();
        mockMCPServers = [connectedFilesystemServer];
    });

    it("uses current filesystem MCP tool names with the arguments envelope", async () => {
        const { calls, revealCalls } = mockToolFetch();

        render(<WorkspaceExplorer onOpenToolsTab={vi.fn()} />);

        await waitFor(() => {
            expect(calls.some((call) => call.tool === "list_directory")).toBe(true);
        });
        expect(calls[0].body).toEqual({ arguments: { path: "workspace" } });
        expect(screen.getByText("Open generated output on this machine")).toBeDefined();
        expect(await screen.findByTestId("workspace-group-output-selector")).toBeDefined();
        expect(screen.getByText("Game Delivery Group (4)")).toBeDefined();
        expect(screen.queryByText("Empty Group")).toBeNull();
        expect(screen.getByRole("link", { name: "Open group outputs" }).getAttribute("href")).toBe(
            "/groups?group_id=group-with-output&panel=outputs",
        );
        expect(screen.getByRole("link", { name: "Workflow log" }).getAttribute("href")).toBe(
            "/groups?group_id=group-with-output&panel=workflow",
        );
        expect(screen.getByRole("link", { name: "Message group" }).getAttribute("href")).toBe(
            "/groups?group_id=group-with-output&panel=message",
        );
        expect(screen.getByRole("tablist", { name: "Workspace output panes" })).toBeDefined();
        expect(screen.getByRole("tab", { name: /Find outputs/i }).getAttribute("aria-selected")).toBe("true");
        expect(screen.queryByPlaceholderText("new directory name")).toBeNull();

        fireEvent.click(screen.getByRole("button", { name: /Open current folder workspace/i }));
        await waitFor(() => {
            expect(revealCalls.some((url) => url.includes("path=workspace"))).toBe(true);
            expect(screen.getByText("Folder opened")).toBeDefined();
        });

        fireEvent.click(await screen.findByText("proof.md"));
        await waitFor(() => {
            expect(calls.some((call) => call.tool === "read_text_file")).toBe(true);
        });
        expect(screen.getByRole("tab", { name: /Preview/i }).getAttribute("aria-selected")).toBe("true");
        expect(screen.getByDisplayValue(/Readable through filesystem MCP/i)).toBeDefined();

        fireEvent.click(screen.getByRole("tab", { name: /Create/i }));
        expect(screen.getByRole("tabpanel", { name: /Create/i })).toBeDefined();
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

    it("opens retained group outputs and can include team source files on demand", async () => {
        const { calls } = mockToolFetch();

        render(<WorkspaceExplorer onOpenToolsTab={vi.fn()} />);

        await screen.findByText("Game Delivery Group (4)");
        fireEvent.click(screen.getByRole("button", { name: /Final Game Brief/i }));
        await waitFor(() => {
            expect(calls.some((call) => call.tool === "read_text_file" && call.body.arguments.path === "workspace/groups/game-delivery/final/game-brief.md")).toBe(true);
        });
        expect(screen.getByRole("tab", { name: /Preview/i }).getAttribute("aria-selected")).toBe("true");

        fireEvent.click(screen.getByLabelText(/Include team source files/i));
        await waitFor(() => {
            expect(calls.some((call) => call.tool === "list_directory" && call.body.arguments.path === "workspace/groups/game-delivery")).toBe(true);
        });
        expect(screen.getByText("workspace/groups/game-delivery")).toBeDefined();
    });

    it("filters retained group outputs by contributor level", async () => {
        mockToolFetch();

        render(<WorkspaceExplorer onOpenToolsTab={vi.fn()} />);

        await screen.findByText("Game Delivery Group (4)");
        expect(screen.getByRole("tablist", { name: /Output contributor level/i })).toBeDefined();
        expect(screen.getByRole("tab", { name: /Team lead 1/i })).toBeDefined();
        expect(screen.getByRole("tab", { name: /Coders 1/i })).toBeDefined();
        expect(screen.getByRole("tab", { name: /Review 1/i })).toBeDefined();
        expect(screen.getByRole("tab", { name: /Media 1/i })).toBeDefined();

        fireEvent.click(screen.getByRole("tab", { name: /Coders 1/i }));
        expect(screen.getByText("Gameplay Loop")).toBeDefined();
        expect(screen.queryByText("Final Game Brief")).toBeNull();

        fireEvent.click(screen.getByRole("tab", { name: /Review 1/i }));
        expect(screen.getByText("QA Review Notes")).toBeDefined();
        expect(screen.queryByText("Gameplay Loop")).toBeNull();
    });

    it("starts browsing at the deep-linked workspace path", async () => {
        const { calls } = mockToolFetch();

        render(<WorkspaceExplorer initialPath="workspace/generated/game" onOpenToolsTab={vi.fn()} />);

        await waitFor(() => {
            expect(calls.some((call) => call.tool === "list_directory")).toBe(true);
        });
        expect(calls[0].body).toEqual({ arguments: { path: "workspace/generated/game" } });
        expect(screen.getByText("workspace/generated/game")).toBeDefined();
    });

    it("normalizes retained output deep links into the workspace browser root", async () => {
        const { calls } = mockToolFetch();

        render(<WorkspaceExplorer initialPath="groups/game-team/generated/first-game" onOpenToolsTab={vi.fn()} />);

        await waitFor(() => {
            expect(calls.some((call) => call.tool === "list_directory")).toBe(true);
        });
        expect(calls[0].body).toEqual({ arguments: { path: "workspace/groups/game-team/generated/first-game" } });
        expect(screen.getByText("workspace/groups/game-team/generated/first-game")).toBeDefined();
    });

    it("guides operators to Capabilities and storage roots when filesystem is missing", async () => {
        mockMCPServers = [];
        const onOpenToolsTab = vi.fn();

        render(<WorkspaceExplorer onOpenToolsTab={onOpenToolsTab} />);

        expect(screen.getByText("Filesystem MCP not installed")).toBeDefined();
        expect(screen.getByRole("link", { name: /View storage roots/i }).getAttribute("href")).toBe("/system?tab=deployments");
        fireEvent.click(screen.getByRole("button", { name: /Open Capabilities/i }));
        expect(onOpenToolsTab).toHaveBeenCalledOnce();
        fireEvent.click(screen.getByRole("button", { name: /Refresh/i }));
        expect(mockFetchMCPServers).toHaveBeenCalled();
    });

    it("keeps storage-root guidance visible while filesystem is disconnected", async () => {
        mockMCPServers = [{ ...connectedFilesystemServer, status: "error" }];

        render(<WorkspaceExplorer onOpenToolsTab={vi.fn()} />);

        expect(screen.getByText("Filesystem MCP not connected")).toBeDefined();
        expect(screen.getByText("error")).toBeDefined();
        expect(screen.getByRole("link", { name: /View storage roots/i })).toBeDefined();
        expect(screen.getByText(/find generated output while the MCP server is recovering/i)).toBeDefined();
    });
});
