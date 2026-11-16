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
        let capturedRefinePrompt = "";
        await page.route("**/api/v1/artifacts/art-ux-101/refine", async (route) => {
            const payload = route.request().postDataJSON() as { prompt?: string };
            capturedRefinePrompt = payload?.prompt ?? "";
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: { id: "art-ux-101", status: "refined" },
                }),
            });
        });
        await page.route("**/api/v1/artifacts**", async (route) => {
            if (route.request().url().includes("/refine")) {
                await route.fallback();
                return;
            }
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

        await page.getByTestId("refine-deliverable-art-ux-101").click();
        await page.getByTestId("refine-input-art-ux-101").fill("Add Q4 enterprise KPI table");
        await page.getByTestId("refine-submit-art-ux-101").click();
        await expect(page.getByTestId("durable-files-open-notice")).toContainText("Refinement request sent to");
        await expect(page.getByTestId("durable-files-open-notice")).toContainText("Add Q4 enterprise KPI table");
        expect(capturedRefinePrompt).toBe("Add Q4 enterprise KPI table");

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

    test("renders Trusted-Memory arbitration badges, layer filter, and contradiction warning on /memory", async ({ page }) => {
        await page.route("**/api/v1/memory/sitreps**", async (route) => {
            await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify([]) });
        });
        await page.route("**/api/v1/artifacts**", async (route) => {
            await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ ok: true, data: [] }) });
        });
        await page.route("**/api/v1/memory/resolve-conflict", async (route) => {
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    id: "mem-cand-2",
                    resolution: "keep",
                    verification_status: "verified",
                    contradicted: false,
                }),
            });
        });
        await page.route("**/api/v1/memory/promote", async (route) => {
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    id: "mem-cand-2",
                    memory_layer: "GOVERNED_DOCTRINE",
                    verification_status: "verified",
                }),
            });
        });
        await page.route("**/api/v1/memory/search**", async (route) => {
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    results: [
                        {
                            id: "mem-doctrine-1",
                            content: "Governed organization doctrine: all external deployments require release approval.",
                            score: 0.94,
                            source: "company_knowledge",
                            created_at: new Date().toISOString(),
                            memory_layer: "GOVERNED_DOCTRINE",
                            verification_status: "verified",
                        },
                        {
                            id: "mem-cand-2",
                            content: "Unverified reflection candidate suggesting auto-deploy on Fridays.",
                            score: 0.71,
                            source: "reflection_synthesis",
                            created_at: new Date().toISOString(),
                            memory_layer: "REFLECTION_MEMORY",
                            verification_status: "contradicted",
                            contradiction_halt: {
                                halted: true,
                                reason: "Contradicts higher-precedence Governed Doctrine mem-doctrine-1",
                                conflicting_id: "mem-doctrine-1",
                                conflicting_layer: "GOVERNED_DOCTRINE",
                            },
                        },
                    ],
                    arbitration: {
                        total_candidates: 2,
                        returned_count: 2,
                        verified_count: 1,
                        contradicted_count: 1,
                        dominant_layer: "GOVERNED_DOCTRINE",
                        has_contradictions: true,
                    },
                }),
            });
        });

        await page.addInitScript(() => {
            window.localStorage.setItem("mycelis-advanced-mode", "true");
        });
        await page.goto("/memory", { waitUntil: "domcontentloaded" });

        const filterGroup = page.getByRole("group", { name: "Filter memory layer" });
        await expect(filterGroup).toBeVisible();
        await filterGroup.getByRole("button", { name: "Governed Doctrine" }).click();

        const searchInput = page.getByLabel("Search semantic memory");
        await searchInput.fill("deployment approval");

        await expect(page.getByTestId("memory-contradiction-banner")).toBeVisible();
        await expect(page.getByText("Governed organization doctrine: all external deployments require release approval.")).toBeVisible();

        await page.getByText("Unverified reflection candidate suggesting auto-deploy on Fridays.").click();
        await expect(page.getByText("Contradicts higher-precedence Governed Doctrine mem-doctrine-1")).toBeVisible();

        await expect(page.getByTestId("memory-promote-doctrine-btn")).toBeVisible();
        await page.getByTestId("memory-resolve-keep-btn").click();
        await expect(page.getByTestId("memory-governance-notice")).toContainText("Conflict resolved: record verified for active recall");

        await page.getByText("Governed organization doctrine: all external deployments require release approval.").click();
        await page.getByText("Unverified reflection candidate suggesting auto-deploy on Fridays.").click();
        await page.getByTestId("memory-promote-doctrine-btn").click();
        await expect(page.getByTestId("memory-governance-notice")).toContainText("Promoted to Governed Doctrine (Verified)");
    });
});
