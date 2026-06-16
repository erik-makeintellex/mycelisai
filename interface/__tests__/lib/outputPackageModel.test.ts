import { describe, expect, it } from "vitest";
import {
  OUTPUT_PACKAGE_FOLDER_LABEL,
  OUTPUT_PACKAGE_OPEN_LABEL,
  OUTPUT_PACKAGE_RESOURCES_LABEL,
  joinWorkspacePath,
  parentWorkspacePath,
  projectPackageOpenPath,
  projectPackageResourcesHref,
  projectPackageRevealPath,
  resourcesWorkspaceHref,
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
});
