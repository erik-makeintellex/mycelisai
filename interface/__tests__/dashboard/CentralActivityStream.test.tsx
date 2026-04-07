import { describe, it, expect, beforeEach, vi } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import CentralActivityStream from "@/components/dashboard/CentralActivityStream";
import { useCortexStore } from "@/store/useCortexStore";

describe("CentralActivityStream", () => {
    beforeEach(() => {
        useCortexStore.setState({
            streamLogs: [],
            teamsDetail: [],
            fetchTeamsDetail: vi.fn().mockResolvedValue(undefined),
            selectSignalDetail: vi.fn(),
        });
    });

    it("shows live stream entries and filters by team and aspect", async () => {
        const selectSignalDetail = vi.fn();
        useCortexStore.setState({
            selectSignalDetail,
            teamsDetail: [
                {
                    id: "marketing-core",
                    name: "Marketing Core",
                    role: "action",
                    type: "standing",
                    mission_id: null,
                    mission_intent: null,
                    inputs: [],
                    deliveries: [],
                    agents: [],
                },
                {
                    id: "research-core",
                    name: "Research Core",
                    role: "action",
                    type: "standing",
                    mission_id: null,
                    mission_intent: null,
                    inputs: [],
                    deliveries: [],
                    agents: [],
                },
            ],
            streamLogs: [
                {
                    type: "artifact",
                    source: "marketing-lead",
                    message: "Launch brief ready for review",
                    timestamp: new Date().toISOString(),
                    team_id: "marketing-core",
                    source_channel: "swarm.team.marketing-core.signal.result",
                },
                {
                    type: "error",
                    source: "research-lead",
                    message: "Research crawl failed",
                    timestamp: new Date().toISOString(),
                    team_id: "research-core",
                    source_channel: "swarm.team.research-core.signal.status",
                },
            ],
        });

        render(<CentralActivityStream />);

        expect(screen.getByText("Live team interaction stream")).toBeDefined();
        expect(screen.getByText("Launch brief ready for review")).toBeDefined();
        expect(screen.getByText("Research crawl failed")).toBeDefined();

        const clickChecklistOption = (testId: string, labelText: string) => {
            const label = within(screen.getByTestId(testId)).getByText(labelText).closest("label");
            const checkbox = label?.querySelector("input[type='checkbox']") as HTMLInputElement | null;
            expect(checkbox).not.toBeNull();
            fireEvent.click(checkbox!);
        };

        clickChecklistOption("activity-team-filter", "Marketing Core");

        await waitFor(() => {
            expect(screen.getByText("Launch brief ready for review")).toBeDefined();
            expect(screen.queryByText("Research crawl failed")).toBeNull();
        });

        fireEvent.click(within(screen.getByTestId("activity-team-filter")).getByText("Clear", { selector: "button" }));
        clickChecklistOption("activity-aspect-filter", "Errors");

        await waitFor(() => {
            expect(screen.getByText("Research crawl failed")).toBeDefined();
            expect(screen.queryByText("Launch brief ready for review")).toBeNull();
        });

        fireEvent.click(screen.getByRole("button", { name: /Research Core/i }));
        expect(selectSignalDetail).toHaveBeenCalled();
    });
});
