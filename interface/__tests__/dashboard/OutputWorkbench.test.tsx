import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import {
  OutputWorkbench,
  mergeOutputWorkbenchItems,
  outputWorkbenchItems,
  projectPackageOutputs,
  teamOutputProjectPackages,
  teamOutputWorkbenchItems,
} from "@/components/soma/OutputWorkbench";
import { outputWorkbenchDigest } from "@/components/soma/OutputWorkbenchDigest";
import type { ExecutionSummaryData, TeamOutputRef } from "@/store/useCortexStore";

describe("OutputWorkbench", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

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
        created_at: "2026-05-17T18:00:00Z",
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
        created_at: "2026-05-17T18:01:00Z",
      },
    ];

    const durableItems = teamOutputWorkbenchItems(teamOutputs);
    expect(durableItems).toEqual([
      {
        text: "Launch brief",
        url: "/api/v1/workspace/files/view?path=generated%2Flaunch%2Fbrief.md",
        storagePath: "generated/launch/brief.md",
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

  it("orders durable team outputs newest-first for the workbench primary output", () => {
    const durableItems = teamOutputWorkbenchItems([
      {
        output_id: "out-old",
        team_id: "team-alpha",
        work_item_id: "work-1",
        kind: "file",
        label: "Older focused brief",
        storage_ref: "generated/launch/older.md",
        created_at: "2026-05-17T18:00:00Z",
      },
      {
        output_id: "out-new",
        team_id: "team-alpha",
        work_item_id: "work-2",
        kind: "file",
        label: "Newest focused brief",
        storage_ref: "generated/launch/newest.md",
        created_at: "2026-05-17T18:05:00Z",
      },
    ]);

    expect(durableItems.map((item) => item.text)).toEqual([
      "Newest focused brief",
      "Older focused brief",
    ]);
    expect(outputWorkbenchDigest({ outputs: durableItems })).toMatchObject({
      text: "Newest focused brief",
    });
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
        storagePath: "saved-media/comic-page.png",
      },
    ]);
  });

  it("selects a compact digest from the latest actionable output", () => {
    expect(outputWorkbenchDigest({
      outputs: [
        { text: "Proof note", url: null, proofArtifactId: "proof-1" },
        { text: "Owner note", url: "/api/v1/workspace/files/view?path=generated%2Fowner-note.md" },
      ],
      projectPackages: [{
        kind: "project_package",
        title: "Owner package",
        folder: "generated/owner-package",
      }],
    })).toEqual({
      text: "Owner note",
      url: "/api/v1/workspace/files/view?path=generated%2Fowner-note.md",
      storagePath: "generated/owner-note.md",
      count: 3,
    });

    expect(outputWorkbenchDigest({
      outputs: [
        {
          text: "New team folder",
          url: "/api/v1/workspace/files/view?path=groups%2Fnew-team",
          storagePath: "groups/new-team",
        },
        {
          text: "Playable output",
          url: "/api/v1/workspace/files/view?path=workspace%2Flogs%2Fplayable.html",
          storagePath: "workspace/logs/playable.html",
        },
      ],
    })).toEqual({
      text: "Playable output",
      url: "/api/v1/workspace/files/view?path=workspace%2Flogs%2Fplayable.html",
      storagePath: "workspace/logs/playable.html",
      count: 2,
    });

    expect(outputWorkbenchDigest({
      outputs: [],
      projectPackages: [{
        kind: "project_package",
        title: "Owner package",
        folder: "generated/owner-package",
      }],
    })).toEqual({
      text: "Owner package",
      url: null,
      storagePath: "generated/owner-package",
      count: 1,
    });

    const packageDigest = outputWorkbenchDigest({
      outputs: [{ text: "Game team folder", url: "/api/v1/workspace/files/view?path=groups%2Fgame-team", storagePath: "groups/game-team" }],
      projectPackages: [{ kind: "project_package", title: "Playable game", folder: "groups/game-team/generated/first-game", entrypoint: "groups/game-team/generated/first-game/index.html" }],
    });
    expect(packageDigest).toEqual({
      text: "Playable game",
      url: "/api/v1/workspace/files/view?path=groups%2Fgame-team%2Fgenerated%2Ffirst-game%2Findex.html",
      storagePath: "groups/game-team/generated/first-game",
      count: 2,
    });
  });

  it("prioritizes generated files over team folders in the main output workbench", () => {
    render(
      <OutputWorkbench
        outputs={[
          {
            text: "New team folder",
            url: "/api/v1/workspace/files/view?path=groups%2Fnew-team",
            storagePath: "groups/new-team",
          },
          {
            text: "Playable output",
            url: "/api/v1/workspace/files/view?path=workspace%2Flogs%2Fplayable.html",
            storagePath: "workspace/logs/playable.html",
          },
        ]}
      />,
    );

    expect(screen.getByText("Latest output").closest("article")?.textContent).toContain("Playable output");
    expect(screen.getByText("Use Open file to view it, or Open folder to show it in the workspace.")).toBeDefined();
    expect(screen.getByRole("button", { name: /Open file Playable output/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Open local folder for Playable output/i })).toBeDefined();
    expect(screen.getByText("More outputs and verification")).toBeDefined();
  });

  it("renders package actions and copyable output quotes", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    const fetchMock = vi.fn(async () => Response.json({ ok: true, data: { workspace_path: "generated/launch" } }));
    vi.stubGlobal("fetch", fetchMock);
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
    expect(screen.getByText("Workspace folder")).toBeDefined();
    expect(screen.getByText("generated/launch")).toBeDefined();
    expect(screen.getAllByText("index.html").length).toBeGreaterThan(0);
    expect(screen.getByText("Smoke test passed")).toBeDefined();
    expect(screen.getByText("Latest output")).toBeDefined();
    expect(screen.getByText("Use Open file to view it, or Open folder to show it in the workspace.")).toBeDefined();
    const verificationDetails = screen.getByText("Verification details").closest("details");
    expect(verificationDetails?.open).toBe(false);
    expect(screen.getByText("path verified")).toBeDefined();
    expect(screen.getByText("readback verified")).toBeDefined();
    expect(screen.getByText("sha256 b94d27b9934d")).toBeDefined();
    expect(screen.getByRole("button", { name: /Open file Launch brief/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Open local folder for Launch microsite/i })).toBeDefined();

    fireEvent.click(screen.getByRole("button", { name: /Open local folder for Launch microsite/i }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith("/api/v1/workspace/files/reveal?path=generated%2Flaunch", { method: "POST" });
      expect(screen.getByText("Folder opened")).toBeDefined();
    });

    fireEvent.click(screen.getByRole("button", { name: /Copy output quote for Launch brief/i }));

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith("> Launch brief\n/runs/run-1");
      expect(screen.getByRole("button", { name: "Copied output quote" })).toBeDefined();
    });
  });
});
