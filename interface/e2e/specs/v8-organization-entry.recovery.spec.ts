import { expect, test } from "@playwright/test";
import { clickStartMode, expectNoForbiddenCopy, mockOrganizationEntryApis, openOrganizationSetup, saveScreenshot } from "../support/organization-entry";

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
});
