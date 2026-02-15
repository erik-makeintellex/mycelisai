"use client";

import React from "react";
import { ZoneA } from "./ZoneA_Rail";
import { ZoneB } from "./ZoneB_Workspace";
import { ZoneC } from "./ZoneC_Stream";
import { ZoneD } from "./ZoneD_Decision";

interface ShellLayoutProps {
    children?: React.ReactNode;
}

export function ShellLayout({ children }: ShellLayoutProps) {
    return (
        <div className="flex h-screen w-screen overflow-hidden bg-zinc-950 text-zinc-900 font-sans selection:bg-indigo-100 selection:text-indigo-900">

            {/* ZONE A: Rail (Navigation & Vitals) */}
            <ZoneA />

            {/* CENTER STAGE */}
            <div className="flex-1 flex flex-col min-w-0 bg-zinc-100 relative">

                {/* ZONE B: Workspace (The Canvas) */}
                {/* We pass children here so Next.js pages render inside Zone B by default */}
                <ZoneB>
                    {children}
                </ZoneB>

            </div>

            {/* ZONE C: Stream (Telemetry & Audit) */}
            <ZoneC />

            {/* ZONE D: Decision (Governance Overlay) */}
            <ZoneD />

        </div>
    );
}
