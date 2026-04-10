import { describe, it, expect, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import TeamLeadInteractionPanel from "@/components/organizations/TeamLeadInteractionPanel";

function jsonResponse(body: unknown, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(body), { status }));
}

describe("TeamLeadInteractionPanel", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        localStorage.clear();
    });

    it("renders guided Soma actions without generic chat wording", () => {
        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        expect(screen.getByText("Create teams with Soma")).toBeDefined();
        expect(screen.getByText("Team creation lane")).toBeDefined();
        expect(screen.getByRole("button", { name: "Start team design" })).toBeDefined();
        expect(screen.getByLabelText("Tell Soma what team or delivery lane you want to create")).toBeDefined();
        expect(screen.getByRole("button", { name: /Run a quick strategy check/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Choose the first priority/i })).toBeDefined();
        expect(screen.getByRole("button", { name: /Review your organization setup/i })).toBeDefined();
        expect(screen.getByText(/Each one should produce a clearer team-creation direction/i)).toBeDefined();
        expect(screen.queryByText(/chat/i)).toBeNull();
        expect(screen.queryByText(/context shell/i)).toBeNull();
        expect(screen.queryByText(/bounded slice/i)).toBeNull();
    });

    it("lets the operator seed the prompt from guided suggestion chips", () => {
        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
                promptSuggestions={[
                    {
                        label: "Marketing launch team",
                        prompt: "Create a temporary marketing launch team for a new product rollout.",
                    },
                ]}
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: "Marketing launch team" }));

        expect(
            (screen.getByLabelText("Tell Soma what team or delivery lane you want to create") as HTMLTextAreaElement).value,
        ).toBe("Create a temporary marketing launch team for a new product rollout.");
    });

    it("triggers a Soma action and renders structured guidance", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Run a quick strategy check",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends a clear next move for Northstar Labs.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                    "Use the first Department as the routing layer for work.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                    "Choose the first priority",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                }),
            );
        });

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
        expect(screen.getByText("Keep moving with")).toBeDefined();
        expect(screen.getAllByText("Review your organization setup").length).toBeGreaterThan(0);
    });

    it("uses the visible Soma prompt to choose a first guided action and shows what the operator asked for", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "focus_first",
                request_label: "Choose the first priority",
                headline: "First focus for Northstar Labs",
                summary: "Start by confirming the first outcome this AI Organization should deliver.",
                priority_steps: [
                    "Confirm the first priority for Northstar Labs.",
                    "Use that priority to create visible movement across the workspace.",
                ],
                suggested_follow_ups: [
                    "Run a quick strategy check",
                    "Review your organization setup",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Help me choose the first priority for this launch." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                    body: JSON.stringify({
                        action: "focus_first",
                        request_context: "Help me choose the first priority for this launch.",
                    }),
                }),
            );
        });

        expect(await screen.findByText("You asked Soma to help with")).toBeDefined();
        expect(screen.getAllByText("Help me choose the first priority for this launch.")).toHaveLength(2);
        expect(screen.getByText("First focus for Northstar Labs")).toBeDefined();
    });

    it("renders a native team execution path for image-oriented team requests", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Plan next steps for this organization",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends moving Northstar Labs into a focused first delivery loop.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
                execution_contract: {
                    execution_mode: "native_team",
                    owner_label: "Native Mycelis team",
                    team_name: "Creative Delivery Team",
                    summary: "Use a bounded creative team inside Northstar Labs so Soma can shape the work and return the generated image as a managed artifact.",
                    target_outputs: [
                        "Reviewable image artifact",
                        "Short concept note",
                    ],
                },
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create a creative team to generate a launch hero image." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    body: JSON.stringify({
                        action: "plan_next_steps",
                        request_context: "Create a creative team to generate a launch hero image.",
                    }),
                }),
            );
        });

        expect(await screen.findByText("Execution path")).toBeDefined();
        expect(screen.getAllByText("Native Mycelis team").length).toBeGreaterThan(0);
        expect(screen.getByText("Creative Delivery Team")).toBeDefined();
        expect(screen.getByText("Reviewable image artifact")).toBeDefined();
    });

    it("renders an external workflow contract path without stripping the contract wording", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Plan next steps for this organization",
                headline: "Team Lead plan for Northstar Labs",
                summary: "Team Lead recommends moving Northstar Labs into a focused first delivery loop.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
                execution_contract: {
                    execution_mode: "external_workflow_contract",
                    owner_label: "External workflow contract",
                    external_target: "n8n workflow contract",
                    summary: "This request is best handled as an external workflow contract so Mycelis can keep the result return clear.",
                    target_outputs: [
                        "Normalized workflow result",
                        "Linked artifact or execution note",
                    ],
                },
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Create an n8n workflow contract for inbound leads." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect((await screen.findAllByText("External workflow contract")).length).toBeGreaterThan(0);
        expect(screen.getByText("n8n workflow contract")).toBeDefined();
        expect(screen.getByText("Normalized workflow result")).toBeDefined();
    });

    it("lets Start with Soma run a quick strategy check even when the prompt is blank", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "Run a quick strategy check",
                headline: "Soma plan for Northstar Labs",
                summary: "Soma recommends the clearest next move for Northstar Labs.",
                priority_steps: [
                    "Align the first outcome with the AI Organization purpose.",
                ],
                suggested_follow_ups: [
                    "Choose the first priority",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith(
                "/api/v1/organizations/org-123/workspace/actions",
                expect.objectContaining({
                    method: "POST",
                    body: JSON.stringify({ action: "plan_next_steps" }),
                }),
            );
        });

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
    });

    it("shows a loading state and prevents duplicate action submission while guidance is loading", async () => {
        let resolveResponse!: (value: Response) => void;
        mockFetch.mockImplementationOnce(
            () =>
                new Promise<Response>((resolve) => {
                    resolveResponse = resolve;
                }),
        );

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        const actionButton = screen.getByRole("button", { name: /Run a quick strategy check/i });
        fireEvent.click(actionButton);
        fireEvent.click(actionButton);

        expect(await screen.findByText("Soma is preparing guidance for this AI Organization...")).toBeDefined();
        expect(actionButton.hasAttribute("disabled")).toBe(true);
        expect(mockFetch).toHaveBeenCalledTimes(1);

        resolveResponse(
            new Response(
                JSON.stringify({
                    ok: true,
                    data: {
                        action: "plan_next_steps",
                        request_label: "Run a quick strategy check",
                        headline: "Team Lead plan for Northstar Labs",
                        summary: "Team Lead recommends a clear next move for Northstar Labs.",
                        priority_steps: ["Align the first outcome with the AI Organization purpose."],
                        suggested_follow_ups: ["Review your organization setup"],
                    },
                }),
                { status: 200 },
            ),
        );

        expect(await screen.findByText("Soma plan for Northstar Labs")).toBeDefined();
    });

    it("persists the Soma draft and last guidance across remount for the same organization", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "focus_first",
                request_label: "Choose the first priority",
                headline: "First focus for Northstar Labs",
                summary: "Start by confirming the first outcome this AI Organization should deliver.",
                priority_steps: [
                    "Confirm the first priority for Northstar Labs.",
                ],
                suggested_follow_ups: [
                    "Review your organization setup",
                ],
            },
        }));

        const { unmount } = render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.change(screen.getByLabelText("Tell Soma what team or delivery lane you want to create"), {
            target: { value: "Help me choose the first priority for this launch." },
        });
        fireEvent.click(screen.getByRole("button", { name: "Start team design" }));

        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        unmount();

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        expect(screen.getByDisplayValue("Help me choose the first priority for this launch.")).toBeDefined();
        expect(screen.getByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("You asked Soma to help with")).toBeDefined();
    });

    it("renders readable fallback guidance when the backend returns a partial malformed payload", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({
            ok: true,
            data: {
                action: "plan_next_steps",
                request_label: "system: debug trace",
                headline: "",
                summary: "{debug:true}",
                priority_steps: [
                    "debug: internal trace",
                    "Turn Northstar Labs into a clear next move.",
                ],
                suggested_follow_ups: [
                    "contract",
                    "Choose the first priority",
                ],
            },
        }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Run a quick strategy check/i }));

        expect(await screen.findByText("Soma for Northstar Labs guidance for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Soma for Northstar Labs has guidance ready for Northstar Labs.")).toBeDefined();
        expect(screen.getByText("Turn Northstar Labs into a clear next move.")).toBeDefined();
        expect(screen.queryByText(/debug/i)).toBeNull();
        expect(screen.queryByText(/^contract$/i)).toBeNull();
    });

    it("shows retry guidance when the Soma action request fails", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));

        expect(await screen.findByText("Soma guidance is unavailable")).toBeDefined();
        expect(screen.getByRole("button", { name: "Retry Soma action" })).toBeDefined();
        expect(screen.getByText(/without leaving Northstar Labs/i)).toBeDefined();
        expect(screen.getByText(/choose another guided Soma action below/i)).toBeDefined();
    });

    it("retries the same Soma action successfully after a failure", async () => {
        mockFetch
            .mockResolvedValueOnce(jsonResponse({ ok: false, error: "Team Lead guidance is unavailable right now." }, 500))
            .mockResolvedValueOnce(jsonResponse({
                ok: true,
                data: {
                    action: "focus_first",
                    request_label: "Choose the first priority",
                    headline: "First focus for Northstar Labs",
                    summary: "Start by confirming the first outcome this AI Organization should deliver.",
                    priority_steps: [
                        "Use 1 Department as the first routing layer for work.",
                        "Use 1 Advisor when the Team Lead needs review or decision support.",
                    ],
                    suggested_follow_ups: [
                        "Run a quick strategy check",
                        "Review your organization setup",
                    ],
                },
            }));

        render(
            <TeamLeadInteractionPanel
                organizationId="org-123"
                organizationName="Northstar Labs"
                somaName="Soma for Northstar Labs"
                teamLeadName="Team Lead for Northstar Labs"
            />,
        );

        fireEvent.click(screen.getByRole("button", { name: /Choose the first priority/i }));
        expect(await screen.findByText("Soma guidance is unavailable")).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: "Retry Soma action" }));

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledTimes(2);
        });
        expect(await screen.findByText("First focus for Northstar Labs")).toBeDefined();
        expect(screen.getByText("Priority steps")).toBeDefined();
    });
});


