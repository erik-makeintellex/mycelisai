import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { formatHumanActivityEvent, formatProposalHeadline } from "@/lib/labels";
import EventCard from "@/components/runs/EventCard";
import type { MissionEvent } from "@/store/useCortexStore";

describe("Activity Jargon Translation Layer & EventCard", () => {
    it("translates delegate_task and tool_completed into human-readable activity sentences", () => {
        const delegated = formatHumanActivityEvent(
            {
                event_type: "tool.invoked",
                source_agent: "admin",
                source_team: "research-team",
                payload: { tool_name: "delegate_task", task: "Draft the Q3 executive brief" },
            },
            "Soma",
        );
        expect(delegated).toContain("Soma");
        expect(delegated).toContain("Assign to Team");

        const fileWritten = formatHumanActivityEvent(
            {
                event_type: "tool.completed",
                source_agent: "coder",
                payload: { tool: "write_file", path: "workspace/reports/q3.md" },
            },
            "Soma",
        );
        expect(fileWritten).toContain("Coder");
        expect(fileWritten).toContain("saved workspace/reports/q3.md to Durable Files");
    });

    it("translates proposals into a plain-English headline", () => {
        const headline = formatProposalHeadline(
            {
                intent: "delegate_task",
                teams: 1,
                agents: 3,
                tools: ["write_file", "read_file"],
                risk_level: "medium",
                confirm_token: "ct-123",
                intent_proof_id: "ip-123",
            },
            "Soma",
        );
        expect(headline).toContain("Soma proposes saving a deliverable to your workspace");
        expect(headline).not.toContain("write_file");
    });

    it("renders EventCard with human-friendly summary and hides raw JSON behind View Technical Envelope", () => {
        const event: MissionEvent = {
            id: "evt-1",
            run_id: "run-123",
            tenant_id: "default",
            event_type: "tool.completed",
            severity: "info",
            source_agent: "admin",
            source_team: "executive-council",
            payload: { tool_name: "write_file", path: "workspace/summary.md", secret_debug_key: "raw_value_99" },
            emitted_at: new Date().toISOString(),
        };

        render(<EventCard event={event} isLast={false} />);

        expect(screen.getByText(/saved workspace\/summary\.md to Durable Files/i)).toBeDefined();
        expect(screen.queryByText(/raw_value_99/)).toBeNull();

        const toggleBtn = screen.getByTestId("toggle-technical-envelope");
        expect(toggleBtn.textContent).toContain("View Technical Envelope");

        fireEvent.click(toggleBtn);
        expect(screen.getByText(/raw_value_99/)).toBeDefined();
    });
});
