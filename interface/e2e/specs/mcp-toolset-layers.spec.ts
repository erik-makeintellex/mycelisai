import { expect, test, type Page } from "@playwright/test";
import { clickVisibleControl } from "../support/click-visible-control";

type RouteLike = {
    fulfill: (options: { status: number; contentType: string; body: string }) => Promise<void>;
};

async function fulfillJSON(route: RouteLike, body: unknown) {
    await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

async function mockSettingsApis(page: Page) {
    const toolSets: Array<Record<string, unknown>> = [
        {
            id: "set-workspace",
            name: "workspace",
            description: "Shared workspace file tools.",
            tool_refs: ["mcp:filesystem/*"],
            scope_kind: "all",
        },
        {
            id: "set-research",
            name: "research",
            tool_refs: ["mcp:fetch/fetch"],
            scope_kind: "group",
            scope_ref: "market-research",
        },
    ];
    let savedPayload: unknown = null;

    await page.route("**/api/v1/user/me", async (route) => {
        await fulfillJSON(route, { ok: true, data: { id: "operator", email: "operator@example.test" } });
    });
    await page.route("**/api/v1/services/status", async (route) => {
        await fulfillJSON(route, { ok: true, data: [{ name: "core", status: "ready" }] });
    });
    await page.route("**/api/v1/mcp/servers", async (route) => {
        await fulfillJSON(route, {
            ok: true,
            data: [{
                id: "mcp-filesystem",
                name: "filesystem",
                transport: "stdio",
                status: "connected",
                tools: [{ id: "read", name: "read_file", description: "Read files." }],
            }],
        });
    });
    await page.route("**/api/v1/mcp/activity?limit=12", async (route) => {
        await fulfillJSON(route, { ok: true, data: [] });
    });
    await page.route("**/api/v1/search/status", async (route) => {
        await fulfillJSON(route, {
            ok: true,
            data: {
                provider: "local_sources",
                enabled: true,
                configured: true,
                supports_local_sources: true,
                supports_public_web: false,
                soma_tool_name: "web_search",
                direct_soma_interaction: true,
                requires_hosted_api_token: false,
                max_results: 8,
            },
        });
    });
    await page.route("**/api/v1/capabilities", async (route) => {
        await fulfillJSON(route, {
            ok: true,
            data: [{
                id: "workspace.files",
                name: "Workspace Files",
                source: "mcp",
                category: "files",
                risk: "medium",
                approval: "required",
                outputs: ["files"],
                availability_status: "available",
            }],
        });
    });
    await page.route("**/api/v1/mcp/toolsets", async (route) => {
        if (route.request().method() === "POST") {
            savedPayload = route.request().postDataJSON();
            toolSets.push({ id: "set-host", ...(savedPayload as Record<string, unknown>) });
            await fulfillJSON(route, { ok: true, data: toolSets.at(-1) });
            return;
        }
        await fulfillJSON(route, { ok: true, data: toolSets });
    });

    return () => savedPayload;
}

test.describe("MCP access layers settings", () => {
    test.skip(({ browserName }) => browserName !== "chromium", "Visible settings proof is stabilized in Chromium.");

    test("creates a host-targeted MCP access layer from Resources", async ({ page }) => {
        const getSavedPayload = await mockSettingsApis(page);

        await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
        await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
        await page.goto("/resources?tab=tools", { waitUntil: "domcontentloaded" });

        await expect(page.getByText("MCP access layers", { exact: true })).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Shared workspace file tools.")).toBeVisible();
        await expect(page.getByText("Group: market-research")).toBeVisible();

        await page.getByLabel("Name").fill("deploy");
        await clickVisibleControl(page, page.getByRole("button", { name: "Host", exact: true }));
        await page.getByLabel("Target Host id").fill("edge-node-2");
        await page.getByLabel("Tool refs").fill("mcp:ssh/*\ntoolset:workspace");
        await page.getByLabel("Description").fill("Deployment host tools");
        await clickVisibleControl(page, page.getByRole("button", { name: "Save layer" }));

        await expect(page.getByText("Access layer saved and refreshed.")).toBeVisible();
        await expect.poll(getSavedPayload).toEqual({
            name: "deploy",
            description: "Deployment host tools",
            tool_refs: ["mcp:ssh/*", "toolset:workspace"],
            scope_kind: "host",
            scope_ref: "edge-node-2",
        });
        await expect(page.getByText("Host: edge-node-2")).toBeVisible();
    });
});
