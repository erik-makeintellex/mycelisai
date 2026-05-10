import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { SomaCausalSummary } from "@/components/soma/SomaCausalSummary";
import type { ChatMessage } from "@/store/useCortexStore";

describe("SomaCausalSummary", () => {
    it("summarizes the latest directed execution as one causal package", () => {
        const messages: ChatMessage[] = [
            { role: "user", content: "Prepare onboarding assets" },
            {
                role: "council",
                content: "Onboarding package is ready.",
                run_id: "run-abc-123",
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
        expect(screen.getByText("Causal package")).toBeDefined();
        expect(screen.getByText("Prepare a reviewed onboarding package")).toBeDefined();
        expect(screen.getByText(/Package the request for the operations team/i)).toBeDefined();
        expect(screen.getByText(/complete: directed_execution/i)).toBeDefined();
        expect(screen.getByText(/Teams: Operations Team/i)).toBeDefined();
        expect(screen.getByText(/Onboarding run package/i)).toBeDefined();
        expect(screen.getByText(/Onboarding brief/i)).toBeDefined();
        expect(screen.getByText(/Run run-abc-123/i)).toBeDefined();
        expect(screen.getByText(/Audit proof/i)).toBeDefined();
        expect(screen.getByText(/Recovery snapshot retained/i)).toBeDefined();
        expect(screen.getByText("Review the package with operators.")).toBeDefined();
    });
});
