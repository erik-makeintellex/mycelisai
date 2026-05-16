"use client";

import React, { useState } from "react";
import Link from "next/link";
import { ArrowRight, CalendarClock, Cable, ShieldCheck, Users, Sparkles } from "lucide-react";
import MissionProfileWizard from "@/components/automations/MissionProfileWizard";

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
    openTab: (tab: "triggers" | "approvals" | "wiring") => void;
    advancedMode: boolean;
}) {
    const [showWizard, setShowWizard] = useState(false);

    return (
        <div className="h-full p-6 overflow-y-auto">
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                <section className="space-y-4">
                    <div data-testid="automations-hub-baseline" className="sr-only">automations-hub-baseline</div>
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
                    {advancedMode && (
                        <>
                            <ActionCard
                                title="Workflow Builder"
                                description="Inspect and adjust advanced workflow structure when you need lower-level control."
                                onCreate={() => openTab("wiring")}
                                onView={() => openTab("wiring")}
                            />
                        </>
                    )}
                    <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-3">
                        <div className="flex items-start gap-2">
                            <Users className="mt-0.5 h-4 w-4 text-cortex-primary" />
                            <div>
                                <h3 className="text-sm font-semibold text-cortex-text-main">Team workstreams</h3>
                                <p className="text-xs text-cortex-text-muted mt-1">
                                    Manage active team leads, inspect outputs, and keep ongoing work visible from the dedicated Teams and Groups workspaces.
                                </p>
                            </div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            <Link href="/teams" className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10">
                                Open Teams
                            </Link>
                            <Link href="/groups" className="px-2.5 py-1.5 rounded border border-cortex-border text-cortex-text-main text-xs font-mono hover:bg-cortex-border">
                                Review Outputs
                            </Link>
                        </div>
                    </div>
                    <div className="rounded-xl border border-cortex-primary/25 bg-cortex-primary/10 p-4 space-y-3">
                        <div className="flex items-start justify-between gap-2">
                            <div>
                                <h3 className="text-sm font-semibold text-cortex-text-main flex items-center gap-2">
                                    <Sparkles className="w-4 h-4 text-cortex-primary" />
                                    Guided Setup
                                </h3>
                                <p className="text-xs text-cortex-text-muted mt-1">
                                    Walk through a guided path from objective to readiness without dropping into advanced workflow editing.
                                </p>
                            </div>
                            <button
                                onClick={() => setShowWizard((s) => !s)}
                                data-testid="open-mission-profile-wizard"
                                className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10"
                            >
                                {showWizard ? "Hide Wizard" : "Open Wizard"}
                            </button>
                        </div>
                        {showWizard ? <MissionProfileWizard openTab={openTab} /> : null}
                    </div>
                </section>

                <section className="space-y-4">
                    <h2 className="text-sm font-mono font-bold uppercase tracking-wide text-cortex-text-main">Actuation Paths</h2>
                    <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-2">
                        <div className="flex items-center gap-2 text-cortex-primary">
                            <CalendarClock className="w-4 h-4" />
                            <h3 className="text-sm font-semibold text-cortex-text-main">Review loop visibility</h3>
                        </div>
                        <p className="text-xs text-cortex-text-muted">
                            Organization workspaces expose real review-loop status. Use Trigger Rules here for mission, tool, artifact, and team events; cadence authoring belongs in the scheduler production lane.
                        </p>
                        <Link href="/dashboard" className="inline-flex px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10">
                            Open Soma workspace
                        </Link>
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
