"use client";

import React from "react";
import { ZoneA } from "./ZoneA_Rail";
import { ZoneB } from "./ZoneB_Workspace";
import { ZoneD } from "./ZoneD_Decision";

interface ShellLayoutProps {
    children?: React.ReactNode;
}

export function ShellLayout({ children }: ShellLayoutProps) {
    return (
        <div className="flex h-screen w-screen overflow-hidden bg-cortex-bg text-cortex-text-main font-sans selection:bg-cortex-primary/30 selection:text-white">

            {/* ZONE A: Rail (Navigation & Vitals) */}
            <ZoneA />

            {/* CENTER STAGE */}
            <div className="flex-1 flex flex-col min-w-0 bg-cortex-bg relative">

                {/* ZONE B: Workspace (The Canvas) */}
                <ZoneB>
                    {children}
                </ZoneB>

            </div>

            {/* ZONE D: Decision (Governance Overlay) */}
            <ZoneD />

        </div>
    );
}
