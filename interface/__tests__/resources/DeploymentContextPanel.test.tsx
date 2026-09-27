import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import DeploymentContextPanel from "@/components/resources/DeploymentContextPanel";

const KEYWORD_ONLY = "Saved. Soma can recall this by keywords; semantic search needs an embedding engine.";

type Reply = { ok: boolean; status?: number; body: unknown };

function routeFetch({ admin = false, save, entries = [] }: { admin?: boolean; save?: Reply; entries?: unknown[] }) {
    return vi.fn(async (url: string, init?: RequestInit) => {
        const reply = (r: Reply) => ({ ok: r.ok, status: r.status ?? (r.ok ? 200 : 500), json: async () => r.body, text: async () => JSON.stringify(r.body) });
        if (url === "/auth/session") {
            return reply({ ok: true, body: { data: { user: { role: admin ? "admin" : "operator" } } } });
        }
        if (init?.method === "POST" && save) {
            return reply(save);
        }
        return reply({ ok: true, body: { entries, count: entries.length, semantic_search: "unavailable" } });
    });
}

const listCalls = (fetchMock: ReturnType<typeof routeFetch>) =>
    fetchMock.mock.calls.filter(([url, init]) => String(url).startsWith("/api/v1/memory/deployment-context") && !init?.method).length;

const postBody = (fetchMock: ReturnType<typeof routeFetch>) =>
    JSON.parse(fetchMock.mock.calls.find(([, init]) => init?.method === "POST")?.[1]?.body as string);

async function fillAndSave(title = "Juniper & Rye Bakery") {
    await waitFor(() => expect(screen.getByText(/No long-term context sources saved yet/i)).toBeDefined());
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: title } });
    fireEvent.change(screen.getByLabelText("Content"), { target: { value: "Weekend special: blueberry-lavender scones, 3 for $10." } });
    fireEvent.click(screen.getByRole("button", { name: /Save context/i }));
}

describe("DeploymentContextPanel", () => {
    let fetchMock: ReturnType<typeof routeFetch>;

    beforeEach(() => sessionStorage.clear());
    afterEach(() => vi.unstubAllGlobals());

    const mount = (options: Parameters<typeof routeFetch>[0]) => {
        fetchMock = routeFetch(options);
        vi.stubGlobal("fetch", fetchMock);
        render(<DeploymentContextPanel />);
    };

    it("shows only title, content, and who can use it by default", async () => {
        mount({});
        await waitFor(() => expect(screen.getByText(/No long-term context sources saved yet/i)).toBeDefined());
        expect(screen.getByLabelText("Title")).toBeDefined();
        expect(screen.getByLabelText("Content")).toBeDefined();
        const audience = screen.getByRole("radiogroup", { name: "Who can use this" });
        expect(audience.textContent).toContain("Whole organization");
        expect(audience.textContent).toContain("My team");
        expect(audience.textContent).toContain("Only me");
        expect(screen.queryByLabelText("Use as")).toBeNull();
        expect(screen.queryByLabelText("Source kind")).toBeNull();
        expect(screen.queryByLabelText("Sensitivity")).toBeNull();
        expect(screen.queryByText(/^Saved. |Saved for future Soma use/)).toBeNull();
    });

    it("reveals the classification behind More options with the renamed labels", async () => {
        mount({});
        await waitFor(() => expect(screen.getByText(/No long-term context sources saved yet/i)).toBeDefined());
        fireEvent.click(screen.getByRole("button", { name: /More options/i }));
        expect(screen.getByLabelText("Use as")).toBeDefined();
        expect(screen.getByLabelText("Sensitivity")).toBeDefined();
        expect(screen.getByLabelText("Target goal sets")).toBeDefined();
        expect(screen.getByRole("option", { name: "Work log entry" })).toBeDefined();
        expect(screen.getByRole("option", { name: "Work log" })).toBeDefined();
        expect(screen.getByRole("option", { name: "Health & safety" })).toBeDefined();
        expect(screen.queryByRole("option", { name: /Diary/i })).toBeNull();
    });

    it("shows a visible error and no Saved state when the save fails, then reloads the list", async () => {
        mount({ save: { ok: false, status: 400, body: { error: "embed deployment context chunk 1: embedding failed" } } });
        await fillAndSave();
        const alert = await screen.findByRole("alert");
        expect(alert.textContent).toContain("embedding failed");
        expect(screen.queryByText(/^Saved. |Saved for future Soma use/)).toBeNull();
        await waitFor(() => expect(listCalls(fetchMock)).toBe(2));
    });

    it("shows a 503 save failure as an alert", async () => {
        mount({ save: { ok: false, status: 503, body: { error: "deployment context store unavailable" } } });
        await fillAndSave();
        expect((await screen.findByRole("alert")).textContent).toContain("store unavailable");
        expect(screen.queryByText(/^Saved. |Saved for future Soma use/)).toBeNull();
    });

    it("shows the honest keyword-only status after a save without embeddings", async () => {
        mount({ save: { ok: true, status: 201, body: {
            artifact_id: "ctx-1", knowledge_class: "customer_context", title: "Juniper & Rye Bakery", chunk_count: 1, vector_count: 0,
            embedding_status: "pending", retrieval_modes: ["keyword"], status_message: KEYWORD_ONLY,
        } } });
        await fillAndSave();
        await waitFor(() => expect(screen.getByText(KEYWORD_ONLY)).toBeDefined());
        expect(screen.queryByRole("alert")).toBeNull();
        await waitFor(() => expect(listCalls(fetchMock)).toBe(2));
    });

    it("defaults a non-admin to customer context", async () => {
        mount({ save: { ok: true, body: { title: "x", status_message: KEYWORD_ONLY } } });
        await fillAndSave();
        await waitFor(() => expect(screen.getByText(KEYWORD_ONLY)).toBeDefined());
        expect(postBody(fetchMock)).toMatchObject({ knowledge_class: "customer_context", visibility: "global", source_kind: "user_note" });
    });

    it("uses admin defaults when the session is a root admin", async () => {
        mount({ admin: true, save: { ok: true, body: { title: "x", status_message: KEYWORD_ONLY } } });
        await waitFor(() => expect(fetchMock.mock.calls.some(([url]) => url === "/auth/session")).toBe(true));
        await fillAndSave();
        await waitFor(() => expect(screen.getByText(KEYWORD_ONLY)).toBeDefined());
        expect(postBody(fetchMock)).toMatchObject({
            knowledge_class: "company_knowledge", source_kind: "user_note", visibility: "global",
            sensitivity_class: "role_scoped", trust_class: "trusted_internal",
        });
    });

    it("maps Only me to private visibility", async () => {
        mount({ save: { ok: true, body: { title: "x", status_message: KEYWORD_ONLY } } });
        await waitFor(() => expect(screen.getByText(/No long-term context sources saved yet/i)).toBeDefined());
        fireEvent.click(screen.getByRole("radio", { name: "Only me" }));
        await fillAndSave();
        await waitFor(() => expect(screen.getByText(KEYWORD_ONLY)).toBeDefined());
        expect(postBody(fetchMock).visibility).toBe("private");
    });

    it("renders saved entries with their honest search status", async () => {
        mount({ entries: [{
            artifact_id: "ctx-1", knowledge_class: "company_knowledge", title: "Juniper & Rye Bakery", source_label: "operator provided",
            source_kind: "worklog_entry", visibility: "global", sensitivity_class: "role_scoped", trust_class: "user_provided",
            chunk_count: 1, vector_count: 0, content_preview: "Weekend special: blueberry-lavender scones.", content_length: 44,
            target_goal_sets: ["spring launch"], created_at: "2026-09-27T12:00:00Z", embedding_status: "pending",
        }] });
        await waitFor(() => expect(screen.getByText("Juniper & Rye Bakery")).toBeDefined());
        expect(screen.getByText("Keyword search")).toBeDefined();
        expect(screen.getByText(/Work log entry/)).toBeDefined();
        expect(screen.getByText(/goal: spring launch/i)).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /Ask Soma with this/i }));
        expect(JSON.parse(sessionStorage.getItem("mycelis:pending-soma-output-continuation") ?? "{}")).toMatchObject({
            title: "Juniper & Rye Bakery", reference: "memory/deployment-context/ctx-1",
        });
    });
});
