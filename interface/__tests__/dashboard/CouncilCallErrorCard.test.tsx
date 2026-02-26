import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import CouncilCallErrorCard from "@/components/dashboard/CouncilCallErrorCard";

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
                member="council-sentry"
                errorMessage="request timeout after 30s"
                onRetry={vi.fn()}
                onSwitchToSoma={vi.fn()}
                onContinueWithSoma={vi.fn()}
            />
        );

        expect(screen.getByText("Council Call Failed")).toBeDefined();
        expect(screen.getByText("timeout")).toBeDefined();
    });

    it("fires retry, switch, and continue handlers", () => {
        const onRetry = vi.fn();
        const onSwitchToSoma = vi.fn();
        const onContinueWithSoma = vi.fn();

        render(
            <CouncilCallErrorCard
                member="council-architect"
                errorMessage="Council agent unreachable (500)"
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
                member="council-coder"
                errorMessage="failed to fetch"
                onRetry={vi.fn()}
                onSwitchToSoma={vi.fn()}
                onContinueWithSoma={vi.fn()}
            />
        );

        fireEvent.click(screen.getByText("Copy Diagnostics"));
        expect(navigator.clipboard.writeText).toHaveBeenCalledWith("failed to fetch");
    });
});

