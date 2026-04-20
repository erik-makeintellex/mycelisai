import { expect, test } from "@playwright/test";
import { installWorkflowOutputShell, openOrganization, sendWorkspaceMessage } from "../support/workflow-output";

test.describe("Workflow output direct answer", () => {
    test("renders the direct answer path without team packaging", async ({ page }) => {
        await installWorkflowOutputShell(page);

        await page.route("**/api/v1/chat", async (route) => {
            const requestBody = route.request().postDataJSON() as {
                messages?: Array<{ content?: string }>;
            };
            const content = requestBody?.messages?.[requestBody.messages.length - 1]?.content ?? "";

            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: {
                        meta: { source_node: "admin", timestamp: "2026-04-15T20:00:00Z" },
                        signal_type: "chat.reply",
                        trust_score: 0.93,
                        template_id: "chat-to-answer",
                        mode: "answer",
                        payload: {
                            text: content.includes("shortest practical recommendation")
                                ? "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint, then run the Windows browser validation flow against the retained output and continuity checks."
                                : "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint.",
                            tools_used: [],
                            consultations: [],
                            artifacts: [],
                        },
                    },
                }),
            });
        });

        await openOrganization(page);
        await sendWorkspaceMessage(
            page,
            "Give me the shortest practical recommendation for how to validate this Windows self-hosted release lane.",
        );

        await expect(
            page.getByText(
                "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint, then run the Windows browser validation flow against the retained output and continuity checks.",
            ),
        ).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
    });
});
