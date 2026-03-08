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
});
