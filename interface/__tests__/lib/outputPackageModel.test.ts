import { describe, expect, it } from "vitest";
import {
  OUTPUT_PACKAGE_FOLDER_LABEL,
  OUTPUT_PACKAGE_OPEN_LABEL,
  OUTPUT_PACKAGE_RESOURCES_LABEL,
  joinWorkspacePath,
  outputCanvasHref,
  outputCanvasSourceHref,
  parentWorkspacePath,
  projectPackageOpenPath,
  projectPackageResourcesHref,
  projectPackageRevealPath,
  resourcesWorkspaceHref,
  somaReturnHref,
  workspaceBrowserPath,
  workspaceFileHref,
} from "@/lib/outputPackageModel";

describe("outputPackageModel", () => {
  it("uses canonical package action labels", () => {
    expect(OUTPUT_PACKAGE_OPEN_LABEL).toBe("Open file");
    expect(OUTPUT_PACKAGE_FOLDER_LABEL).toBe("Open folder");
    expect(OUTPUT_PACKAGE_RESOURCES_LABEL).toBe("Open in Resources");
  });

  it("normalizes package entrypoints into open, reveal, and resources paths", () => {
    expect(joinWorkspacePath("workspace/generated/coin-runner", "index.html")).toBe("workspace/generated/coin-runner/index.html");
    expect(joinWorkspacePath("workspace/generated/coin-runner", "workspace/generated/coin-runner/index.html")).toBe("workspace/generated/coin-runner/index.html");
    expect(parentWorkspacePath("workspace/generated/coin-runner/index.html")).toBe("workspace/generated/coin-runner");
    expect(projectPackageOpenPath({ folder: "workspace/generated/coin-runner", entrypoint: "index.html" })).toBe("workspace/generated/coin-runner/index.html");
    expect(projectPackageOpenPath({ folder: "workspace/generated/coin-runner", entrypoint: "dist/index.html" })).toBe("workspace/generated/coin-runner/dist/index.html");
    expect(projectPackageOpenPath({ folder: "workspace/generated/coin-runner", entrypoint: "workspace/generated/coin-runner/dist/index.html" })).toBe("workspace/generated/coin-runner/dist/index.html");
    expect(projectPackageRevealPath({ folder: "workspace/generated/coin-runner", entrypoint: "index.html" })).toBe("workspace/generated/coin-runner");
    expect(workspaceFileHref("workspace/generated/coin-runner/index.html")).toBe("/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fcoin-runner%2Findex.html");
    expect(resourcesWorkspaceHref("workspace/generated/coin-runner")).toBe("/resources?tab=workspace&path=workspace%2Fgenerated%2Fcoin-runner");
    expect(projectPackageResourcesHref({ folder: "workspace/generated/coin-runner", entrypoint: "index.html" })).toBe("/resources?tab=workspace&path=workspace%2Fgenerated%2Fcoin-runner");
    expect(workspaceBrowserPath("groups/game-team/generated/first-game")).toBe("workspace/groups/game-team/generated/first-game");
    expect(projectPackageResourcesHref({ folder: "groups/game-team/generated/first-game", entrypoint: "index.html" })).toBe("/resources?tab=workspace&path=workspace%2Fgroups%2Fgame-team%2Fgenerated%2Ffirst-game");
  });

  it("falls back to file paths when package metadata is sparse", () => {
    expect(projectPackageOpenPath({ filePath: "workspace/logs/playable.html" })).toBe("workspace/logs/playable.html");
    expect(projectPackageRevealPath({ filePath: "workspace/logs/playable.html" })).toBe("workspace/logs");
    expect(projectPackageResourcesHref({ filePath: "workspace/logs/playable.html" })).toBe("/resources?tab=workspace&path=workspace%2Flogs");
  });

  it("builds a retained-output canvas link with the exact Soma context", () => {
    const href = outputCanvasHref({
      label: "Playable game",
      url: "/api/v1/workspace/files/view?path=groups%2Fgame-team%2Fgenerated%2Fgame%2Findex.html",
      storagePath: "groups/game-team/generated/game/index.html",
      returnTo: "/dashboard?team_id=game-team&outcome_id=outcome-7#latest",
      proofArtifactId: "proof-7",
      teamId: "game-team",
      runId: "run-7",
      workItemId: "work-7",
      outputId: "output-7",
      contentDigest: "sha256:moonlit",
    });
    const parsed = new URL(href!, "http://mycelis.local");

    expect(parsed.pathname).toBe("/outputs/view");
    expect(parsed.searchParams.get("label")).toBe("Playable game");
    expect(parsed.searchParams.get("path")).toBe("groups/game-team/generated/game/index.html");
    expect(parsed.searchParams.get("return_to")).toBe("/dashboard?team_id=game-team&outcome_id=outcome-7#latest");
    expect(parsed.searchParams.get("proof")).toBe("proof-7");
    expect(parsed.searchParams.get("team_id")).toBe("game-team");
    expect(parsed.searchParams.get("run_id")).toBe("run-7");
    expect(parsed.searchParams.get("work_item_id")).toBe("work-7");
    expect(parsed.searchParams.get("output_id")).toBe("output-7");
    expect(parsed.searchParams.get("digest")).toBe("sha256:moonlit");
    expect(parsed.searchParams.get("source")).toBe("/api/v1/workspace/files/view?path=groups%2Fgame-team%2Fgenerated%2Fgame%2Findex.html");
  });

  it("limits the canvas to retained workspace files and safe Soma return URLs", () => {
    expect(outputCanvasSourceHref("https://example.com/untrusted.html")).toBeNull();
    expect(outputCanvasHref({ label: "External", url: "https://example.com/untrusted.html" })).toBeNull();
    expect(somaReturnHref("https://example.com/dashboard?team_id=other")).toBe("/dashboard");
    expect(somaReturnHref("/groups?group_id=other")).toBe("/dashboard");
  });
});
