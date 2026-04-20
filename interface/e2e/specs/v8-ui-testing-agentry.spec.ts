import { expect, test } from "@playwright/test";
import {
    answerEnvelope,
    chatPlaceholder,
    lastUserMessage,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("V8 UI testing agentry product contract", () => {
    test("keeps Soma primary, preserves continuity on reload, and contains oversized markdown output", async ({ page }) => {
        const directAnswer =
            "Workspace V8 keeps Soma at the center of the AI Organization while recent activity, retained knowledge, and quick checks explain what changed and why.";
        const hugeTable = [
            "| C1 | C2 | C3 | C4 | C5 | C6 | C7 | C8 | C9 | C10 | C11 | C12 | C13 | C14 | C15 | C16 | C17 | C18 | C19 | C20 |",
            "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |",
            "| v1 | v2 | v3 | v4 | v5 | v6 | v7 | v8 | v9 | v10 | v11 | v12 | v13 | v14 | v15 | v16 | v17 | v18 | v19 | v20 |",
        ].join("\n");

        await mockOrganizationWorkspace(page, (requestBody) => {
            const content = lastUserMessage(requestBody);
            if (/huge markdown table/i.test(content)) {
                return answerEnvelope(hugeTable);
            }
            return answerEnvelope(directAnswer);
        });

        await openOrganization(page);
        await expect(page.getByRole("heading", { name: "Recent Activity" })).toBeVisible();
        await expect(page.getByText("Choose a starter prompt")).toBeVisible();
        await page.getByRole("button", { name: "Plan the next move" }).click();
        await expect(page.getByPlaceholder(chatPlaceholder)).toHaveValue("Plan the next move");

        await sendWorkspaceMessage(page, "Summarize the current Workspace V8 design objectives.");
        await expect(page.getByText(directAnswer)).toBeVisible({ timeout: 20_000 });

        await page.reload({ waitUntil: "domcontentloaded" });
        await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
        await expect(page.getByText("Summarize the current Workspace V8 design objectives.")).toBeVisible();
        await expect(page.getByText(directAnswer)).toBeVisible();

        await sendWorkspaceMessage(page, "Generate a huge markdown table with 20 columns.");
        const table = page.locator('[data-testid="mission-chat"] table').last();
        await expect(table).toBeVisible({ timeout: 20_000 });
        await expect
            .poll(async () => table.evaluate((node) => node.parentElement?.className ?? ""))
            .toContain("overflow-x-auto");

        const body = await page.content();
        expect(body).not.toContain("bg-white");
    });
});
