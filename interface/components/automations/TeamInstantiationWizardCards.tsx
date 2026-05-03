"use client";

import { Check } from "lucide-react";
import type { TeamProfileTemplate } from "@/lib/workflowContracts";

type StepBadgeProps = { active: boolean; complete: boolean; label: string };

export function StepBadge({ active, complete, label }: StepBadgeProps) {
    return (
        <div
            className={`px-2.5 py-1 rounded-md text-[10px] font-mono border ${
                active
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-primary"
                    : complete
                      ? "border-cortex-success/40 bg-cortex-success/10 text-cortex-success"
                      : "border-cortex-border text-cortex-text-muted"
            }`}
        >
            {label}
        </div>
    );
}

export function ProfileCard({
    profile,
    selected,
    onSelect,
}: {
    profile: TeamProfileTemplate;
    selected: boolean;
    onSelect: () => void;
}) {
    return (
        <button
            onClick={onSelect}
            className={`w-full text-left rounded-lg border p-3 transition-colors ${
                selected
                    ? "border-cortex-primary/40 bg-cortex-primary/10"
                    : "border-cortex-border bg-cortex-surface hover:bg-cortex-bg"
            }`}
        >
            <div className="flex items-start justify-between gap-2">
                <div>
                    <p className="text-sm font-semibold text-cortex-text-main">{profile.name}</p>
                    <p className="text-[11px] text-cortex-text-muted mt-1">{profile.description}</p>
                </div>
                {selected && <Check className="w-4 h-4 text-cortex-primary mt-0.5" />}
            </div>
            <div className="mt-2 flex flex-wrap gap-1">
                {profile.requiredCapabilities.map((cap) => (
                    <span key={cap} className="px-1.5 py-0.5 rounded border border-cortex-border text-[10px] font-mono text-cortex-text-muted">
                        {cap}
                    </span>
                ))}
            </div>
        </button>
    );
}
