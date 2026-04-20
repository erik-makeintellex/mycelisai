import { expect, test } from "@playwright/test";
import {
    createdTemplateOrganization,
    mockOrganizationEntryApis,
    openCreatedOrganization,
    openDashboard,
    openOrganizationSetup,
    recentOrganizationLink,
    recentOrganizationOpenButton,
} from "../support/organization-entry";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - workspace persistence", () => {
    test("keeps the Soma draft and last guidance visible after leaving and returning to the workspace", async ({ page }) => {
        test.slow();
        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
            homeResponsesById: {
                [createdTemplateOrganization.id]: createdTemplateOrganization,
            },
        });

        await openDashboard(page);
        await openOrganizationSetup(page);
        await expect(recentOrganizationOpenButton(page, "Northstar Labs")).toBeVisible();
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await expect(page.getByRole("button", { name: "Create teams with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        const prompt = page.getByLabel("Tell Soma what team or delivery lane you want to create");
        await prompt.fill("Help me choose the first priority for this launch.");
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("You asked Soma to help with")).toBeVisible();

        await openDashboard(page);
        await expect(page).toHaveURL(/\/dashboard$/);
        const returnToOrganizationLink = recentOrganizationLink(page, "Northstar Labs");
        await expect(returnToOrganizationLink).toBeVisible();
        await expect(returnToOrganizationLink).toHaveAttribute("href", "/organizations/org-123");

        await Promise.all([
            page.waitForURL(/\/organizations\/org-123$/),
            returnToOrganizationLink.click(),
        ]);

        await expect(page).toHaveURL(/\/organizations\/org-123$/);
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await expect(page.getByRole("button", { name: "Create teams with Soma" })).toBeVisible();
        await page.getByRole("button", { name: "Create teams with Soma" }).click();
        await expect(page.getByLabel("Tell Soma what team or delivery lane you want to create")).toHaveValue("Help me choose the first priority for this launch.");
        await expect(page.getByText("Soma plan for Northstar Labs")).toBeVisible();
        await expect(page.getByText("You asked Soma to help with")).toBeVisible();
    });
});
