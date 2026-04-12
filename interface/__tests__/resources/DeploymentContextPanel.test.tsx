import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import DeploymentContextPanel from "@/components/resources/DeploymentContextPanel";

describe("DeploymentContextPanel", () => {
    const fetchMock = vi.fn();

    beforeEach(() => {
        vi.stubGlobal("fetch", fetchMock);
    });

    afterEach(() => {
        vi.unstubAllGlobals();
        fetchMock.mockReset();
    });

    it("renders existing deployment context entries", async () => {
        fetchMock.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                entries: [{
                    artifact_id: "ctx-1",
                    knowledge_class: "customer_context",
                    title: "Deployment Brief",
                    source_label: "operator provided",
                    source_kind: "user_document",
                    visibility: "global",
                    sensitivity_class: "role_scoped",
                    trust_class: "user_provided",
                    chunk_count: 2,
                    vector_count: 2,
                    content_preview: "Mycelis should run with governed MCP access.",
                    content_length: 58,
                    content_domain: "operations",
                    target_goal_sets: ["deployment readiness"],
                    created_at: "2026-04-04T12:00:00Z",
                }],
            }),
        });

        render(<DeploymentContextPanel />);

        await waitFor(() => {
            expect(screen.getByText("Deployment Brief")).toBeDefined();
            expect(screen.getByText(/governed MCP access/i)).toBeDefined();
            expect(screen.getByText(/2 vectors/i)).toBeDefined();
            expect(screen.getByText(/goal: deployment readiness/i)).toBeDefined();
        });
    });

    it("submits new deployment context and refreshes the list", async () => {
        fetchMock
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ entries: [] }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    artifact_id: "ctx-2",
                    knowledge_class: "company_knowledge",
                    title: "Security Notes",
                    chunk_count: 1,
                    vector_count: 1,
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    entries: [{
                        artifact_id: "ctx-2",
                        knowledge_class: "company_knowledge",
                        title: "Security Notes",
                        source_label: "operator provided",
                        source_kind: "user_document",
                        visibility: "global",
                        sensitivity_class: "role_scoped",
                        trust_class: "user_provided",
                        chunk_count: 1,
                        vector_count: 1,
                        content_preview: "Restrict web access by trust class.",
                        content_length: 36,
                        created_at: "2026-04-04T12:30:00Z",
                    }],
                }),
            });

        render(<DeploymentContextPanel />);

        await waitFor(() => {
            expect(screen.getByText(/No deployment context loaded yet/i)).toBeDefined();
        });

        fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Security Notes" } });
        fireEvent.change(screen.getByLabelText("Knowledge Class"), { target: { value: "company_knowledge" } });
        fireEvent.change(screen.getByLabelText("Content"), { target: { value: "Restrict web access by trust class." } });
        fireEvent.click(screen.getByRole("button", { name: /Load Context/i }));

        await waitFor(() => {
            expect(fetchMock).toHaveBeenCalledTimes(3);
            expect(screen.getByText("Security Notes")).toBeDefined();
            expect(screen.getByText(/Loaded Security Notes as approved company knowledge into 1 vectors across 1 chunks/i)).toBeDefined();
            expect(screen.getAllByText(/company knowledge/i).length).toBeGreaterThan(0);
        });

        const submitCall = fetchMock.mock.calls[1];
        expect(submitCall?.[0]).toBe("/api/v1/memory/deployment-context");
        expect(submitCall?.[1]?.method).toBe("POST");
    });

    it("submits private user content with goal-set metadata", async () => {
        fetchMock
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ entries: [] }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    artifact_id: "ctx-private",
                    knowledge_class: "user_private_context",
                    title: "Finance Notes",
                    chunk_count: 1,
                    vector_count: 1,
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ entries: [] }),
            });

        render(<DeploymentContextPanel />);

        await waitFor(() => {
            expect(screen.getByText(/No deployment context loaded yet/i)).toBeDefined();
        });

        fireEvent.change(screen.getByLabelText("Source Kind"), { target: { value: "finance_record" } });
        fireEvent.change(screen.getByLabelText("Content Domain"), { target: { value: "finance" } });
        fireEvent.change(screen.getByLabelText("Target Goal Sets"), { target: { value: "tax planning, cash flow" } });
        const file = new File(["Invoice timing and savings goals."], "Finance Notes.md", { type: "text/markdown" });
        fireEvent.change(screen.getByLabelText("Upload Text File"), { target: { files: [file] } });
        await waitFor(() => {
            expect(screen.getAllByDisplayValue("Finance Notes.md").length).toBeGreaterThan(0);
        });
        fireEvent.click(screen.getByRole("button", { name: /Load Context/i }));

        await waitFor(() => {
            expect(screen.getByText(/Loaded Finance Notes as private user content/i)).toBeDefined();
        });

        const submitBody = JSON.parse(fetchMock.mock.calls[1]?.[1]?.body as string);
        expect(submitBody.knowledge_class).toBe("user_private_context");
        expect(submitBody.visibility).toBe("private");
        expect(submitBody.sensitivity_class).toBe("restricted");
        expect(submitBody.content_domain).toBe("finance");
        expect(submitBody.target_goal_sets).toEqual(["tax planning", "cash flow"]);
        expect(submitBody.content).toBe("Invoice timing and savings goals.");
    });
});
