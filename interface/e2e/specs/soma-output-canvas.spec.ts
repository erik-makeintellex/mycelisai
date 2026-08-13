import { expect, test, type Page } from "@playwright/test";
import { expectNoHorizontalOverflow, fulfillJSON } from "../support/finalization-proof";

const outputPath = "groups/delivery-team/generated/demo/index.html";
const somaReturn = "/dashboard?team_id=delivery-team&outcome_id=outcome-7#latest";

async function installOutputCanvasFixture(page: Page) {
  await page.route("**/api/v1/**", async (route) => {
    await fulfillJSON(route, 200, { ok: true, data: [] });
  });
  await page.route("**/api/v1/stream**", async (route) => {
    await route.fulfill({ status: 200, contentType: "text/event-stream", body: ": connected\n\n" });
  });
  await page.route("**/api/v1/user/me", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: { id: "operator-output", name: "Output Reviewer", email: "output@example.test" },
    });
  });
  await page.route("**/api/v1/services/status", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: [
        { name: "nats", status: "online" },
        { name: "postgres", status: "online" },
      ],
    });
  });
  await page.route("**/api/v1/workspace/files/view?path=**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "text/html",
      body: `<!doctype html><html><body><main><h1>Playable delivery</h1><button id="advance">Advance</button><p id="state">Ready</p><script>document.getElementById('advance').onclick=()=>document.getElementById('state').textContent='Played';</script></main></body></html>`,
    });
  });
}

for (const viewport of [
  { name: "desktop", width: 1366, height: 768 },
  { name: "compact", width: 390, height: 844 },
]) {
  test(`opens a retained output and returns to Soma at ${viewport.name} size`, async ({ page }, testInfo) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height });
    await installOutputCanvasFixture(page);

    const consoleErrors: string[] = [];
    const pageErrors: string[] = [];
    page.on("console", (message) => {
      if (message.type() === "error") consoleErrors.push(message.text());
    });
    page.on("pageerror", (error) => pageErrors.push(error.message));

    const source = `/api/v1/workspace/files/view?path=${encodeURIComponent(outputPath)}`;
    const params = new URLSearchParams({
      source,
      label: "Playable delivery",
      path: outputPath,
      return_to: somaReturn,
      proof: "proof-output-7",
    });
    await page.goto(`/outputs/view?${params.toString()}`);

    await expect(page.getByRole("heading", { name: "Playable delivery" })).toBeVisible();
    const canvas = page.frameLocator('iframe[title="Playable delivery"]');
    await expect(canvas.getByRole("heading", { name: "Playable delivery" })).toBeVisible();
    await canvas.getByRole("button", { name: "Advance" }).click();
    await expect(canvas.getByText("Played")).toBeVisible();
    await expectNoHorizontalOverflow(page);
    await page.screenshot({ path: testInfo.outputPath(`output-canvas-${viewport.name}.png`), fullPage: true });

    const back = page.getByRole("link", { name: "Back to Soma" });
    await expect(back).toHaveAttribute("href", somaReturn);
    await back.click();
    await expect(page).toHaveURL(new RegExp(`/dashboard\\?team_id=delivery-team&outcome_id=outcome-7#latest$`));

    expect(consoleErrors).toEqual([]);
    expect(pageErrors).toEqual([]);
  });
}
