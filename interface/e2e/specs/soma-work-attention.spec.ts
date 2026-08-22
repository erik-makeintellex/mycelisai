import { expect, test, type Page } from "@playwright/test";
import { expectNoHorizontalOverflow } from "../support/finalization-proof";

const repeatedWorkUpdates = [
  threadMessage("execution_started", "Work started", "Work is underway."),
  threadMessage("attention_required", "Work needs attention", "The configured cognitive engine did not return a readable reply."),
  threadMessage("attention_required", "Work needs attention", "The configured provider timed out before returning a usable result."),
];

function threadMessage(kind: "execution_started" | "attention_required", label: string, detail: string) {
  return {
    role: "system",
    content: `${label} - ${detail}`,
    mode: kind === "attention_required" ? "blocker" : "execution_result",
    run_id: "run-attention-proof-1",
    thread_event: {
      kind,
      label,
      detail,
      tone: kind === "attention_required" ? "warning" : "info",
      status: kind === "attention_required" ? "degraded" : "running",
      run_id: "run-attention-proof-1",
      source_kind: "system",
      source_channel: "team-work.result-projection",
      payload_kind: "thread_event",
    },
  };
}

async function installWorkAttentionFixture(page: Page) {
  await page.addInitScript((messages) => {
    window.localStorage.setItem("mycelis-advanced-mode", "false");
    window.localStorage.setItem("mycelis-workspace-chat", JSON.stringify(messages));
  }, repeatedWorkUpdates);
}

test.describe("Soma work attention UX", () => {
  for (const viewport of [
    { name: "desktop", width: 1366, height: 768 },
    { name: "compact", width: 390, height: 844 },
  ]) {
    test(`${viewport.name} shows one conversational recovery direction`, async ({ page }) => {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await installWorkAttentionFixture(page);
      await page.goto("/dashboard", { waitUntil: "domcontentloaded" });

      await expect(page.getByTestId("soma-thread-state-card")).toHaveCount(1);
      await expect(page.getByText("Soma needs your direction")).toBeVisible();
      await expect(page.getByText(/Tell Soma to try again, use another available service, or change the request/i)).toBeVisible();
      await expect(page.getByRole("button", { name: /Continue with Soma/i })).toHaveCount(0);
      await expect(page.getByText("degraded", { exact: true })).toHaveCount(0);

      const composer = page.getByTestId("central-soma-chat-frame").getByRole("textbox");
      await expect(composer).toBeVisible();
      await expect(composer).toBeInViewport();
      await page.getByText("What happened", { exact: true }).click();
      await expect(page.getByText(/configured provider timed out/i)).toBeVisible();
      await expectNoHorizontalOverflow(page);
    });
  }
});
