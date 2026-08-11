import { expect, test, type Browser, type Page, type TestInfo } from "@playwright/test";

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
    affordances: [/Start with/i, /Tell Soma what outcome you want/i, /Outcomes/i],
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
  const networkIssues: string[] = [];
  page.on("console", (message) => {
    const text = message.text();
    if (message.type() === "error" && !/^Failed to load resource:/i.test(text)) {
      consoleIssues.push(text);
    }
    if (message.type() === "warning" && /same key|unique key|hydration|validateDOMNesting/i.test(text)) {
      consoleIssues.push(text);
    }
  });
  page.on("pageerror", (error) => pageErrors.push(error.message));
  page.on("response", (response) => {
    if (response.status() >= 500) {
      networkIssues.push(`${response.status()} ${response.request().method()} ${response.url()}`);
    }
  });
  return { consoleIssues, networkIssues, pageErrors };
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
    await expect
      .poll(async () => {
        const matches = await page.getByText(affordance).all();
        return (await Promise.all(matches.map((match) => match.isVisible()))).some(Boolean);
      }, { timeout: 15_000 })
      .toBe(true);
  }
  const metrics = await collectLayoutMetrics(page);
  expect(metrics.widthOverflow, `${route.name} should not create horizontal document overflow`).toBeLessThanOrEqual(4);
  await attachReviewArtifacts(page, testInfo, route.name);
}

async function signInFromStaleWorkUrl(browser: Browser, testInfo: TestInfo, viewport: { width: number; height: number }) {
  const context = await browser.newContext({
    baseURL: String(testInfo.project.use.baseURL),
    storageState: { cookies: [], origins: [] },
    viewport,
  });
  const page = await context.newPage();
  await page.goto("/groups", { waitUntil: "domcontentloaded" });
  await expect(page).toHaveURL(/\/login\?next=%2Fdashboard$/);
  await page.getByLabel(/Local admin username/i).fill(process.env.MYCELIS_LOCAL_ADMIN_USERNAME || "admin");
  await page
    .getByLabel(/Password or local API key/i)
    .fill(process.env.MYCELIS_LOCAL_ADMIN_PASSWORD || process.env.MYCELIS_API_KEY || "playwright-admin");
  await page.getByRole("button", { name: /Sign in as local admin/i }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByRole("heading", { name: /Talk to Soma/i })).toBeVisible();
  return { context, page };
}

async function expectNoDocumentOverflow(page: Page, label: string) {
  await expect.poll(
    () => page.evaluate(() => {
      const root = document.documentElement;
      const body = document.body;
      if (!root || !body) return Number.POSITIVE_INFINITY;
      return Math.max(root.scrollWidth, body.scrollWidth) - window.innerWidth;
    }),
    { message: `${label} should not create horizontal document overflow` },
  ).toBeLessThanOrEqual(4);
  await expect(page.locator("nextjs-portal")).not.toBeVisible();
}

async function openNav(page: Page, testId: string, path: RegExp, heading: RegExp | string) {
  await page.getByTestId(testId).click();
  await expect(page).toHaveURL(path);
  await expect(page.getByRole("heading", { name: heading }).first()).toBeVisible({ timeout: 20_000 });
  await expectNoDocumentOverflow(page, testId);
}

test.describe("New user UI sweep", () => {
  test.skip(({ browserName }) => browserName !== "chromium", "The broad UX sweep is stabilized in Chromium.");

  test("primary routes present obvious next actions without page errors", async ({ page }, testInfo) => {
    const { consoleIssues, networkIssues, pageErrors } = installErrorGuards(page);
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });
    await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "false"));

    for (const route of primaryRoutes) {
      await expectRoute(page, route, testInfo);
    }

    expect(pageErrors).toEqual([]);
    expect(consoleIssues).toEqual([]);
    expect(networkIssues).toEqual([]);
  });

  test("admin routes are understandable from the default gate and after enabling Admin tools", async ({ page }, testInfo) => {
    const { consoleIssues, networkIssues, pageErrors } = installErrorGuards(page);
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
    expect(networkIssues).toEqual([]);
  });

  for (const viewport of [
    { name: "desktop", width: 1366, height: 768 },
    { name: "compact", width: 390, height: 844 },
  ]) {
    test(`authenticated ${viewport.name} journey stays understandable through visible navigation`, async ({ browser }, testInfo) => {
      testInfo.setTimeout(90_000);
      const { context, page } = await signInFromStaleWorkUrl(browser, testInfo, viewport);
      const { consoleIssues, networkIssues, pageErrors } = installErrorGuards(page);

      await expectNoDocumentOverflow(page, `${viewport.name} Soma`);
      await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeVisible();

      await openNav(page, "nav-groups", /\/groups$/, /Manage focused collaboration lanes|Groups/i);
      await page.getByRole("link", { name: "Create group", exact: true }).click();
      await page.waitForTimeout(500);
      expect(pageErrors, `${viewport.name} Groups page errors`).toEqual([]);
      expect(consoleIssues, `${viewport.name} Groups console errors`).toEqual([]);
      expect(networkIssues, `${viewport.name} Groups network errors`).toEqual([]);
      await expect(page.getByLabel("Name")).toBeVisible();
      await expect(page.getByLabel("Goal Statement")).toBeVisible();
      await expectNoDocumentOverflow(page, `${viewport.name} Groups create`);

      await openNav(page, "nav-resources", /\/resources$/, "Resources");
      for (const resourceName of ["Capabilities", "Exchange", "Output Files"]) {
        await page.getByRole("tab", { name: new RegExp(resourceName, "i") }).first().click();
        await expectNoDocumentOverflow(page, `${viewport.name} Resources ${resourceName}`);
      }

      await page.evaluate(() => window.localStorage.setItem("mycelis-advanced-mode", "true"));
      await page.reload({ waitUntil: "domcontentloaded" });
      await openNav(page, "nav-memory", /\/memory$/, /Memory/i);
      await openNav(page, "nav-docs", /\/docs$/, /Docs|Documentation|Help/i);
      await openNav(page, "nav-settings", /\/settings$/, "Settings");
      await openNav(page, "nav-dashboard", /\/dashboard$/, /Talk to Soma/i);
      await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeVisible();

      await attachReviewArtifacts(page, testInfo, `authenticated-${viewport.name}-journey`);
      expect(pageErrors).toEqual([]);
      expect(consoleIssues).toEqual([]);
      expect(networkIssues).toEqual([]);
      await context.close();
    });
  }
});
