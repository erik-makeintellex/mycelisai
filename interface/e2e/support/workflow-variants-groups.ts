import { expect, type Locator, type Page } from "@playwright/test";
import {
  type APIEnvelope,
  type GroupRecord,
  gotoWithColdStartRetry,
  parseJSONIfPossible,
} from "./workflow-variants-live";

export async function reviewArchivedOutputs(
  page: Page,
  group: GroupRecord,
  artifactTitles: string[],
  contributingLeads: number,
) {
  await gotoWithColdStartRetry(
    page,
    `/groups?group_id=${encodeURIComponent(group.group_id)}`,
  );
  await expect(
    page.getByRole("heading", { name: "Manage focused collaboration lanes." }),
  ).toBeVisible();
  const groupHeading = page.getByRole("heading", { name: group.name });
  try {
    await expect(groupHeading).toBeVisible({ timeout: 5_000 });
  } catch {
    await retryUntilGroupVisible(page, group, groupHeading);
  }
  await assertOutputSummary(page, artifactTitles, contributingLeads);
  await page.getByRole("button", { name: "Archive temporary group" }).click();
  await expect(page.getByTestId("groups-notice")).toContainText(
    "Temporary group archived.",
  );
  await expect(
    page.getByText("Archived temporary group", { exact: true }),
  ).toBeVisible();
  await expect(page.getByTestId("groups-archived-readonly-note")).toContainText(
    "retained output review",
  );
  await assertOutputSummary(page, artifactTitles, contributingLeads);
  await page.reload({ waitUntil: "domcontentloaded" });
  await expect(
    page.getByText("Archived temporary group", { exact: true }),
  ).toBeVisible();
  for (const artifactTitle of artifactTitles) {
    await expect(page.getByText(artifactTitle, { exact: true })).toBeVisible();
  }
}

async function retryUntilGroupVisible(
  page: Page,
  group: GroupRecord,
  groupHeading: Locator,
) {
  for (
    let attempt = 0;
    attempt < 5 && !(await groupHeading.isVisible().catch(() => false));
    attempt += 1
  ) {
    const response = await page.request.get("/api/v1/groups");
    const parsed = await parseJSONIfPossible<APIEnvelope<GroupRecord[]>>(
      response,
    );
    expect(
      response.ok(),
      parsed.body ? JSON.stringify(parsed.body) : parsed.raw,
    ).toBeTruthy();
    expect(
      (parsed.body?.data ?? []).some(
        (candidate) => candidate.group_id === group.group_id,
      ),
    ).toBeTruthy();
    await gotoWithColdStartRetry(
      page,
      `/groups?group_id=${encodeURIComponent(group.group_id)}`,
    );
  }
  await expect(groupHeading).toBeVisible({ timeout: 30_000 });
}

async function assertOutputSummary(
  page: Page,
  artifactTitles: string[],
  contributingLeads: number,
) {
  await expect(
    page.getByText("Temporary group", { exact: true }),
  ).toBeVisible();
  await expect(page.getByTestId("groups-output-summary")).toContainText(
    `${artifactTitles.length} outputs`,
  );
  await expect(page.getByTestId("groups-output-summary")).toContainText(
    `${contributingLeads} contributing leads`,
  );
  for (const artifactTitle of artifactTitles) {
    await expect(page.getByText(artifactTitle, { exact: true })).toBeVisible();
  }
}
