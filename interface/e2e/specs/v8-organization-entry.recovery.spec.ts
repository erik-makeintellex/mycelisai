import { expect, test } from "@playwright/test";
import {
    clickStartMode,
    createdTemplateOrganization,
    expectNoForbiddenCopy,
    mockOrganizationEntryApis,
    openCreatedOrganization,
    openOrganizationSetup,
    saveScreenshot,
} from "../support/organization-entry";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - recovery", () => {
    test("keeps creation available when recent organizations fail and shows retry guidance", async ({ page }, testInfo) => {
        await mockOrganizationEntryApis(page, {
            organizationsSummaryResponses: [
                {
                    status: 500,
                    body: { ok: false, error: "Recent AI Organizations are unavailable right now." },
                },
                {
                    status: 200,
                    body: {
                        ok: true,
                        data: [
                            {
                                id: "org-999",
                                name: "Atlas",
                                purpose: "Resume me later",
                                start_mode: "empty",
                                team_lead_label: "Team Lead",
                                advisor_count: 0,
                                department_count: 0,
                                specialist_count: 0,
                                status: "ready",
                            },
                        ],
                    },
                },
            ],
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);

        await expect(page.getByText("Recent AI Organizations are unavailable", { exact: true })).toBeVisible();
        await expect(page.getByText("You can still create a new AI Organization above while we retry your recent organizations.")).toBeVisible();
        await expect(page.getByRole("button", { name: "Retry recent AI Organizations" })).toBeVisible();
        await clickStartMode(page, /Start from template/i);
        await expect(page.getByRole("button", { name: /Engineering Starter/i })).toBeVisible();

        await page.getByRole("button", { name: "Retry recent AI Organizations" }).click();
        await expect(page.getByText("Atlas")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "retry-recovery.png");
    });

    test.skip("preserves organization context when a guided Soma action fails and then succeeds on retry", async ({ page }, testInfo) => {
        let actionAttempts = 0;

        await mockOrganizationEntryApis(page, {
            homeResponsesById: {
                [createdTemplateOrganization.id]: createdTemplateOrganization,
            },
            createHandler: () => ({
                status: 201,
                body: { ok: true, data: createdTemplateOrganization },
            }),
            actionHandler: (requestBody) => {
                actionAttempts += 1;
                if (actionAttempts === 1) {
                    return {
                        status: 500,
                        body: { ok: false, error: "Team Lead guidance is unavailable right now." },
                    };
                }

                return {
                    status: 200,
                    body: {
                        ok: true,
                        data: {
                            action: requestBody.action ?? "plan_next_steps",
                            request_label: "Run a quick strategy check",
                            headline: "Team Lead plan for Northstar Labs",
                            summary: "Team Lead recommends a clear next move for Northstar Labs.",
                            priority_steps: [
                                "Align the first outcome with the AI Organization purpose.",
                                "Use the first Department as the routing layer for work.",
                            ],
                            suggested_follow_ups: [
                                "Review your organization setup",
                                "Choose the first priority",
                            ],
                        },
                    },
                };
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await clickStartMode(page, /Start from template/i);
        await page.getByLabel("AI Organization name").fill(createdTemplateOrganization.name);
        await page.getByLabel("Purpose").fill(createdTemplateOrganization.purpose);
        await Promise.all([
            page.waitForResponse((response) => response.url().endsWith("/api/v1/organizations") && response.request().method() === "POST"),
            page.getByRole("button", { name: "Create AI Organization", exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdTemplateOrganization.id);
        await page.getByRole("button", { name: /Run a quick strategy check/i }).click();

        await expect(page.getByText("Soma guidance is unavailable", { exact: true })).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByText("Soma ready")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Response Style" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Continuity" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Retry Soma action" })).toBeVisible();
        await expectNoForbiddenCopy(page);

        await page.getByRole("button", { name: "Retry Soma action" }).click();

        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("Priority steps")).toBeVisible();
        await expect(page.getByText("Keep moving with")).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "guided-retry-recovery.png");
    });
});
