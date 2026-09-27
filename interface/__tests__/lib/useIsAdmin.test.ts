import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useIsAdmin } from "@/lib/useIsAdmin";
import { mockFetch } from "../setup";

describe("useIsAdmin", () => {
    it("reports admin true when the session role is admin", async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role: "admin" } } }) });
        const { result } = renderHook(() => useIsAdmin());
        await waitFor(() => expect(result.current.checked).toBe(true));
        expect(result.current.isAdmin).toBe(true);
    });

    it("defaults to non-admin for a standard user", async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ data: { user: { role: "standard" } } }) });
        const { result } = renderHook(() => useIsAdmin());
        await waitFor(() => expect(result.current.checked).toBe(true));
        expect(result.current.isAdmin).toBe(false);
    });

    it("fails closed (non-admin) when the session fetch fails", async () => {
        mockFetch.mockResolvedValue({ ok: false, json: async () => ({}) });
        const { result } = renderHook(() => useIsAdmin());
        await waitFor(() => expect(result.current.checked).toBe(true));
        expect(result.current.isAdmin).toBe(false);
    });

    it("fails closed when fetch rejects or returns a non-promise", async () => {
        mockFetch.mockImplementation(() => undefined as unknown as Promise<Response>);
        const { result } = renderHook(() => useIsAdmin());
        await waitFor(() => expect(result.current.checked).toBe(true));
        expect(result.current.isAdmin).toBe(false);
    });
});
