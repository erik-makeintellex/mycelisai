import { expect, test, type Page } from "@playwright/test";

type RouteLike = {
    fulfill: (options: { status: number; contentType: string; body: string }) => Promise<void>;
    request: () => { method: () => string };
};

async function fulfillJSON(route: RouteLike, status: number, body: unknown) {
    await route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

type MockConnectedToolsOptions = {
    servers?: Array<{
        id: string;
        name: string;
        transport: string;
        status: string;
        tools: Array<{
            id: string;
            name: string;
            description: string;
        }>;
    }>;
    activity?: Array<{
        id: string;
        server_id?: string;
        server_name: string;
        tool_name: string;
        state: string;
        summary: string;
        message: string;
        timestamp: string;
    }>;
};

async function mockConnectedToolsApis(page: Page, options: MockConnectedToolsOptions = {}) {
    const filesystemServer = {
        id: "mcp-filesystem",
        name: "filesystem",
        transport: "stdio",
        status: "connected",
        tools: [
            {
                id: "mcp-filesystem-list-directory",
                name: "list_directory",
                description: "List files available to workspace agents.",
            },
            {
                id: "mcp-filesystem-read-file",
                name: "read_file",
                description: "Read a workspace file for an assigned agent task.",
            },
        ],
    };
    const fetchServer = {
        id: "mcp-fetch",
        name: "fetch",
        transport: "stdio",
        status: "connected",
        tools: [
            {
                id: "mcp-fetch-fetch",
                name: "fetch",
                description: "Fetch a URL and return normalized page content.",
            },
        ],
    };
    const servers = options.servers ?? [filesystemServer];
    const activity = options.activity ?? [
        {
            id: "activity-read-file",
            server_id: "mcp-filesystem",
            server_name: "filesystem",
            tool_name: "read_file",
            state: "success",
            summary: "Soma used filesystem.read_file while preparing the launch brief.",
            message: "Soma used filesystem.read_file while preparing the launch brief.",
            timestamp: "2026-04-11T12:10:00Z",
        },
    ];
    const library = [
        {
            name: "Research",
            servers: [
                {
                    name: "fetch",
                    title: "Fetch",
                    description: "Fetch web pages for research-backed agent work.",
                    tags: ["research", "web", "mcp"],
                    version: "latest",
                    packages: [
                        {
                            identifier: "@modelcontextprotocol/server-fetch",
                            version: "latest",
                            transport: { type: "stdio" },
                        },
                    ],
                    repository: "https://github.com/modelcontextprotocol/servers",
                },
            ],
        },
    ];

    await page.route("**/api/v1/user/me", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: {
                id: "operator-1",
                name: "Operator",
                email: "operator@example.test",
            },
        });
    });

    await page.route("**/api/v1/services/status", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: [
                { name: "core", status: "ready" },
                { name: "frontend", status: "ready" },
            ],
        });
    });

    await page.route("**/api/v1/mcp/servers", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: servers });
    });

    await page.route("**/api/v1/mcp/activity?limit=12", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: activity });
    });

    await page.route("**/api/v1/mcp/library", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: library });
    });

    await page.route("**/api/v1/mcp/library/inspect", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            decision: "allow",
            reasons: ["Local-first curated MCP entry can install into the operator-owned group."],
            data: {
                approved: true,
                governance: {
                    allowed: true,
                    reason: "Local-first curated MCP entry can install into the operator-owned group.",
                },
            },
        });
    });

    await page.route("**/api/v1/mcp/library/install", async (route) => {
        if (!servers.some((server) => server.id === fetchServer.id)) {
            servers.push(fetchServer);
            activity.unshift({
                id: "activity-fetch-install",
                server_id: "mcp-fetch",
                server_name: "fetch",
                tool_name: "fetch",
                state: "installed",
                summary: "Fetch MCP server installed for the current user-owned group.",
                message: "Fetch MCP server installed for the current user-owned group.",
                timestamp: "2026-04-11T12:12:00Z",
            });
        }
        await fulfillJSON(route, 200, {
            ok: true,
            data: {
                ok: true,
                message: "Installed into your current MCP group without an extra approval step.",
            },
        });
    });
}

async function openConnectedTools(page: Page) {
    for (let attempt = 0; attempt < 2; attempt += 1) {
        await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
        await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
        await page.waitForFunction(() => window.localStorage.getItem("mycelis-advanced-mode") === "true");
        await page.reload({ waitUntil: "domcontentloaded" });
        if (await page.getByTestId("nav-resources").isVisible({ timeout: 5_000 }).catch(() => false)) {
            break;
        }
    }

    await expect(page.getByTestId("nav-resources")).toBeVisible({ timeout: 10_000 });
    await page.goto("/resources?tab=tools", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 20_000 });
}

test.describe("Connected Tools MCP workflow", () => {
    test.skip(({ browserName }) => browserName !== "chromium", "Connected Tools browser workflow proof is stabilized in Chromium for MVP review.");

    test("shows active MCP usage and installs a curated server from the library", async ({ page }) => {
        await mockConnectedToolsApis(page);
        await openConnectedTools(page);

        await expect(page.getByRole("button", { name: "Connected Tools" })).toBeVisible();
        await expect(page.getByText("Connected Tools Workflow")).toBeVisible();
        await expect(page.getByText("Recent MCP Activity", { exact: true })).toBeVisible();
        await expect(page.getByText("filesystem · read_file")).toBeVisible();
        await expect(page.getByText("Soma used filesystem.read_file while preparing the launch brief.").first()).toBeVisible();
        await expect(page.getByText("filesystem").first()).toBeVisible();
        await expect(page.getByText("2 tools")).toBeVisible();

        await page.getByRole("button", { name: /filesystem.*2 tools/i }).click();
        await expect(page.getByText("Live Usage")).toBeVisible();
        await expect(page.getByText("list_directory", { exact: true })).toBeVisible();
        await expect(page.getByText("read_file", { exact: true }).last()).toBeVisible();

        await page.getByRole("button", { name: "BROWSE LIBRARY" }).click();
        await expect(page.getByText("Current Group MCP Config")).toBeVisible();
        await expect(page.getByText("Fetch", { exact: true })).toBeVisible();
        await expect(page.getByText("Fetch web pages for research-backed agent work.")).toBeVisible();

        await page.getByRole("button", { name: "INSTALL", exact: true }).click();
        await expect(page.getByText("Installed fetch. Check the connected server card and live MCP activity below.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("fetch").first()).toBeVisible();
        await expect(page.getByText("Fetch MCP server installed for the current user-owned group.").first()).toBeVisible();
    });

    test("shows the bootstrap-disabled empty state and sends the operator to the library", async ({ page }) => {
        await mockConnectedToolsApis(page, { servers: [], activity: [] });
        await openConnectedTools(page);

        await expect(page.getByText("No MCP servers installed.", { exact: true })).toBeVisible({ timeout: 20_000 });
        await expect(page.getByRole("button", { name: "OPEN LIBRARY" })).toBeVisible();

        await page.getByRole("button", { name: "OPEN LIBRARY" }).click();
        await expect(page.getByText("Current Group MCP Config")).toBeVisible();
        await expect(page.getByText("Fetch", { exact: true })).toBeVisible();
    });
});
