import { describe, expect, it } from "vitest";
import { buildMissionChatFailure } from "@/lib/missionChatFailure";

describe("buildMissionChatFailure", () => {
    it("builds a workspace server-error contract", () => {
        const failure = buildMissionChatFailure({
            assistantName: "Soma",
            targetId: "admin",
            message: "Soma chat blocked (500)",
            statusCode: 500,
        });

        expect(failure.routeKind).toBe("workspace");
        expect(failure.type).toBe("server_error");
        expect(failure.bannerLabel).toBe("Workspace chat server error");
        expect(failure.title).toBe("Soma Chat Blocked");
    });

    it("classifies a 503 workspace failure as unreachable when the route says it is unreachable", () => {
        const failure = buildMissionChatFailure({
            assistantName: "Soma",
            targetId: "admin",
            message: "Soma chat unreachable (503)",
            statusCode: 503,
        });

        expect(failure.type).toBe("unreachable");
        expect(failure.bannerLabel).toBe("Workspace chat unreachable");
    });

    it("builds a council timeout contract", () => {
        const failure = buildMissionChatFailure({
            assistantName: "Soma",
            targetId: "council-coder",
            message: "deadline exceeded",
        });

        expect(failure.routeKind).toBe("council");
        expect(failure.type).toBe("timeout");
        expect(failure.targetLabel).toBe("council-coder");
        expect(failure.recommendedAction).toMatch(/continue with Soma/i);
    });

    it("builds a setup-required contract from structured availability data", () => {
        const failure = buildMissionChatFailure({
            assistantName: "Soma",
            targetId: "admin",
            message: "Soma does not have an available cognitive engine right now.",
            statusCode: 503,
            availability: {
                code: "provider_disabled",
                summary: "Soma is routed to an AI Engine that is configured but disabled.",
                recommended_action: "Open Settings and enable a reachable AI Engine for Soma.",
                setup_required: true,
                setup_path: "/settings",
            },
        });

        expect(failure.type).toBe("setup_required");
        expect(failure.title).toBe("Soma Setup Required");
        expect(failure.bannerLabel).toBe("AI engine setup required");
        expect(failure.setupPath).toBe("/settings");
    });
});
