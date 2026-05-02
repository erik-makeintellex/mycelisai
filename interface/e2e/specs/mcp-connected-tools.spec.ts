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
        channel_name?: string;
        run_id?: string;
        team_id?: string;
        agent_id?: string;
        timestamp: string;
    }>;
    searchStatus?: {
        provider: string;
        enabled: boolean;
        configured: boolean;
        supports_local_sources: boolean;
        supports_public_web: boolean;
        soma_tool_name: string;
        direct_soma_interaction: boolean;
        requires_hosted_api_token: boolean;
        max_results: number;
        blocker?: {
            code: string;
            message: string;
            next_action: string;
        };
        next_actions?: string[];
    };
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
            channel_name: "api.data.output",
            run_id: "run-launch-brief",
            team_id: "soma-launch-lane",
            agent_id: "soma",
            timestamp: "2026-04-11T12:10:00Z",
        },
    ];
    const searchStatus = options.searchStatus ?? {
        provider: "searxng",
        enabled: true,
        configured: true,
        supports_local_sources: false,
        supports_public_web: true,
        soma_tool_name: "web_search",
        direct_soma_interaction: true,
        requires_hosted_api_token: false,
        max_results: 8,
        next_actions: ["Ask Soma to search the public web through the self-hosted SearXNG provider."],
    };
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

    await page.route("**/api/v1/search/status", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: searchStatus });
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

type APIEnvelope<T> = {
    ok?: boolean;
    data?: T;
    error?: string;
};

type MCPServerRecord = {
    id: string;
    name: string;
    status?: string;
    tools?: Array<{
        id?: string;
        name: string;
        description?: string;
    }>;
};

type MCPActivityRecord = {
    id: string;
    server_id?: string;
    server_name?: string;
    tool_name?: string;
    state?: string;
    message?: string;
    summary?: string;
    team_id?: string;
    agent_id?: string;
    run_id?: string;
    timestamp?: string;
};

type GroupRecord = {
    group_id: string;
};

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return {
            raw,
            body: JSON.parse(raw) as T,
        };
    } catch {
        return {
            raw,
            body: null as T | null,
        };
    }
}

function unwrapData<T>(body: APIEnvelope<T> | T): T {
    if (body && typeof body === "object" && "data" in body) {
        return (body as APIEnvelope<T>).data as T;
    }
    return body as T;
}

async function listMCPServers(page: Page): Promise<MCPServerRecord[]> {
    const response = await page.request.get("/api/v1/mcp/servers");
    const parsed = await parseJSONIfPossible<APIEnvelope<MCPServerRecord[]> | MCPServerRecord[]>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    return unwrapData<MCPServerRecord[]>(parsed.body ?? []);
}

async function ensureFilesystemMCP(page: Page): Promise<MCPServerRecord> {
    const hasReadFile = (server: MCPServerRecord) => server.name === "filesystem"
        && server.status !== "error"
        && (server.tools ?? []).some((tool) => tool.name === "read_file");

    let servers = await listMCPServers(page);
    const existing = servers.find(hasReadFile);
    if (existing) {
        return existing;
    }

    const installResponse = await page.request.post("/api/v1/mcp/library/install", {
        data: {
            name: "filesystem",
            governance_context: {
                source_surface: "mcp_connected_tools_live_e2e",
                config_scope: "user_group",
            },
        },
    });
    const installed = await parseJSONIfPossible<APIEnvelope<unknown>>(installResponse);
    if (!installResponse.ok()) {
        throw new Error(`filesystem MCP install unavailable for live correlation proof: ${installed.raw}`);
    }

    for (let attempt = 0; attempt < 10; attempt += 1) {
        servers = await listMCPServers(page);
        const server = servers.find(hasReadFile);
        if (server) {
            return server;
        }
        await page.waitForTimeout(1_000);
    }

    throw new Error("filesystem MCP server did not expose read_file after install.");
}

async function createMCPOnlyTeam(page: Page, teamID: string, agentID: string) {
    const response = await page.request.post("/api/v1/teams", {
        data: {
            id: teamID,
            name: "Slice 3 MCP Correlation Lane",
            type: "action",
            description: "Temporary live-browser proof lane for MCP-backed workflow correlation.",
            inputs: [`swarm.team.${teamID}.internal.command`],
            deliveries: [`swarm.team.${teamID}.signal.result`],
            members: [
                {
                    id: agentID,
                    role: "mcp-proof-worker",
                    system_prompt: [
                        "You are the Slice 3 MCP correlation proof worker.",
                        "When asked to inspect README.md through the connected filesystem MCP capability, call the read_file tool with the smallest valid arguments.",
                        "Do not use internal filesystem tools, do not ask follow-up questions, and do not describe the tool-call JSON to the operator.",
                        "Use exactly one tool call first, then summarize that README.md was inspected through the MCP-backed lane.",
                        "Tool execution format: output only {\"tool_call\":{\"name\":\"read_file\",\"arguments\":{\"path\":\"README.md\"}}}.",
                    ].join(" "),
                    max_iterations: 4,
                    tools: ["mcp:filesystem/*"],
                },
            ],
        },
    });
    const parsed = await parseJSONIfPossible<unknown>(response);
    expect(response.status(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeLessThan(300);
}

async function createCorrelationGroup(page: Page, teamID: string) {
    const expiry = new Date(Date.now() + 60 * 60 * 1000).toISOString();
    const response = await page.request.post("/api/v1/groups", {
        data: {
            name: `Slice 3 MCP Correlation Group ${teamID}`,
            goal_statement: "Prove that a real team lane can use an MCP-backed capability and surface recent MCP activity.",
            work_mode: "read_only",
            allowed_capabilities: ["artifact.review"],
            member_user_ids: ["owner"],
            team_ids: [teamID],
            coordinator_profile: "browser-live-proof",
            approval_policy_ref: "slice-3-mcp-correlation",
            expiry,
        },
    });
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord>>(response);
    expect(response.status(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(201);
    expect(parsed.body?.data?.group_id).toBeTruthy();
    return parsed.body!.data!.group_id;
}

async function broadcastMCPAsk(page: Page, groupID: string, marker: string) {
    const response = await page.request.post(`/api/v1/groups/${encodeURIComponent(groupID)}/broadcast`, {
        data: {
            message: [
                `Slice 3 live MCP correlation marker: ${marker}.`,
                "Use the installed filesystem MCP capability to read README.md.",
                "Return a short result after the tool completes.",
            ].join(" "),
        },
    });
    const parsed = await parseJSONIfPossible<unknown>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
}

async function waitForMCPActivityForTeam(page: Page, teamID: string, agentID: string) {
    for (let attempt = 0; attempt < 30; attempt += 1) {
        const response = await page.request.get("/api/v1/mcp/activity?limit=12");
        const parsed = await parseJSONIfPossible<APIEnvelope<MCPActivityRecord[]>>(response);
        expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
        const activity = (parsed.body?.data ?? []).find((entry) => (
            entry.team_id === teamID
            && entry.agent_id === agentID
            && entry.tool_name === "read_file"
            && (entry.state === "completed" || entry.state === "success")
        ));
        if (activity) {
            return activity;
        }
        await page.waitForTimeout(2_000);
    }
    throw new Error(`Timed out waiting for MCP read_file activity from team ${teamID} / agent ${agentID}.`);
}

async function gotoWithColdStartRetry(page: Page, path: string) {
    try {
        await page.goto(path, { waitUntil: "domcontentloaded" });
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (!message.includes("net::ERR_CONNECTION_RESET") && !message.includes("net::ERR_ABORTED") && !message.includes("frame was detached")) {
            throw error;
        }
        await page.goto(path, { waitUntil: "domcontentloaded" });
    }
}

async function openConnectedTools(page: Page) {
    for (let attempt = 0; attempt < 2; attempt += 1) {
        await gotoWithColdStartRetry(page, "/dashboard");
        await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
        await page.waitForFunction(() => window.localStorage.getItem("mycelis-advanced-mode") === "true");
        await page.reload({ waitUntil: "domcontentloaded" });
        if (await page.getByTestId("nav-resources").isVisible({ timeout: 5_000 }).catch(() => false)) {
            break;
        }
    }

    await expect(page.getByTestId("nav-resources")).toBeVisible({ timeout: 10_000 });
    await gotoWithColdStartRetry(page, "/resources?tab=tools");
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 20_000 });
}

test.describe("Connected Tools MCP workflow", () => {
    test.skip(({ browserName }) => browserName !== "chromium", "Connected Tools browser workflow proof is stabilized in Chromium for MVP review.");

    test("shows active MCP usage and installs a curated server from the library", async ({ page }) => {
        await mockConnectedToolsApis(page);
        await openConnectedTools(page);

        await expect(page.getByRole("button", { name: "Connected Tools" })).toBeVisible();
        await expect(page.getByText("Connected Tools Workflow")).toBeVisible();
        await expect(page.getByText("Mycelis Search Capability")).toBeVisible();
        await expect(page.getByText("Soma search is ready")).toBeVisible();
        await expect(page.getByText("Soma direct: web_search")).toBeVisible();
        await expect(page.getByText("Public web", { exact: true })).toBeVisible();
        await expect(page.getByText("No hosted Brave token required for local_sources, local_api, or self-hosted SearXNG.")).toBeVisible();
        await expect(page.getByText("Recent MCP Activity", { exact: true })).toBeVisible();
        await expect(page.getByText("filesystem · read_file")).toBeVisible();
        await expect(page.getByText("Soma used filesystem.read_file while preparing the launch brief.").first()).toBeVisible();
        await expect(page.getByText("Team soma-launch-lane · Agent soma · Run run-launch-brief").first()).toBeVisible();
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

    test("correlates a live team MCP-backed capability with recent Connected Tools activity", async ({ page }) => {
        test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
        test.slow();
        test.setTimeout(180_000);

        await ensureFilesystemMCP(page);
        const stamp = `${Date.now()}-${Math.floor(Math.random() * 10_000)}`;
        const teamID = `slice3-mcp-${stamp}`;
        const agentID = `slice3-worker-${stamp}`;
        const marker = `slice3-${stamp}`;

        await createMCPOnlyTeam(page, teamID, agentID);
        const groupID = await createCorrelationGroup(page, teamID);
        await broadcastMCPAsk(page, groupID, marker);

        const activity = await waitForMCPActivityForTeam(page, teamID, agentID);
        expect(activity.server_id || activity.server_name).toBeTruthy();

        await openConnectedTools(page);
        await expect(page.getByText("Recent MCP Activity", { exact: true })).toBeVisible();
        await expect(page.getByText(/filesystem · read_file/).first()).toBeVisible({ timeout: 30_000 });
        await expect(page.getByText(`Team ${teamID} · Agent ${agentID}`).first()).toBeVisible();
    });

    test("shows the bootstrap-disabled empty state and sends the operator to the library", async ({ page }) => {
        await mockConnectedToolsApis(page, {
            servers: [],
            activity: [],
            searchStatus: {
                provider: "disabled",
                enabled: false,
                configured: false,
                supports_local_sources: false,
                supports_public_web: false,
                soma_tool_name: "web_search",
                direct_soma_interaction: true,
                requires_hosted_api_token: false,
                max_results: 8,
                blocker: {
                    code: "search_provider_disabled",
                    message: "Mycelis Search is disabled.",
                    next_action: "Set MYCELIS_SEARCH_PROVIDER=local_sources for governed local-source search or searxng for self-hosted web search.",
                },
            },
        });
        await openConnectedTools(page);

        await expect(page.getByText("Soma search needs configuration")).toBeVisible();
        await expect(page.getByText("Mycelis Search is disabled.")).toBeVisible();
        await expect(page.getByText("No MCP servers installed.", { exact: true })).toBeVisible({ timeout: 20_000 });
        await expect(page.getByRole("button", { name: "OPEN LIBRARY" })).toBeVisible();

        await page.getByRole("button", { name: "OPEN LIBRARY" }).click();
        await expect(page.getByText("Current Group MCP Config")).toBeVisible();
        await expect(page.getByText("Fetch", { exact: true })).toBeVisible();
    });
});
