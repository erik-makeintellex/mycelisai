import { expect, test } from "@playwright/test";
import {
    createdTemplateOrganization,
    mockOrganizationEntryApis,
    openDashboard,
    openCreatedOrganization,
    openOrganizationSetup,
    recentOrganizationLink,
    recentOrganizationOpenButton,
} from "../support/organization-entry";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - continuity", () => {
    test("reopens a recent AI Organization and lands back in the Soma workspace", async ({ page }) => {
        test.slow();
        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
        });

        await openDashboard(page);
        await openOrganizationSetup(page);
        await expect(page.getByText("Northstar Labs")).toBeVisible();
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await expect(page.getByRole("link", { name: "Start with Soma" })).toBeVisible();
        await page.waitForFunction(() => window.localStorage.getItem("mycelis-last-organization-id") === "org-123");
    });

    test("returns to the current AI Organization after leaving the workspace", async ({ page }) => {
        test.slow();
        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
        });

        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await page.waitForFunction(() => window.localStorage.getItem("mycelis-last-organization-id") === "org-123");
        await openDashboard(page);
        await expect(page).toHaveURL(/\/dashboard$/);
        const returnToOrganizationLink = recentOrganizationLink(page, "Northstar Labs");
        await expect(returnToOrganizationLink).toBeVisible();
        await expect(returnToOrganizationLink).toHaveAttribute("href", "/organizations/org-123");

        await Promise.all([
            page.waitForURL(/\/organizations\/org-123$/),
            returnToOrganizationLink.click(),
        ]);
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
    });
});
