import path from "node:path";
import { afterEach, describe, expect, it, vi } from "vitest";
import { backendWorkspaceRoots } from "@/e2e/support/finalization-browser-package";

const repoRoot = path.resolve(__dirname, "../../..");

describe("finalization browser package workspace roots", () => {
    afterEach(() => {
        vi.unstubAllEnvs();
    });

    it("checks both task-runner and source-Core roots for a relative workspace", () => {
        vi.stubEnv("MYCELIS_WORKSPACE", "./workspace");

        expect(backendWorkspaceRoots()).toEqual([
            path.join(repoRoot, "workspace"),
            path.join(repoRoot, "core", "workspace"),
        ]);
    });

    it("keeps an explicitly absolute backend workspace exact", () => {
        vi.stubEnv("PLAYWRIGHT_BACKEND_WORKSPACE_ROOT", "/data/workspace");

        expect(backendWorkspaceRoots()).toEqual(["/data/workspace"]);
    });
});
