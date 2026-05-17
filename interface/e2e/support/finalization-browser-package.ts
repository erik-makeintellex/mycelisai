import fs from "node:fs";
import path from "node:path";
import { execFileSync } from "node:child_process";
import { expect, type Page, type Route } from "@playwright/test";
import type { RouteResponse } from "./soma-ui-testing";
import { liveAPIHeaders, liveAPIURL } from "./live-api-auth";

const repoRoot = path.resolve(__dirname, "../../..");
export const liveTimeoutMs = 180_000;
export const chatTimeoutMs = 120_000;

export type APIEnvelope<T> = { ok?: boolean; data?: T; error?: string };

export type ChatEnvelope = {
  data?: { mode?: string; payload?: { tools_used?: string[] } };
};

export type ConfirmEnvelope = {
  data?: {
    run_id?: string;
    verified?: boolean;
    execution_state?: string;
    execution_summary?: { outputs?: ExecutionOutput[] };
  };
};

export type ExecutionOutput = {
  kind?: string;
  title?: string;
  id?: string;
  href?: string;
  retained?: boolean;
  entrypoint?: string;
  folder?: string;
  files?: string[];
  validation?: string;
};

export type GroupRecord = {
  group_id: string;
  name: string;
  work_mode?: string;
  status?: string;
  team_ids?: string[];
};

export type ArtifactRecord = {
  id: string;
  title: string;
  artifact_type: string;
  team_id?: string;
  agent_id?: string;
  content_type?: string;
  file_path?: string;
  metadata?: Record<string, unknown>;
  status?: string;
  created_at?: string;
};

type OrganizationEnvelope = { data?: { id?: string } };

export async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
  const raw = await response.text();
  try {
    return { raw, body: JSON.parse(raw) as T };
  } catch {
    return { raw, body: null as T | null };
  }
}

function backendWorkspaceRoots() {
  const configuredRoot = process.env.PLAYWRIGHT_BACKEND_WORKSPACE_ROOT ?? process.env.MYCELIS_BACKEND_WORKSPACE_ROOT;
  if (configuredRoot?.trim()) {
    return [path.isAbsolute(configuredRoot) ? configuredRoot : path.join(repoRoot, configuredRoot)];
  }
  return [
    path.join(repoRoot, "core", "workspace"),
    path.join(repoRoot, "workspace", "docker-compose", "data", "workspace"),
  ];
}

let cachedK8sPodName: string | null = null;

function k8sWorkspaceProbeEnabled() {
  return process.env.PLAYWRIGHT_BACKEND_WORKSPACE_PROBE === "k8s";
}

function shellQuote(value: string) {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

function k8sCorePodName() {
  if (cachedK8sPodName) return cachedK8sPodName;
  const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? "mycelis";
  const selector = process.env.PLAYWRIGHT_K8S_CORE_SELECTOR ?? "app=mycelis-core";
  cachedK8sPodName = execFileSync(
    "kubectl",
    ["get", "pods", "-n", namespace, "-l", selector, "-o", "jsonpath={.items[0].metadata.name}"],
    { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"], timeout: 10_000 },
  ).trim();
  return cachedK8sPodName;
}

export function targetExists(relativePath: string) {
  const normalized = relativePath.replace(/^workspace[\\/]/, "").replace(/\\/g, "/");
  if (backendWorkspaceRoots().some((workspaceRoot) => fs.existsSync(path.join(workspaceRoot, normalized)))) return true;
  if (!k8sWorkspaceProbeEnabled()) return false;
  const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? "mycelis";
  const root = (process.env.PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT ?? "/data/workspace").replace(/\/$/, "");
  try {
    execFileSync(
      "kubectl",
      ["exec", "-n", namespace, k8sCorePodName(), "--", "sh", "-c", `test -f ${shellQuote(`${root}/${normalized}`)}`],
      { stdio: "ignore", timeout: 10_000 },
    );
    return true;
  } catch {
    return false;
  }
}

export function removeTarget(relativePath: string) {
  const normalized = relativePath.replace(/^workspace[\\/]/, "").replace(/\\/g, "/");
  for (const workspaceRoot of backendWorkspaceRoots()) {
    fs.rmSync(path.join(workspaceRoot, normalized), { force: true });
  }
  if (!k8sWorkspaceProbeEnabled()) return;
  const namespace = process.env.PLAYWRIGHT_K8S_NAMESPACE ?? "mycelis";
  const root = (process.env.PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT ?? "/data/workspace").replace(/\/$/, "");
  try {
    execFileSync(
      "kubectl",
      ["exec", "-n", namespace, k8sCorePodName(), "--", "sh", "-c", `rm -f ${shellQuote(`${root}/${normalized}`)}`],
      { stdio: "ignore", timeout: 10_000 },
    );
  } catch {
    // Cleanup is best-effort; the browser assertion should remain the useful failure.
  }
}

export async function createOrganization(page: Page, name: string) {
  const response = await page.request.post(liveAPIURL("/api/v1/organizations"), {
    headers: liveAPIHeaders(),
    data: { name, purpose: "Exact UI finalization browser package proof", start_mode: "empty" },
  });
  const parsed = await parseJSONIfPossible<OrganizationEnvelope>(response);
  expect(response.ok(), parsed.body ? JSON.stringify(parsed.body) : parsed.raw).toBeTruthy();
  expect(parsed.body?.data?.id).toBeTruthy();
  return parsed.body!.data!.id!;
}

export async function liveAPIGet(page: Page, url: string) {
  return page.request.get(liveAPIURL(url), { headers: liveAPIHeaders() });
}

export async function openLiveWorkspace(page: Page, organizationId: string) {
  await page.goto(`/organizations/${organizationId}`, { waitUntil: "domcontentloaded" });
  await page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i).waitFor({ timeout: 30_000 });
}

export async function submitLiveWorkspaceChat(page: Page, content: string) {
  const input = page.getByPlaceholder(/Tell Soma what you want to plan, review, create, or execute/i);
  await input.fill(content);
  const responsePromise = page.waitForResponse(
    (response) => response.url().includes("/api/v1/chat") && response.request().method() === "POST",
    { timeout: chatTimeoutMs },
  );
  await input.press("Enter");
  const response = await responsePromise;
  const parsed = await parseJSONIfPossible<ChatEnvelope>(response);
  return { response, raw: parsed.raw, body: parsed.body };
}

export async function confirmProposal(page: Page) {
  const responsePromise = page.waitForResponse(
    (response) => response.url().includes("/api/v1/intent/confirm-action") && response.request().method() === "POST",
    { timeout: chatTimeoutMs },
  );
  await page.getByRole("button", { name: /Approve & Execute|Execute/i }).last().click();
  const response = await responsePromise;
  const parsed = await parseJSONIfPossible<ConfirmEnvelope>(response);
  return { response, raw: parsed.raw, body: parsed.body };
}

export function expectProjectPackageMetadata(projectPackage: ExecutionOutput, expected: {
  title: string;
  entrypoint: string;
  folder: string;
}) {
  expect(projectPackage.kind).toBe("project_package");
  expect(projectPackage.title).toContain(expected.title);
  expect(projectPackage.retained).toBeTruthy();
  expect(projectPackage.href).toBe(`/api/v1/workspace/files/view?path=${encodeURIComponent(expected.entrypoint)}`);
  expect(projectPackage.entrypoint).toBe(expected.entrypoint);
  expect(projectPackage.folder).toBe(expected.folder);
  expect(projectPackage.files ?? []).toEqual(expect.arrayContaining(["index.html", "README.md"]));
  expect(projectPackage.validation).toMatch(/browser|validation|opened|play/i);
}

export async function expectProjectPackageVisible(page: Page, expected: {
  title: string;
  entrypoint: string;
  folder: string;
}) {
  await expect(page.getByText("Operator trust package").last()).toBeVisible({ timeout: 30_000 });
  await expect(page.getByText("Run proof + retained output").last()).toBeVisible();
  await expect(page.getByText(expected.title).last()).toBeVisible();
  await expect(page.getByText(`entry: ${expected.entrypoint}`).last()).toBeVisible();
  await expect(page.getByText(`folder: ${expected.folder}`).last()).toBeVisible();
  await expect(page.getByText("README.md").last()).toBeVisible();
  await expect(page.getByText(/browser|validation|opened|play/i).last()).toBeVisible();
}

export async function fulfillJSON(route: Route, status: number, body: unknown) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

export function firstDemoPackageProposal(): RouteResponse {
  return {
    status: 200,
    body: {
      ok: true,
      data: {
        meta: { source_node: "admin", timestamp: "2026-05-16T20:00:00Z" },
        signal_type: "chat.reply",
        trust_score: 0.9,
        template_id: "chat-to-proposal",
        mode: "proposal",
        payload: {
          text: "I can create the first-demo game package after approval.",
          tools_used: ["create_team", "write_file"],
          consultations: [],
          artifacts: [],
          proposal: {
            intent: "create_first_demo_browser_game_package",
            operator_summary: "create a playable browser game package with README and validation notes.",
            expected_result: "A retained project package will include an entrypoint, folder, README, file list, and browser validation notes.",
            affected_resources: ["workspace/generated/coin-runner/index.html", "workspace/generated/coin-runner/README.md"],
            teams: 1,
            agents: 1,
            tools: ["create_team", "write_file"],
            risk_level: "medium",
            confirm_token: "confirm-first-demo-package",
            intent_proof_id: "proof-first-demo-package",
            approval_required: false,
            approval_reason: "capability_risk",
            approval_mode: "optional",
            capability_risk: "medium",
            capability_ids: ["create_team", "write_file"],
            external_data_use: false,
            estimated_cost: 0,
          },
        },
      },
    },
  };
}
