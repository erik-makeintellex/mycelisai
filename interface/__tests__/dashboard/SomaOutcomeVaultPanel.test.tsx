import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SomaOutcomeVaultPanel } from "@/components/soma/SomaOutcomeVaultPanel";

describe("SomaOutcomeVaultPanel", () => {
  it("renders active operations, recovery alerts, and deliverables as accessible links when targets exist", () => {
    render(
      <SomaOutcomeVaultPanel
        operationCount={2}
        recoveryCount={1}
        projectSummary={{
          title: "Media Pack outcome workspace",
          detail: "Soma has assigned work toward this outcome and will keep the thread updated.",
          ownerLabel: "Soma",
          leadLabel: "Media Lead, coordinator",
          registryOwnerLabel: "Media Lead lead",
          teamCount: 1,
          workCount: 2,
          outputCount: 1,
          recoveryCount: 1,
          href: "/teams?view=work",
          hrefLabel: "Open work",
          outputHref: "/resources?tab=workspace",
          outputLabel: "Open outputs",
        }}
        alerts={[{
          id: "work:review-1",
          kind: "recovery",
          severity: "warning",
          title: "Recover browser proof",
          detail: "Review the failed run and choose a safe retry.",
          href: "/teams?view=work&work_item_id=review-1",
          actionLabel: "Open work item",
          targetReference: "work:review-1",
          target: {
            type: "recovery",
            id: "review-1",
            href: "/teams?view=work&work_item_id=review-1",
            label: "Recovery item",
          },
        }]}
        latestOutput={{
          text: "Launch package",
          url: "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Flaunch%2Findex.html",
          storagePath: "workspace/generated/launch",
          count: 1,
        }}
      />,
    );

    expect(screen.getByText("Saved results, work in progress, and anything that needs your attention.")).toBeDefined();
    expect(screen.getByText("2 items need your attention")).toBeDefined();
    expect(screen.getByText("1 item needs recovery attention.")).toBeDefined();

    const reviewItems = screen.getByRole("link", { name: "Review next step" });
    expect(reviewItems.getAttribute("href")).toBe("/teams?view=work");
    expect(reviewItems.getAttribute("data-target-reference")).toBe("/teams?view=work");

    const typedAlert = screen.getByRole("link", { name: "Review recovery: Recover browser proof" });
    expect(typedAlert.getAttribute("href")).toBe("/teams?view=work&work_item_id=review-1");
    expect(typedAlert.getAttribute("data-alert-id")).toBe("work:review-1");
    expect(typedAlert.getAttribute("data-alert-kind")).toBe("recovery");
    expect(typedAlert.getAttribute("data-target-reference")).toBe("work:review-1");
    expect(typedAlert.getAttribute("data-target-type")).toBe("recovery");
    expect(typedAlert.getAttribute("data-target-id")).toBe("review-1");
    expect(screen.queryByText("Recovery item")).toBeNull();

    expect(screen.getByText("Outcome ready to revisit")).toBeDefined();
    expect(screen.getByText("Media Pack outcome workspace")).toBeDefined();
    expect(screen.getByText("Recovery attention")).toBeDefined();
    expect(screen.queryByText("Lead:")).toBeNull();
    expect(screen.queryByText("Media Lead, coordinator")).toBeNull();
    expect(screen.queryByText("OutcomeProject owner:")).toBeNull();
    expect(screen.queryByText("TeamRegistry owner:")).toBeNull();
    expect(screen.queryByText("Media Lead lead")).toBeNull();
    expect(screen.getByRole("link", { name: "Review recovery" }).getAttribute("href")).toBe("/teams?view=work");
    expect(screen.getByRole("link", { name: "Open saved outcomes" }).getAttribute("href")).toBe("/resources?tab=workspace");

    const deliverable = screen.getByRole("link", { name: "Open latest deliverable Launch package" });
    expect(deliverable.getAttribute("href")).toBe("/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Flaunch%2Findex.html");
    expect(deliverable.getAttribute("data-target-reference")).toBe("workspace/generated/launch");
    expect(screen.getByText("File details")).toBeDefined();
    expect(screen.getByRole("button", { name: /Open Launch package in a new browser window/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Open local folder for Launch package at workspace\/generated\/launch/i })).toBeDefined();
  });

  it("opens a storage-only deliverable in the workspace vault", () => {
    render(
      <SomaOutcomeVaultPanel
        operationCount={0}
        latestOutput={{
          text: "Generated brief",
          url: null,
          storagePath: "generated/brief.md",
          count: 1,
        }}
      />,
    );

    const deliverable = screen.getByRole("link", { name: "Open latest deliverable Generated brief" });
    expect(deliverable.getAttribute("href")).toBe("/resources?tab=workspace&path=workspace%2Fgenerated%2Fbrief.md");
    expect(deliverable.getAttribute("data-target-reference")).toBe("generated/brief.md");
    expect(screen.queryByRole("link", { name: /Open review lane/i })).toBeNull();
  });

  it("keeps untargeted alerts and deliverables non-clickable", () => {
    render(
      <SomaOutcomeVaultPanel
        operationCount={0}
        latestOutput={{
          text: "Untargeted digest",
          url: null,
          count: 1,
        }}
      />,
    );

    expect(screen.getByText("Untargeted digest")).toBeDefined();
    expect(screen.queryByRole("link", { name: /Open latest deliverable/i })).toBeNull();
    expect(screen.queryByRole("link", { name: /items needing attention/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /Open Untargeted digest/i })).toBeNull();
  });
});
