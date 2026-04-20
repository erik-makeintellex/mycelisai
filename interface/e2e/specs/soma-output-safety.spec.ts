import { expect, test } from "@playwright/test";
import {
    answerEnvelope,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma output safety", () => {
    test("silently retries a first transient Soma failure and recovers without surfacing a blocker", async ({ page }) => {
        let attempts = 0;

        await mockOrganizationWorkspace(page, () => {
            attempts += 1;
            if (attempts === 1) {
                return {
                    status: 500,
                    body: {
                        ok: false,
                        error: "Soma chat unreachable (500)",
                    },
                };
            }
            return answerEnvelope("Recovered answer after startup wobble.");
        });

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Summarize the current Workspace V8 design objectives.");

        await expect(page.getByText("Recovered answer after startup wobble.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(/Soma Chat Blocked/i)).toHaveCount(0);
        await expect(page.getByTestId("mission-chat").getByRole("button", { name: /^Retry$/i })).toHaveCount(0);
        expect(attempts).toBe(2);
    });
});
