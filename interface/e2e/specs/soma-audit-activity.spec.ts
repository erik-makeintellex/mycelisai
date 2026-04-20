import { expect, test } from "@playwright/test";
import { mockApprovalsAudit } from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma audit activity", () => {
    test("shows inspect-only audit activity in Automations approvals", async ({ page }) => {
        await mockApprovalsAudit(page);

        await page.goto("/automations?tab=approvals");
        await page.waitForLoadState("domcontentloaded");
        await expect(page.getByRole("button", { name: "Approvals" })).toBeVisible({ timeout: 20_000 });
        await page.getByRole("button", { name: "Audit" }).click();

        await expect(page.getByText("Activity Log")).toBeVisible();
        await expect(page.getByText(/Inspect recent approvals, execution outcomes, capability use/i)).toBeVisible();
        await expect(page.getByText("proposal generated")).toBeVisible();
        await expect(page.getByText("approval required")).toBeVisible();
        await expect(page.getByText("Capability: write_file")).toBeVisible();
        await expect(page.getByText("Resource: workspace/logs/hello_world.py")).toBeVisible();
    });
});
