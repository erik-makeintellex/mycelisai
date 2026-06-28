import { expect, test } from "@playwright/test";

test.describe("Web access setup user path", () => {
  test.skip(({ browserName }) => browserName !== "chromium", "Live UX review is stabilized in Chromium.");

  test("reviews the Settings to Resources path for enabling web access", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    const pageErrors: string[] = [];
    page.on("console", (message) => {
      if (message.type() === "error") consoleErrors.push(message.text());
    });
    page.on("pageerror", (error) => pageErrors.push(error.message));

    await page.goto("/settings", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();
    await testInfo.attach("settings-default", {
      body: await page.screenshot({ fullPage: true }),
      contentType: "image/png",
    });

    const capabilityTab = page.getByRole("tab", { name: /Capabilities/i });
    if ((await capabilityTab.count()) === 0) {
      const adminToggle = page.getByRole("button", { name: /Admin tools/i });
      if ((await adminToggle.count()) > 0) await adminToggle.click();
    }

    await expect(page.getByRole("tab", { name: /Capabilities/i })).toBeVisible();
    await page.getByRole("tab", { name: /Capabilities/i }).click();
    await expect(page.getByRole("link", { name: /Open web access setup/i })).toBeVisible();
    await testInfo.attach("settings-capabilities-redirect", {
      body: await page.screenshot({ fullPage: true }),
      contentType: "image/png",
    });

    await page.getByRole("link", { name: /Open web access setup/i }).click();
    await expect(page).toHaveURL(/\/resources\?tab=tools#web-access/);
    await expect(page.getByText("Capabilities", { exact: true }).first()).toBeVisible();
    await expect(page.getByRole("region", { name: /Web access setup/i })).toBeVisible();
    await expect(page.getByText(/Public web|Local-source search|Web access needs setup|Checking web access/i).first()).toBeVisible();
    await expect(page.getByRole("button", { name: /Add web capability/i })).toBeVisible();
    await expect(page.getByText(/Mycelis Search Capability/i)).toBeVisible();
    await testInfo.attach("resources-capabilities", {
      body: await page.screenshot({ fullPage: true }),
      contentType: "image/png",
    });

    await page.getByRole("button", { name: /Add web capability/i }).click();
    await expect(page.getByPlaceholder(/Search MCP servers/i)).toHaveValue("fetch");
    await expect(page.getByText(/Add MCP Server/i)).toBeVisible();
    await testInfo.attach("resources-web-library-filter", {
      body: await page.screenshot({ fullPage: true }),
      contentType: "image/png",
    });

    expect(consoleErrors, "No console errors during web-access setup path").toEqual([]);
    expect(pageErrors, "No page errors during web-access setup path").toEqual([]);
  });
});
