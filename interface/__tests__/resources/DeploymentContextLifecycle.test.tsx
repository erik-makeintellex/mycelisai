import { describe, it, expect, vi, afterEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import DeploymentContextPanel from "@/components/resources/DeploymentContextPanel";

type Reply = { ok: boolean; status?: number; body: unknown };

const entry = (overrides: Record<string, unknown> = {}) => ({
    artifact_id: "ctx-1", knowledge_class: "customer_context", title: "Juniper & Rye Bakery", source_label: "operator provided",
    source_kind: "user_note", visibility: "private", sensitivity_class: "role_scoped", trust_class: "user_provided",
    chunk_count: 1, vector_count: 0, content_preview: "Weekend special: blueberry-lavender scones.", content_length: 44,
    created_at: "2026-09-27T12:00:00Z", embedding_status: "pending", lifecycle_state: "active", can_manage: true,
    ...overrides,
});

function routeFetch(entries: unknown[], change: Reply = { ok: true, body: { ok: true, data: { changed: true } } }, edit: Reply = { ok: true, body: { ok: true, data: { changed: true } } }) {
    return vi.fn(async (url: string, init?: RequestInit) => {
        const reply = (r: Reply) => ({ ok: r.ok, status: r.status ?? (r.ok ? 200 : 500), json: async () => r.body });
        if (url === "/auth/session") return reply({ ok: true, body: { data: { user: { role: "operator" } } } });
        if (init?.method === "PATCH") return reply(edit);
        if (init?.method === "POST" || init?.method === "DELETE") return reply(change);
        const archived = url.includes("include_archived=true");
        const visible = entries.filter((e) => archived || (e as { lifecycle_state?: string }).lifecycle_state !== "archived");
        return reply({ ok: true, body: { entries: visible, count: visible.length, semantic_search: "unavailable" } });
    });
}

const changeCalls = (fetchMock: ReturnType<typeof routeFetch>) =>
    fetchMock.mock.calls.filter(([, init]) => init?.method === "POST" || init?.method === "DELETE").map(([url, init]) => `${init?.method} ${url}`);
const editCalls = (fetchMock: ReturnType<typeof routeFetch>) =>
    fetchMock.mock.calls.filter(([, init]) => init?.method === "PATCH").map(([url, init]) => ({ url: String(url), body: JSON.parse(String(init?.body)) }));
const listUrls = (fetchMock: ReturnType<typeof routeFetch>) =>
    fetchMock.mock.calls.filter(([url, init]) => String(url).startsWith("/api/v1/memory/deployment-context?") && !init?.method).map(([url]) => String(url));

describe("DeploymentContextPanel lifecycle actions", () => {
    let fetchMock: ReturnType<typeof routeFetch>;
    afterEach(() => vi.unstubAllGlobals());

    const mount = (entries: unknown[], change?: Reply, edit?: Reply) => {
        fetchMock = routeFetch(entries, change, edit);
        vi.stubGlobal("fetch", fetchMock);
        render(<DeploymentContextPanel />);
    };

    it("archives an entry the viewer manages and reloads the list", async () => {
        mount([entry()]);
        fireEvent.click(await screen.findByRole("button", { name: "Archive Juniper & Rye Bakery" }));
        await waitFor(() => expect(changeCalls(fetchMock)).toEqual(["POST /api/v1/memory/deployment-context/ctx-1/archive"]));
        await waitFor(() => expect(listUrls(fetchMock).length).toBe(2));
        expect(screen.queryByRole("alert")).toBeNull();
    });

    it("hides the actions for entries the viewer cannot manage", async () => {
        mount([entry({ can_manage: false })]);
        await screen.findByText("Juniper & Rye Bakery");
        expect(screen.queryByRole("button", { name: /Archive Juniper/ })).toBeNull();
        expect(screen.queryByRole("button", { name: /Delete Juniper/ })).toBeNull();
    });

    it("asks before a permanent delete and does nothing on cancel", async () => {
        mount([entry()]);
        fireEvent.click(await screen.findByRole("button", { name: "Delete Juniper & Rye Bakery" }));
        expect(screen.getByText("Delete permanently? Soma will no longer use this.")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
        expect(screen.queryByText("Delete permanently? Soma will no longer use this.")).toBeNull();
        expect(changeCalls(fetchMock)).toEqual([]);

        fireEvent.click(screen.getByRole("button", { name: "Delete Juniper & Rye Bakery" }));
        fireEvent.click(screen.getByRole("button", { name: "Delete permanently" }));
        await waitFor(() => expect(changeCalls(fetchMock)).toEqual(["DELETE /api/v1/memory/deployment-context/ctx-1"]));
        await waitFor(() => expect(listUrls(fetchMock).length).toBe(2));
    });

    it("shows archived entries on request and restores them", async () => {
        mount([entry({ lifecycle_state: "archived", archived_at: "2026-09-27T13:00:00Z" })]);
        await waitFor(() => expect(screen.getByText(/No long-term context sources saved yet/i)).toBeDefined());
        fireEvent.click(screen.getByRole("checkbox", { name: "Show archived" }));
        await screen.findByText("Juniper & Rye Bakery");
        expect(listUrls(fetchMock).at(-1)).toContain("include_archived=true");
        expect(screen.getByText("Archived")).toBeDefined();
        expect(screen.queryByRole("button", { name: /Archive Juniper/ })).toBeNull();
        fireEvent.click(screen.getByRole("button", { name: "Restore Juniper & Rye Bakery" }));
        await waitFor(() => expect(changeCalls(fetchMock)).toEqual(["POST /api/v1/memory/deployment-context/ctx-1/restore"]));
    });

    it("shows plain blocker copy when the server refuses the change", async () => {
        mount([entry()], { ok: false, status: 403, body: { ok: false, error: "raw", data: { code: "memory_entry_not_owned" } } });
        fireEvent.click(await screen.findByRole("button", { name: "Archive Juniper & Rye Bakery" }));
        const alert = await screen.findByRole("alert");
        expect(alert.textContent).toContain("Only the person who saved this can change it");
        expect(alert.textContent).toContain("Nothing was changed");
    });

    it("explains that organization memory needs an admin", async () => {
        mount([entry({ visibility: "global" })], { ok: false, status: 403, body: { ok: false, error: "raw", data: { code: "admin_required" } } });
        fireEvent.click(await screen.findByRole("button", { name: "Delete Juniper & Rye Bakery" }));
        fireEvent.click(screen.getByRole("button", { name: "Delete permanently" }));
        expect((await screen.findByRole("alert")).textContent).toContain("Organization memory is managed by admins");
    });
});

describe("DeploymentContextPanel edit-in-place", () => {
    let fetchMock: ReturnType<typeof routeFetch>;
    afterEach(() => vi.unstubAllGlobals());

    const mount = (entries: unknown[], edit?: Reply) => {
        fetchMock = routeFetch(entries, undefined, edit);
        vi.stubGlobal("fetch", fetchMock);
        render(<DeploymentContextPanel />);
    };

    it("opens an inline editor prefilled with the title and source label, sends only changed fields, and refreshes on success", async () => {
        mount([entry()]);
        fireEvent.click(await screen.findByRole("button", { name: "Edit Juniper & Rye Bakery" }));

        const titleInput = screen.getByDisplayValue("Juniper & Rye Bakery");
        const labelInput = screen.getByDisplayValue("operator provided");
        fireEvent.change(titleInput, { target: { value: "Juniper & Rye Bakehouse" } });

        fireEvent.click(screen.getByRole("button", { name: "Save Juniper & Rye Bakery" }));
        await waitFor(() => expect(editCalls(fetchMock)).toEqual([{ url: "/api/v1/memory/deployment-context/ctx-1", body: { title: "Juniper & Rye Bakehouse" } }]));
        // Untouched fields (source label, content) are never sent.
        expect(labelInput).toBeDefined();
        await waitFor(() => expect(listUrls(fetchMock).length).toBe(2));
        expect(screen.queryByRole("alert")).toBeNull();
    });

    it("sends content only when the operator types replacement text, never the truncated preview", async () => {
        mount([entry()]);
        fireEvent.click(await screen.findByRole("button", { name: "Edit Juniper & Rye Bakery" }));
        fireEvent.change(screen.getByPlaceholderText(/Leave blank to keep the saved content/i), { target: { value: "New scone flavor: fig and rosemary." } });
        fireEvent.click(screen.getByRole("button", { name: "Save Juniper & Rye Bakery" }));
        await waitFor(() => expect(editCalls(fetchMock)).toEqual([{ url: "/api/v1/memory/deployment-context/ctx-1", body: { content: "New scone flavor: fig and rosemary." } }]));
    });

    it("disables Save until a field actually changes, so an empty patch is never sent", async () => {
        mount([entry()]);
        fireEvent.click(await screen.findByRole("button", { name: "Edit Juniper & Rye Bakery" }));
        expect((screen.getByRole("button", { name: "Save Juniper & Rye Bakery" }) as HTMLButtonElement).disabled).toBe(true);
    });

    it("hides Edit for archived entries and shows a restore-to-edit hint instead", async () => {
        mount([entry({ lifecycle_state: "archived" })]);
        fireEvent.click(await screen.findByRole("checkbox", { name: "Show archived" }));
        await screen.findByText("Juniper & Rye Bakery");
        expect(screen.queryByRole("button", { name: /^Edit /i })).toBeNull();
        expect(screen.getByText(/restore to edit/i)).toBeDefined();
    });

    it("shows honest blocker copy and keeps the draft open on a 409 memory_entry_archived race", async () => {
        mount([entry()], { ok: false, status: 409, body: { ok: false, error: "raw", data: { code: "memory_entry_archived" } } });
        fireEvent.click(await screen.findByRole("button", { name: "Edit Juniper & Rye Bakery" }));
        fireEvent.change(screen.getByDisplayValue("Juniper & Rye Bakery"), { target: { value: "Renamed" } });
        fireEvent.click(screen.getByRole("button", { name: "Save Juniper & Rye Bakery" }));
        const alert = await screen.findByRole("alert");
        expect(alert.textContent).toContain("Restore this before editing");
        expect(alert.textContent).toContain("Nothing was saved");
        // Never a silent success: the editor form is still present.
        expect(screen.getByDisplayValue("Renamed")).toBeDefined();
    });

    it("offers Reload on a 409 memory_entry_changed and refreshes the list when chosen", async () => {
        mount([entry()], { ok: false, status: 409, body: { ok: false, error: "raw", data: { code: "memory_entry_changed" } } });
        fireEvent.click(await screen.findByRole("button", { name: "Edit Juniper & Rye Bakery" }));
        fireEvent.change(screen.getByDisplayValue("Juniper & Rye Bakery"), { target: { value: "Renamed" } });
        fireEvent.click(screen.getByRole("button", { name: "Save Juniper & Rye Bakery" }));
        const alert = await screen.findByRole("alert");
        expect(alert.textContent).toContain("changed since you opened it");
        fireEvent.click(screen.getByRole("button", { name: "Reload" }));
        await waitFor(() => expect(listUrls(fetchMock).length).toBe(2));
        // The editor closes after a reload; it does not resubmit the stale draft.
        expect(screen.queryByRole("button", { name: "Save Juniper & Rye Bakery" })).toBeNull();
    });

    it("cancels without saving and discards the draft", async () => {
        mount([entry()]);
        fireEvent.click(await screen.findByRole("button", { name: "Edit Juniper & Rye Bakery" }));
        fireEvent.change(screen.getByDisplayValue("Juniper & Rye Bakery"), { target: { value: "Abandoned edit" } });
        fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
        expect(screen.queryByDisplayValue("Abandoned edit")).toBeNull();
        expect(editCalls(fetchMock)).toEqual([]);
        expect(await screen.findByText("Juniper & Rye Bakery")).toBeDefined();
    });
});
