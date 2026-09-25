import { describe, it, expect } from "vitest";
import { classifyServicesStatusFailure, deriveCoreReachability } from "@/lib/coreReachability";

const now = new Date().toISOString();
const healthy = [
    { name: "nats", status: "online" as const },
    { name: "postgres", status: "online" as const },
    { name: "cognitive", status: "online" as const },
];

describe("deriveCoreReachability", () => {
    it("is unknown before any health check", () => {
        expect(deriveCoreReachability({})).toBe("unknown");
    });

    it("is unreachable on network error or proxy 5xx", () => {
        expect(deriveCoreReachability({ servicesStatusFailure: classifyServicesStatusFailure() })).toBe("unreachable");
        expect(deriveCoreReachability({ servicesStatusFailure: classifyServicesStatusFailure(500) })).toBe("unreachable");
    });

    it("is degraded on a 4xx health response", () => {
        expect(deriveCoreReachability({ servicesStatusFailure: classifyServicesStatusFailure(401) })).toBe("degraded");
    });

    it("is degraded when a core service or the stream is down", () => {
        expect(
            deriveCoreReachability({
                servicesStatusUpdatedAt: now,
                servicesStatus: [...healthy.slice(1), { name: "nats", status: "offline" }],
            }),
        ).toBe("degraded");
        expect(
            deriveCoreReachability({ servicesStatusUpdatedAt: now, servicesStatus: healthy, streamConnectionState: "offline" }),
        ).toBe("degraded");
    });

    it("is online when all core services report online", () => {
        expect(
            deriveCoreReachability({ servicesStatusUpdatedAt: now, servicesStatus: healthy, streamConnectionState: "online" }),
        ).toBe("online");
    });
});
