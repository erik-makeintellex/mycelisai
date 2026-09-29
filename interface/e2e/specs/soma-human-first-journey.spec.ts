import { expect, test, type Page } from "@playwright/test";
import {
  expectNoHorizontalOverflow,
  mockTeamsWorkspace,
} from "../support/finalization-proof";

const focusedTeamID = "active-demo-team";
const runtimeVocabulary = /\b(?:NATS|MCP|execution contract|event chain|run[_ -]?id)\b/i;
const primaryNavigation = [
  { testID: "nav-dashboard", label: "Soma" },
  { testID: "nav-groups", label: "Work" },
  { testID: "nav-resources", label: "Resources" },
  { testID: "nav-docs", label: "Help" },
] as const;

type ViewportTarget = {
  name: string;
  width: number;
  height: number;
  desktop: boolean;
};

const viewportTargets: ViewportTarget[] = [
  { name: "desktop", width: 1366, height: 768, desktop: true },
  { name: "compact", width: 390, height: 844, desktop: false },
];

async function installHumanFirstFixture(page: Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem("mycelis-advanced-mode", "false");

    const staleChatKeys: string[] = [];
    for (let index = 0; index < window.localStorage.length; index += 1) {
      const key = window.localStorage.key(index);
      if (key?.startsWith("mycelis-workspace-chat")) staleChatKeys.push(key);
    }
    staleChatKeys.forEach((key) => window.localStorage.removeItem(key));
  });

  await mockTeamsWorkspace(page);
  await page.route("**/api/v1/workspace/files/view?path=**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "text/html",
      body: "<!doctype html><html><body><h1>Playable Coin Runner</h1></body></html>",
    });
  });
}

async function expectHumanFirstInitialState(page: Page, target: ViewportTarget) {
  const surface = page.getByTestId("soma-operating-surface");
  const frame = page.getByTestId("soma-workspace-frame");
  const composer = page.getByTestId("central-soma-chat-frame").getByRole("textbox");

  await expect(surface).toBeVisible();
  await expect(frame).toBeVisible();
  await expect(composer).toHaveCount(1);
  await expect(composer).toBeInViewport();
  expect(await page.evaluate(() => window.scrollY)).toBe(0);
  await expect(page.getByRole("dialog")).toHaveCount(0);

  const visibleSurfaceText = await surface.evaluate((node) => (node as HTMLElement).innerText);
  expect(visibleSurfaceText).not.toMatch(runtimeVocabulary);

  for (const item of primaryNavigation) {
    const link = page.getByTestId(item.testID);
    await expect(link).toHaveAttribute("title", item.label);
    if (target.desktop) {
      await expect(link.getByText(item.label, { exact: true })).toBeVisible();
    }
  }

  return { frame, surface };
}

async function openAndMeasureOutputReview(page: Page) {
  const toggle = page.getByRole("button", { name: "Open Outcome Vault" });
  await expect(toggle).toBeVisible();
  await expect(toggle).toHaveAttribute("aria-expanded", "false");
  // E2E-T1: the toggle is interactive (visible, hit-testable, not
  // intercepted -- verified with document.elementFromPoint) the moment the
  // fixture navigates, but a click that lands before this client component
  // finishes hydrating is a no-op: nothing replays it once React attaches
  // its listener. Retry the click instead of trusting a single dispatch.
  await expect
    .poll(
      async () => {
        if ((await toggle.getAttribute("aria-expanded")) === "true") return true;
        await toggle.click();
        return false;
      },
      { timeout: 10_000 },
    )
    .toBe(true);

  const outputSurface = page.getByTestId("soma-outcome-vault-overlay").getByRole("dialog", { name: "Outcome Vault" });
  await expect(outputSurface).toHaveAttribute("role", "dialog");
  await expect(outputSurface).toHaveAttribute("aria-modal", "true");
  await expect(outputSurface).toBeInViewport();
  await expect(page.getByRole("dialog")).toHaveCount(1);
  await expect(outputSurface.getByText("Coin Runner game", { exact: true }).first()).toBeVisible();
  await expect(outputSurface.getByRole("heading", { name: "Saved results" })).toBeVisible();

  const bounds = await outputSurface.boundingBox();
  expect(bounds).not.toBeNull();

  const overflow = await outputSurface.evaluate((node) => ({
    clientWidth: node.clientWidth,
    scrollWidth: node.scrollWidth,
  }));
  expect(overflow.scrollWidth).toBeLessThanOrEqual(overflow.clientWidth + 1);
  await expectNoHorizontalOverflow(page);

  return bounds!;
}

test.describe("Soma human-first journey gate", () => {
  for (const target of viewportTargets) {
    test(`${target.name} keeps Soma primary and gives output review usable space`, async ({ page }) => {
      await page.setViewportSize({ width: target.width, height: target.height });
      await installHumanFirstFixture(page);
      await page.goto(`/dashboard?fresh=1&team_id=${focusedTeamID}`, { waitUntil: "domcontentloaded" });

      const { frame } = await expectHumanFirstInitialState(page, target);
      const frameBefore = await frame.boundingBox();
      expect(frameBefore).not.toBeNull();

      const outputBounds = await openAndMeasureOutputReview(page);
      const frameAfter = await frame.boundingBox();
      const actualViewport = await page.evaluate(() => ({ width: window.innerWidth, height: window.innerHeight }));
      expect(frameAfter).not.toBeNull();
      expect(Math.abs(frameAfter!.width - frameBefore!.width)).toBeLessThanOrEqual(1);

      if (target.desktop) {
        expect(outputBounds.width).toBeGreaterThanOrEqual(360);
        expect(outputBounds.x + outputBounds.width).toBeLessThanOrEqual(actualViewport.width);
      } else {
        expect(outputBounds.width / actualViewport.width).toBeGreaterThanOrEqual(0.75);
        expect(outputBounds.x).toBeGreaterThanOrEqual(0);
        expect(outputBounds.y).toBeGreaterThanOrEqual(0);
        expect(outputBounds.x + outputBounds.width).toBeLessThanOrEqual(actualViewport.width);
        expect(outputBounds.y + outputBounds.height).toBeLessThanOrEqual(actualViewport.height);
      }

      await page.getByRole("button", { name: "Close Outcome Vault", exact: true }).click();
      await expect(page.getByRole("dialog")).toHaveCount(0);
      await expect(frame.getByRole("button", { name: "Play game Coin Runner game in Mycelis" })).toBeVisible();
      await frame.getByRole("button", { name: "Play game Coin Runner game in Mycelis" }).click();
      await expect(page).toHaveURL(/\/outputs\/view\?/);
      await expect(page.frameLocator('iframe[title="Coin Runner game"]').getByRole("heading", { name: "Playable Coin Runner" })).toBeVisible();

      const back = page.getByRole("link", { name: "Back to Soma" });
      await expect(back).toHaveAttribute("href", `/dashboard?team_id=${focusedTeamID}`);
      await back.click();
      await expect(page).toHaveURL(new RegExp(`/dashboard\\?team_id=${focusedTeamID}$`));
      await expect(page.getByTestId("central-soma-chat-frame").getByRole("textbox")).toBeInViewport();
    });
  }
});
