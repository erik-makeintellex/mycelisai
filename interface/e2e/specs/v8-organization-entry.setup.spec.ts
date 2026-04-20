import { expect, test } from "@playwright/test";
import {
    clickStartMode,
    expectNoForbiddenCopy,
    mockOrganizationEntryApis,
    openOrganizationSetup,
    saveScreenshot,
} from "../support/organization-entry";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - setup", () => {
    test("lands on the AI Organization setup flow with a dominant creation entrypoint", async ({ page }, testInfo) => {
        await mockOrganizationEntryApis(page);

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");

        await expect(page.getByRole("heading", { name: /Work directly with .* from the admin home\./i })).toBeVisible();
        await expect(page.getByText("Central Soma")).toBeVisible();
        await expect(page.getByText("The root workspace should feel like a direct conversation with Soma.")).toBeVisible();
        await expect(page.getByRole("link", { name: "Open groups workspace" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Create or open AI Organizations" })).toBeVisible();
        await page.getByRole("button", { name: "Create or open AI Organizations" }).click();
        await expect(page.getByText("AI Organization Setup", { exact: true })).toBeVisible();
        await expect(page.getByRole("button", { name: "Explore Templates" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Start Empty", exact: true })).toBeVisible();
        await expect(page.getByText("Keep the main admin home centered on Soma.")).toBeVisible();
        await expect(page.getByText("Mission Control")).toHaveCount(0);
        await expect(page.getByText("New Chat")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "dominant-entrypoint.png");
    });

    test("shows the starter template path with user-facing terminology only", async ({ page }, testInfo) => {
        await mockOrganizationEntryApis(page);

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await clickStartMode(page, /Start from template/i);

        const starterCard = page.getByRole("button", { name: /Engineering Starter/i });
        await expect(starterCard).toBeVisible();
        await expect(starterCard).toContainText("AI Organization starter");
        await expect(starterCard).toContainText("Team Lead");
        await expect(starterCard).toContainText("Advisors");
        await expect(starterCard).toContainText("Departments");
        await expect(starterCard).toContainText("Specialists");
        await expect(starterCard).toContainText("AI Engine Settings");
        await expect(starterCard).toContainText("Memory & Continuity");
        await expect(page.getByText("Hidden until Advanced mode")).toBeVisible();
        await expect(page.getByText("Learn about AI Organizations")).toBeVisible();
        await expect(page.getByText("Inception")).toHaveCount(0);
        await expect(page.getByText("Soma Kernel")).toHaveCount(0);
        await expect(page.getByText("Central Council")).toHaveCount(0);
        await expect(page.getByText("Provider Policy")).toHaveCount(0);
        await expect(page.getByText("Identity / Continuity")).toHaveCount(0);
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "template-mode.png");
    });
});
