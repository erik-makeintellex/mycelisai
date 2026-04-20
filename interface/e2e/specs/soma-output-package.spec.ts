import { expect, test } from "@playwright/test";
import {
    answerEnvelope,
    lastUserMessage,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma output package separation", () => {
    test("distinguishes direct Soma answers from team-managed output packages", async ({ page }) => {
        await mockOrganizationWorkspace(page, (requestBody) => {
            const content = lastUserMessage(requestBody);
            if (/team-managed/i.test(content)) {
                return answerEnvelope("Team-managed output package ready: Marketing Delivery Team produced the reviewable campaign package.", {
                    askClass: "governed_artifact",
                    consultations: [
                        {
                            member: "council-creative",
                            summary: "Marketing Delivery Team should return reviewable creative output and a concise usage note.",
                        },
                    ],
                    artifacts: [
                        {
                            id: "campaign-brief-1",
                            type: "document",
                            title: "Marketing Delivery Team output brief",
                            content_type: "text/markdown",
                            content: "# Campaign package\n\n- Hero concept\n- Channel copy\n- Review checklist",
                        },
                    ],
                });
            }

            return answerEnvelope("Direct Soma answer: lead with the customer outcome and keep the positioning line short.");
        });

        await openOrganization(page);

        await sendWorkspaceMessage(page, "Give me a direct single-agent positioning line.");
        await expect(page.getByText("Direct Soma answer: lead with the customer outcome and keep the positioning line short.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Team-managed output package ready")).toHaveCount(0);

        await sendWorkspaceMessage(page, "Now create a team-managed output package for the marketing launch.");
        await expect(page.getByText("Team-managed output package ready: Marketing Delivery Team produced the reviewable campaign package.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Artifact result")).toBeVisible();
        await expect(page.getByText("Soma prepared 1 artifact for review: Marketing Delivery Team output brief.")).toBeVisible();
        await expect(page.getByText(/Soma consulted/i)).toBeVisible();
        await expect(page.getByTestId("mission-chat").getByText("Creative", { exact: true }).last()).toBeVisible();
        await expect(page.getByText("Marketing Delivery Team should return reviewable creative output and a concise usage note.")).toBeVisible();
        await expect(page.getByTestId("mission-chat").getByText("Marketing Delivery Team output brief").first()).toBeVisible();
    });
});
