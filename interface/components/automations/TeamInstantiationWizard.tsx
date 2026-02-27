"use client";

import React, { useMemo, useState } from "react";
import { ArrowRight, Check, Play, Save } from "lucide-react";
import CapabilityReadinessGateCard from "@/components/automations/CapabilityReadinessGateCard";
import {
    TEAM_PROFILE_TEMPLATES,
    type ReadinessSnapshot,
    type TeamProfileTemplate,
} from "@/lib/workflowContracts";

const STEPS = ["Objective", "Profile", "Readiness", "Launch"] as const;
type StepIndex = 0 | 1 | 2 | 3;

interface TeamInstantiationWizardProps {
    openTab: (tab: "triggers" | "approvals" | "teams" | "wiring") => void;
}

function StepBadge({ active, complete, label }: { active: boolean; complete: boolean; label: string }) {
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

function ProfileCard({
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
                {selected ? <Check className="w-4 h-4 text-cortex-primary mt-0.5" /> : null}
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

export default function TeamInstantiationWizard({ openTab }: TeamInstantiationWizardProps) {
    const [step, setStep] = useState<StepIndex>(0);
    const [objective, setObjective] = useState("");
    const [profileId, setProfileId] = useState<string>(TEAM_PROFILE_TEMPLATES[0].id);
    const [readiness, setReadiness] = useState<ReadinessSnapshot | null>(null);
    const [lastAction, setLastAction] = useState("");

    const selectedProfile = useMemo(
        () => TEAM_PROFILE_TEMPLATES.find((p) => p.id === profileId) ?? TEAM_PROFILE_TEMPLATES[0],
        [profileId]
    );

    const nextDisabled = useMemo(() => {
        if (step === 0) return objective.trim().length < 12;
        return false;
    }, [step, objective]);

    const canLaunchNow = readiness ? readiness.blockers.length === 0 : false;

    const next = () => setStep((prev) => (prev < 3 ? ((prev + 1) as StepIndex) : prev));
    const back = () => setStep((prev) => (prev > 0 ? ((prev - 1) as StepIndex) : prev));

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-4">
            <div className="flex items-center justify-between gap-2">
                <div>
                    <h3 className="text-sm font-semibold text-cortex-text-main">Team Instantiation Wizard</h3>
                    <p className="text-[11px] text-cortex-text-muted mt-1">
                        Sprint 0 scaffold: objective, profile, readiness, and governed launch review.
                    </p>
                </div>
                <div className="flex items-center gap-1.5">
                    {STEPS.map((label, idx) => (
                        <StepBadge key={label} label={label} active={idx === step} complete={idx < step} />
                    ))}
                </div>
            </div>

            {step === 0 && (
                <div className="space-y-2">
                    <label className="text-[11px] font-mono text-cortex-text-muted">Objective brief</label>
                    <textarea
                        value={objective}
                        onChange={(e) => setObjective(e.target.value)}
                        placeholder="Describe the outcome you want this team to execute."
                        className="w-full min-h-[110px] rounded-lg border border-cortex-border bg-cortex-bg p-3 text-sm text-cortex-text-main focus:outline-none focus:border-cortex-primary"
                    />
                    <p className="text-[10px] text-cortex-text-muted">
                        Tip: include desired output and any governance constraints.
                    </p>
                </div>
            )}

            {step === 1 && (
                <div className="space-y-2">
                    <p className="text-[11px] font-mono text-cortex-text-muted">Select a starter profile</p>
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-2">
                        {TEAM_PROFILE_TEMPLATES.map((profile) => (
                            <ProfileCard
                                key={profile.id}
                                profile={profile}
                                selected={profile.id === profileId}
                                onSelect={() => setProfileId(profile.id)}
                            />
                        ))}
                    </div>
                </div>
            )}

            {step === 2 && (
                <div className="space-y-2">
                    <CapabilityReadinessGateCard onSnapshotChange={setReadiness} />
                    <div className="rounded-lg border border-cortex-border bg-cortex-bg p-3">
                        <p className="text-[11px] font-semibold text-cortex-text-main mb-1">Suggested routes</p>
                        <div className="flex flex-wrap gap-1">
                            {selectedProfile.suggestedRoutes.map((route) => (
                                <span key={route} className="px-1.5 py-0.5 rounded border border-cortex-border text-[10px] font-mono text-cortex-text-muted">
                                    {route}
                                </span>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            {step === 3 && (
                <div className="space-y-3">
                    <div className="rounded-lg border border-cortex-border bg-cortex-bg p-3 space-y-2">
                        <p className="text-[11px] font-semibold text-cortex-text-main">Launch review</p>
                        <p className="text-xs text-cortex-text-main">
                            <span className="text-cortex-text-muted">Objective:</span> {objective || "(not set)"}
                        </p>
                        <p className="text-xs text-cortex-text-main">
                            <span className="text-cortex-text-muted">Profile:</span> {selectedProfile.name}
                        </p>
                        <p className="text-xs text-cortex-text-main">
                            <span className="text-cortex-text-muted">Inputs:</span> {selectedProfile.inputChannels.join(", ")}
                        </p>
                        <p className="text-xs text-cortex-text-main">
                            <span className="text-cortex-text-muted">Outputs:</span> {selectedProfile.outputChannels.join(", ")}
                        </p>
                        <p className="text-xs text-cortex-text-main">
                            <span className="text-cortex-text-muted">Readiness blockers:</span>{" "}
                            {readiness?.blockers.length ? readiness.blockers.length : 0}
                        </p>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
                        <button
                            disabled={!canLaunchNow}
                            onClick={() => {
                                setLastAction("Launch-now scaffolded. Next: open Teams and monitor run creation.");
                                openTab("teams");
                            }}
                            className="px-3 py-2 rounded border border-cortex-success/30 text-cortex-success text-xs font-mono hover:bg-cortex-success/10 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-1"
                        >
                            <Play className="w-3.5 h-3.5" />
                            Launch Now
                        </button>
                        <button
                            onClick={() => {
                                setLastAction("Propose-only route selected. Next: open Approvals to review release.");
                                openTab("approvals");
                            }}
                            className="px-3 py-2 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10 inline-flex items-center justify-center gap-1"
                        >
                            <ArrowRight className="w-3.5 h-3.5" />
                            Launch Propose-Only
                        </button>
                        <button
                            onClick={() => {
                                setLastAction("Template save scaffolded. Next: finalize profile persistence API wiring.");
                            }}
                            className="px-3 py-2 rounded border border-cortex-border text-cortex-text-main text-xs font-mono hover:bg-cortex-border inline-flex items-center justify-center gap-1"
                        >
                            <Save className="w-3.5 h-3.5" />
                            Save Template
                        </button>
                    </div>
                    {lastAction ? (
                        <div className="rounded-md border border-cortex-primary/30 bg-cortex-primary/10 px-2.5 py-2 text-[11px] text-cortex-text-main">
                            {lastAction}
                        </div>
                    ) : null}
                </div>
            )}

            <div className="flex items-center justify-between pt-1">
                <button
                    onClick={back}
                    disabled={step === 0}
                    className="px-2.5 py-1.5 rounded border border-cortex-border text-[11px] font-mono text-cortex-text-main hover:bg-cortex-border disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    Back
                </button>
                <button
                    onClick={next}
                    disabled={step === 3 || nextDisabled}
                    className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-[11px] font-mono text-cortex-primary hover:bg-cortex-primary/10 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    Continue
                </button>
            </div>
        </div>
    );
}
