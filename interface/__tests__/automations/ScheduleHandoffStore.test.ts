import { beforeEach, describe, expect, it, vi } from "vitest";
import { mockFetch } from "../setup";
import { resetCortexStore } from "../store/useCortexStoreTestSupport";
import { useCortexStore, type TriggerRule } from "@/store/useCortexStore";

const now = new Date().toISOString();

function buildScheduleRule(patch: Partial<TriggerRule>): TriggerRule {
    return {
        id: "r-schedule",
        tenant_id: "default",
        name: "Weekly evidence review",
        trigger_kind: "schedule",
        event_pattern: "scheduler.due",
        condition: {},
        target_mission_id: "mission-review",
        mode: "propose",
        cooldown_seconds: 3600,
        schedule_interval_seconds: 3600,
        next_run_at: now,
        proof_expectations: "Visible audit and retained proof",
        recovery_behavior: "Pause and inspect the failed proposal",
        max_depth: 5,
        max_active_runs: 1,
        is_active: true,
        created_at: now,
        updated_at: now,
        ...patch,
    };
}

describe("schedule handoff store behavior", () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it("posts handoff resolution status and updates the matching schedule rule from the API response", async () => {
        const execution = {
            id: "exec-awaiting",
            rule_id: "r-awaiting",
            event_id: "schedule:r-awaiting:due",
            status: "proposed" as const,
            proposal_status: "approved" as const,
            schedule_handoff_state: "approved" as const,
            handoff_key: "schedule:r-awaiting:due",
            handoff_payload: { autonomous_execution: false },
            executed_at: now,
        };
        useCortexStore.setState({
            triggerRules: [
                buildScheduleRule({
                    id: "r-awaiting",
                    latest_execution: {
                        ...execution,
                        proposal_status: "awaiting_approval",
                        schedule_handoff_state: "awaiting_approval",
                    },
                }),
                buildScheduleRule({ id: "r-other", name: "Other schedule" }),
            ],
        });
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true, data: { execution } }),
        });

        await useCortexStore.getState().resolveScheduleHandoff("r-awaiting", "exec-awaiting", "approved");

        expect(mockFetch).toHaveBeenCalledWith(
            "/api/v1/triggers/r-awaiting/history/exec-awaiting/approval",
            expect.objectContaining({
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ status: "approved" }),
            }),
        );
        const [updatedRule, otherRule] = useCortexStore.getState().triggerRules;
        expect(updatedRule.schedule_handoff_state).toBe("approved");
        expect(updatedRule.latest_execution).toEqual(execution);
        expect(otherRule.id).toBe("r-other");
        expect(otherRule.schedule_handoff_state).toBeUndefined();
    });

    it("leaves schedule handoff state unchanged when the API rejects the transition", async () => {
        vi.spyOn(console, "error").mockImplementation(() => {});
        useCortexStore.setState({
            triggerRules: [
                buildScheduleRule({
                    id: "r-awaiting",
                    latest_execution: {
                        id: "exec-awaiting",
                        rule_id: "r-awaiting",
                        event_id: "schedule:r-awaiting:due",
                        status: "proposed",
                        proposal_status: "awaiting_approval",
                        schedule_handoff_state: "awaiting_approval",
                        handoff_payload: {},
                        executed_at: now,
                    },
                }),
            ],
        });
        mockFetch.mockResolvedValue({
            ok: false,
            text: async () => "transition denied",
        });

        await useCortexStore.getState().resolveScheduleHandoff("r-awaiting", "exec-awaiting", "rejected");

        const rule = useCortexStore.getState().triggerRules[0];
        expect(rule.schedule_handoff_state).toBeUndefined();
        expect(rule.latest_execution?.proposal_status).toBe("awaiting_approval");
        expect(rule.latest_execution?.schedule_handoff_state).toBe("awaiting_approval");
    });
});
