import { expect, type Page, type Route } from "@playwright/test";

type ServiceStatus = {
  name: string;
  status: "online" | "offline" | "degraded";
  detail?: string;
  latency_ms?: number;
};

type TeamDetailEntry = {
  id: string;
  name: string;
  role: string;
  type: "standing" | "mission";
  mission_id: string | null;
  mission_intent: string | null;
  inputs: string[];
  deliveries: string[];
  agents: Array<{
    id: string;
    role: string;
    status: number;
    last_heartbeat: string;
    tools: string[];
    model: string;
    system_prompt?: string;
  }>;
};

type CatalogueAgent = {
  id: string;
  name: string;
  role: string;
  tools: string[];
  inputs: string[];
  outputs: string[];
  verification_rubric: string[];
  created_at: string;
  updated_at: string;
};

export async function fulfillJSON(route: Route, status: number, body: unknown) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

export async function enableAdvancedMode(page: Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem("mycelis-advanced-mode", "true");
  });
}

export async function mockOperatorShell(page: Page, services: ServiceStatus[] = defaultServices) {
  await page.route("**/api/v1/user/me", async (route) => {
    await fulfillJSON(route, 200, {
      ok: true,
      data: { id: "operator-qa", name: "QA Operator", email: "qa@example.test" },
    });
  });

  await page.route("**/api/v1/services/status", async (route) => {
    await fulfillJSON(route, 200, { ok: true, data: services });
  });
}

export async function mockTeamsWorkspace(
  page: Page,
  teams: TeamDetailEntry[] = defaultTeams,
  catalogueAgents: CatalogueAgent[] = defaultCatalogueAgents,
) {
  await mockOperatorShell(page);

  await page.route("**/api/v1/teams/detail", async (route) => {
    await fulfillJSON(route, 200, teams);
  });

  await page.route("**/api/v1/catalogue/agents", async (route) => {
    await fulfillJSON(route, 200, catalogueAgents);
  });
}

export async function expectNoHorizontalOverflow(page: Page) {
  const widths = await page.evaluate(() => ({
    documentClientWidth: document.documentElement.clientWidth,
    bodyScrollWidth: document.body.scrollWidth,
    documentScrollWidth: document.documentElement.scrollWidth,
  }));
  expect(widths.bodyScrollWidth).toBeLessThanOrEqual(widths.documentClientWidth + 1);
  expect(widths.documentScrollWidth).toBeLessThanOrEqual(widths.documentClientWidth + 1);
}

export const defaultServices: ServiceStatus[] = [
  { name: "nats", status: "online", detail: "JetStream accepting team signals.", latency_ms: 8 },
  { name: "postgres", status: "online", detail: "Persistence responding.", latency_ms: 11 },
  { name: "scheduler", status: "online", detail: "Automation timing loop is current.", latency_ms: 5 },
  { name: "comms", status: "degraded", detail: "Optional providers are not configured." },
];

export const defaultTeams: TeamDetailEntry[] = [
  {
    id: "active-demo-team",
    name: "First Demo Game Team",
    role: "builder",
    type: "mission",
    mission_id: "mission-first-demo",
    mission_intent: "Create the canonical first-demo playable browser game package.",
    inputs: ["operator expression", "first-demo acceptance contract"],
    deliveries: ["workspace/generated/coin-runner/index.html", "workspace/generated/coin-runner/README.md"],
    agents: [
      {
        id: "active-demo-lead",
        role: "lead",
        status: 2,
        last_heartbeat: "2026-05-17T12:00:00Z",
        tools: ["create_team", "write_file", "store_artifact"],
        model: "balanced",
        system_prompt: "Own visible first-demo output and proof.",
      },
    ],
  },
  {
    id: "degraded-proof-team",
    name: "Recovery Proof Team",
    role: "reviewer",
    type: "standing",
    mission_id: null,
    mission_intent: null,
    inputs: ["failed run proof"],
    deliveries: [],
    agents: [
      {
        id: "recovery-reviewer",
        role: "reviewer",
        status: 3,
        last_heartbeat: "2026-05-17T12:01:00Z",
        tools: ["review_run"],
        model: "balanced",
      },
    ],
  },
];

export const defaultCatalogueAgents: CatalogueAgent[] = [
  {
    id: "template-builder",
    name: "Browser Game Builder",
    role: "builder",
    tools: ["write_file", "store_artifact"],
    inputs: ["acceptance criteria"],
    outputs: ["project_package"],
    verification_rubric: ["README present", "browser output opens"],
    created_at: "2026-05-17T12:00:00Z",
    updated_at: "2026-05-17T12:00:00Z",
  },
];
