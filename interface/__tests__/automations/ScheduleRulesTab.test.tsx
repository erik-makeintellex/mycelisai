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

describe("ScheduleRulesTab", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        triggerRules = [];
    });

    it("shows propose-only schedule rules with proof and recovery", async () => {
        triggerRules = [{
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
        }];

        await act(async () => { render(<ScheduleRulesTab />); });

        expect(screen.getByText("Weekly evidence review")).toBeDefined();
        expect(screen.getByText("propose only")).toBeDefined();
        expect(screen.getByText("Visible audit and retained proof")).toBeDefined();
        expect(screen.getByText("Pause and inspect the failed proposal")).toBeDefined();
        expect(mockFetchTriggerRules).toHaveBeenCalled();
    });

    it("creates a propose-only schedule rule", async () => {
        mockCreateTriggerRule.mockResolvedValue(null);
        await act(async () => { render(<ScheduleRulesTab />); });

        await act(async () => { fireEvent.click(screen.getByText("New Schedule")); });
        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Daily proof review" } });
        fireEvent.change(screen.getByLabelText("Target mission ID"), { target: { value: "mission-daily" } });

        await act(async () => { fireEvent.click(screen.getByText("Create Schedule")); });

        expect(mockCreateTriggerRule).toHaveBeenCalledWith(expect.objectContaining({
            trigger_kind: "schedule",
            event_pattern: "scheduler.due",
            mode: "propose",
            target_mission_id: "mission-daily",
        }));
    });

    it("shows disabled schedule state with resume action", async () => {
        triggerRules = [{
            id: "r-disabled",
            tenant_id: "default",
            name: "Paused quarterly evidence review",
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
            is_active: false,
            created_at: now,
            updated_at: now,
        }];

        await act(async () => { render(<ScheduleRulesTab />); });

        expect(screen.getByText("disabled")).toBeDefined();
        expect(screen.getByRole("button", { name: "Resume" })).toBeDefined();
    });

    it("surfaces explicit schedule handoff approval state variants when present", async () => {
        triggerRules = [
            buildScheduleRule({ id: "r-pending", name: "Pending review", schedule_handoff_state: "pending" }),
            buildScheduleRule({ id: "r-proposed", name: "Proposed review", handoff_status: "proposed" }),
            buildScheduleRule({ id: "r-approved", name: "Approved review", approval_state: "approved" }),
            buildScheduleRule({
                id: "r-rejected",
                name: "Rejected review",
                latest_execution: {
                    id: "exec-rejected",
                    rule_id: "r-rejected",
                    event_id: "schedule:r-rejected:due",
                    status: "proposed",
                    proposal_status: "rejected",
                    handoff_payload: {},
                    executed_at: now,
                },
            }),
            buildScheduleRule({
                id: "r-executed",
                name: "Executed review",
                latest_execution: {
                    id: "exec-executed",
                    rule_id: "r-executed",
                    event_id: "schedule:r-executed:due",
                    status: "fired",
                    schedule_handoff_state: "executed",
                    handoff_payload: {},
                    executed_at: now,
                },
            }),
        ];

        await act(async () => { render(<ScheduleRulesTab />); });

        expect(screen.getByText("handoff pending")).toBeDefined();
        expect(screen.getByText("handoff proposed")).toBeDefined();
        expect(screen.getByText("handoff approved")).toBeDefined();
        expect(screen.getByText("handoff rejected")).toBeDefined();
        expect(screen.getByText("handoff executed")).toBeDefined();
    });

    it("disables pause or resume while toggle is in flight", async () => {
        triggerRules = [{
            id: "r-pending",
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
        }];
        let resolveToggle: () => void = () => {};
        mockToggleTriggerRule.mockReturnValue(new Promise<void>((resolve) => {
            resolveToggle = resolve;
        }));

        await act(async () => { render(<ScheduleRulesTab />); });

        await act(async () => { fireEvent.click(screen.getByRole("button", { name: "Pause" })); });

        const pendingButton = screen.getByRole("button", { name: "Updating..." }) as HTMLButtonElement;
        expect(pendingButton.disabled).toBe(true);
        expect(mockToggleTriggerRule).toHaveBeenCalledWith("r-pending", false);

        await act(async () => { resolveToggle(); });
        await waitFor(() => expect(screen.getByRole("button", { name: "Pause" })).toBeDefined());
    });
});
