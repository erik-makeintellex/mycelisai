import { expect, test, type Page } from "@playwright/test";

async function seedMalformedPersistedChat(page: Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem(
      "mycelis-workspace-chat",
      JSON.stringify([
        { role: "user", content: { text: "legacy object prompt" } },
        {
          role: "architect",
          source_node: "admin",
          content: "Legacy result with duplicate fields",
          tools_used: ["web_search", "web_search"],
          response_depth: "not-a-depth",
          proposal: {
            intent: "legacy proposal",
            risk_level: "medium",
            confirm_token: "token-1",
            intent_proof_id: "proof-1",
            tools: ["delegate_task", "delegate_task"],
            affected_resources: ["team:Game Team", "team:Game Team"],
          },
          execution_summary: {
            execution: { status: "complete", shape: "team_execution", summary: "legacy complete" },
            understanding: { summary: "legacy", assumptions: "not an array" },
            outputs: { label: "not an array" },
            proof: [
              { label: "Audit", run_id: "run-1" },
              { label: "Audit", run_id: "run-1" },
            ],
            capability_use: {
              tools: "not an array",
              used: ["web_search", "web_search"],
            },
          },
        },
        { role: "not-a-role", content: "drop me" },
      ]),
    );
  });
}

test.describe("Dashboard persisted chat regression", () => {
  test("dirty legacy chat state cannot crash the Soma dashboard", async ({ page }) => {
    const consoleIssues: string[] = [];
    const pageErrors: string[] = [];
    page.on("console", (message) => {
      if (message.type() === "error") consoleIssues.push(message.text());
      if (message.type() === "warning" && /same key|unique key/i.test(message.text())) {
        consoleIssues.push(message.text());
      }
    });
    page.on("pageerror", (error) => pageErrors.push(error.message));

    await seedMalformedPersistedChat(page);
    await page.goto("/dashboard", { waitUntil: "domcontentloaded" });

    await expect(page.getByTestId("central-soma-chat-frame")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/legacy object prompt/i)).toBeVisible();
    await expect(page.getByText(/Legacy result with duplicate fields/i)).toBeVisible();
    await expect(page.getByText("drop me")).toHaveCount(0);
    await expect(page.getByPlaceholder(/Tell Soma what you want/i)).toBeVisible();

    expect(pageErrors).toEqual([]);
    expect(consoleIssues).toEqual([]);
  });
});
