import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor, within } from "@testing-library/react";
import DeploymentsPage from "@/components/deployments/DeploymentsPage";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";
import type { DeploymentRecord } from "@/components/deployments/deploymentTypes";

function jsonResponse(status: number, body: unknown) {
    return { ok: status >= 200 && status < 300, status, json: async () => body };
}

function member(overrides: Partial<DeploymentRecord["members"][number]> = {}) {
    return {
        id: "member-1",
        role: "lead",
        provider: "ollama",
        model: "qwen2.5-coder:7b",
        provider_origin: "manifest",
        status: "idle",
        tools: [],
        ...overrides,
    };
}

function record(overrides: Partial<DeploymentRecord>): DeploymentRecord {
    return {
        id: "rec-1",
        kind: "standing_team",
        name: "Standing Team One",
        what_runs: "3 agents answering product requests",
        scope: { kind: "global" },
        origin: "bundle",
        lifecycle_state: "running",
        members: [member()],
        tools: [],
        config_ref: { kind: "team", id: "rec-1", path: "/api/v1/config/teams/rec-1" },
        ...overrides,
    };
}

const standingTeam = record({ id: "standing-1", kind: "standing_team", name: "Standing Delivery" });
const missionTeam = record({
    id: "mission-1",
    kind: "mission_team",
    name: "Mission Bravo",
    scope: { kind: "mission", id: "mission-1" },
    origin: "mission:mission-1",
});
const group = record({
    id: "group-1",
    kind: "group",
    name: "Client Engagement Group",
    scope: { kind: "global" },
    work_mode: "propose_only",
    approval_posture: "Every client-facing output needs approval before it leaves the org.",
    outputs_summary: { retained: 4, by_status: { verified: 4 }, proof_status: "verified" },
});
const plannedDepartment = record({
    id: "org-1:dept-1",
    kind: "planned_department",
    name: "Planned Ops Desk",
    scope: { kind: "organization", id: "org-1", name: "Northstar Labs" },
    origin: "template:starter-operations-desk",
    lifecycle_state: "planned",
    members: [],
    outputs_summary: undefined,
});

const allDeployments = [standingTeam, missionTeam, group, plannedDepartment];

function deploymentsResponse(deployments: DeploymentRecord[], sources: any[] = []) {
    return jsonResponse(200, {
        ok: true,
        data: { deployments, sources, generated_at: new Date().toISOString() },
    });
}

describe("DeploymentsPage", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        useCortexStore.setState({
            advancedMode: false,
            servicesStatusFailure: null,
            servicesStatusUpdatedAt: null,
        });
    });

    it("renders one row for each of the four deployment kinds", async () => {
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));

        render(<DeploymentsPage />);

        const list = await screen.findByTestId("deployments-list");
        expect(within(list).getByText("Standing Delivery")).toBeDefined();
        expect(within(list).getByText("Mission Bravo")).toBeDefined();
        expect(within(list).getByText("Client Engagement Group")).toBeDefined();
        expect(within(list).getByText("Planned Ops Desk")).toBeDefined();
        expect(within(list).getAllByTestId("deployments-kind-badge")).toHaveLength(4);
    });

    it("requests the filtered kind and state from Core when a filter chip is clicked", async () => {
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));
        render(<DeploymentsPage />);
        await screen.findByTestId("deployments-list");

        fireEvent.click(screen.getByTestId("deployments-filter-kind-group"));

        await waitFor(() => {
            const lastCall = mockFetch.mock.calls.at(-1)?.[0] as string;
            expect(lastCall).toContain("kind=group");
        });

        fireEvent.click(screen.getByTestId("deployments-filter-state-running"));
        await waitFor(() => {
            const lastCall = mockFetch.mock.calls.at(-1)?.[0] as string;
            expect(lastCall).toContain("state=running");
        });
    });

    it("shows the detail pane for a selected row", async () => {
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));
        render(<DeploymentsPage />);
        await screen.findByTestId("deployments-list");

        fireEvent.click(screen.getByText("Mission Bravo"));

        const detail = await screen.findByTestId("deployments-detail");
        expect(within(detail).getByText("Mission Bravo")).toBeDefined();
    });

    it("shows 'unknown' retained outputs when outputs_summary is absent", async () => {
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));
        render(<DeploymentsPage />);
        await screen.findByTestId("deployments-list");

        fireEvent.click(screen.getByText("Planned Ops Desk"));

        const summary = await screen.findByTestId("deployments-outputs-summary");
        expect(summary.textContent).toMatch(/unknown/);
    });

    it("shows a retained-output count and proof status when outputs_summary is present", async () => {
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));
        render(<DeploymentsPage />);
        await screen.findByTestId("deployments-list");

        fireEvent.click(screen.getByText("Client Engagement Group"));

        const summary = await screen.findByTestId("deployments-outputs-summary");
        expect(summary.textContent).toMatch(/4 retained output/);
        expect(summary.textContent).toMatch(/verified/);
    });

    it("shows a non-blocking notice when a source is degraded, unavailable, or restricted", async () => {
        mockFetch.mockResolvedValue(
            deploymentsResponse(allDeployments, [
                { source: "groups", status: "restricted", detail: "Groups hidden: requires admin access", count: 0 },
            ]),
        );
        render(<DeploymentsPage />);

        const notice = await screen.findByTestId("deployments-source-notice-groups");
        expect(notice.textContent).toMatch(/Groups hidden: requires admin access/);
        // The list still renders; the notice is non-blocking.
        await screen.findByTestId("deployments-list");
    });

    it("hides View config when Advanced mode is off", async () => {
        useCortexStore.setState({ advancedMode: false });
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));
        render(<DeploymentsPage />);
        await screen.findByTestId("deployments-detail");

        expect(screen.queryByTestId("deployments-view-config")).toBeNull();
    });

    it("shows View config when Advanced mode is on", async () => {
        useCortexStore.setState({ advancedMode: true });
        mockFetch.mockResolvedValue(deploymentsResponse(allDeployments));
        render(<DeploymentsPage />);
        await screen.findByTestId("deployments-detail");

        const link = await screen.findByTestId("deployments-view-config");
        expect(link.getAttribute("href")).toBe(allDeployments[0].config_ref.path);
    });
});
