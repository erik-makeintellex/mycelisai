import { describe, expect, it } from "vitest";
import { normalizeIncomingSignal } from "@/lib/signalNormalize";
import { activeWorkContextFromMessages, chatMessageFromThreadSignal } from "@/store/cortexStoreThreadEvents";
import { executionStartedEvent } from "@/store/cortexStoreProposalThreadEvents";

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

    it("retains active team work identity without exposing it as a new interface object", () => {
        const event = executionStartedEvent("11111111-1111-4111-8111-111111111111", [{
            team_id: "launch-team",
            work_item_id: "22222222-2222-4222-8222-222222222222",
            state: "running",
        }]);

        expect(event).toMatchObject({
            run_id: "11111111-1111-4111-8111-111111111111",
            team_id: "launch-team",
            work_item_id: "22222222-2222-4222-8222-222222222222",
        });
        expect(activeWorkContextFromMessages([{
            role: "system",
            content: "Work started",
            mode: "execution_result",
            run_id: event.run_id,
            thread_events: [event],
        }])).toEqual({
            type: "team_work",
            id: "22222222-2222-4222-8222-222222222222",
            run_id: "11111111-1111-4111-8111-111111111111",
            team_id: "launch-team",
            work_item_id: "22222222-2222-4222-8222-222222222222",
        });
    });

    it("stops attaching conversation after the correlated result is ready", () => {
        expect(activeWorkContextFromMessages([{
            role: "system",
            content: "Work complete",
            mode: "execution_result",
            run_id: "11111111-1111-4111-8111-111111111111",
            thread_event: {
                kind: "result_ready",
                label: "Work complete",
                tone: "success",
                run_id: "11111111-1111-4111-8111-111111111111",
                team_id: "launch-team",
                work_item_id: "22222222-2222-4222-8222-222222222222",
            },
        }])).toBeNull();
    });
});
