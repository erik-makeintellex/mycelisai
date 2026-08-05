import { expect, test } from "@playwright/test";

test.describe("Canonical documentation navigation", () => {
  for (const viewport of [
    { name: "desktop", width: 1366, height: 768 },
    { name: "mobile", width: 390, height: 844 },
  ]) {
    test(`${viewport.name} exposes one product architecture authority`, async ({ page }) => {
      const consoleErrors: string[] = [];
      const pageErrors: string[] = [];
      page.on("console", (message) => {
        if (message.type() === "error") consoleErrors.push(message.text());
      });
      page.on("pageerror", (error) => pageErrors.push(error.message));

      await page.setViewportSize(viewport);
      await page.goto("/docs?doc=mycelis-canonical-prd", { waitUntil: "domcontentloaded" });

      await expect(page.getByRole("heading", { name: "Mycelis Canonical PRD" })).toBeVisible();
      await expect(page.getByText("Architecture Docs Index", { exact: true })).toHaveCount(0);
      await expect(page.getByText("Worker Library Source Map", { exact: true })).toHaveCount(0);
      await expect(page.getByTestId("docs-article-pane")).toBeVisible();

      const horizontalOverflow = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      );
      expect(horizontalOverflow).toBe(false);
      expect(consoleErrors).toEqual([]);
      expect(pageErrors).toEqual([]);
    });
  }
});
