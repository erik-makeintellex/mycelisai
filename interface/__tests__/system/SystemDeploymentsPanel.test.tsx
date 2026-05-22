import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import SystemDeploymentsPanel from "@/components/system/SystemDeploymentsPanel";

describe("SystemDeploymentsPanel", () => {
    beforeEach(() => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue(
                new Response(
                    JSON.stringify({
                        ok: true,
                        data: {
                            deployment_root: "E:/mycelis",
                            execution_root: "E:/mycelis/core",
                            workspace_root: "E:/mycelis/core/workspace",
                            artifact_root: "E:/mycelis/core/workspace/artifacts",
                            current_commit: "20a8a070",
                            image_tag: "unknown",
                            chart_version: "unknown",
                            deployment_lane: "native-source",
                            endpoint_posture: "local",
                            runtime_health: {
                                status: "online",
                                online: 4,
                                degraded: 0,
                                offline: 0,
                                total: 4,
                            },
                            proof_lane: "local",
                            recovery_posture: "ready",
                            checked_at: "2026-05-22T20:00:00.000Z",
                        },
                    }),
                    { status: 200, headers: { "Content-Type": "application/json" } },
                ),
            ),
        );
        Object.defineProperty(navigator, "clipboard", {
            value: { writeText: vi.fn() },
            configurable: true,
        });
    });

    it("makes output root configuration visible without logs", async () => {
        render(<SystemDeploymentsPanel />);

        await waitFor(() => {
            expect(screen.getByText("Output root configuration")).toBeDefined();
        });

        expect(screen.getByText("Workspace files")).toBeDefined();
        expect(screen.getByText("MYCELIS_WORKSPACE")).toBeDefined();
        expect(screen.getAllByText("E:/mycelis/core/workspace").length).toBeGreaterThan(0);

        expect(screen.getByText("Artifacts and media")).toBeDefined();
        expect(screen.getByText("MYCELIS_ARTIFACT_ROOT")).toBeDefined();
        expect(screen.getAllByText("E:/mycelis/core/workspace/artifacts").length).toBeGreaterThan(0);

        expect(screen.getByText("Compose/Helm output block")).toBeDefined();
        expect(screen.getByText("MYCELIS_OUTPUT_HOST_PATH")).toBeDefined();
    });
});
