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

const createdEmptyOrganization = {
    id: "org-456",
    name: "Skylight Works",
    purpose: "Create an AI operations group from a clean starting point.",
    start_mode: "empty",
    team_lead_label: "Team Lead",
    advisor_count: 0,
    department_count: 0,
    specialist_count: 0,
    ai_engine_settings_summary: "Set up later in Advanced mode",
    response_contract_profile_id: "clear_balanced",
    response_contract_summary: "Clear & Balanced",
    memory_personality_summary: "Set up later in Advanced mode",
    status: "ready",
    departments: [],
};

test.skip(({ browserName }) => browserName !== "chromium", "Deep organization-entry workflow coverage is stabilized in Chromium for the MVP audit.");
test.describe.configure({ mode: "serial" });

test.describe("V8 AI Organization entry flow - empty start", () => {
    test("creates an empty-start AI Organization and keeps the organization frame after success", async ({ page }, testInfo) => {
        test.slow();
        let capturedRequestBody: Record<string, unknown> | null = null;

        await mockOrganizationEntryApis(page, {
            homeResponsesById: {
                [createdTemplateOrganization.id]: createdTemplateOrganization,
                [createdEmptyOrganization.id]: createdEmptyOrganization,
            },
            createHandler: (requestBody) => {
                capturedRequestBody = requestBody;
                return {
                    status: 201,
                    body: { ok: true, data: createdEmptyOrganization },
                };
            },
        });

        await page.goto("/dashboard");
        await page.waitForLoadState("domcontentloaded");
        await openOrganizationSetup(page);
        await clickStartMode(page, /Start Empty$/i);
        await page.getByLabel("AI Organization name").fill(createdEmptyOrganization.name);
        await page.getByLabel("Purpose").fill(createdEmptyOrganization.purpose);
        await Promise.all([
            page.waitForResponse((response) => response.url().endsWith("/api/v1/organizations") && response.request().method() === "POST"),
            page.getByRole("button", { name: "Create AI Organization", exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdEmptyOrganization.id);
        expect(capturedRequestBody).toEqual({
            name: createdEmptyOrganization.name,
            purpose: createdEmptyOrganization.purpose,
            start_mode: "empty",
        });

        await expect(page.getByText("AI Organization Home")).toBeVisible();
        await expect(page.getByRole("heading", { name: createdEmptyOrganization.name, exact: true })).toBeVisible();
        await expect(page.getByText("Soma ready")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Talk with Soma" })).toBeVisible();
        await expect(page.getByRole("link", { name: "Start with Soma" })).toBeVisible();
        await expect(page.getByPlaceholder("Tell Soma what you want to plan, review, create, or execute")).toBeFocused();
        await expect(page.getByText("Started from", { exact: true })).toBeVisible();
        await expect(page.getByText("Empty", { exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Advisors" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Departments" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Automations" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "What the Organization Is Retaining" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "AI Engine Settings" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Response Style" })).toBeVisible();
        await expect(page.getByRole("heading", { name: "Memory & Continuity" })).toBeVisible();
        await expectNoForbiddenCopy(page);

        await saveScreenshot(page, testInfo, "empty-success-home.png");
    });
});
