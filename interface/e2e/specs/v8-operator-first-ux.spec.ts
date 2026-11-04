import { expect, test } from "@playwright/test";

test.skip(({ browserName }) => browserName !== "chromium", "Operator-First UX certification runs in Chromium.");

test.describe("V8 Operator-First UX Upgrade Certification", () => {
    test.describe.configure({ mode: "serial" });

    test("renders WorkspaceContextHeader and dynamic Soma prompt placeholder on /dashboard", async ({ page }) => {
        await page.goto("/dashboard?fresh=1", { waitUntil: "domcontentloaded" });
        const header = page.getByTestId("workspace-context-header");
        await expect(header).toBeVisible();
        await expect(header).toContainText("[Soma Executive Workspace]");
        await expect(header).toContainText("Standing Council Active");

        const promptInput = page.getByPlaceholder(/Describe the outcome you need from Soma/i);
        await expect(promptInput).toBeVisible();
    });

    test("renders Durable Files & Deliverables operator view on /resources when advancedMode is off and switches to Advanced Resources", async ({ page }) => {
        await page.route("**/api/v1/artifacts**", async (route) => {
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: "art-ux-101",
                            agent_id: "admin",
                            artifact_type: "document",
                            title: "Quarterly Launch Plan",
                            content_type: "text/markdown",
                            content: "# Quarterly Launch Plan\nReady for operator review.",
                            file_path: "workspace/deliverables/quarterly-launch-plan.md",
                            status: "approved",
                            created_at: new Date().toISOString(),
                        },
                    ],
                }),
            });
        });

        await page.addInitScript(() => {
            window.localStorage.setItem("mycelis-advanced-mode", "false");
        });
        await page.goto("/resources", { waitUntil: "domcontentloaded" });

        await expect(page.getByTestId("durable-files-operator-view")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Durable Files & Deliverables" })).toBeVisible();
        await expect(page.getByText("Quarterly Launch Plan")).toBeVisible();
        await expect(page.getByText("workspace/deliverables/quarterly-launch-plan.md")).toBeVisible();

        const previewBtn = page.getByTestId("preview-deliverable-art-ux-101");
        await expect(previewBtn).toBeVisible();
        await previewBtn.click();
        await expect(page.getByTestId("deliverable-preview-pane-art-ux-101")).toContainText("Ready for operator review.");

        const openBtn = page.getByTestId("open-deliverable-art-ux-101");
        await openBtn.click();
        await expect(page.getByTestId("durable-files-open-notice")).toContainText("workspace/deliverables/quarterly-launch-plan.md");

        const switchBtn = page.getByTestId("resources-switch-advanced-footer");
        await expect(switchBtn).toContainText("Switch to Advanced Mode to configure AI Engines & Connected MCP Tools");
        await switchBtn.click();

        await expect(page.getByRole("heading", { name: "Advanced Resources" })).toBeVisible();
    });

    test("renders System Status Summary Card on /system when advancedMode is off and opens Advanced Diagnostics on click", async ({ page }) => {
        await page.route("**/api/v1/services/status", async (route) => {
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: [
                        { name: "nats", status: "online" },
                        { name: "postgres", status: "online" },
                        { name: "cognitive", status: "online" },
                        { name: "reactive", status: "online" },
                        { name: "groups_bus", status: "online" },
                    ],
                }),
            });
        });

        await page.addInitScript(() => {
            window.localStorage.setItem("mycelis-advanced-mode", "false");
        });
        await page.goto("/system", { waitUntil: "domcontentloaded" });

        await expect(page.getByTestId("system-status-summary-card")).toBeVisible();
        await expect(page.getByRole("heading", { name: "System Status: Online (5/5 services active)" })).toBeVisible();
        await expect(page.getByText("Event Bus & Team Coordination")).toBeVisible();
        await expect(page.getByText("Workspace Memory & Storage")).toBeVisible();

        const openDiagnosticsBtn = page.getByTestId("open-advanced-diagnostics");
        await expect(openDiagnosticsBtn).toBeVisible();
        await openDiagnosticsBtn.click();

        await expect(page.getByRole("heading", { name: "System", exact: true })).toBeVisible();
        await expect(page.getByRole("button", { name: "Runtime Health" })).toBeVisible();
    });
});
