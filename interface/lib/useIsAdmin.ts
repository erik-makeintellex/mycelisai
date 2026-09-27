"use client";

import { useEffect, useState } from "react";

/**
 * Fetches the viewer's role once. Fails closed: any non-ok response, thrown
 * error, or non-JSON body defaults to "standard", never "admin". `checked`
 * stays false until the request settles, so callers can avoid flashing the
 * wrong state before the first response.
 *
 * A small local helper rather than a shared session hook: `useWebSession`
 * only exists on U1's branch, not on dev.
 */
export function useIsAdmin(): { isAdmin: boolean; checked: boolean } {
    const [isAdmin, setIsAdmin] = useState(false);
    const [checked, setChecked] = useState(false);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            try {
                const res = await fetch("/auth/session", { cache: "no-store" });
                const body = res?.ok ? await res.json().catch(() => null) : null;
                if (!cancelled) setIsAdmin(body?.data?.user?.role === "admin");
            } catch {
                if (!cancelled) setIsAdmin(false);
            } finally {
                if (!cancelled) setChecked(true);
            }
        })();
        return () => {
            cancelled = true;
        };
    }, []);

    return { isAdmin, checked };
}
