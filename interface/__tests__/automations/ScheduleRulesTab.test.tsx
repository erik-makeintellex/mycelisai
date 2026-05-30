import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { TriggerRule } from "@/store/useCortexStore";

const mockFetchTriggerRules = vi.fn();
const mockCreateTriggerRule = vi.fn();
const mockToggleTriggerRule = vi.fn();
let triggerRules: TriggerRule[] = [];

vi.mock("@/store/useCortexStore", () => ({
    useCortexStore: (selector: any) =>
        selector({
            triggerRules,
            isFetchingTriggers: false,
            fetchTriggerRules: mockFetchTriggerRules,
            createTriggerRule: mockCreateTriggerRule,
            toggleTriggerRule: mockToggleTriggerRule,
        }),
}));

import ScheduleRulesTab from "@/components/automations/ScheduleRulesTab";

const now = new Date().toISOString();

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
});
