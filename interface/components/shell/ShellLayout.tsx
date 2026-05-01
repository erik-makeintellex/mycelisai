"use client";

import React, { useEffect } from "react";
import { ZoneA } from "./ZoneA_Rail";
import { ZoneB } from "./ZoneB_Workspace";
import DegradedModeBanner from "@/components/dashboard/DegradedModeBanner";
import StatusDrawer from "@/components/dashboard/StatusDrawer";
import SignalDetailDrawer from "@/components/stream/SignalDetailDrawer";
import { useCortexStore } from "@/store/useCortexStore";
import { ThemeSync } from "@/components/shell/ThemeSync";

interface ShellLayoutProps {
    children?: React.ReactNode;
}

export function ShellLayout({ children }: ShellLayoutProps) {
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);
    const fetchUserSettings = useCortexStore((s) => s.fetchUserSettings);
    const initializeStream = useCortexStore((s) => s.initializeStream);

    useEffect(() => {
        fetchUserSettings();
        fetchServicesStatus();
        initializeStream();
        const interval = setInterval(() => {
            fetchServicesStatus();
        }, 6000);
        return () => clearInterval(interval);
    }, [fetchServicesStatus, fetchUserSettings, initializeStream]);

    return (
        <div className="flex h-screen w-screen overflow-hidden bg-cortex-bg text-cortex-text-main font-sans selection:bg-cortex-primary/30 selection:text-white">
            <ThemeSync />

            {/* ZONE A: Rail (Navigation & Vitals) */}
            <ZoneA />

            {/* CENTER STAGE */}
            <div className="flex-1 flex flex-col min-w-0 bg-cortex-bg relative">
                <DegradedModeBanner />

                {/* ZONE B: Workspace (The Canvas) */}
                <ZoneB>
                    {children}
                </ZoneB>

                <SignalDetailDrawer />
            </div>

            <StatusDrawer />

        </div>
    );
}
