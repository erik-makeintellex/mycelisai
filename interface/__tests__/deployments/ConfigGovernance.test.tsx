import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { CloneTemplateDialog } from "@/components/deployments/CloneTemplateDialog";
import { ApplyConfigDialog } from "@/components/deployments/ApplyConfigDialog";
import { mockFetch } from "../setup";

function jsonResponse(status: number, body: unknown) {
    return { ok: status >= 200 && status < 300, status, json: async () => body };
}

const previewToken = {
    requires_approval: true,
    blocked: false,
    confirm_token: { token: "tok-1", intent_proof_id: "ip-1", expires_at: "2026-09-25T18:30:00Z" },
    intent_proof: { id: "ip-1", resolved_intent: "template.clone.template.custom-x@sha256:aaa", status: "pending" },
    audit_event_id: "audit-1",
    preview: {
        kind: "template",
        id: "custom-x",
        document_sha256: "aaa",
        valid: true,
        validation_errors: [] as string[],
        diff: { changed: [] as string[], added: ["id"], removed: [] as string[] },
    },
};

describe("CloneTemplateDialog", () => {
    beforeEach(() => mockFetch.mockReset());

    it("previews then approves a clone and shows success", async () => {
        mockFetch
            .mockResolvedValueOnce(jsonResponse(202, { ok: true, data: previewToken }))
            .mockResolvedValueOnce(
                jsonResponse(201, {
                    ok: true,
                    data: { template: { id: "custom-x", name: "Acme", source: "operator", document_sha256: "aaa" }, intent_proof_id: "ip-1", audit_event_id: "audit-1" },
                }),
            );

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));

        await screen.findByTestId("config-governance-preview-summary");
        expect(screen.getByTestId("config-governance-diff-added").textContent).toMatch("id");
        expect((screen.getByTestId("clone-template-approve") as HTMLButtonElement).disabled).toBe(false);

        fireEvent.click(screen.getByTestId("clone-template-approve"));
        const success = await screen.findByTestId("clone-template-success");
        expect(success.textContent).toMatch("custom-x");
    });

    it("disables approve when the preview is blocked", async () => {
        mockFetch.mockResolvedValueOnce(
            jsonResponse(202, {
                ok: true,
                data: { ...previewToken, blocked: true, confirm_token: undefined, preview: { ...previewToken.preview, valid: false, validation_errors: ["id: required"] } },
            }),
        );

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));

        await screen.findByTestId("config-governance-blocked");
        expect((screen.getByTestId("clone-template-approve") as HTMLButtonElement).disabled).toBe(true);
    });

    it("requires a new preview after editing the name post-preview", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse(202, { ok: true, data: previewToken }));

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));
        await screen.findByTestId("config-governance-preview-summary");

        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme 2" } });
        expect(screen.queryByTestId("config-governance-preview-summary")).toBeNull();
        expect((screen.getByTestId("clone-template-approve") as HTMLButtonElement).disabled).toBe(true);
    });

    it("re-previews automatically on a 409 during approve", async () => {
        mockFetch
            .mockResolvedValueOnce(jsonResponse(202, { ok: true, data: previewToken }))
            .mockResolvedValueOnce(jsonResponse(409, { ok: false, error: "document changed" }))
            .mockResolvedValueOnce(
                jsonResponse(202, { ok: true, data: { ...previewToken, preview: { ...previewToken.preview, document_sha256: "bbb" } } }),
            );

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));
        await screen.findByTestId("config-governance-preview-summary");

        fireEvent.click(screen.getByTestId("clone-template-approve"));

        await waitFor(() => expect(mockFetch).toHaveBeenCalledTimes(3));
        expect(await screen.findByTestId("config-governance-preview-summary")).toBeTruthy();
        expect(screen.queryByTestId("clone-template-success")).toBeNull();
    });

    it("shows write_back_violations", async () => {
        mockFetch.mockResolvedValueOnce(
            jsonResponse(400, { ok: false, data: { write_back_violations: ["spec.provider_policy.openai.api_key"] } }),
        );

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));

        const banner = await screen.findByTestId("config-governance-violations");
        expect(banner.textContent).toMatch(/api_key/);
    });

    it("shows a 403 state plainly", async () => {
        mockFetch.mockResolvedValueOnce(jsonResponse(403, { ok: false, error: "forbidden" }));

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));

        expect(await screen.findByTestId("config-governance-denied")).toBeTruthy();
    });

    it("shows a 503 state naming the migration command", async () => {
        mockFetch.mockResolvedValueOnce(
            jsonResponse(503, { ok: false, error: "migration 041 is missing; run uv run inv compose.migrate" }),
        );

        render(<CloneTemplateDialog sourceTemplateId="starter-x" sourceTemplateName="Starter X" onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("clone-template-name"), { target: { value: "Acme" } });
        fireEvent.click(screen.getByTestId("clone-template-preview"));

        const banner = await screen.findByTestId("config-governance-unavailable");
        expect(banner.textContent).toMatch(/compose.migrate/);
    });
});

describe("ApplyConfigDialog", () => {
    beforeEach(() => mockFetch.mockReset());

    it("edits, previews with a diff, and approves", async () => {
        const applyPreview = {
            requires_approval: true,
            blocked: false,
            confirm_token: { token: "tok-2", intent_proof_id: "ip-2", expires_at: "2026-09-25T18:30:00Z" },
            intent_proof: { id: "ip-2", resolved_intent: "config.apply.groups.group-1@sha256:ccc", status: "pending" },
            audit_event_id: "audit-2",
            preview: {
                kind: "group",
                id: "group-1",
                document_sha256: "ccc",
                valid: true,
                validation_errors: [] as string[],
                diff: { changed: ["spec.name"], added: [] as string[], removed: [] as string[] },
            },
        };
        mockFetch
            .mockResolvedValueOnce(jsonResponse(202, { ok: true, data: applyPreview }))
            .mockResolvedValueOnce(
                jsonResponse(200, { ok: true, data: { applied: true, diff: applyPreview.preview.diff, intent_proof_id: "ip-2", audit_event_id: "audit-2" } }),
            );

        render(<ApplyConfigDialog kind="groups" id="group-1" initialDocument={"id: group-1\nspec:\n  name: Ops\n"} onClose={() => {}} />);
        fireEvent.change(screen.getByTestId("apply-config-editor"), { target: { value: "id: group-1\nspec:\n  name: Revenue Ops\n" } });
        fireEvent.click(screen.getByTestId("apply-config-preview"));

        await screen.findByTestId("config-governance-preview-summary");
        expect(screen.getByTestId("config-governance-diff-changed").textContent).toMatch("spec.name");

        fireEvent.click(screen.getByTestId("apply-config-approve"));
        expect(await screen.findByTestId("apply-config-success")).toBeTruthy();
    });
});
