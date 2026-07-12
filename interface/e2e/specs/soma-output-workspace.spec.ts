import { expect, test, type Page } from "@playwright/test";

const tableAsk = "Compare output shapes Soma can deliver.";
const tableReply = [
  "Soma can shape this as a compact output plan.",
  "",
  "| Output | Support | Proof |",
  "| --- | --- | --- |",
  "| Table / report | Columns, rows, source notes | Schema and assumptions |",
  "| Code / app | Openable package and launch path | Validation notes |",
  "| Media | Preview and saved artifact | Provider boundary |",
].join("\n");

async function seedSomaTableConversation(page: Page) {
  await page.addInitScript(
    ({ ask, reply }) => {
      const keys = [];
      for (let index = 0; index < window.localStorage.length; index += 1) {
        const key = window.localStorage.key(index);
        if (key?.startsWith("mycelis-workspace-chat")) keys.push(key);
      }
      keys.forEach((key) => window.localStorage.removeItem(key));
      window.localStorage.removeItem("mycelis-rail-collapsed");
      window.localStorage.setItem(
        "mycelis-workspace-chat",
        JSON.stringify([
          {
            role: "user",
            content: ask,
            timestamp: "2026-07-09T12:00:00Z",
          },
          {
            role: "council",
            content: reply,
            source_node: "admin",
            mode: "answer",
            timestamp: "2026-07-09T12:00:02Z",
            execution_summary: {
              execution: {
                status: "completed",
                shape: "tool_assisted_work",
                summary: "Soma shaped the expected output into a readable table plan.",
              },
              intent: {
                original: ask,
                resolved: "Show a table-style output plan for deliverable support.",
              },
              outputs: [
                {
                  kind: "file",
                  title: "Output shape comparison",
                  content_type: "text/csv",
                  path: "generated/proof/output-shape-comparison.csv",
                  retained: true,
                },
              ],
              proof: [{ label: "Run proof", run_id: "run-output-table", verified: true }],
            },
          },
        ]),
      );
    },
    { ask: tableAsk, reply: tableReply },
  );
}

async function visualLayoutMetrics(page: Page) {
  return page.evaluate(() => {
    const rail = document.querySelector('[data-testid="zone-a-rail"]')?.getBoundingClientRect();
    const frame = document.querySelector('[data-testid="central-soma-chat-frame"]')?.getBoundingClientRect();
    const composer = document.querySelector('[data-testid="central-soma-chat-frame"] textarea')?.getBoundingClientRect();
    return {
      railWidth: rail?.width ?? 0,
      frameWidth: frame?.width ?? 0,
      composerBottom: composer?.bottom ?? 0,
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight,
    };
  });
}

test.describe("Soma output workspace UX", () => {
  test("collapses the left rail and renders table-like Soma output as a table", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (message) => {
      if (message.type() === "error") consoleErrors.push(message.text());
    });

    await seedSomaTableConversation(page);
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });

    await expect(page.getByTestId("soma-operating-surface")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByTestId("central-soma-chat-frame")).toBeVisible();
    await expect(page.getByText(tableAsk)).toBeVisible();
    await expect(page.getByRole("table")).toBeVisible();
    await expect(page.getByRole("columnheader", { name: "Output" })).toBeVisible();
    await expect(page.getByRole("cell", { name: "Code / app" })).toBeVisible();
    await expect(page.getByText("Output plan")).toBeVisible();
    await expect(page.getByText("Table / report")).toBeVisible();

    const before = await visualLayoutMetrics(page);
    await page.getByTestId("rail-collapse-toggle").click();
    await expect(page.getByTestId("zone-a-rail")).toHaveAttribute("data-collapsed", "true");
    await expect(page.getByTestId("nav-dashboard")).toBeVisible();
    await expect.poll(async () => (await visualLayoutMetrics(page)).railWidth).toBeLessThan(before.railWidth);
    const after = await visualLayoutMetrics(page);
    expect(after.railWidth).toBeLessThan(before.railWidth);
    expect(after.frameWidth).toBeGreaterThan(before.frameWidth);

    const input = page.getByPlaceholder(/Tell Soma what you want/i);
    await expect(input).toBeVisible();
    await input.fill("Make this a retained CSV as well.");
    await expect(input).toHaveValue("Make this a retained CSV as well.");
    expect(after.composerBottom).toBeLessThanOrEqual(after.viewportHeight);

    await page.screenshot({ path: testInfo.outputPath("soma-output-table-collapsed-rail.png"), fullPage: true });
    expect(consoleErrors).toEqual([]);
  });
});
