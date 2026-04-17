import path from "path";
import { describe, it, expect } from "vitest";
import { resolveDocFilePath } from "@/lib/docsPathResolver";

describe("docs-api route path resolution", () => {
    it("finds repo-root docs from an interface-like working directory", () => {
        const resolved = resolveDocFilePath("docs/user/README.md", [
            path.join(process.cwd(), "interface"),
        ]);

        expect(resolved.replace(/\\/g, "/")).toMatch(/docs\/user\/README\.md$/);
    });

    it("finds repo-root docs from a nested build output directory", () => {
        const resolved = resolveDocFilePath("docs/user/README.md", [
            path.join(process.cwd(), "interface", ".next", "server", "app", "docs-api", "[slug]"),
        ]);

        expect(resolved.replace(/\\/g, "/")).toMatch(/docs\/user\/README\.md$/);
    });
});
