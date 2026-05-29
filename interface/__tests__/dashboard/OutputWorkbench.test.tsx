import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import {
  OutputWorkbench,
  mergeOutputWorkbenchItems,
  outputWorkbenchItems,
  projectPackageOutputs,
  teamOutputProjectPackages,
  teamOutputWorkbenchItems,
} from "@/components/soma/OutputWorkbench";
import type { ExecutionSummaryData, TeamOutputRef } from "@/store/useCortexStore";

describe("OutputWorkbench", () => {
  it("extracts project packages separately from retained outputs and artifacts", () => {
    const summary: ExecutionSummaryData = {
      outputs: [
        {
          kind: "project_package",
          title: "Launch microsite",
          folder: "generated/launch",
          entrypoint: "index.html",
          files: ["index.html", "styles.css"],
        },
        { title: "Launch brief", url: "/runs/run-1" },
      ],
    };

    expect(projectPackageOutputs(summary.outputs)).toHaveLength(1);
    expect(outputWorkbenchItems(summary, [{ type: "document", title: "Launch brief", url: "/runs/run-1" }])).toEqual([
      { text: "Launch brief", url: "/runs/run-1" },
    ]);
    expect(outputWorkbenchItems(summary, [{ type: "document", title: "Operator notes", url: "/notes/1" }])).toEqual([
      { text: "Launch brief", url: "/runs/run-1" },
      { text: "Operator notes", url: "/notes/1" },
    ]);
  });

  it("projects durable team outputs into the workbench without duplicating chat outputs", () => {
    const teamOutputs: TeamOutputRef[] = [
      {
        output_id: "out-1",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "file",
        label: "Launch brief",
        storage_ref: "generated/launch/brief.md",
      },
      {
        output_id: "out-2",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "project_package",
        label: "Launch package",
        storage_ref: "generated/launch",
        entrypoint: "index.html",
        proof_ref: "proof-1",
      },
    ];

    const durableItems = teamOutputWorkbenchItems(teamOutputs);
    expect(durableItems).toEqual([
      {
        text: "Launch brief",
        url: "/api/v1/workspace/files/view?path=generated%2Flaunch%2Fbrief.md",
      },
    ]);
    expect(teamOutputProjectPackages(teamOutputs)).toEqual([
      expect.objectContaining({
        kind: "project_package",
        title: "Launch package",
        folder: "generated/launch",
        entrypoint: "index.html",
        validation: "Linked proof or validation record",
      }),
    ]);
    expect(mergeOutputWorkbenchItems([{ text: "Launch brief", url: durableItems[0].url }], durableItems)).toHaveLength(1);
  });

  it("normalizes retained media paths so Soma can preview and open generated content", () => {
    const summary: ExecutionSummaryData = {
      outputs: [
        {
          kind: "image",
          title: "Comic page",
          href: "saved-media/comic-page.png",
          retained: true,
        },
      ],
    };

    expect(outputWorkbenchItems(summary)).toEqual([
      {
        text: "Comic page",
        url: "/api/v1/workspace/files/view?path=saved-media%2Fcomic-page.png",
      },
    ]);
  });

  it("renders package actions and copyable output quotes", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });

    render(
      <OutputWorkbench
        outputs={[{
          text: "Launch brief",
          url: "/runs/run-1",
          proofArtifactId: "proof-artifact-1",
          proof: {
            path_boundary_status: "verified",
            readback_status: "verified",
            checksum_algorithm: "sha256",
            checksum: "b94d27b9934d3e08a52e52d7da7dabfadebf1fde558f6ad0845e1274f7f9cde9",
          },
        }]}
        projectPackages={[
          {
            kind: "project_package",
            title: "Launch microsite",
            summary: "Reviewable output package",
            folder: "generated/launch",
            entrypoint: "index.html",
            files: ["index.html"],
            validation: "Smoke test passed",
          },
        ]}
      />,
    );

    expect(screen.getByTestId("output-workbench")).toBeDefined();
    expect(screen.getByText("Launch microsite")).toBeDefined();
    expect(screen.getByText("Reviewable output package")).toBeDefined();
    expect(screen.getByText("entry: index.html")).toBeDefined();
    expect(screen.getByText("Smoke test passed")).toBeDefined();
    expect(screen.getByText("path verified")).toBeDefined();
    expect(screen.getByText("readback verified")).toBeDefined();
    expect(screen.getByText("sha256 b94d27b9934d")).toBeDefined();
    expect(screen.getByRole("button", { name: /Open Launch brief/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Open local folder for Launch microsite/i })).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Copy output quote for Launch brief/i }));

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith("> Launch brief\n/runs/run-1");
      expect(screen.getByRole("button", { name: "Copied output quote" })).toBeDefined();
    });
  });
});
