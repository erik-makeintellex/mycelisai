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
  const toggle = page.getByTestId("soma-workbench-panel-toggle");
  await expect(toggle).toBeVisible();
  await expect(toggle).toHaveAttribute("aria-expanded", "false");
  await toggle.click();

  const outputSurface = page.getByTestId("soma-workbench-side-rail");
  await expect(outputSurface).toHaveAttribute("role", "dialog");
  await expect(outputSurface).toHaveAttribute("aria-hidden", "false");
  await expect(outputSurface).toBeInViewport();
  await expect(page.getByRole("dialog")).toHaveCount(1);
  await page.waitForTimeout(250);

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
        expect(outputBounds.width / actualViewport.width).toBeGreaterThanOrEqual(0.65);
        expect(Math.abs(actualViewport.width - outputBounds.x - outputBounds.width)).toBeLessThanOrEqual(24);
      } else {
        expect(outputBounds.x).toBeGreaterThanOrEqual(0);
        expect(outputBounds.y).toBeGreaterThanOrEqual(0);
        expect(outputBounds.x + outputBounds.width).toBeLessThanOrEqual(actualViewport.width);
        expect(outputBounds.y + outputBounds.height).toBeLessThanOrEqual(actualViewport.height);
      }
    });
  }
});
