import { expect, test, type Page } from "@playwright/test";
import { fulfillJSON, gotoWithColdStartRetry } from "../support/workflow-output";

const sampleRecord = {
    record_id: "7d9f4c1e-2a3b-4c5d-8e9f-0a1b2c3d4e5f",
    document: {
        kind: "WorkerProfile",
        metadata: {
            id: "seeded-profile",
            version: "1",
            scope: { kind: "workspace", ref: "primary" },
            enabled: true,
        },
    },
};

function sessionBody(role: "admin" | "standard") {
    return {
        ok: true,
        data: {
            user: {
                email: role === "admin" ? "operator@example.test" : "member@example.test",
                name: role === "admin" ? "QA Operator" : "QA Member",
                role,
                provider: "local",
            },
        },
    };
}

async function mockSession(page: Page, role: "admin" | "standard") {
    await page.route("**/auth/session", async (route) => fulfillJSON(route, 200, sessionBody(role)));
}

async function mockRecordsList(page: Page, records: unknown[] = [sampleRecord]) {
    await page.route("**/api/v1/config-documents?limit=100", async (route) => fulfillJSON(route, 200, { ok: true, data: records }));
}

async function mockExport(page: Page, status: number, data: unknown = null) {
    await page.route("**/api/v1/config-documents/*/export*", async (route) =>
        fulfillJSON(route, status, {
            ok: status < 300,
            data,
            error: status >= 300 ? "raw backend stack trace should never render" : undefined,
        }),
    );
}

async function openResources(page: Page) {
    await gotoWithColdStartRetry(page, "/dashboard");
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
    await page.waitForFunction(() => window.localStorage.getItem("mycelis-advanced-mode") === "true");
    await gotoWithColdStartRetry(page, "/resources?tab=roles");
    await expect(page.getByRole("heading", { name: "Resources" })).toBeVisible({ timeout: 20_000 });
}

test.describe("Config document view (mocked)", () => {
    test.skip(({ browserName }) => browserName !== "chromium", "Config document view proof is stabilized in Chromium for MVP review.");

    test("an admin sees the saved configurations list", async ({ page }) => {
        await mockSession(page, "admin");
        await mockRecordsList(page);
        await openResources(page);

        await expect(page.getByText("Saved configurations")).toBeVisible();
        await expect(page.getByText(/WorkerProfile.*seeded-profile/)).toBeVisible();
        await expect(page.getByText(/v1.*workspace\/primary.*Active/)).toBeVisible();
        await expect(page.getByRole("button", { name: /View config/i })).toBeVisible();
    });

    test("opening a record shows the exported content as text", async ({ page }) => {
        await mockSession(page, "admin");
        await mockRecordsList(page);
        await mockExport(page, 200, { content: "apiVersion: mycelis.ai/v1\nkind: WorkerProfile", redaction_applied: false });
        await openResources(page);

        await page.getByRole("button", { name: /View config/i }).click();

        const dialog = page.getByRole("dialog", { name: "View config" });
        await expect(dialog).toBeVisible();
        await expect(dialog.getByText("apiVersion: mycelis.ai/v1")).toBeVisible();
        await expect(dialog.locator("pre")).toContainText("apiVersion: mycelis.ai/v1");
        await expect(dialog.getByText("Some values were hidden.")).toHaveCount(0);
    });

    test("shows the redaction notice when values were hidden", async ({ page }) => {
        await mockSession(page, "admin");
        await mockRecordsList(page);
        await mockExport(page, 200, { content: "spec:\n  api_key: '[redacted]'", redaction_applied: true });
        await openResources(page);

        await page.getByRole("button", { name: /View config/i }).click();

        const dialog = page.getByRole("dialog", { name: "View config" });
        await expect(dialog.getByText("Some values were hidden.")).toBeVisible();
        await expect(dialog.getByText(/re-importable/i)).toHaveCount(0);
    });

    test("the copy button copies the exported content to the clipboard", async ({ page, context, browserName }) => {
        test.skip(browserName !== "chromium", "Clipboard permission grants are Chromium-only.");
        await context.grantPermissions(["clipboard-read", "clipboard-write"]);

        await mockSession(page, "admin");
        await mockRecordsList(page);
        await mockExport(page, 200, { content: "apiVersion: mycelis.ai/v1", redaction_applied: false });
        await openResources(page);

        await page.getByRole("button", { name: /View config/i }).click();
        const dialog = page.getByRole("dialog", { name: "View config" });
        await expect(dialog.getByText("apiVersion: mycelis.ai/v1")).toBeVisible();

        await dialog.getByRole("button", { name: /^copy$/i }).click();
        await expect(dialog.getByRole("button", { name: /^copied$/i })).toBeVisible();

        const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
        expect(clipboardText).toBe("apiVersion: mycelis.ai/v1");
    });

    test("a 403 shows the normalized restricted state on the list", async ({ page }) => {
        await mockSession(page, "admin");
        await page.route("**/api/v1/config-documents?limit=100", async (route) =>
            fulfillJSON(route, 403, { ok: false, error: "raw backend stack trace should never render" }),
        );
        await openResources(page);

        await expect(page.getByText(/do not have access to saved configurations/i)).toBeVisible();
        await expect(page.getByText(/raw backend stack trace/i)).toHaveCount(0);
    });

    test("a 403 shows the normalized restricted state on the viewer", async ({ page }) => {
        await mockSession(page, "admin");
        await mockRecordsList(page);
        await mockExport(page, 403);
        await openResources(page);

        await page.getByRole("button", { name: /View config/i }).click();
        const dialog = page.getByRole("dialog", { name: "View config" });
        await expect(dialog.getByText(/do not have access to this configuration/i)).toBeVisible();
        await expect(dialog.getByText(/raw backend stack trace/i)).toHaveCount(0);
    });

    test("a 503 shows the failure state on the list", async ({ page }) => {
        await mockSession(page, "admin");
        await page.route("**/api/v1/config-documents?limit=100", async (route) =>
            fulfillJSON(route, 503, { ok: false, error: "raw backend stack trace should never render" }),
        );
        await openResources(page);

        await expect(page.getByText(/configuration storage is unavailable right now/i)).toBeVisible();
        await expect(page.getByText(/raw backend stack trace/i)).toHaveCount(0);
    });

    test("a 503 shows the failure state on the viewer", async ({ page }) => {
        await mockSession(page, "admin");
        await mockRecordsList(page);
        await mockExport(page, 503);
        await openResources(page);

        await page.getByRole("button", { name: /View config/i }).click();
        const dialog = page.getByRole("dialog", { name: "View config" });
        await expect(dialog.getByText(/unavailable right now/i)).toBeVisible();
        await expect(dialog.getByText(/raw backend stack trace/i)).toHaveCount(0);
    });

    test("a raw error body is never rendered anywhere on the page", async ({ page }) => {
        await mockSession(page, "admin");
        await page.route("**/api/v1/config-documents?limit=100", async (route) =>
            fulfillJSON(route, 500, { ok: false, error: "raw backend stack trace should never render" }),
        );
        await openResources(page);

        await expect(page.getByText(/saved configurations could not be loaded/i)).toBeVisible();
        const bodyText = await page.locator("body").innerText();
        expect(bodyText).not.toContain("raw backend stack trace");
    });

    test("a non-admin session never sees the list and never calls the list API", async ({ page }) => {
        await mockSession(page, "standard");
        let listCalled = false;
        await page.route("**/api/v1/config-documents?limit=100", async (route) => {
            listCalled = true;
            await fulfillJSON(route, 200, { ok: true, data: [sampleRecord] });
        });

        await openResources(page);
        await page.waitForTimeout(500);

        await expect(page.getByText("Saved configurations")).toHaveCount(0);
        await expect(page.getByRole("button", { name: /View config/i })).toHaveCount(0);
        expect(listCalled).toBe(false);
    });
});
