import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { SomaCausalSummary } from "@/components/soma/SomaCausalSummary";
import type { ChatMessage } from "@/store/useCortexStore";

const user = (content: string): ChatMessage => ({ role: "user", content }) as ChatMessage;

describe("SomaCausalSummary", () => {
    it("shows a neutral empty state before the first turn and claims no work", () => {
        render(<SomaCausalSummary messages={[]} />);

        expect(screen.getByTestId("soma-causal-summary").getAttribute("data-phase")).toBe("idle");
        expect(screen.getByText(/Nothing yet/i)).toBeDefined();
        expect(screen.queryByText("Coordinated")).toBeNull();
        expect(screen.queryByText("Conversation guidance")).toBeNull();
        expect(screen.queryByText("Soma thread")).toBeNull();
    });

    it("shows a blocked summary after a failed turn instead of claiming coordination", () => {
        render(
            <SomaCausalSummary
                messages={[
                    user("Plan the launch"),
                    {
                        role: "council",
                        content: "Soma or the local API proxy is currently unreachable from this client.",
                        source_node: "admin",
                        mode: "blocker",
                    } as ChatMessage,
                ]}
            />,
        );

        expect(screen.getByTestId("soma-causal-summary").getAttribute("data-phase")).toBe("failed");
        expect(screen.getByText("Soma could not complete this")).toBeDefined();
        expect(screen.getByText("Plan the launch")).toBeDefined();
        expect(screen.getByText(/Blocked/)).toBeDefined();
        expect(screen.queryByText("Coordinated")).toBeNull();
        expect(screen.queryByText("Soma thread")).toBeNull();
    });

    it("shows a waiting state while a request is in flight", () => {
        render(<SomaCausalSummary messages={[user("Draft a memo")]} />);

        expect(screen.getByTestId("soma-causal-summary").getAttribute("data-phase")).toBe("pending");
        expect(screen.getByText(/Waiting for Soma/)).toBeDefined();
    });

    it("summarizes a successful answer with what was actually produced", () => {
        render(
            <SomaCausalSummary
                messages={[
                    user("Draft a memo"),
                    { role: "admin", content: "Here it is", source_node: "admin", mode: "answer" } as ChatMessage,
                ]}
            />,
        );

        expect(screen.getByText("Soma just did this")).toBeDefined();
        expect(screen.getByText("Coordinated")).toBeDefined();
        expect(screen.getByText("Response in thread")).toBeDefined();
    });
});
