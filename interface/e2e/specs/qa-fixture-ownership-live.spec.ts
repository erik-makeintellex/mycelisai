import { expect, test } from "@playwright/test";

import {
  createOrganization,
  liveAPIGet,
  openLiveWorkspace,
} from "../support/finalization-browser-package";
import {
  createQAFixtureScope,
  purgeDeliveryFixture,
} from "../support/qa-fixture-ownership";

test.describe("QA fixture ownership", () => {
  test.skip(!process.env.PLAYWRIGHT_LIVE_BACKEND, "requires a live Core backend");

  test("opens and deterministically removes an exactly scoped organization", async ({ page }) => {
    const stamp = Date.now();
    const fixture = await createQAFixtureScope(page, `fixture-ownership-${stamp}`);
    const organizationID = await createOrganization(
      page,
      `Fixture Ownership ${stamp}`,
      { fixtureScopeID: fixture.id },
    );

    await openLiveWorkspace(page, organizationID);
    await expect(page.getByText(`Fixture Ownership ${stamp}`, { exact: true }).first()).toBeVisible();

    await purgeDeliveryFixture(page, fixture, { organizationID });

    const removed = await liveAPIGet(page, `/api/v1/organizations/${organizationID}/home`);
    expect(removed.status()).toBe(404);
  });
});
