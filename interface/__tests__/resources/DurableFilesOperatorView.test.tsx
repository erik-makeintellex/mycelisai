import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import DurableFilesOperatorView, { isPlayableHtmlDeliverable } from "@/components/resources/DurableFilesOperatorView";
import { useCortexStore } from "@/store/useCortexStore";

describe("DurableFilesOperatorView", () => {
    const fetchArtifactsMock = vi.fn();
    const toggleAdvancedModeMock = vi.fn();

    beforeEach(() => {
        fetchArtifactsMock.mockReset();
        toggleAdvancedModeMock.mockReset();
        useCortexStore.setState({
            artifacts: [
                {
                    id: "art-puzzle-html",
                    agent_id: "coding-specialist",
                    artifact_type: "code",
                    title: "Sliding Tile Puzzle Game",
                    content_type: "text/html",
                    content: "<!DOCTYPE html><html><body><h1>Sliding Puzzle</h1><script>console.log('ready')</script></body></html>",
                    file_path: "workspace/deliverables/sliding-puzzle/index.html",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-12T12:00:00Z",
                },
                {
                    id: "art-launch-brief",
                    agent_id: "marketing-lead",
                    artifact_type: "document",
                    title: "Product Launch Brief",
                    content_type: "text/markdown",
                    content: "# Launch Positioning\nKey value propositions and FAQ.",
                    file_path: "workspace/deliverables/launch/brief.md",
                    metadata: {},
                    status: "approved",
                    created_at: "2026-04-12T12:05:00Z",
                },
            ],
            isFetchingArtifacts: false,
            fetchArtifacts: fetchArtifactsMock,
            missionChat: [],
            assistantName: "Soma",
            toggleAdvancedMode: toggleAdvancedModeMock,
        });
    });

    it("detects playable HTML deliverables by contentType or .html path", () => {
        expect(
            isPlayableHtmlDeliverable({
                id: "1",
                title: "Game",
                contentType: "text/html",
                producer: "Soma",
                saved: true,
            }),
        ).toBe(true);
        expect(
            isPlayableHtmlDeliverable({
                id: "2",
                title: "Game File",
                contentType: "text/plain",
                filePath: "workspace/game/index.html",
                producer: "Soma",
                saved: true,
            }),
        ).toBe(true);
        expect(
            isPlayableHtmlDeliverable({
                id: "3",
                title: "Brief",
                contentType: "text/markdown",
                filePath: "workspace/docs/brief.md",
                producer: "Soma",
                saved: true,
            }),
        ).toBe(false);
    });

    it("renders Play / Live Preview for HTML deliverables and mounts a sandboxed iframe when clicked", () => {
        render(<DurableFilesOperatorView />);

        expect(screen.getByTestId("durable-file-card-art-puzzle-html")).toBeDefined();
        expect(screen.getByTestId("durable-file-card-art-launch-brief")).toBeDefined();

        const playBtn = screen.getByTestId("play-deliverable-art-puzzle-html");
        expect(playBtn).toBeDefined();
        expect(playBtn.textContent).toContain("Play / Live Preview");

        // Non-HTML markdown deliverable should not show Play / Live Preview
        expect(screen.queryByTestId("play-deliverable-art-launch-brief")).toBeNull();

        // Clicking Play mounts the sandboxed iframe and opens preview
        fireEvent.click(playBtn);

        const iframe = screen.getByTestId("deliverable-iframe-art-puzzle-html") as HTMLIFrameElement;
        expect(iframe).toBeDefined();
        expect(iframe.getAttribute("sandbox")).toBe("allow-scripts");
        expect(iframe.getAttribute("srcDoc") ?? iframe.srcdoc).toContain("Sliding Puzzle");
        expect(screen.getByTestId("deliverable-preview-pane-art-puzzle-html")).toBeDefined();

        // Clicking Play again toggles the live iframe off
        fireEvent.click(playBtn);
        expect(screen.queryByTestId("deliverable-iframe-art-puzzle-html")).toBeNull();
    });
});
