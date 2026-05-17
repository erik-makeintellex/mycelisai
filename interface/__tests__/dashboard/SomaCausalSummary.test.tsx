import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { SomaCausalSummary } from "@/components/soma/SomaCausalSummary";
import type { ChatMessage } from "@/store/useCortexStore";

describe("SomaCausalSummary", () => {
    it("uses a neutral first-request state before Soma activity exists", () => {
        render(<SomaCausalSummary messages={[]} />);

        expect(screen.getByText("Ready for your first request")).toBeDefined();
        expect(screen.getByText("First request")).toBeDefined();
        expect(screen.queryByText("Soma just did this")).toBeNull();
        expect(screen.queryByText("Trust package")).toBeNull();
        expect(screen.queryByText("Proof")).toBeNull();
    });

    it("summarizes the latest directed execution as one causal package", () => {
        const writeText = vi.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, "clipboard", {
            value: { writeText },
            configurable: true,
        });

        const messages: ChatMessage[] = [
            { role: "user", content: "Prepare onboarding assets" },
            {
                role: "council",
                content: "Onboarding package is ready.",
                run_id: "run-abc-123",
                ui_response_state: {
                    kind: "execution_result",
                    label: "Output ready",
                    tone: "success",
                },
                artifacts: [
                    {
                        id: "artifact-1",
                        type: "document",
                        title: "Onboarding brief",
                    },
                ],
                execution_summary: {
                    intent: {
                        original: "Prepare onboarding assets",
                        resolved: "Prepare a reviewed onboarding package",
                    },
                    understanding: {
                        summary: "Package the request for the operations team.",
                        assumptions: ["Use the active organization context"],
                    },
                    execution: {
                        shape: "directed_execution",
                        status: "complete",
                        summary: "Soma coordinated the operations lane and produced reviewable output.",
                    },
                    capability_use: {
                        teams: ["Operations Team"],
                        capabilities: ["artifact.compose"],
                    },
                    outputs: [{ title: "Onboarding run package", url: "/runs/run-abc-123" }],
                    proof: [{ label: "Audit proof", url: "/proof/proof-123" }],
                    audit_recovery: "Recovery snapshot retained.",
                    next_step: "Review the package with operators.",
                },
            },
        ];

        render(<SomaCausalSummary messages={messages} />);

        expect(screen.getByText("Soma just did this")).toBeDefined();
        expect(screen.getByText("Trust package")).toBeDefined();
        expect(screen.getByText("Output ready")).toBeDefined();
        expect(screen.getByText(/Soma coordinated the operations lane and produced reviewable output/i)).toBeDefined();
        expect(screen.getByText(/Onboarding run package/i)).toBeDefined();
        expect(screen.getByText(/Onboarding brief/i)).toBeDefined();
        expect(screen.getByText(/Run run-abc-123/i)).toBeDefined();
        expect(screen.getByText(/Audit proof/i)).toBeDefined();
        expect(screen.getByText(/Recovery snapshot retained/i)).toBeDefined();
        expect(screen.getByText("Review the package with operators.")).toBeDefined();
        expect(screen.queryByText("Prepare a reviewed onboarding package")).toBeNull();

        fireEvent.click(screen.getByRole("button", { name: /Show trust details/i }));

        expect(screen.getByText("Prepare a reviewed onboarding package")).toBeDefined();
        expect(screen.getByText(/Package the request for the operations team/i)).toBeDefined();
        expect(screen.getByText(/complete: directed_execution/i)).toBeDefined();
        expect(screen.getByText(/Teams: Operations Team/i)).toBeDefined();

        fireEvent.click(screen.getAllByRole("button", { name: /Copy output quote for Onboarding run package, Onboarding brief/i })[0]);

        return waitFor(() => {
            expect(writeText).toHaveBeenCalledWith("> Onboarding run package, Onboarding brief");
            expect(screen.getAllByRole("button", { name: /Copied output quote/i }).length).toBeGreaterThan(0);
        });
    });
});
