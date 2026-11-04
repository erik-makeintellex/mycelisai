import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import WorkspaceContextHeader from "@/components/dashboard/WorkspaceContextHeader";
import { somaPlaceholder } from "@/components/dashboard/missionControlChatUi";

describe("WorkspaceContextHeader & Contextual Placeholders", () => {
    it("frames the default Soma Executive Workspace with Standing Council Active", () => {
        render(<WorkspaceContextHeader assistantName="Soma" />);

        expect(screen.getByText("[Soma Executive Workspace]")).toBeDefined();
        expect(screen.getByText("Standing Council Active")).toBeDefined();
    });

    it("frames a focused team workspace with active agent count", () => {
        render(
            <WorkspaceContextHeader
                assistantName="Soma"
                currentTeamName="Launch Strategy"
                agentCount={4}
            />,
        );

        expect(screen.getByText("[Launch Strategy Workspace]")).toBeDefined();
        expect(screen.getByText("4 Agents Active")).toBeDefined();
    });

    it("updates somaPlaceholder dynamically for team focus vs executive workspace", () => {
        const teamPrompt = somaPlaceholder({
            assistantName: "Soma",
            broadcastMode: false,
            currentTeamName: "Launch Strategy",
            directTarget: null,
            showAdvancedRouting: false,
            simpleMode: true,
        });
        expect(teamPrompt).toContain("Direct the Launch Strategy workspace");

        const rootPrompt = somaPlaceholder({
            assistantName: "Soma",
            broadcastMode: false,
            currentTeamName: undefined,
            directTarget: null,
            showAdvancedRouting: false,
            simpleMode: true,
        });
        expect(rootPrompt).toContain("Describe the outcome you need from Soma");
    });
});
