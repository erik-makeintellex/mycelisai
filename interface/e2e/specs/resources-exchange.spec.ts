import { expect, test, type Page } from "@playwright/test";
import { clickVisibleControl } from "../support/click-visible-control";
import { fulfillJSON, gotoWithColdStartRetry } from "../support/workflow-output";

async function openResourcesTab(page: Page, tab: string) {
    await gotoWithColdStartRetry(page, "/dashboard");
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
    await page.waitForFunction(() => window.localStorage.getItem("mycelis-advanced-mode") === "true");
    await gotoWithColdStartRetry(page, `/resources?tab=${tab}`);
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 20_000 });
}

async function mockExchangeApis(page: Page) {
    await page.route("**/api/v1/exchange/channels", async (route) =>
        fulfillJSON(route, 200, [
            {
                id: "c1",
                name: "research.results",
                type: "output",
                schema_id: "ToolResult",
                visibility: "advanced",
                sensitivity_class: "team_scoped",
                owner: "soma",
            },
        ]),
    );
    await page.route("**/api/v1/exchange/threads?limit=12", async (route) =>
        fulfillJSON(route, 200, [
            {
                id: "t1",
                title: "Launch research handoff",
                thread_type: "review_thread",
                status: "active",
                channel_name: "research.results",
                participants: ["soma", "team_lead"],
                allowed_reviewers: ["operator"],
            },
        ]),
    );
    await page.route("**/api/v1/exchange/items?limit=12", async (route) =>
        fulfillJSON(route, 200, [
            {
                id: "i1",
                summary: "Research team handed launch evidence to marketing.",
                schema_id: "ToolResult",
                channel_name: "research.results",
                created_by: "team:research",
                created_at: "2026-07-14T12:00:00Z",
                sensitivity_class: "team_scoped",
                trust_class: "bounded_internal",
                capability_id: "research",
                review_required: true,
            },
        ]),
    );
}

test.describe("Resources exchange workflow", () => {
    test.skip(({ browserName }) => browserName !== "chromium", "Resources browser workflow proof is stabilized in Chromium for MVP review.");

    test("shows exchange handoffs before advanced lanes", async ({ page }) => {
        await mockExchangeApis(page);
        await openResourcesTab(page, "exchange");

        await expect(page.getByText("Team handoffs")).toBeVisible();
        await expect(page.getByText("Research team handed launch evidence to marketing.")).toBeVisible();
        await expect(page.getByRole("button", { name: /Work threads/i })).toBeVisible();
        await clickVisibleControl(page, page.getByRole("button", { name: /Work threads/i }));
        await expect(page.getByText("Launch research handoff")).toBeVisible();
        await clickVisibleControl(page, page.getByRole("button", { name: /Source lanes/i }));
        await expect(page.getByText("research.results")).toBeVisible();
    });
});
