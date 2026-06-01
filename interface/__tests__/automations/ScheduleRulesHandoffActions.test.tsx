import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { TriggerRule } from "@/store/useCortexStore";

const mockFetchTriggerRules = vi.fn();
const mockCreateTriggerRule = vi.fn();
const mockToggleTriggerRule = vi.fn();
const mockResolveScheduleHandoff = vi.fn();
let triggerRules: TriggerRule[] = [];

vi.mock("@/store/useCortexStore", () => ({
    useCortexStore: (selector: any) =>
        selector({
            triggerRules,
            isFetchingTriggers: false,
            fetchTriggerRules: mockFetchTriggerRules,
            createTriggerRule: mockCreateTriggerRule,
            toggleTriggerRule: mockToggleTriggerRule,
            resolveScheduleHandoff: mockResolveScheduleHandoff,
        }),
}));

import ScheduleRulesTab from "@/components/automations/ScheduleRulesTab";

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

function awaitingRule(patch: Partial<TriggerRule> = {}) {
    return buildScheduleRule({
        id: "r-awaiting",
        name: "Awaiting handoff review",
        latest_execution: {
            id: "exec-awaiting",
            rule_id: "r-awaiting",
            event_id: "schedule:r-awaiting:due",
            status: "proposed",
            proposal_status: "awaiting_approval",
            handoff_key: "schedule:r-awaiting:due",
            handoff_payload: { autonomous_execution: false },
            executed_at: now,
        },
        ...patch,
    });
}

describe("ScheduleRulesTab handoff actions", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        triggerRules = [];
    });

    it.each([
        ["Approve handoff", "approved"],
        ["Reject", "rejected"],
        ["Cancel", "cancelled"],
    ] as const)("resolves a persisted schedule handoff as %s without implying execution", async (buttonName, status) => {
        mockResolveScheduleHandoff.mockResolvedValue(undefined);
        triggerRules = [awaitingRule()];

        await act(async () => { render(<ScheduleRulesTab />); });

        expect(screen.getByText(/approval records trust intent only/i)).toBeDefined();
        await act(async () => { fireEvent.click(screen.getByRole("button", { name: buttonName })); });

        expect(mockResolveScheduleHandoff).toHaveBeenCalledWith("r-awaiting", "exec-awaiting", status);
    });

    it("disables schedule handoff controls while a resolution is in flight", async () => {
        let resolveHandoff: () => void = () => {};
        mockResolveScheduleHandoff.mockReturnValue(new Promise<void>((resolve) => {
            resolveHandoff = resolve;
        }));
        triggerRules = [awaitingRule()];

        await act(async () => { render(<ScheduleRulesTab />); });
        await act(async () => { fireEvent.click(screen.getByRole("button", { name: "Reject" })); });

        for (const name of ["Approve handoff", "Reject", "Cancel"]) {
            expect((screen.getByRole("button", { name }) as HTMLButtonElement).disabled).toBe(true);
        }
        expect(mockResolveScheduleHandoff).toHaveBeenCalledWith("r-awaiting", "exec-awaiting", "rejected");

        await act(async () => { resolveHandoff(); });
        await waitFor(() => {
            expect((screen.getByRole("button", { name: "Reject" }) as HTMLButtonElement).disabled).toBe(false);
        });
    });

    it("hides controls when the latest execution is not awaiting approval", async () => {
        triggerRules = [awaitingRule({
            id: "r-approved",
            name: "Approved handoff review",
            latest_execution: {
                id: "exec-approved",
                rule_id: "r-approved",
                event_id: "schedule:r-approved:due",
                status: "proposed",
                proposal_status: "approved",
                handoff_key: "schedule:r-approved:due",
                handoff_payload: { autonomous_execution: false },
                executed_at: now,
            },
        })];

        await act(async () => { render(<ScheduleRulesTab />); });

        expect(screen.getByText("handoff approved")).toBeDefined();
        expect(screen.queryByRole("button", { name: "Approve handoff" })).toBeNull();
        expect(screen.queryByRole("button", { name: "Reject" })).toBeNull();
        expect(screen.queryByRole("button", { name: "Cancel" })).toBeNull();
    });

    it("hides controls when awaiting approval has no persisted execution", async () => {
        triggerRules = [buildScheduleRule({
            id: "r-missing-execution",
            name: "Missing execution review",
            schedule_handoff_state: "awaiting_approval",
        })];

        await act(async () => { render(<ScheduleRulesTab />); });

        expect(screen.getByText("handoff awaiting approval")).toBeDefined();
        expect(screen.queryByRole("button", { name: "Approve handoff" })).toBeNull();
        expect(screen.queryByRole("button", { name: "Reject" })).toBeNull();
        expect(screen.queryByRole("button", { name: "Cancel" })).toBeNull();
    });
});
