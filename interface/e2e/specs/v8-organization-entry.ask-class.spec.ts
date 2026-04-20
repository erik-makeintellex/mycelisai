import { expect, test } from "@playwright/test";
import {
    createdTemplateOrganization,
    expectNoForbiddenCopy,
    mockOrganizationEntryApis,
    openCreatedOrganization,
    openOrganizationSetup,
    recentOrganizationOpenButton,
} from "../support/organization-entry";

function answerEnvelope(
    text: string,
    options?: {
        askClass?: string;
        consultations?: Array<{ member: string; summary: string }>;
        artifacts?: Array<Record<string, unknown>>;
    },
) {
    return {
        status: 200,
        body: {
            ok: true,
            data: {
                meta: { source_node: "admin", timestamp: "2026-03-19T18:00:00Z" },
                signal_type: "chat_response",
                trust_score: 0.9,
                template_id: "chat-to-answer",
                mode: "answer",
                payload: {
                    text,
                    ask_class: options?.askClass,
                    consultations: options?.consultations ?? [],
                    tools_used: [],
                    artifacts: options?.artifacts ?? [],
                },
            },
        },
    };
}

function lastUserMessage(requestBody: { messages?: Array<{ content?: string }> }): string {
    const messages = Array.isArray(requestBody.messages) ? requestBody.messages : [];
    return messages[messages.length - 1]?.content ?? "";
}

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - ask-class output cues", () => {
    test("keeps ask-class output cues visible inside the organization workspace chat", async ({ page }) => {
        test.slow();
        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
            homeResponsesById: {
                [createdTemplateOrganization.id]: createdTemplateOrganization,
            },
            chatHandler: (requestBody) => {
                const content = lastUserMessage(requestBody as { messages?: Array<{ content?: string }> });
                if (/artifact/i.test(content)) {
                    return answerEnvelope("I prepared a launch brief for review inside Northstar Labs.", {
                        askClass: "governed_artifact",
                        artifacts: [
                            {
                                id: "artifact-launch-brief",
                                type: "document",
                                title: "Launch brief",
                                content_type: "text/markdown",
                                content: "# Launch brief",
                            },
                        ],
                    });
                }

                return answerEnvelope("The architect reviewed the tradeoffs and recommends the safer rollout path.", {
                    askClass: "specialist_consultation",
                    consultations: [
                        {
                            member: "council-architect",
                            summary: "Recommend the safer rollout path.",
                        },
                    ],
                });
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();
        await expect(page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i)).toBeVisible();

        const input = page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i);
        await input.fill("Create an artifact brief for this launch.");
        await input.press("Enter");
        await expect(page.getByTestId("mission-chat").getByText("Artifact result")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma prepared 1 artifact for review: Launch brief.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByTestId("mission-chat").getByText("Launch brief").first()).toBeVisible();
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();

        await input.fill("Get specialist advice on the architecture tradeoffs.");
        await input.press("Enter");
        await expect(page.getByTestId("mission-chat").getByText("Specialist support")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText("Soma checked with Architect while shaping this answer: Recommend the safer rollout path.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(/Soma consulted/i)).toBeVisible();
        await expect(page.getByTestId("mission-chat").getByText("Architect", { exact: true }).last()).toBeVisible();
        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expectNoForbiddenCopy(page);
    });
});
