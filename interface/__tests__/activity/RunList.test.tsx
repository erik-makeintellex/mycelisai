import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { RunList } from "@/components/activity/ActivityPanels";
import { useCortexStore } from "@/store/useCortexStore";

describe("RunList", () => {
    beforeEach(() => {
        useCortexStore.setState({
            runsLoadError: null,
            servicesStatusFailure: null,
            fetchRecentRuns: vi.fn().mockResolvedValue(undefined),
        });
    });

    it("shows No runs yet only when the fetch succeeded with zero runs", () => {
        render(<RunList runs={[]} selectedRunId={null} isFetching={false} onSelect={vi.fn()} />);
        expect(screen.getByText("No runs yet.")).toBeDefined();
    });

    it("shows a load failure with Retry when the runs fetch failed", () => {
        const fetchRecentRuns = vi.fn().mockResolvedValue(undefined);
        useCortexStore.setState({ runsLoadError: "Could not load runs (HTTP 500).", fetchRecentRuns });

        render(<RunList runs={[]} selectedRunId={null} isFetching={false} onSelect={vi.fn()} />);

        expect(screen.queryByText("No runs yet.")).toBeNull();
        expect(screen.getByTestId("activity-runs-load-failure").textContent).toMatch(/HTTP 500/);
        fireEvent.click(screen.getByRole("button", { name: /Retry/i }));
        expect(fetchRecentRuns).toHaveBeenCalledTimes(1);
    });
});
