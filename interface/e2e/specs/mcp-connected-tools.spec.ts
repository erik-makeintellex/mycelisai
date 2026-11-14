import { expect, test, type Page } from "@playwright/test";
import {
    createCorrelationGroup,
    createMCPOnlyTeam,
    ensureFilesystemMCP,
    mockConnectedToolsApis,
    type MCPActivityRecord,
} from "../support/mcp-connected-tools-fixtures";

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
    const raw = await response.text();
    expect(response.ok(), raw).toBeTruthy();
}

async function waitForMCPActivityForTeam(page: Page, teamID: string, agentID: string) {
    for (let attempt = 0; attempt < 30; attempt += 1) {
        const response = await page.request.get("/api/v1/mcp/activity?limit=12");
        const raw = await response.text();
        let body: { data?: MCPActivityRecord[] } | null = null;
        try {
            body = JSON.parse(raw) as { data?: MCPActivityRecord[] };
        } catch {
            body = null;
        }
        expect(response.ok(), body ? JSON.stringify(body) : raw).toBeTruthy();
        const activity = (body?.data ?? []).find((entry) => (
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
