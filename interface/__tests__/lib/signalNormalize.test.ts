import { describe, expect, it } from "vitest";
import { normalizeIncomingSignal, streamSignalToDetail } from "@/lib/signalNormalize";

describe("signal normalization", () => {
    it("preserves legacy stream signals", () => {
        const signal = normalizeIncomingSignal({
            type: "artifact",
            source: "scanner-1",
            message: "artifact ready",
            timestamp: "2026-03-06T10:00:00Z",
            payload: { title: "Report" },
            topic: "swarm.team.alpha.signal.status",
        });

        expect(signal.type).toBe("artifact");
        expect(signal.source).toBe("scanner-1");
        expect(signal.message).toBe("artifact ready");
        expect(signal.topic).toBe("swarm.team.alpha.signal.status");
    });

    it("normalizes standardized signal envelopes", () => {
        const signal = normalizeIncomingSignal({
            meta: {
                timestamp: "2026-03-06T10:00:00Z",
                source_kind: "system",
                source_channel: "swarm.team.alpha.signal.result",
                payload_kind: "result",
                team_id: "alpha",
                run_id: "run-123",
            },
            text: "execution complete",
            payload: { summary: "done" },
        });

        expect(signal.type).toBe("result");
        expect(signal.source).toBe("alpha");
        expect(signal.message).toBe("execution complete");
        expect(signal.topic).toBe("swarm.team.alpha.signal.result");
        expect(signal.payload_kind).toBe("result");
        expect(signal.source_kind).toBe("system");
        expect(signal.team_id).toBe("alpha");
        expect(signal.run_id).toBe("run-123");
    });

    it("carries normalized metadata into signal detail", () => {
        const detail = streamSignalToDetail(
            normalizeIncomingSignal({
                meta: {
                    timestamp: "2026-03-06T10:00:00Z",
                    source_kind: "web_api",
                    source_channel: "swarm.team.beta.signal.status",
                    payload_kind: "status",
                    team_id: "beta",
                    agent_id: "beta-agent",
                },
                payload: { state: "ready" },
            }),
        );

        expect(detail.type).toBe("status");
        expect(detail.source).toBe("beta");
        expect(detail.payload_kind).toBe("status");
        expect(detail.source_kind).toBe("web_api");
        expect(detail.team_id).toBe("beta");
        expect(detail.agent_id).toBe("beta-agent");
        expect(detail.source_channel).toBe("swarm.team.beta.signal.status");
    });
});
