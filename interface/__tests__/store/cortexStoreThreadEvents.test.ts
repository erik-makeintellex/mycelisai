import { describe, expect, it } from "vitest";
import { normalizeIncomingSignal } from "@/lib/signalNormalize";
import { chatMessageFromThreadSignal } from "@/store/cortexStoreThreadEvents";

describe("cortexStoreThreadEvents", () => {
    it("maps typed stream events into compact Soma thread messages", () => {
        const signal = normalizeIncomingSignal({
            type: "thread_event",
            meta: {
                timestamp: "2026-06-27T12:00:00Z",
                source_kind: "web_api",
                source_channel: "api.intent.confirm-action",
                payload_kind: "thread_event",
                run_id: "run-123",
                team_id: "team-alpha",
            },
            payload: {
                kind: "execution_started",
                label: "Execution started",
                detail: "Soma accepted the approved work.",
                tone: "info",
                status: "running",
                href: "/runs/run-123",
                href_label: "Open run receipt",
                target_reference: "run:run-123",
            },
        });

        const message = chatMessageFromThreadSignal(signal);

        expect(message).toMatchObject({
            role: "system",
            mode: "execution_result",
            run_id: "run-123",
            thread_events: [{
                kind: "execution_started",
                label: "Execution started",
                detail: "Soma accepted the approved work.",
                tone: "info",
                status: "running",
                href: "/runs/run-123",
                href_label: "Open run receipt",
                target_reference: "run:run-123",
                source_kind: "web_api",
                source_channel: "api.intent.confirm-action",
            }],
        });
    });

    it("ignores ordinary stream signals so Soma does not become a raw bus log", () => {
        const signal = normalizeIncomingSignal({
            meta: {
                source_kind: "system",
                source_channel: "swarm.team.alpha.signal.status",
                payload_kind: "status",
            },
            payload: { summary: "Raw team status." },
        });

        expect(chatMessageFromThreadSignal(signal)).toBeNull();
    });
});
