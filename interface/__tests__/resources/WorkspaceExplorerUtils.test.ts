import { describe, expect, it } from "vitest";
import { parseListOutput } from "@/components/resources/WorkspaceExplorerUtils";

describe("WorkspaceExplorerUtils", () => {
    it("dedupes listing rows and drops structured metadata fragments", () => {
        const entries = parseListOutput(
            [
                "[FILE] index.html",
                "[FILE] index.html",
                '"exchange_item_id": "a3f9dc91-afae-4dbf-8dc6-88b8e52574cc",',
                '"run_id": "run-1",',
                "{",
                "}",
                "[DIR] assets",
            ].join("\n"),
            "workspace/groups/trusted-outcome-live/generated/first-game",
        );

        expect(entries).toEqual([
            {
                name: "index.html",
                path: "workspace/groups/trusted-outcome-live/generated/first-game/index.html",
                type: "file",
            },
            {
                name: "assets",
                path: "workspace/groups/trusted-outcome-live/generated/first-game/assets",
                type: "dir",
            },
        ]);
    });
});
