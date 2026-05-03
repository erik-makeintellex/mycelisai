"use client";

import { Clock3, RadioTower, RefreshCcw } from "lucide-react";
import type { ProposalData } from "@/store/useCortexStore";

function cadenceLabel(proposal: ProposalData): string {
    if (proposal.task_cadence === "scheduled") return "Scheduled";
    if (proposal.task_cadence === "continuous") return "Keep running";
    if (proposal.task_cadence === "event_driven") return "Event-driven";
    if (proposal.tools.some((tool) => /schedule|trigger|review_loop/i.test(tool))) return "Scheduled";
    if (proposal.tools.some((tool) => /monitor|watch|listen/i.test(tool))) return "Keep running";
    return "Run once";
}

function cadenceSummary(proposal: ProposalData): string {
    if (proposal.schedule_summary?.trim()) return proposal.schedule_summary.trim();
    if (proposal.runtime_posture?.trim()) return proposal.runtime_posture.trim();
    switch (cadenceLabel(proposal)) {
        case "Scheduled":
            return "Soma should create or update a scheduled automation only after approval.";
        case "Keep running":
            return "Soma should leave an ongoing monitor or worker active until an admin stops it.";
        case "Event-driven":
            return "Soma should react to matching platform events instead of running immediately.";
        default:
            return "Soma should execute once and return the result here.";
    }
}

function busLabel(scope?: ProposalData["bus_scope"]): string {
    switch (scope) {
        case "current_team":
            return "Current team bus";
        case "multi_team":
            return "Multi-team bus";
        case "global":
            return "Global bus";
        default:
            return "No bus connection";
    }
}

export default function ProposalRunIntent({ proposal }: { proposal: ProposalData }) {
    const subjects = proposal.nats_subjects ?? [];
    const showBus = proposal.bus_scope && proposal.bus_scope !== "none";

    return (
        <div className="grid gap-3 md:grid-cols-2">
            <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                <div className="flex items-center gap-1.5 text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
                    {cadenceLabel(proposal) === "Run once" ? <Clock3 className="h-3 w-3" /> : <RefreshCcw className="h-3 w-3" />}
                    Task lifecycle
                </div>
                <p className="mt-1.5 text-sm font-medium text-cortex-text-main">{cadenceLabel(proposal)}</p>
                <p className="mt-1 text-xs leading-5 text-cortex-text-muted">{cadenceSummary(proposal)}</p>
            </div>
            <div className="rounded border border-cortex-border bg-cortex-bg/40 px-3 py-2.5">
                <div className="flex items-center gap-1.5 text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">
                    <RadioTower className="h-3 w-3" />
                    Team / NATS connection
                </div>
                <p className="mt-1.5 text-sm font-medium text-cortex-text-main">{busLabel(proposal.bus_scope)}</p>
                <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                    {showBus ? "Soma will connect this work to the configured agentry bus after approval." : "This work stays in the chat/run path unless you approve bus wiring."}
                </p>
                {subjects.length > 0 ? (
                    <div className="mt-2 flex flex-wrap gap-1">
                        {subjects.slice(0, 3).map((subject) => (
                            <span key={subject} className="rounded border border-cortex-primary/20 bg-cortex-primary/10 px-1.5 py-0.5 text-[9px] font-mono text-cortex-primary">
                                {subject}
                            </span>
                        ))}
                    </div>
                ) : null}
            </div>
        </div>
    );
}
