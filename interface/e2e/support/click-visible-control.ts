import { expect, type Locator, type Page } from "@playwright/test";

export async function clickVisibleControl(
    page: Page,
    locator: Locator,
    options: { timeout?: number } = {},
) {
    const timeout = options.timeout ?? 10_000;
    await expect(locator).toBeVisible({ timeout });
    await expect(locator).toBeEnabled({ timeout });
    await locator.evaluate((element) => {
        element.scrollIntoView({ block: "center", inline: "center" });
    });
    const box = await locator.boundingBox({ timeout });
    expect(box, "Expected clickable control to have a browser-visible bounding box").toBeTruthy();
    const center = {
        x: box!.x + box!.width / 2,
        y: box!.y + box!.height / 2,
    };
    const isHitTarget = await locator.evaluate((element, point) => {
        const hit = document.elementFromPoint(point.x, point.y);
        return hit === element || Boolean(hit?.closest("button, a, [role='button']") === element);
    }, center);
    expect(isHitTarget, "Expected clickable control center to resolve to the intended control").toBeTruthy();
    await page.mouse.click(center.x, center.y);
}
