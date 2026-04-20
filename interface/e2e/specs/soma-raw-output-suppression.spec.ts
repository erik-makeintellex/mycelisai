import { expect, test } from "@playwright/test";
import {
    answerEnvelope,
    lastUserMessage,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma raw output suppression", () => {
    test("keeps raw tool JSON and raw transport strings out of the visible Soma transcript", async ({ page }) => {
        const rawCouncilPayload = '{"error":"consult_council requires \\"member\\" and \\"question\\"","tool":"consult_council"}';

        await mockOrganizationWorkspace(page, (requestBody) => {
            const content = lastUserMessage(requestBody);
            if (/raw tool json/i.test(content)) {
                return answerEnvelope(rawCouncilPayload);
            }
            return {
                status: 500,
                body: {
                    ok: false,
                    error: "Internal Server Error",
                },
            };
        });

        await openOrganization(page);

        await sendWorkspaceMessage(page, "Trigger the raw tool json regression.");
        await expect(page.getByText(/could not produce a readable reply/i)).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(rawCouncilPayload)).toHaveCount(0);
        await expect(page.getByText(/consult_council requires/i)).toHaveCount(0);

        await sendWorkspaceMessage(page, "Trigger the raw transport failure regression.");
        await expect(page.getByText(/Soma Chat Blocked/i)).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(/Internal Server Error/i)).toHaveCount(0);
    });
});
