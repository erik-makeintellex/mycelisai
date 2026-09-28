import { afterEach, describe, expect, it, vi } from "vitest";
import { editDeploymentContextEntry } from "@/lib/deploymentContextEdit";

type Reply = { ok: boolean; status?: number; body: unknown };

function stubFetch(reply: Reply) {
    const fetchMock = vi.fn(async (_url: string, _init?: RequestInit) => ({
        ok: reply.ok,
        status: reply.status ?? (reply.ok ? 200 : 500),
        json: async () => reply.body,
    }));
    vi.stubGlobal("fetch", fetchMock);
    return fetchMock;
}

describe("editDeploymentContextEntry", () => {
    afterEach(() => vi.unstubAllGlobals());

    it("PATCHes only the given fields and reports changed:true on success", async () => {
        const fetchMock = stubFetch({ ok: true, body: { ok: true, data: { artifact_id: "ctx-1", changed: true, title: "New title" } } });
        const result = await editDeploymentContextEntry("ctx-1", { title: "New title" });
        expect(result).toEqual({ ok: true, changed: true });
        const [url, init] = fetchMock.mock.calls[0];
        expect(url).toBe("/api/v1/memory/deployment-context/ctx-1");
        expect(init?.method).toBe("PATCH");
        expect(init?.headers).toMatchObject({ "Content-Type": "application/json" });
        expect(JSON.parse(String(init?.body))).toEqual({ title: "New title" });
    });

    it("reports changed:false for an identical patch", async () => {
        stubFetch({ ok: true, body: { ok: true, data: { artifact_id: "ctx-1", changed: false } } });
        const result = await editDeploymentContextEntry("ctx-1", { title: "Same title" });
        expect(result).toEqual({ ok: true, changed: false });
    });

    it("encodes the artifact id in the URL", async () => {
        const fetchMock = stubFetch({ ok: true, body: { ok: true, data: { changed: true } } });
        await editDeploymentContextEntry("ctx/weird id", { title: "x" });
        expect(fetchMock.mock.calls[0][0]).toBe("/api/v1/memory/deployment-context/ctx%2Fweird%20id");
    });

    it.each([
        [400, "memory_entry_not_owned"],
        [403, "admin_required"],
        [404, "memory_entry_not_found"],
        [409, "memory_entry_archived"],
        [409, "memory_entry_changed"],
        [503, "service_unavailable"],
    ])("surfaces the blocker code on a %s response", async (status, code) => {
        stubFetch({ ok: false, status, body: { ok: false, error: "raw backend text", data: { code } } });
        const result = await editDeploymentContextEntry("ctx-1", { content: "new text" });
        expect(result).toEqual({ ok: false, code, httpStatus: status });
    });

    it("falls back to request_failed when the error body has no code", async () => {
        stubFetch({ ok: false, status: 500, body: {} });
        const result = await editDeploymentContextEntry("ctx-1", { title: "x" });
        expect(result).toEqual({ ok: false, code: "request_failed", httpStatus: 500 });
    });

    it("falls back to request_failed on a network exception, never throwing", async () => {
        vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("network down"); }));
        const result = await editDeploymentContextEntry("ctx-1", { title: "x" });
        expect(result).toEqual({ ok: false, code: "request_failed", httpStatus: 0 });
    });
});
