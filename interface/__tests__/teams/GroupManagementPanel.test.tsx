import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import { mockFetch } from "../setup";

function jsonResponse(payload: unknown, ok = true, status = 200): Response {
    return {
        ok,
        status,
        json: async () => payload,
    } as Response;
}

function urlFromInput(input: RequestInfo | URL): string {
    if (typeof input === "string") return input;
    if (input instanceof URL) return input.pathname;
    const requestLike = input as Request;
    if (typeof requestLike.url === "string") {
        try {
            return new URL(requestLike.url).pathname;
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
            if (url === "/api/v1/groups" && (!init || init.method === "GET")) {
                return jsonResponse({ ok: true, data: [] });
            }
            if (url === "/api/v1/groups/monitor") {
                return jsonResponse({ ok: true, data: { status: "online", published_count: 0 } });
            }
            if (url === "/api/v1/groups" && init?.method === "POST") {
                return jsonResponse(
                    {
                        ok: true,
                        data: {
                            requires_approval: true,
                            confirm_token: { token: "tok-123" },
                            intent_proof: { id: "proof-1" },
                        },
                    },
                    true,
                    202,
                );
            }
            return jsonResponse({ error: "not found" }, false, 404);
        });

        render(<GroupManagementPanel />);

        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Ops Group" } });
        fireEvent.change(screen.getByLabelText("Goal Statement"), { target: { value: "Coordinate operational runs" } });
        fireEvent.click(screen.getByTestId("groups-create-button"));

        await waitFor(() => {
            expect(screen.getByTestId("groups-approval-card")).toBeDefined();
        });

        expect(screen.getByTestId("groups-confirm-token-input")).toHaveProperty("value", "tok-123");
        expect(screen.getByTestId("groups-notice").textContent).toContain("Approval required");
    });

    it("resubmits create request with confirm_token after approval prompt", async () => {
        const postBodies: Array<Record<string, unknown>> = [];
        mockFetch.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
            const url = urlFromInput(input);
            if (url === "/api/v1/groups" && (!init || init.method === "GET")) {
                return jsonResponse({ ok: true, data: [] });
            }
            if (url === "/api/v1/groups/monitor") {
                return jsonResponse({ ok: true, data: { status: "online", published_count: 0 } });
            }
            if (url === "/api/v1/groups" && init?.method === "POST") {
                postBodies.push(JSON.parse(String(init.body)));
                if (postBodies.length === 1) {
                    return jsonResponse(
                        {
                            ok: true,
                            data: {
                                requires_approval: true,
                                confirm_token: { token: "tok-123" },
                                intent_proof: { id: "proof-1" },
                            },
                        },
                        true,
                        202,
                    );
                }
                return jsonResponse({ ok: true, data: { group_id: "group-1" } }, true, 201);
            }
            return jsonResponse({ error: "not found" }, false, 404);
        });

        render(<GroupManagementPanel />);

        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Ops Group" } });
        fireEvent.change(screen.getByLabelText("Goal Statement"), { target: { value: "Coordinate operational runs" } });
        fireEvent.click(screen.getByTestId("groups-create-button"));

        await waitFor(() => {
            expect(screen.getByTestId("groups-approval-card")).toBeDefined();
        });

        fireEvent.click(screen.getByTestId("groups-create-button"));

        await waitFor(() => {
            expect(screen.getByTestId("groups-notice").textContent).toContain("Group created successfully");
        });

        expect(postBodies).toHaveLength(2);
        expect(postBodies[1].confirm_token).toBe("tok-123");
    });
});
