import { beforeEach, describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

vi.mock("reactflow", async () => {
    const mock = await import("../mocks/reactflow");
    return mock;
});

vi.mock("@/components/automations/CapabilityReadinessGateCard", async () => {
    const React = await import("react");
    return {
        __esModule: true,
        default: ({ onSnapshotChange }: { onSnapshotChange?: (s: any) => void }) => {
            React.useEffect(() => {
                onSnapshotChange?.({
                    providerReady: true,
                    mcpReady: true,
                    governanceReady: true,
                    natsReady: true,
                    sseReady: true,
                    dbReady: true,
                    blockers: [],
                });
            }, [onSnapshotChange]);
            return <div data-testid="readiness-card">Readiness</div>;
        },
    };
});

import TeamInstantiationWizard from "@/components/automations/TeamInstantiationWizard";
import { useCortexStore } from "@/store/useCortexStore";

describe("TeamInstantiationWizard", () => {
    const createMissionProfile = vi.fn();
    const activateMissionProfile = vi.fn();
    const fetchMissionProfiles = vi.fn();

    beforeEach(() => {
        createMissionProfile.mockReset();
        activateMissionProfile.mockReset();
        fetchMissionProfiles.mockReset();
        createMissionProfile.mockResolvedValue({ id: "profile-1" });
        activateMissionProfile.mockResolvedValue(undefined);
        fetchMissionProfiles.mockResolvedValue(undefined);

        useCortexStore.setState({
            missionProfiles: [],
            createMissionProfile,
            activateMissionProfile,
            fetchMissionProfiles,
        });
    });

    it("requires objective text before step progression", () => {
        render(<TeamInstantiationWizard openTab={vi.fn()} />);
        const continueButton = screen.getByRole("button", { name: "Continue" });
        expect((continueButton as HTMLButtonElement).disabled).toBe(true);

        fireEvent.change(screen.getByPlaceholderText("Describe the outcome you want this team to execute."), {
            target: { value: "Build a governed delivery workflow for sprint zero." },
        });

        expect((continueButton as HTMLButtonElement).disabled).toBe(false);
    });

    it("reaches launch step and runs propose-only launch flow", async () => {
        const openTab = vi.fn();
        render(<TeamInstantiationWizard openTab={openTab} />);

        fireEvent.change(screen.getByPlaceholderText("Describe the outcome you want this team to execute."), {
            target: { value: "Build a governed delivery workflow for sprint zero." },
        });

        fireEvent.click(screen.getByRole("button", { name: "Continue" })); // profile
        fireEvent.click(screen.getByRole("button", { name: "Continue" })); // readiness
        fireEvent.click(screen.getByRole("button", { name: "Continue" })); // launch

        expect(screen.getByText("Launch review")).toBeDefined();
        expect(screen.getByRole("button", { name: "Launch Now" })).toBeDefined();
        expect(screen.getByRole("button", { name: "Launch Propose-Only" })).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Launch Propose-Only" }));
        await screen.findByText(/Propose-only profile activated/);
        expect(openTab).toHaveBeenCalledWith("approvals");
        expect(createMissionProfile).toHaveBeenCalledTimes(1);
        expect(activateMissionProfile).toHaveBeenCalledWith("profile-1");
    });
});
