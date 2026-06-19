import { expect, test } from "@playwright/test";
import { expectProjectPackageVisible } from "../support/finalization-browser-package";
import { clickVisibleControl } from "../support/click-visible-control";
import { openOrganization, sendWorkspaceMessage } from "../support/soma-ui-testing";
import {
  apiFetch,
  installTrustedOutcomeJourneyMocks,
  trustedJourney,
} from "../support/trusted-outcome-journey";

type Envelope<T> = { ok?: boolean; data?: T };

test.describe("Trusted Outcome Journey", () => {
  test("proves Ask, Understand, Approve, Execute, Deliver, Trust, Recover, and Revisit", async ({ page }) => {
    const j = trustedJourney;
    await installTrustedOutcomeJourneyMocks(page);

    await openOrganization(page);
    await expect(page.getByRole("heading", { name: /What do you want Soma to do/i })).toBeVisible();

    await sendWorkspaceMessage(page, "Create a retained trusted outcome kit with proof and recovery notes.");
    await expect(page.getByText("I can create a retained journey kit after approval.")).toBeVisible();
    await expect(page.getByText("A retained Trusted Outcome Kit will open from Soma")).toBeVisible();

    await expect(page.getByText("RUN CONFIRMATION")).toBeVisible();
    const confirmResponse = page.waitForResponse((response) =>
      response.url().includes("/api/v1/intent/confirm-action") && response.request().method() === "POST",
    );
    await clickVisibleControl(page, page.getByRole("button", { name: /Run now|Approve|Execute/i }));
    const confirmBody = (await (await confirmResponse).json()) as Envelope<{ run_id?: string; verified?: boolean }>;
    expect(confirmBody.ok).toBeTruthy();
    expect(confirmBody.data?.run_id).toBe(j.runId);
    expect(confirmBody.data?.verified).toBeTruthy();

    await expect(page.getByText(/Action completed|verified|Output package/i).first()).toBeVisible();
    await expectProjectPackageVisible(page, {
      title: j.packageTitle,
      entrypoint: j.entrypoint,
      folder: j.folder,
    });

    const popupPromise = page.waitForEvent("popup", { timeout: 5_000 }).catch(() => null);
    await clickVisibleControl(page, page.getByRole("button", { name: new RegExp(`Open file .*${j.packageTitle}`, "i") }).last());
    const outputPage = (await popupPromise) ?? page;
    if (!outputPage.url().includes("/api/v1/workspace/files/view")) {
      await outputPage.goto(`/api/v1/workspace/files/view?path=${encodeURIComponent(j.entrypoint)}`);
    }
    await expect(outputPage.getByRole("heading", { name: j.packageTitle })).toBeVisible();
    await expect(outputPage.getByText("Recover, trust, and revisit this output.")).toBeVisible();
    if (outputPage !== page) await outputPage.close();

    await clickVisibleControl(page, page.getByRole("button", { name: /Open .*folder/i }).last());
    await expect(page.getByText(/Folder opened|Folder access blocked|Opened folder/i).last()).toBeVisible();

    const proof = await apiFetch<Envelope<{ run_id?: string; confidence?: string; summary?: string }>>(
      page,
      `/api/v1/trust/proof-artifacts/${j.proofArtifactId}`,
    );
    expect(proof.ok).toBeTruthy();
    expect(proof.body.data?.run_id).toBe(j.runId);
    expect(proof.body.data?.confidence).toBe("high");

    const contract = await apiFetch<Envelope<{ approved?: boolean; output_refs?: string[] }>>(
      page,
      `/api/v1/trust/execution-contracts/${j.executionContractId}`,
    );
    expect(contract.ok).toBeTruthy();
    expect(contract.body.data?.approved).toBeTruthy();
    expect(contract.body.data?.output_refs).toContain(j.entrypoint);

    const events = await apiFetch<Envelope<Array<{ event_type?: string }>>>(page, `/api/v1/runs/${j.runId}/events`);
    expect(events.body.data?.map((event) => event.event_type)).toEqual(
      expect.arrayContaining(["mission.started", "artifact.created", "mission.completed"]),
    );

    const workAsk = await apiFetch<Envelope<{ proof?: { status?: string } }>>(page, `/api/v1/teams/${j.teamId}/work/ask`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ request: "Have the team verify the retained output." }),
    });
    expect(workAsk.body.data?.proof?.status).toBe("verified");

    const workItems = await apiFetch<Envelope<Array<{ status?: string; recovery?: Record<string, unknown> }>>>(
      page,
      `/api/v1/teams/${j.teamId}/work`,
    );
    const recoveryItem = workItems.body.data?.find((item) => item.status === "needs_recovery");
    expect(recoveryItem?.recovery?.safe_next).toContain("Repair dependency");

    await page.reload();
    await expectProjectPackageVisible(page, {
      title: j.packageTitle,
      entrypoint: j.entrypoint,
      folder: j.folder,
    });

    await page.goto("/resources?tab=outputs");
    await expect(page.getByRole("heading", { name: /Resources/i })).toBeVisible();
    await page.getByRole("button", { name: new RegExp(`${j.packageTitle}.*${j.entrypoint}`, "i") }).click();
    await expect(page.getByRole("button", { name: "Preview output file index.html" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Preview output file PROOF.md" })).toBeVisible();

    await page.goto(`/groups?group_id=${j.groupId}&panel=workflow&advanced=1`);
    await expect(page.getByRole("heading", { name: "Trusted Outcome Delivery Lane" })).toBeVisible();
    await expect(page.getByText("Stored output is available at workspace/generated/trusted-outcome-kit/index.html.")).toBeVisible();
    await page.getByRole("tab", { name: /Outputs/i }).click();
    await expect(page.getByText(j.packageTitle)).toBeVisible();

    await page.goto(`/runs/${j.runId}?tab=events`);
    await expect(page.getByRole("heading", { name: "Run completed" })).toBeVisible();
    await expect(page.getByText("Trusted Outcome Kit delivered.")).toBeVisible();
    await expect(page.getByText(/Completed run evidence is available/i)).toBeVisible();
  });
});
