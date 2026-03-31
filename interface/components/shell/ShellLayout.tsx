"use client";

import React, { useEffect } from "react";
import { ZoneA } from "./ZoneA_Rail";
import { ZoneB } from "./ZoneB_Workspace";
import DegradedModeBanner from "@/components/dashboard/DegradedModeBanner";
import StatusDrawer from "@/components/dashboard/StatusDrawer";
import { Activity } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { ThemeSync } from "@/components/shell/ThemeSync";

interface ShellLayoutProps {
    children?: React.ReactNode;
}

export function ShellLayout({ children }: ShellLayoutProps) {
    const setStatusDrawerOpen = useCortexStore((s) => s.setStatusDrawerOpen);
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

            </div>

            <StatusDrawer />
            <button
                onClick={() => setStatusDrawerOpen(true)}
                className="fixed right-4 bottom-4 z-30 px-2.5 py-1.5 rounded-lg border border-cortex-primary/30 bg-cortex-surface text-cortex-primary hover:bg-cortex-primary/10 text-[10px] font-mono flex items-center gap-1.5"
                title="Open Status Drawer"
            >
                <Activity className="w-3.5 h-3.5" />
                Status
            </button>

        </div>
    );
}
