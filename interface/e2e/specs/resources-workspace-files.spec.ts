import { expect, test, type Page } from "@playwright/test";
import { clickVisibleControl } from "../support/click-visible-control";

type ToolCallRecord = {
    tool: string;
    arguments: Record<string, unknown>;
};

async function enableAdvancedMode(page: Page) {
    await page.addInitScript(() => {
        window.localStorage.setItem("mycelis-advanced-mode", "true");
    });
}

async function openWorkspaceFiles(page: Page) {
    await enableAdvancedMode(page);
    await page.goto("/resources?tab=workspace", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("button", { name: /Output Files/i })).toBeVisible();
}

async function mockWorkspaceMCP(page: Page) {
    const calls: ToolCallRecord[] = [];
    const files = new Map<string, string>([
        ["workspace/proof.md", "# Existing Proof\nReadable through filesystem MCP."],
        ["workspace/groups/game-delivery/final/game-brief.md", "# Final Game Brief\nRetained user output."],
        ["workspace/groups/game-delivery/source/gameplay.js", "export const loop = 'playable';"],
    ]);
    const directories = new Set<string>([
        "workspace/logs",
        "workspace/groups",
        "workspace/groups/game-delivery",
        "workspace/groups/game-delivery/final",
        "workspace/groups/game-delivery/source",
        "workspace/groups/game-delivery/review",
        "workspace/groups/game-delivery/media",
    ]);

    await page.route("**/api/v1/groups", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
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
            }),
        });
    });

    await page.route("**/api/v1/groups/group-with-output/outputs?limit=20", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
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
                ],
            }),
        });
    });

    await page.route("**/api/v1/groups/group-empty/outputs?limit=20", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({ data: [] }),
        });
    });

    await page.route("**/api/v1/mcp/servers", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify([
                {
                    id: "filesystem-server",
                    name: "filesystem",
                    status: "connected",
                    transport: "stdio",
                    tools: [
                        { id: "list-directory", name: "list_directory", description: "List workspace files" },
                        { id: "read-text-file", name: "read_text_file", description: "Read workspace files" },
                        { id: "create-directory", name: "create_directory", description: "Create workspace folders" },
                        { id: "write-file", name: "write_file", description: "Write workspace files" },
                    ],
                },
            ]),
        });
    });

    await page.route("**/api/v1/mcp/servers/filesystem-server/tools/*/call", async (route) => {
        const tool = route.request().url().match(/\/tools\/([^/]+)\/call$/)?.[1] ?? "";
        let body: Record<string, unknown> = {};
        try {
            body = route.request().postDataJSON() as Record<string, unknown>;
        } catch {
            body = {};
        }
        const args = (body.arguments ?? {}) as Record<string, unknown>;
        calls.push({ tool, arguments: args });

        if (tool === "list_directory") {
            const currentPath = String(args.path ?? "workspace").replaceAll("\\", "/").replace(/\/$/, "");
            const prefix = `${currentPath}/`;
            const childDirs = Array.from(directories)
                .filter((name) => name.startsWith(prefix))
                .map((name) => name.slice(prefix.length))
                .filter((name) => name.length > 0 && !name.includes("/"))
                .sort();
            const childFiles = Array.from(files.keys())
                .filter((name) => name.startsWith(prefix))
                .map((name) => name.slice(prefix.length))
                .filter((name) => name.length > 0 && !name.includes("/"))
                .sort();
            const listing = [
                ...childDirs.map((name) => `[DIR] ${name}`),
                ...childFiles.map((name) => `[FILE] ${name}`),
            ].join("\n");
            await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ content: [{ type: "text", text: listing }] }) });
            return;
        }
        if (tool === "read_text_file") {
            const path = String(args.path ?? "");
            await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ content: [{ type: "text", text: files.get(path) ?? "" }] }) });
            return;
        }
        if (tool === "create_directory") {
            directories.add(String(args.path ?? ""));
            await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ content: [{ type: "text", text: "created" }] }) });
            return;
        }
        if (tool === "write_file") {
            files.set(String(args.path ?? ""), String(args.content ?? ""));
            await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ content: [{ type: "text", text: "written" }] }) });
            return;
        }
        await route.fulfill({ status: 404, contentType: "application/json", body: JSON.stringify({ error: `unexpected tool ${tool}` }) });
    });

    await page.route("**/api/v1/workspace/files/reveal?*", async (route) => {
        await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({ ok: true, data: { workspace_path: "workspace", folder_path: "workspace" } }),
        });
    });

    return calls;
}

async function waitForToolCall(
    page: Page,
    calls: ToolCallRecord[],
    predicate: (call: ToolCallRecord) => boolean,
) {
    for (let attempt = 0; attempt < 30; attempt += 1) {
        if (calls.some(predicate)) return;
        await page.waitForTimeout(250);
    }
    expect(calls).toContainEqual(expect.objectContaining({ tool: "expected tool call was not recorded" }));
}

test.describe("Resources workspace files", () => {
    test.skip(({ browserName }) => browserName !== "chromium", "Workspace Files proof is stabilized in Chromium for MVP review.");

    test("lets an operator browse, read, create, and write through filesystem MCP", async ({ page }) => {
        const calls = await mockWorkspaceMCP(page);
        await openWorkspaceFiles(page);

        await expect(page.getByTestId("workspace-group-output-selector")).toBeVisible();
        await expect(page.getByLabel("Select group")).toHaveValue("group-with-output");
        await expect(page.getByRole("tab", { name: /Team lead 1/i })).toBeVisible();
        await expect(page.getByRole("tab", { name: /Coders 1/i })).toBeVisible();
        await clickVisibleControl(page, page.getByRole("tab", { name: /Coders 1/i }));
        await expect(page.getByRole("button", { name: /Gameplay Loop/i })).toBeVisible();
        await expect(page.getByRole("button", { name: /Final Game Brief/i })).toHaveCount(0);
        await clickVisibleControl(page, page.getByRole("tab", { name: /All 3/i }));
        await expect(page.getByText("Empty Group")).toHaveCount(0);
        await clickVisibleControl(page, page.getByRole("button", { name: /Final Game Brief/i }));
        await expect(page.locator("textarea").first()).toHaveValue(/Retained user output/i);
        await page.getByLabel("Include team source files").check();
        await waitForToolCall(page, calls, (call) => call.tool === "list_directory" && call.arguments.path === "workspace/groups/game-delivery");

        await openWorkspaceFiles(page);
        await expect(page.getByText("proof.md")).toBeVisible({ timeout: 15_000 });
        await clickVisibleControl(page, page.getByRole("button", { name: "Preview output file proof.md" }));
        await expect(page.locator("textarea").first()).toHaveValue(/Readable through filesystem MCP/i);
        expect(calls[0]).toEqual({ tool: "list_directory", arguments: { path: "workspace" } });
        expect(calls.some((call) => call.tool === "read_text_file" && call.arguments.path === "workspace/proof.md")).toBeTruthy();

        await clickVisibleControl(page, page.getByRole("button", { name: /Open current folder workspace/i }));
        await expect(page.getByText(/Opened local folder for workspace/i)).toBeVisible();

        await clickVisibleControl(page, page.getByRole("tab", { name: /Create/i }));
        await page.getByPlaceholder("new directory name").fill("generated");
        await clickVisibleControl(page, page.getByRole("button", { name: /Create Dir/i }));
        await waitForToolCall(page, calls, (call) => call.tool === "create_directory" && call.arguments.path === "workspace/generated");
        await clickVisibleControl(page, page.getByRole("tab", { name: /Find outputs/i }));
        await expect(page.getByRole("button", { name: "Open folder generated" })).toBeVisible();

        await clickVisibleControl(page, page.getByRole("tab", { name: /Create/i }));
        await page.getByPlaceholder("new file name").fill("generated-proof.md");
        await page.getByPlaceholder("Optional content for new file").fill("# Generated Proof\nCreated from the Workspace Files GUI.");
        await clickVisibleControl(page, page.getByRole("button", { name: /Write File/i }));
        await waitForToolCall(page, calls, (call) => call.tool === "write_file" && call.arguments.path === "workspace/generated-proof.md");
        await clickVisibleControl(page, page.getByRole("tab", { name: /Find outputs/i }));
        await expect(page.getByText("generated-proof.md")).toBeVisible();

        await clickVisibleControl(page, page.getByRole("button", { name: "Preview output file generated-proof.md" }));
        await expect(page.locator("textarea").first()).toHaveValue(/Created from the Workspace Files GUI/i);
    });

    test("writes a retained workspace output through the live filesystem MCP", async ({ page }) => {
        test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend with filesystem MCP connected");
        test.slow();
        test.setTimeout(120_000);

        const stamp = Date.now();
        const filename = `mcp_gui_output_${stamp}.md`;
        const marker = `MCP GUI Output Proof ${stamp}`;

        await openWorkspaceFiles(page);

        await clickVisibleControl(page, page.getByRole("tab", { name: /Create/i }));
        await expect(page.getByPlaceholder("new file name")).toBeVisible({ timeout: 20_000 });
        await page.getByPlaceholder("new file name").fill(filename);
        await page.getByPlaceholder("Optional content for new file").fill(`# ${marker}\n\nCreated through Resources Workspace Files using filesystem MCP.`);
        await clickVisibleControl(page, page.getByRole("button", { name: /Write File/i }));

        await clickVisibleControl(page, page.getByRole("tab", { name: /Find outputs/i }));
        await expect(page.getByText(filename)).toBeVisible({ timeout: 30_000 });
        await clickVisibleControl(page, page.getByRole("button", { name: `Preview output file ${filename}` }));
        await expect(page.locator("textarea").first()).toHaveValue(new RegExp(marker), { timeout: 20_000 });

        const viewResponse = await page.request.get(`/api/v1/workspace/files/view?path=${encodeURIComponent(`workspace/${filename}`)}`);
        expect(viewResponse.ok()).toBeTruthy();
        expect(await viewResponse.text()).toContain(marker);
    });
});
