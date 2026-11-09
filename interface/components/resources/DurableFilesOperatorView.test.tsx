import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import DurableFilesOperatorView from "./DurableFilesOperatorView";
import { useCortexStore } from "@/store/useCortexStore";

describe("DurableFilesOperatorView Deliverable Iteration & Bundle Export", () => {
    beforeEach(() => {
        useCortexStore.setState({
            assistantName: "Soma",
            isFetchingArtifacts: false,
            fetchArtifacts: vi.fn(),
            toggleAdvancedMode: vi.fn(),
            artifacts: [
                {
                    id: "deliv-1",
                    title: "Sliding Tile Puzzle Game",
                    content_type: "text/html",
                    file_path: "workspace/deliverables/puzzle.html",
                    content: "<!doctype html><html><body>Puzzle</body></html>",
                    agent_id: "admin",
                    status: "approved",
                    created_at: "2026-04-12T10:00:00Z",
                } as any,
                {
                    id: "deliv-2",
                    title: "Game Rules & Controls",
                    content_type: "text/markdown",
                    file_path: "workspace/deliverables/rules.md",
                    content: "# Puzzle Rules\nUse arrow keys.",
                    agent_id: "admin",
                    status: "approved",
                    created_at: "2026-04-12T10:01:00Z",
                } as any,
            ],
            missionChat: [],
        });
    });

    it("exports all deliverables as a bundle and shows export notice", () => {
        render(<DurableFilesOperatorView />);

        const exportBtn = screen.getByTestId("export-deliverable-bundle");
        expect(exportBtn).toBeDefined();

        fireEvent.click(exportBtn);

        const notice = screen.getByTestId("durable-files-open-notice");
        expect(notice.textContent).toContain("Exported bundle (2 deliverables)");
    });

    it("opens inline refinement bar for a deliverable and sends refinement prompt to Soma", () => {
        render(<DurableFilesOperatorView />);

        const refineBtn = screen.getByTestId("refine-deliverable-deliv-1");
        expect(refineBtn).toBeDefined();
        fireEvent.click(refineBtn);

        const refineInput = screen.getByTestId("refine-input-deliv-1") as HTMLInputElement;
        expect(refineInput).toBeDefined();

        fireEvent.change(refineInput, {
            target: { value: "Add timer and 4x4 difficulty toggle" },
        });

        const submitBtn = screen.getByTestId("refine-submit-deliv-1");
        fireEvent.click(submitBtn);

        const notice = screen.getByTestId("durable-files-open-notice");
        expect(notice.textContent).toContain(
            'Refinement request sent to Soma for Sliding Tile Puzzle Game: "Add timer and 4x4 difficulty toggle"',
        );
        expect(screen.queryByTestId("refine-input-deliv-1")).toBeNull();
    });
});
