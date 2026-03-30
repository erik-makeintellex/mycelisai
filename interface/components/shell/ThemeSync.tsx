"use client";

import { useEffect } from "react";
import { useCortexStore } from "@/store/useCortexStore";

function resolveAppliedTheme(theme: "aero-light" | "midnight-cortex" | "system") {
    if (theme === "system" && typeof window !== "undefined") {
        return window.matchMedia("(prefers-color-scheme: dark)").matches ? "midnight-cortex" : "aero-light";
    }
    return theme;
}

export function ThemeSync() {
    const theme = useCortexStore((s) => s.theme);

    useEffect(() => {
        const root = document.documentElement;
        const applyTheme = () => {
            root.dataset.theme = resolveAppliedTheme(theme);
        };

        applyTheme();

        if (theme !== "system" || typeof window === "undefined") {
            return;
        }

        const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
        const onChange = () => applyTheme();
        mediaQuery.addEventListener("change", onChange);
        return () => mediaQuery.removeEventListener("change", onChange);
    }, [theme]);

    return null;
}
