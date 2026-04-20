import { expect, test } from "@playwright/test";
import {
    mockOrganizationWorkspace,
    openOrganization,
    proposalEnvelope,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma proposal mode", () => {
    test("routes mutating requests through proposal mode and keeps cancel explicit", async ({ page }) => {
        const workspace = await mockOrganizationWorkspace(page, () => proposalEnvelope());

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Create a simple python file named hello_world.py in the workspace.");

        await expect(page.getByText("PROPOSED ACTION")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma wants to")).toBeVisible();
        await expect(page.getByText("create a hello_world.py file in your workspace.")).toBeVisible();
        await expect(page.getByText("A new Python file will be saved to workspace/logs/hello_world.py after approval.")).toBeVisible();
        await expect(page.getByText("workspace/logs/hello_world.py", { exact: true })).toBeVisible();
        await expect(page.getByText("Approval optional")).toBeVisible();
        await expect(page.getByText(/RISK MEDIUM/i)).toBeVisible();
        await expect(page.getByRole("button", { name: /Show details/i })).toBeVisible();

        await page.getByRole("button", { name: /^Cancel$/i }).click();
        await expect(page.getByText(/Proposal cancelled\. No action executed\./i)).toBeVisible({ timeout: 20_000 });
        await expect.poll(() => workspace.cancelCalls()).toBe(1);
    });
});
