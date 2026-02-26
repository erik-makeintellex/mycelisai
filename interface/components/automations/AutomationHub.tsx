"use client";

import React from "react";
import { ArrowRight, Cable, ShieldCheck, Users, Clock3 } from "lucide-react";

function ActionCard({
    title,
    description,
    onCreate,
    onView,
}: {
    title: string;
    description: string;
    onCreate: () => void;
    onView: () => void;
}) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-3">
            <div>
                <h3 className="text-sm font-semibold text-cortex-text-main">{title}</h3>
                <p className="text-xs text-cortex-text-muted mt-1">{description}</p>
            </div>
            <div className="flex gap-2">
                <button onClick={onCreate} className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10">
                    Create
                </button>
                <button onClick={onView} className="px-2.5 py-1.5 rounded border border-cortex-border text-cortex-text-main text-xs font-mono hover:bg-cortex-border">
                    View Existing
                </button>
            </div>
        </div>
    );
}

export default function AutomationHub({
    openTab,
    advancedMode,
}: {
    openTab: (tab: "triggers" | "approvals" | "teams" | "wiring") => void;
    advancedMode: boolean;
}) {
    return (
        <div className="h-full p-6 overflow-y-auto">
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                <section className="space-y-4">
                    <h2 className="text-sm font-mono font-bold uppercase tracking-wide text-cortex-text-main">Available Now</h2>
                    <ActionCard
                        title="Trigger Rules"
                        description="Route mission events into controlled downstream actions."
                        onCreate={() => openTab("triggers")}
                        onView={() => openTab("triggers")}
                    />
                    <ActionCard
                        title="Approvals"
                        description="Review governance-gated proposals and release execution."
                        onCreate={() => openTab("approvals")}
                        onView={() => openTab("approvals")}
                    />
                    <ActionCard
                        title="Teams"
                        description="Inspect team readiness, online agents, and operating roles."
                        onCreate={() => openTab("teams")}
                        onView={() => openTab("teams")}
                    />
                    {advancedMode && (
                        <ActionCard
                            title="Neural Wiring"
                            description="Design and iterate team-agent wiring in advanced mode."
                            onCreate={() => openTab("wiring")}
                            onView={() => openTab("wiring")}
                        />
                    )}
                </section>

                <section className="space-y-4">
                    <h2 className="text-sm font-mono font-bold uppercase tracking-wide text-cortex-text-main">Coming Soon</h2>
                    <div className="rounded-xl border border-cortex-warning/30 bg-cortex-warning/10 p-4 space-y-2">
                        <div className="flex items-center gap-2 text-cortex-warning">
                            <Clock3 className="w-4 h-4" />
                            <h3 className="text-sm font-semibold">Scheduler</h3>
                        </div>
                        <p className="text-xs text-cortex-text-muted">Cron scheduling and recurring execution are planned for the next V7 milestone.</p>
                    </div>

                    <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4">
                        <button
                            onClick={() => openTab("triggers")}
                            className="w-full px-3 py-2 rounded border border-cortex-primary/30 text-cortex-primary font-mono text-sm hover:bg-cortex-primary/10 flex items-center justify-center gap-2"
                        >
                            Set Up Your First Automation Chain
                            <ArrowRight className="w-4 h-4" />
                        </button>
                        <ol className="mt-3 text-xs font-mono text-cortex-text-muted space-y-1.5">
                            <li className="flex items-center gap-2"><Cable className="w-3 h-3" />1. Create Trigger</li>
                            <li className="flex items-center gap-2"><ShieldCheck className="w-3 h-3" />2. Set Propose Mode</li>
                            <li className="flex items-center gap-2"><Users className="w-3 h-3" />3. Route to Team</li>
                            <li>4. Review Approval</li>
                            <li>5. Execute</li>
                        </ol>
                    </div>
                </section>
            </div>
        </div>
    );
}

