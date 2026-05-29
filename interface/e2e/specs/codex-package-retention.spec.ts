import { expect, test, type Route } from "@playwright/test";
import {
    chatPlaceholder,
    mockOrganizationWorkspace,
    openOrganization,
    organizationId,
    sendWorkspaceMessage,
    type ChatRequestBody,
    type RouteResponse,
} from "../support/soma-ui-testing";

const runId = "run-codex-package-1";
const confirmToken = "confirm-codex-package-1";
const proofId = "proof-codex-package-1";
const basePath = "codex-task-packages/myc-rc-001";

const deliveryPackage = {
    package_kind: "codex_task_bundle",
    package_slug: "myc-rc-001",
    base_path: basePath,
    proof_path: `${basePath}/proof.md`,
    run_id: runId,
    proof_id: proofId,
    files: [
        { path: `${basePath}/implementation_task.md`, title: "Implementation Task", bytes: 120, sha256: "abc1230000000000" },
        { path: `${basePath}/validation_checklist.md`, title: "Validation Checklist", bytes: 90, sha256: "abc4560000000000" },
        { path: `${basePath}/affected_files_summary.md`, title: "Affected Files Summary", bytes: 80, sha256: "abc7890000000000" },
        { path: `${basePath}/next_recommended_slices.md`, title: "Next Recommended Slices", bytes: 110, sha256: "def1230000000000" },
        { path: `${basePath}/proof.md`, title: "Package Proof", bytes: 140, sha256: "def4560000000000" },
    ],
};

const packageArtifacts = deliveryPackage.files.map((file) => ({
    type: "document",
    title: file.title,
    content_type: "text/markdown",
    content: `# ${file.title}\n\nCodex package proof fixture.`,
    saved_path: file.path,
}));

function fulfillJSON(route: Route, status: number, body: unknown) {
    return route.fulfill({
        status,
        contentType: "application/json",
        body: JSON.stringify(body),
    });
}

function codexProposalEnvelope(_requestBody: ChatRequestBody): RouteResponse {
    return {
        status: 200,
        body: {
            ok: true,
            data: {
                meta: { source_node: "admin", timestamp: "2026-03-26T16:00:00Z" },
                signal_type: "chat_response",
                trust_score: 0.88,
                template_id: "chat-to-proposal",
                mode: "proposal",
                payload: {
                    text: "I can create a Codex task bundle, but I need your confirmation before execution.",
                    tools_used: ["write_file"],
                    proposal: {
                        intent: "chat-action",
                        operator_summary: `Create a Codex task bundle under "${basePath}".`,
                        expected_result: "A multi-file Codex-ready implementation package will be created with durable run proof after approval.",
                        affected_resources: deliveryPackage.files.map((file) => file.path),
                        teams: 1,
                        agents: 1,
                        tools: ["write_file"],
                        risk_level: "medium",
                        confirm_token: confirmToken,
                        intent_proof_id: proofId,
                    },
                },
            },
        },
    };
}

async function mockConfirmAction(page: import("@playwright/test").Page) {
    await page.route("**/api/v1/intent/confirm-action", async (route) => {
        await fulfillJSON(route, 200, {
            ok: true,
            data: {
                confirmed: true,
                verified: true,
                execution_state: "verified",
                proof_id: proofId,
                run_id: runId,
                run_status: "completed",
                message: `Codex task bundle created under ${basePath} with proof at ${basePath}/proof.md.`,
                package: deliveryPackage,
                artifacts: packageArtifacts,
            },
        });
    });
}

test.skip(({ browserName }) => browserName !== "chromium", "Codex package retention proof is stabilized in Chromium.");

test.describe("Codex task package retention", () => {
    test("retains a confirmed Codex package after refresh and organization re-entry", async ({ page }) => {
        await page.route("**/api/v1/council/members", async (route) => {
            await fulfillJSON(route, 200, { ok: true, data: [{ id: "admin", role: "admin", team: "admin-core" }] });
        });
        await mockOrganizationWorkspace(page, codexProposalEnvelope);
        await mockConfirmAction(page);

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Generate a Codex task package for MYC-RC-001.");

        await expect(page.getByText("PROPOSED ACTION")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByText(`Create a Codex task bundle under "${basePath}".`)).toBeVisible();
        await page.getByRole("button", { name: /Approve & Execute|Execute/i }).last().click();

        const workbench = page.getByTestId("output-workbench");
        await expect(workbench).toBeVisible({ timeout: 20_000 });
        await expect(workbench.getByText("Codex task bundle")).toBeVisible();
        await expect(workbench.getByText("Implementation Task", { exact: true })).toBeVisible();
        await expect(workbench.getByText("Package Proof", { exact: true })).toBeVisible();
        await expect(page.locator(`a[href="/runs/${runId}"]`).first()).toBeVisible();

        await page.reload({ waitUntil: "domcontentloaded" });
        await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
        await expect(workbench).toBeVisible();
        await expect(workbench.getByText("Package Proof", { exact: true })).toBeVisible();

        await page.goto(`/organizations/${organizationId}`);
        await page.getByPlaceholder(chatPlaceholder).waitFor({ timeout: 20_000 });
        const reentryWorkbench = page.getByTestId("output-workbench");
        await expect(reentryWorkbench).toBeVisible();
        await expect(reentryWorkbench.getByText("Validation Checklist", { exact: true })).toBeVisible();
    });
});
