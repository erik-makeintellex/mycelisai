import { expect, test } from "@playwright/test";
import {
    createdTemplateOrganization,
    mockOrganizationEntryApis,
    openTeamDesignLane,
    openCreatedOrganization,
    openOrganizationSetup,
    recentOrganizationOpenButton,
} from "../support/organization-entry";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - native vs external workflow output", () => {
    test("keeps native team output and external workflow contract paths visibly separated in team design", async ({ page }) => {
        test.slow();
        const capturedActionBodies: Record<string, unknown>[] = [];

        await mockOrganizationEntryApis(page, {
            organizations: [createdTemplateOrganization],
            homeResponsesById: {
                [createdTemplateOrganization.id]: createdTemplateOrganization,
            },
            actionHandler: (requestBody) => {
                capturedActionBodies.push(requestBody);
                const requestContext = String(requestBody.request_context ?? "");

                if (/n8n/i.test(requestContext)) {
                    return {
                        status: 200,
                        body: {
                            ok: true,
                            data: {
                                action: requestBody.action ?? "plan_next_steps",
                                request_label: "Plan next steps for this organization",
                                headline: "Soma plan for Northstar Labs",
                                summary: "Soma is ready to keep the external workflow path clear for this organization.",
                                priority_steps: [
                                    "Confirm the workflow trigger and expected return shape.",
                                    "Keep the external automation separate from native team execution.",
                                ],
                                suggested_follow_ups: [
                                    "Review your organization setup",
                                    "Choose the first priority",
                                ],
                                execution_contract: {
                                    execution_mode: "external_workflow_contract",
                                    owner_label: "External workflow contract",
                                    external_target: "n8n workflow contract",
                                    summary: "Use an external workflow contract here so Mycelis can invoke it cleanly and return a normalized result without treating it like a native team.",
                                    target_outputs: [
                                        "Normalized workflow result",
                                        "Linked artifact or execution note",
                                    ],
                                },
                            },
                        },
                    };
                }

                return {
                    status: 200,
                    body: {
                        ok: true,
                        data: {
                            action: requestBody.action ?? "plan_next_steps",
                            request_label: "Plan next steps for this organization",
                            headline: "Soma plan for Northstar Labs",
                            summary: "Soma is ready to shape a native creative delivery path for this organization.",
                            priority_steps: [
                                "Align the image goal with the organization purpose.",
                                "Stand up the creative team inside the organization before generating output.",
                            ],
                            suggested_follow_ups: [
                                "Review your organization setup",
                                "Choose the first priority",
                            ],
                            execution_contract: {
                                execution_mode: "native_team",
                                owner_label: "Native Mycelis team",
                                team_name: "Creative Delivery Team",
                                summary: "Use a native creative team here so Soma can coordinate the work and return a reviewable image artifact inside the organization.",
                                target_outputs: [
                                    "Reviewable image artifact",
                                    "Short concept note",
                                ],
                            },
                        },
                    },
                };
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await recentOrganizationOpenButton(page, "Northstar Labs").click();
        await openCreatedOrganization(page, createdTemplateOrganization.id);

        await openTeamDesignLane(page);
        const prompt = page.getByLabel("Tell Soma what team or delivery lane you want to create");

        await prompt.fill("Create a creative team to generate a launch hero image.");
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("Execution path")).toBeVisible();
        await expect(page.getByText("Native Mycelis team").first()).toBeVisible();
        await expect(page.getByText("Creative Delivery Team")).toBeVisible();
        await expect(page.getByText("Reviewable image artifact", { exact: true })).toBeVisible();

        await prompt.fill("Create an n8n workflow contract for inbound leads.");
        await page.getByRole("button", { name: "Start team design" }).click();

        await expect(page.getByText("External workflow contract").first()).toBeVisible();
        await expect(page.getByText("n8n workflow contract", { exact: true })).toBeVisible();
        await expect(page.getByText("Normalized workflow result", { exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Soma for Northstar Labs" })).toBeVisible();

        expect(capturedActionBodies).toEqual([
            {
                action: "plan_next_steps",
                request_context: "Create a creative team to generate a launch hero image.",
            },
            {
                action: "plan_next_steps",
                request_context: "Create an n8n workflow contract for inbound leads.",
            },
        ]);
    });
});
