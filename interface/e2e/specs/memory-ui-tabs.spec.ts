import { expect, test, type Page } from "@playwright/test";

function installErrorGuards(page: Page) {
  const consoleIssues: string[] = [];
  const pageErrors: string[] = [];
  page.on("console", (message) => {
    const text = message.text();
    if (message.type() === "error") consoleIssues.push(text);
    if (
      message.type() === "warning" &&
      /same key|unique key|hydration|validateDOMNesting/i.test(text)
    ) {
      consoleIssues.push(text);
    }
  });
  page.on("pageerror", (error) => pageErrors.push(error.message));
  return { consoleIssues, pageErrors };
}

async function expectNoHorizontalOverflow(page: Page) {
  const overflow = await page.evaluate(() => {
    const doc = document.documentElement;
    const body = document.body;
    return Math.max(doc.scrollWidth, body.scrollWidth) - window.innerWidth;
  });
  expect(overflow).toBeLessThanOrEqual(4);
}

test.describe("Memory focused tab workspace", () => {
  test.skip(({ browserName }) => browserName !== "chromium", "Chromium live UI proof");

  test("lets operators move through recent work, search, and details without squished columns", async ({ page }) => {
    const { consoleIssues, pageErrors } = installErrorGuards(page);
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));

    await page.goto("/memory", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "Memory" })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Recent Work/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Search Memory/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Details/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Warm/i })).toBeVisible();

    await page.getByRole("tab", { name: /SitReps/i }).click();
    await expect(page.getByRole("tab", { name: /SitReps/i })).toHaveAttribute("aria-selected", "true");
    await page.getByRole("tab", { name: /Artifacts/i }).click();
    await expect(page.getByRole("tab", { name: /Artifacts/i })).toHaveAttribute("aria-selected", "true");
    await page.getByRole("tab", { name: /Warm/i }).click();
    await expect(page.getByRole("tab", { name: /Warm/i })).toHaveAttribute("aria-selected", "true");

    await page.getByRole("tab", { name: /Search Memory/i }).click();
    await page.getByRole("textbox", { name: /Search semantic memory/i }).fill("Soma research");
    await expect(page.getByText(/Searching vectors|Memory search needs attention|No results found|%/i).first()).toBeVisible();
    const degradedNotice = page.getByText(/Memory search needs attention/i);
    if (await degradedNotice.isVisible().catch(() => false)) {
      await expect(page.getByText(/embedding-capable|embedding provider|vector recall/i).first()).toBeVisible();
    }

    await page.getByRole("tab", { name: /Details/i }).click();
    await expect(page.getByText(/Select a memory search result or artifact/i)).toBeVisible();
    await expectNoHorizontalOverflow(page);
    expect(pageErrors).toEqual([]);
    expect(consoleIssues).toEqual([]);
  });
});
