import { expect, test, type Route } from "@playwright/test";

const marketingRunId = "run-marketing-kit-1";
const marketingConfirmToken = "confirm-marketing-kit-1";
const marketingProofId = "proof-marketing-kit-1";
const marketingBasePath = "deliverables/marketing/launch-kit";

const marketingDeliveryPackage = {
    package_kind: "marketing_launch_kit",
    package_slug: "launch-kit",
    base_path: marketingBasePath,
    proof_path: `${marketingBasePath}/proof.md`,
    run_id: marketingRunId,
    proof_id: marketingProofId,
    files: [
        { path: `${marketingBasePath}/positioning_brief.md`, title: "Positioning Brief", bytes: 420, sha256: "mkt1110000000000" },
        { path: `${marketingBasePath}/launch_email_and_copy.md`, title: "Launch Email & Copy", bytes: 510, sha256: "mkt2220000000000" },
        { path: `${marketingBasePath}/customer_faq.md`, title: "Customer FAQ", bytes: 380, sha256: "mkt3330000000000" },
        { path: `${marketingBasePath}/proof.md`, title: "Marketing Launch Kit Proof", bytes: 260, sha256: "mkt4440000000000" },
    ],
};

function marketingProofEnvelopeFor(file: { path: string; bytes: number; sha256: string }) {
    const fileName = file.path.split("/").pop() ?? "file.md";
    return {
        proof_id: `${marketingProofId}:${fileName}`,
        output_ref_id: `marketing_launch_kit:launch-kit:${fileName}`,
        artifact_id: `artifact:${fileName}`,
        storage_ref: file.path,
        source_run_id: marketingRunId,
        source_contract_id: marketingProofId,
        execution_status: "completed",
        path_boundary_status: "ok",
        checksum: file.sha256,
        checksum_algorithm: "sha256",
        bytes: file.bytes,
        content_type: "text/markdown",
        created_at: "2026-11-08T19:00:00Z",
        readback_status: "verified",
    };
}

const marketingOutputRefs = marketingDeliveryPackage.files.map((file, index) => {
    const fileName = file.path.split("/").pop() ?? "file.md";
    return {
        id: `marketing_launch_kit:launch-kit:${fileName}`,
        kind: "document",
        label: file.title,
        storage_ref: file.path,
        open_url: `/api/v1/artifacts/artifact:${fileName}/download`,
        source_run_id: marketingRunId,
        source_contract_id: marketingProofId,
        artifact_id: `artifact:${fileName}`,
        package_id: "marketing_launch_kit:launch-kit",
        role: file.path.endsWith("proof.md") ? "proof" : index === 0 ? "primary" : "supporting",
        status: "final",
        bytes: file.bytes,
        sha256: file.sha256,
        proof: marketingProofEnvelopeFor(file),
    };
});

const marketingArtifacts = marketingDeliveryPackage.files.map((file) => ({
    type: "document",
    title: file.title,
    content_type: "text/markdown",
    content: `# ${file.title}\n\nVerified marketing deliverable stored at ${file.path}.`,
    saved_path: file.path,
}));

function fulfillJSON(route: Route, status: number, body: unknown) {
    return route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

test.skip(({ browserName }) => browserName !== "chromium", "Non-expert deliverable journeys run in Chromium.");

test.describe("Non-expert user deliverable journeys", () => {
    test.describe.configure({ mode: "serial" });

    test("non-expert user launches Marketing Launch Kit starter on /dashboard, approves proposal, and reviews multi-document marketing bundle", async ({ page }) => {
        await page.route("**/api/v1/user/me", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: { id: "operator-1", name: "Operator", email: "operator@example.test" },
            });
        });
        await page.route("**/api/v1/services/status", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: [
                    { name: "nats", status: "online" },
                    { name: "postgres", status: "online" },
                    { name: "cognitive", status: "online" },
                ],
            });
        });
        await page.route("**/api/v1/council/members", async (route) => {
            await fulfillJSON(route, 200, { ok: true, data: [{ id: "admin", role: "admin", team: "admin-core" }] });
        });
        await page.route("**/api/v1/chat", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    meta: { source_node: "admin", timestamp: "2026-11-08T19:00:00Z" },
                    signal_type: "chat_response",
                    trust_score: 0.92,
                    template_id: "chat-to-proposal",
                    mode: "proposal",
                    payload: {
                        text: "I can create your Marketing Launch Kit bundle, including positioning, email copy, and a customer FAQ.",
                        tools_used: ["write_file"],
                        proposal: {
                            intent: "chat-action",
                            operator_summary: `Create a Marketing Launch Kit under "${marketingBasePath}".`,
                            expected_result: "A multi-document marketing bundle with durable proof will be saved after approval.",
                            affected_resources: marketingDeliveryPackage.files.map((file) => file.path),
                            teams: 1,
                            agents: 1,
                            tools: ["write_file"],
                            risk_level: "medium",
                            confirm_token: marketingConfirmToken,
                            intent_proof_id: marketingProofId,
                        },
                    },
                },
            });
        });
        await page.route("**/api/v1/intent/confirm-action", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: {
                    confirmed: true,
                    verified: true,
                    execution_state: "verified",
                    proof_id: marketingProofId,
                    run_id: marketingRunId,
                    run_status: "completed",
                    message: `Marketing Launch Kit created under ${marketingBasePath} with proof at ${marketingBasePath}/proof.md.`,
                    package: marketingDeliveryPackage,
                    artifacts: marketingArtifacts,
                    output_refs: marketingOutputRefs,
                },
            });
        });

        await page.goto("/dashboard?fresh=1", { waitUntil: "domcontentloaded" });

        const starterBtn = page.getByRole("button", { name: /Marketing Launch Kit/i });
        await expect(starterBtn).toBeVisible();
        await expect(page.getByRole("button", { name: /Playable Puzzle Game/i })).toBeVisible();
        await expect(page.getByRole("button", { name: /Customer Onboarding Kit/i })).toBeVisible();
        await expect(page.getByRole("button", { name: /Interactive KPI Dashboard/i })).toBeVisible();
        await starterBtn.click();

        const expectedPrompt = "Create marketing documentation for our product launch including a positioning brief, launch email, and customer FAQ.";
        const promptInput = page.getByPlaceholder(/Describe the outcome you need from Soma/i);
        await expect(promptInput).toHaveValue(expectedPrompt);

        await promptInput.press("Enter");

        await expect(page.getByText("PROPOSED ACTION")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(`Create a Marketing Launch Kit under "${marketingBasePath}".`)).toBeVisible();

        await page.getByRole("button", { name: /Approve & Execute|Execute/i }).last().click();

        const workbench = page.getByTestId("output-workbench");
        await expect(workbench).toBeVisible({ timeout: 20_000 });
        await expect(workbench.getByText("Positioning Brief", { exact: true })).toBeVisible();
        await expect(workbench.getByText("Launch Email & Copy", { exact: true })).toBeVisible();
        await expect(workbench.getByText("Customer FAQ", { exact: true })).toBeVisible();
        await expect(workbench.getByText("Marketing Launch Kit Proof", { exact: true })).toBeVisible();
        await expect(workbench.getByText(`${marketingBasePath}/positioning_brief.md`).first()).toBeVisible();
        await expect(workbench.getByText(`${marketingBasePath}/launch_email_and_copy.md`).first()).toBeVisible();
        await expect(workbench.getByText(`${marketingBasePath}/customer_faq.md`).first()).toBeVisible();
        await expect(workbench.getByText(`${marketingBasePath}/proof.md`).first()).toBeVisible();
        await expect(workbench.getByText("verified").first()).toBeVisible();
    });

    test("non-expert user opens Durable Files on /resources and plays the generated HTML5 Sliding-Tile Puzzle Game inside the live preview iframe", async ({ page }) => {
        await page.route("**/api/v1/artifacts**", async (route) => {
            await fulfillJSON(route, 200, {
                ok: true,
                data: [
                    {
                        id: "art-puzzle-game-1",
                        agent_id: "admin",
                        artifact_type: "code",
                        title: "Sliding Tile Puzzle (HTML5)",
                        content_type: "text/html",
                        file_path: "deliverables/games/sliding-tile-puzzle/index.html",
                        content: `<!DOCTYPE html><html><body><h1>Sliding Tile Puzzle</h1><span id="move-counter">Moves: 0</span><button data-testid="puzzle-tile" onclick="document.getElementById('move-counter').textContent='Moves: 1'">1</button></body></html>`,
                        status: "approved",
                        created_at: "2026-11-08T19:05:00Z",
                    },
                ],
            });
        });

        await page.addInitScript(() => {
            window.localStorage.setItem("mycelis-advanced-mode", "false");
        });
        await page.goto("/resources", { waitUntil: "domcontentloaded" });

        await expect(page.getByTestId("durable-files-operator-view")).toBeVisible();
        await expect(page.getByRole("heading", { name: "Durable Files & Deliverables" })).toBeVisible();
        await expect(page.getByText("Sliding Tile Puzzle (HTML5)")).toBeVisible();
        await expect(page.getByText("deliverables/games/sliding-tile-puzzle/index.html")).toBeVisible();

        const playBtn = page.getByTestId("play-deliverable-art-puzzle-game-1");
        await expect(playBtn).toBeVisible();
        await playBtn.click();

        const gameIframe = page.getByTestId("deliverable-iframe-art-puzzle-game-1");
        await expect(gameIframe).toBeVisible();

        const frame = page.frameLocator('[data-testid="deliverable-iframe-art-puzzle-game-1"]');
        await expect(frame.getByRole("heading", { name: "Sliding Tile Puzzle" })).toBeVisible();
        await expect(frame.getByText("Moves: 0")).toBeVisible();

        await frame.getByTestId("puzzle-tile").click();
        await expect(frame.getByText("Moves: 1")).toBeVisible();
    });
});
