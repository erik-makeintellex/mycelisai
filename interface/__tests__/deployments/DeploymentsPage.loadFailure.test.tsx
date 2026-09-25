import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import DeploymentsPage from "@/components/deployments/DeploymentsPage";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";

function jsonResponse(status: number, body: unknown) {
    return { ok: status >= 200 && status < 300, status, json: async () => body };
}

describe("DeploymentsPage load failure", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        useCortexStore.setState({
            advancedMode: false,
            servicesStatusFailure: null,
            servicesStatusUpdatedAt: null,
        });
    });

    it("replaces the list with a visible error and Retry when deployments fail to load", async () => {
        mockFetch.mockResolvedValue(jsonResponse(500, {}));

        render(<DeploymentsPage />);

        const failure = await screen.findByTestId("deployments-load-failure");
        expect(failure.textContent).toMatch(/Could not load deployments/);
        expect(failure.textContent).toMatch(/HTTP 500/);

        mockFetch.mockResolvedValue(
            jsonResponse(200, { ok: true, data: { deployments: [], sources: [], generated_at: new Date().toISOString() } }),
        );
        fireEvent.click(screen.getByRole("button", { name: /Retry/i }));

        await waitFor(() => expect(screen.queryByTestId("deployments-load-failure")).toBeNull());
    });

    it("names the Core API root cause when Core is unreachable", async () => {
        useCortexStore.setState({
            servicesStatusFailure: { kind: "unreachable", at: new Date().toISOString() },
        });
        mockFetch.mockRejectedValue(new TypeError("Failed to fetch"));

        render(<DeploymentsPage />);

        const failure = await screen.findByTestId("deployments-load-failure");
        expect(failure.textContent).toMatch(/Core API unreachable/);
        expect(failure.textContent).toMatch(/uv run inv lifecycle\.up --frontend/);
    });
});
