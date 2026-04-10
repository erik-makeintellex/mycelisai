import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";

function jsonResponse(payload: unknown, ok = true, status = 200): Response {
    return { ok, status, json: async () => payload } as Response;
}

function urlFromInput(input: RequestInfo | URL): string {
    if (typeof input === "string") return input;
    if (input instanceof URL) return input.pathname + input.search;
    const requestLike = input as Request;
    if (typeof requestLike.url === "string") {
        try {
            const parsed = new URL(requestLike.url);
            return parsed.pathname + parsed.search;
        } catch {
            return requestLike.url;
        }
    }
    return String(input);
}

describe("GroupManagementPanel", () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it("shows approval-required state with confirm token", async () => {
        mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
            const url = urlFromInput(input);
            if (url === "/api/v1/groups" && (!init?.method || init.method === "GET")) {
                return jsonResponse({ ok: true, data: [] });
            }
            if (url === "/api/v1/groups/monitor") {
                return jsonResponse({ ok: true, data: { status: "online", published_count: 0 } });
            }
            if (url === "/api/v1/groups" && init?.method === "POST") {
                return jsonResponse({ ok: true, data: { requires_approval: true, confirm_token: { token: "tok-123" }, intent_proof: { id: "proof-1" } } }, true, 202);
            }
            return jsonResponse({ error: "not found" }, false, 404);
        });

        render(<GroupManagementPanel />);
        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Ops Group" } });
        fireEvent.change(screen.getByLabelText("Goal Statement"), { target: { value: "Coordinate operational runs" } });
        fireEvent.click(screen.getByTestId("groups-create-button"));

        await waitFor(() => expect(screen.getByTestId("groups-approval-card")).toBeDefined());
        expect(screen.getByTestId("groups-confirm-token-input")).toHaveProperty("value", "tok-123");
        expect(screen.getByTestId("groups-notice").textContent).toContain("Approval required");
    });

    it("resubmits create request with confirm_token after approval prompt", async () => {
        const postBodies: Array<Record<string, unknown>> = [];
        mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
            const url = urlFromInput(input);
            if (url === "/api/v1/groups" && (!init?.method || init.method === "GET")) {
                return jsonResponse({ ok: true, data: [] });
            }
            if (url === "/api/v1/groups/monitor") {
                return jsonResponse({ ok: true, data: { status: "online", published_count: 0 } });
            }
            if (url === "/api/v1/groups" && init?.method === "POST") {
                postBodies.push(JSON.parse(String(init.body)));
                if (postBodies.length === 1) {
                    return jsonResponse({ ok: true, data: { requires_approval: true, confirm_token: { token: "tok-123" }, intent_proof: { id: "proof-1" } } }, true, 202);
                }
                return jsonResponse({ ok: true, data: { group_id: "group-1" } }, true, 201);
            }
            return jsonResponse({ error: "not found" }, false, 404);
        });

        render(<GroupManagementPanel />);
        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Ops Group" } });
        fireEvent.change(screen.getByLabelText("Goal Statement"), { target: { value: "Coordinate operational runs" } });
        fireEvent.click(screen.getByTestId("groups-create-button"));

        await waitFor(() => expect(screen.getByTestId("groups-approval-card")).toBeDefined());
        fireEvent.click(screen.getByTestId("groups-create-button"));

        await waitFor(() => expect(screen.getByTestId("groups-notice").textContent).toContain("Group created successfully"));
        expect(postBodies).toHaveLength(2);
        expect(postBodies[1].confirm_token).toBe("tok-123");
    });

    it("archives temporary groups and keeps retained outputs reviewable", async () => {
        const groups = [
            {
                group_id: "group-standing",
                name: "Standing Ops",
                goal_statement: "Run durable operations",
                work_mode: "propose_only",
                member_user_ids: [],
                team_ids: ["team-ops"],
                coordinator_profile: "ops-lead",
                approval_policy_ref: "",
                status: "active",
                created_by: "admin",
                created_at: new Date().toISOString(),
            },
            {
                group_id: "group-temp",
                name: "Temp Campaign",
                goal_statement: "Produce one campaign package",
                work_mode: "execute_with_approval",
                member_user_ids: [],
                team_ids: ["team-marketing"],
                coordinator_profile: "marketing-lead",
                approval_policy_ref: "",
                status: "active",
                expiry: new Date(Date.now() + 60_000).toISOString(),
                created_by: "admin",
                created_at: new Date().toISOString(),
            },
        ];

        mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
            const url = urlFromInput(input);
            if (url === "/api/v1/groups" && (!init?.method || init.method === "GET")) {
                return jsonResponse({
                    ok: true,
                    data: groups.map((group) => ({ ...group })),
                });
            }
            if (url === "/api/v1/groups/monitor") {
                return jsonResponse({ ok: true, data: { status: "online", published_count: 2, last_group_id: "group-temp" } });
            }
            if (url === "/api/v1/groups/group-temp/status" && init?.method === "PATCH") {
                groups[1] = {
                    ...groups[1],
                    status: "archived",
                };
                return jsonResponse({ ok: true, data: groups[1] });
            }
            if (url === "/api/v1/groups/group-standing/outputs?limit=8") {
                return jsonResponse({ ok: true, data: [] });
            }
            if (url === "/api/v1/groups/group-temp/outputs?limit=8") {
                return jsonResponse({
                    ok: true,
                    data: [
                        {
                            id: "artifact-1",
                            agent_id: "marketing-lead",
                            artifact_type: "document",
                            title: "Launch brief",
                            content_type: "text/markdown",
                            content: "Campaign summary",
                            metadata: {},
                            status: "approved",
                            created_at: new Date().toISOString(),
                        },
                    ],
                });
            }
            return jsonResponse({ error: "not found" }, false, 404);
        });

        render(<GroupManagementPanel />);

        await waitFor(() => expect(screen.getByRole("button", { name: /Temp Campaign/i })).toBeDefined());
        fireEvent.click(screen.getByRole("button", { name: /Temp Campaign/i }));

        await waitFor(() => expect(screen.getByText("Launch brief")).toBeDefined());
        fireEvent.click(screen.getByRole("button", { name: "Archive temporary group" }));

        await waitFor(() => expect(screen.getByTestId("groups-notice").textContent).toContain("Temporary group archived"));
        await waitFor(() => expect(screen.getByText("Archived temporary groups")).toBeDefined());
        fireEvent.click(screen.getByRole("button", { name: /Temp Campaign.*Produce one campaign package/i }));
        await waitFor(() => expect(screen.getByTestId("groups-archived-readonly-note").textContent).toContain("retained output review"));
        await waitFor(() => expect(screen.getByTestId("groups-retained-outputs-note").textContent).toContain("Downloads remain available"));
        expect(screen.getByText("Campaign summary")).toBeDefined();
        expect(screen.getByRole("link", { name: /Download/i }).getAttribute("href")).toBe("/api/v1/artifacts/artifact-1/download");
        expect(screen.getByRole("link", { name: /Open team-marketing lead/i }).getAttribute("href")).toBe("/dashboard?team_id=team-marketing");
        await waitFor(() => expect(screen.queryByRole("button", { name: "Broadcast to group" })).toBeNull());
    });

    it("honors an initially selected group id from the route", async () => {
        const groups = [
            {
                group_id: "group-standing",
                name: "Standing Ops",
                goal_statement: "Run durable operations",
                work_mode: "propose_only",
                member_user_ids: [],
                team_ids: ["team-ops"],
                coordinator_profile: "ops-lead",
                approval_policy_ref: "",
                status: "active",
                created_by: "admin",
                created_at: new Date().toISOString(),
            },
            {
                group_id: "group-temp",
                name: "Temp Campaign",
                goal_statement: "Produce one campaign package",
                work_mode: "propose_only",
                member_user_ids: [],
                team_ids: [],
                coordinator_profile: "marketing-lead",
                approval_policy_ref: "",
                status: "active",
                expiry: new Date(Date.now() + 60_000).toISOString(),
                created_by: "admin",
                created_at: new Date().toISOString(),
            },
        ];

        mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
            const url = urlFromInput(input);
            if (url === "/api/v1/groups" && (!init?.method || init.method === "GET")) {
                return jsonResponse({ ok: true, data: groups });
            }
            if (url === "/api/v1/groups/monitor") {
                return jsonResponse({ ok: true, data: { status: "online", published_count: 0 } });
            }
            if (url === "/api/v1/groups/group-temp/outputs?limit=8") {
                return jsonResponse({ ok: true, data: [] });
            }
            if (url === "/api/v1/groups/group-standing/outputs?limit=8") {
                return jsonResponse({ ok: true, data: [] });
            }
            return jsonResponse({ error: "not found" }, false, 404);
        });

        render(<GroupManagementPanel initialSelectedGroupId="group-temp" />);

        await waitFor(() => expect(screen.getByRole("heading", { name: "Temp Campaign" })).toBeDefined());
    });
});
