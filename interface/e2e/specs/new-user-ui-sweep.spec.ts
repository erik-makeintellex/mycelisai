import { expect, test, type Page, type TestInfo } from "@playwright/test";

type RouteCheck = {
  path: string;
  name: string;
  heading: RegExp | string;
  affordances: Array<RegExp | string>;
};

const primaryRoutes: RouteCheck[] = [
  {
    path: "/dashboard?fresh=1",
    name: "soma",
    heading: /Talk to Soma/i,
    affordances: [/Quick asks/i, /Tell Soma what outcome you want/i, /Outcomes/i],
  },
  {
    path: "/groups",
    name: "groups",
    heading: /Manage focused collaboration lanes|Groups/i,
    affordances: [/Open Soma/i, /Group records/i, /Outputs|Message|Settings/i],
  },
  {
    path: "/resources",
    name: "resources",
    heading: "Resources",
    affordances: [/Output Files/i, /Capabilities/i, /Exchange/i],
  },
  {
    path: "/docs",
    name: "docs",
    heading: /Docs|Documentation|Help/i,
    affordances: [/Soma|Resources|Architecture|Search|Read/i],
  },
  {
    path: "/settings",
    name: "settings",
    heading: "Settings",
    affordances: [/Open web access setup/i, /Assistant|Theme|Connected|Access/i],
  },
];

const adminRoutes: RouteCheck[] = [
  {
    path: "/activity",
    name: "activity",
    heading: /Progress, runs, and bus review|Activity review/i,
    affordances: [/Review|Runs|Events|Open admin tools/i],
  },
  {
    path: "/runs",
    name: "runs",
    heading: /Run history|Run lists are in Admin tools/i,
    affordances: [/Workspace|Proof|completed|No runs yet|Open admin tools/i],
  },
  {
    path: "/memory",
    name: "memory",
    heading: /Memory|Memory is in Admin tools/i,
    affordances: [/Search|Recall|Open admin tools/i],
  },
  {
    path: "/system",
    name: "system",
    heading: /System|System is in Admin tools/i,
    affordances: [/Status|Health|Open admin tools/i],
  },
];

function installErrorGuards(page: Page) {
  const consoleIssues: string[] = [];
  const pageErrors: string[] = [];
  page.on("console", (message) => {
    const text = message.text();
    if (message.type() === "error") consoleIssues.push(text);
    if (message.type() === "warning" && /same key|unique key|hydration|validateDOMNesting/i.test(text)) {
      consoleIssues.push(text);
    }
  });
  page.on("pageerror", (error) => pageErrors.push(error.message));
  return { consoleIssues, pageErrors };
}

async function collectLayoutMetrics(page: Page) {
  return page.evaluate(() => {
    const doc = document.documentElement;
    const body = document.body;
    const widthOverflow = Math.max(doc.scrollWidth, body.scrollWidth) - window.innerWidth;
    return {
      widthOverflow,
      documentHeight: doc.scrollHeight,
      viewportHeight: window.innerHeight,
      scrollRatio: Number((doc.scrollHeight / window.innerHeight).toFixed(2)),
      activeElement: document.activeElement?.tagName ?? null,
    };
  });
}

async function attachReviewArtifacts(page: Page, testInfo: TestInfo, routeName: string) {
  await testInfo.attach(`${routeName}-screenshot`, {
    body: await page.screenshot({ fullPage: true }),
    contentType: "image/png",
  });
  await testInfo.attach(`${routeName}-layout.json`, {
    body: JSON.stringify(await collectLayoutMetrics(page), null, 2),
    contentType: "application/json",
  });
}

async function expectRoute(page: Page, route: RouteCheck, testInfo: TestInfo) {
  await page.goto(route.path, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("heading", { name: route.heading }).first()).toBeVisible({ timeout: 20_000 });
  for (const affordance of route.affordances) {
    await expect(page.getByText(affordance).first()).toBeVisible({ timeout: 15_000 });
  }
  const metrics = await collectLayoutMetrics(page);
  expect(metrics.widthOverflow, `${route.name} should not create horizontal document overflow`).toBeLessThanOrEqual(4);
  await attachReviewArtifacts(page, testInfo, route.name);
}

test.describe("New user UI sweep", () => {
  test.skip(({ browserName }) => browserName !== "chromium", "The broad UX sweep is stabilized in Chromium.");

  test("primary routes present obvious next actions without page errors", async ({ page }, testInfo) => {
    const { consoleIssues, pageErrors } = installErrorGuards(page);
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "false"));

    for (const route of primaryRoutes) {
      await expectRoute(page, route, testInfo);
    }

    expect(pageErrors).toEqual([]);
    expect(consoleIssues).toEqual([]);
  });

  test("admin routes are understandable from the default gate and after enabling Admin tools", async ({ page }, testInfo) => {
    const { consoleIssues, pageErrors } = installErrorGuards(page);
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "false"));

    for (const route of adminRoutes) {
      await page.goto(route.path, { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("link", { name: /Open admin tools/i })).toBeVisible({ timeout: 15_000 });
      await attachReviewArtifacts(page, testInfo, `${route.name}-gate`);
    }

    await page.getByRole("link", { name: /Open admin tools/i }).click();
    await page.waitForFunction(() => window.localStorage.getItem("mycelis-advanced-mode") === "true");

    for (const route of adminRoutes) {
      await expectRoute(page, route, testInfo);
    }

    expect(pageErrors).toEqual([]);
    expect(consoleIssues).toEqual([]);
  });
});
