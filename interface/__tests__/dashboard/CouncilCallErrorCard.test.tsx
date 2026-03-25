import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import CouncilCallErrorCard from "@/components/dashboard/CouncilCallErrorCard";
import { buildMissionChatFailure } from "@/lib/missionChatFailure";

describe("CouncilCallErrorCard", () => {
    beforeEach(() => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, "clipboard", {
            value: { writeText },
            configurable: true,
        });
    });

    it("classifies timeout failures", () => {
        render(
            <CouncilCallErrorCard
                failure={buildMissionChatFailure({
                    assistantName: "Soma",
                    targetId: "council-sentry",
                    message: "request timeout after 30s",
                })}
                onRetry={vi.fn()}
                onSwitchToSoma={vi.fn()}
                onContinueWithSoma={vi.fn()}
            />
        );

        expect(screen.getByText("Council Call Failed")).toBeDefined();
        expect(screen.getByText("timeout")).toBeDefined();
    });

    it("classifies mixed unreachable 500 errors as server errors", () => {
        render(
            <CouncilCallErrorCard
                failure={buildMissionChatFailure({
                    assistantName: "Soma",
                    targetId: "council-architect",
                    message: "Council agent unreachable (500)",
                    statusCode: 500,
                })}
                onRetry={vi.fn()}
                onSwitchToSoma={vi.fn()}
                onContinueWithSoma={vi.fn()}
            />
        );

        expect(screen.getByText("server_error")).toBeDefined();
        expect(screen.getByText(/returned an internal error/i)).toBeDefined();
    });

    it("fires retry, switch, and continue handlers", () => {
        const onRetry = vi.fn();
        const onSwitchToSoma = vi.fn();
        const onContinueWithSoma = vi.fn();

        render(
            <CouncilCallErrorCard
                failure={buildMissionChatFailure({
                    assistantName: "Soma",
                    targetId: "council-architect",
                    message: "Council agent unreachable (500)",
                    statusCode: 500,
                })}
                onRetry={onRetry}
                onSwitchToSoma={onSwitchToSoma}
                onContinueWithSoma={onContinueWithSoma}
            />
        );

        fireEvent.click(screen.getByText("Retry"));
        fireEvent.click(screen.getByText("Switch to Soma"));
        fireEvent.click(screen.getByText("Continue with Soma Only"));

        expect(onRetry).toHaveBeenCalledTimes(1);
        expect(onSwitchToSoma).toHaveBeenCalledTimes(1);
        expect(onContinueWithSoma).toHaveBeenCalledTimes(1);
    });

    it("copies diagnostics", async () => {
        render(
            <CouncilCallErrorCard
                failure={buildMissionChatFailure({
                    assistantName: "Soma",
                    targetId: "council-coder",
                    message: "failed to fetch",
                })}
                onRetry={vi.fn()}
                onSwitchToSoma={vi.fn()}
                onContinueWithSoma={vi.fn()}
            />
        );

        fireEvent.click(screen.getByText("Copy Diagnostics"));
        expect(navigator.clipboard.writeText).toHaveBeenCalledWith("failed to fetch");
    });

    it("offers settings navigation for setup-required blockers", () => {
        const assign = vi.fn();
        Object.defineProperty(window, "location", {
            value: { assign },
            configurable: true,
        });

        render(
            <CouncilCallErrorCard
                failure={buildMissionChatFailure({
                    assistantName: "Soma",
                    targetId: "admin",
                    message: "Soma does not have an available cognitive engine right now.",
                    availability: {
                        code: "provider_disabled",
                        summary: "Soma is routed to an AI Engine that is configured but disabled.",
                        recommended_action: "Open Settings and enable a reachable AI Engine for Soma.",
                        setup_required: true,
                        setup_path: "/settings",
                    },
                })}
                onRetry={vi.fn()}
                onSwitchToSoma={vi.fn()}
                onContinueWithSoma={vi.fn()}
            />
        );

        fireEvent.click(screen.getByText("Open Settings"));
        expect(assign).toHaveBeenCalledWith("/settings");
    });
});
