import { expect, test, type Page } from "@playwright/test";
import { clickVisibleControl } from "../support/click-visible-control";

type RouteLike = { fulfill: (options: { status: number; contentType: string; body: string }) => Promise<void> };
type APIEnvelope<T> = { ok?: boolean; data?: T; error?: string };
type MCPServerRecord = { id: string; name: string; status?: string; tools?: Array<{ name: string }> };
type MCPActivityRecord = {
    id: string;
    server_id?: string;
    server_name: string;
    tool_name: string;
    state: string;
    team_id?: string;
    agent_id?: string;
};

async function fulfillJSON(route: RouteLike, status: number, body: unknown) {
    await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return { raw, body: JSON.parse(raw) as T };
    } catch {
        return { raw, body: null as T | null };
    }
}

function unwrapData<T>(body: APIEnvelope<T> | T): T {
    return body && typeof body === "object" && "data" in body ? (body as APIEnvelope<T>).data as T : body as T;
}

async function mockDisabledConnectedTools(page: Page) {
    await page.route("**/api/v1/user/me", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: { id: "operator-1", name: "Operator" } });
    });
    await page.route("**/api/v1/services/status", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: [{ name: "core", status: "ready" }] });
    });
    await page.route("**/api/v1/mcp/servers", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: [] });
    });
    await page.route("**/api/v1/mcp/activity?limit=12", async (route) => {
        await fulfillJSON(route, 200, { ok: true, data: [] });
    });
    await page.route("**/api/v1/search/status", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: {
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
                    next_action: "Set MYCELIS_SEARCH_PROVIDER=builtin_web for built-in web search, local_sources for governed local-source search, or searxng for self-hosted web search.",
                },
            },
        });
    });
    await page.route("**/api/v1/mcp/library", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: [{
                category: "Data & Search",
                name: "Data & Search",
                servers: [{
                    name: "fetch",
                    title: "Fetch",
                    description: "Read a specific URL Soma or a team has been given and convert it for analysis",
                    tags: ["web"],
                    packages: [{ identifier: "@modelcontextprotocol/server-fetch", version: "latest" }],
                }],
            }],
        });
    });
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
    if (existing) return existing;

    const response = await page.request.post("/api/v1/mcp/library/install", {
        data: { name: "filesystem", governance_context: { source_surface: "mcp_connected_tools_live_e2e" } },
    });
    const parsed = await parseJSONIfPossible<unknown>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
    for (let attempt = 0; attempt < 10; attempt += 1) {
        servers = await listMCPServers(page);
        const server = servers.find(hasReadFile);
        if (server) return server;
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
            inputs: [`swarm.team.${teamID}.internal.command`],
            deliveries: [`swarm.team.${teamID}.signal.result`],
            members: [{
                id: agentID,
                role: "mcp-proof-worker",
                system_prompt: "Call read_file with README.md through filesystem MCP, then summarize the result.",
                max_iterations: 4,
                tools: ["mcp:filesystem/*"],
            }],
        },
    });
    const parsed = await parseJSONIfPossible<unknown>(response);
    expect(response.status(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeLessThan(300);
}

async function createCorrelationGroup(page: Page, teamID: string) {
    const response = await page.request.post("/api/v1/groups", {
        data: {
            name: `Slice 3 MCP Correlation Group ${teamID}`,
            goal_statement: "Prove a team lane can use an MCP-backed capability and surface activity.",
            work_mode: "read_only",
            allowed_capabilities: ["artifact.review"],
            member_user_ids: ["owner"],
            team_ids: [teamID],
            coordinator_profile: "browser-live-proof",
            approval_policy_ref: "slice-3-mcp-correlation",
            expiry: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
        },
    });
    const parsed = await parseJSONIfPossible<APIEnvelope<{ group_id: string }>>(response);
    expect(response.status(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBe(201);
    return parsed.body!.data!.group_id;
}

async function broadcastMCPAsk(page: Page, groupID: string, marker: string) {
    const response = await page.request.post(`/api/v1/groups/${encodeURIComponent(groupID)}/broadcast`, {
        data: { message: `Slice 3 marker ${marker}. Use filesystem MCP to read README.md and return a short result.` },
    });
    const parsed = await parseJSONIfPossible<unknown>(response);
    expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
}

async function waitForMCPActivityForTeam(page: Page, teamID: string, agentID: string) {
    for (let attempt = 0; attempt < 30; attempt += 1) {
        const response = await page.request.get("/api/v1/mcp/activity?limit=12");
        const parsed = await parseJSONIfPossible<APIEnvelope<MCPActivityRecord[]>>(response);
        const activity = (parsed.body?.data ?? []).find((entry) => entry.team_id === teamID
            && entry.agent_id === agentID
            && entry.tool_name === "read_file"
            && (entry.state === "completed" || entry.state === "success"));
        if (activity) return activity;
        await page.waitForTimeout(2_000);
    }
    throw new Error(`Timed out waiting for MCP read_file activity from team ${teamID} / agent ${agentID}.`);
}

async function openConnectedTools(page: Page) {
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
    await page.reload({ waitUntil: "domcontentloaded" });
    await clickVisibleControl(page, page.getByTestId("nav-resources"));
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 20_000 });
    await clickVisibleControl(page, page.getByText("Capabilities", { exact: true }));
}

test.describe("MCP connected tools edge states", () => {
    test("correlates a live team MCP-backed capability with recent Capabilities activity", async ({ page }) => {
        test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");
        test.slow();
        test.setTimeout(180_000);
        await ensureFilesystemMCP(page);
        const stamp = `${Date.now()}-${Math.floor(Math.random() * 10_000)}`;
        const teamID = `slice3-mcp-${stamp}`;
        const agentID = `slice3-worker-${stamp}`;

        await createMCPOnlyTeam(page, teamID, agentID);
        const groupID = await createCorrelationGroup(page, teamID);
        await broadcastMCPAsk(page, groupID, `slice3-${stamp}`);

        const activity = await waitForMCPActivityForTeam(page, teamID, agentID);
        expect(activity.server_id || activity.server_name).toBeTruthy();
        await openConnectedTools(page);
        await clickVisibleControl(page, page.getByRole("button", { name: /Servers/i }));
        await expect(page.getByText("Recent MCP Activity", { exact: true })).toBeVisible();
        await expect(page.getByText(/filesystem · read_file/).first()).toBeVisible({ timeout: 30_000 });
        await expect(page.getByText(`Team ${teamID} · Agent ${agentID}`).first()).toBeVisible();
    });

    test("shows the bootstrap-disabled empty state and sends the operator to the library", async ({ page }) => {
        await mockDisabledConnectedTools(page);
        await openConnectedTools(page);
        await expect(page.getByText("Web access needs setup")).toBeVisible();
        await expect(page.getByText("Mycelis Search is disabled.").first()).toBeVisible();
        await clickVisibleControl(page, page.getByRole("button", { name: /^Servers$/i }));
        await expect(page.getByText("No MCP servers installed.", { exact: true })).toBeVisible({ timeout: 20_000 });
        await clickVisibleControl(page, page.getByRole("button", { name: "Add connector" }).first());
        await expect(page.getByPlaceholder(/Search MCP servers/i)).toBeVisible();
        await expect(page.getByText("Fetch", { exact: true })).toBeVisible();
    });
});
