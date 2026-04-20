import { expect, test } from "@playwright/test";
import {
    clickStartMode,
    expectNoForbiddenCopy,
    mockOrganizationEntryApis,
    openCreatedOrganization,
    openOrganizationSetup,
    saveScreenshot,
    createdTemplateOrganization,
} from "../support/organization-entry";

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - template creation", () => {
    test("creates an AI Organization from a template and starts a guided Soma workflow", async ({ page }, testInfo) => {
        test.slow();
        let capturedRequestBody: Record<string, unknown> | null = null;
        let capturedActionBody: Record<string, unknown> | null = null;

        await mockOrganizationEntryApis(page, {
            homeResponsesById: {
                [createdTemplateOrganization.id]: createdTemplateOrganization,
            },
            createHandler: (requestBody) => {
                capturedRequestBody = requestBody;
                return {
                    status: 201,
                    body: { ok: true, data: createdTemplateOrganization },
                };
            },
            actionHandler: (requestBody) => {
                capturedActionBody = requestBody;
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
        await clickStartMode(page, /Start from template/i);
        await page.getByLabel("AI Organization name").fill(createdTemplateOrganization.name);
        await page.getByLabel("Purpose").fill(createdTemplateOrganization.purpose);
        await Promise.all([
            page.waitForResponse((response) => response.url().endsWith("/api/v1/organizations") && response.request().method() === "POST"),
            page.getByRole("button", { name: "Create AI Organization", exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdTemplateOrganization.id);
        expect(capturedRequestBody).toEqual({
            name: createdTemplateOrganization.name,
            purpose: createdTemplateOrganization.purpose,
            start_mode: "template",
            template_id: createdTemplateOrganization.template_id,
        });

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: createdTemplateOrganization.name, exact: true })).toBeVisible();
        await expect(page.getByText("Soma ready")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await expect(page.getByRole("link", { name: "Start with Soma" })).toBeVisible();
        await expect(page.getByPlaceholder("Tell Soma what you want to plan, review, create, or execute")).toBeFocused();
        await expect(page.getByText("Started from", { exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Automations" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "What the Organization Is Retaining" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Response Style" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Continuity" })).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "template-guided-workflow.png");
    });
});
