import { expect, test } from "@playwright/test";
import {
    answerEnvelope,
    lastUserMessage,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma ask-class cues", () => {
    test("shows visible ask-class cues for artifact and specialist answers", async ({ page }) => {
        await mockOrganizationWorkspace(page, (requestBody) => {
            const content = lastUserMessage(requestBody);
            if (/artifact/i.test(content)) {
                return answerEnvelope("I prepared a brief for review.", {
                    askClass: "governed_artifact",
                    artifacts: [
                        {
                            id: "doc-qa-1",
                            type: "document",
                            title: "Creative Brief",
                            content_type: "text/markdown",
                            content: "# Brief",
                        },
                    ],
                });
            }

            return answerEnvelope("The architect reviewed the tradeoffs and recommends the safer route.", {
                askClass: "specialist_consultation",
                consultations: [
                    {
                        member: "council-architect",
                        summary: "Recommend the safer route.",
                    },
                ],
            });
        });

        await openOrganization(page);

        await sendWorkspaceMessage(page, "Create an artifact for this launch.");
        await expect(page.getByText("Artifact result")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma prepared 1 artifact for review: Creative Brief.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByTestId("mission-chat").getByText("Creative Brief").first()).toBeVisible();

        await sendWorkspaceMessage(page, "Get specialist advice on the architecture tradeoffs.");
        await expect(page.getByText("Specialist support")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma checked with Architect while shaping this answer: Recommend the safer route.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(/Soma consulted/i)).toBeVisible();
        await expect(page.getByTestId("mission-chat").getByText("Architect", { exact: true }).last()).toBeVisible();
    });
});
