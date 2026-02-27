"use client";

import React, { useEffect, useMemo, useState } from "react";
import { RotateCcw, Wand2 } from "lucide-react";
import {
    BUS_EXPOSURE_MODES,
    ROUTE_TEMPLATE_OPTIONS,
    type BusExposureMode,
    type TeamProfileTemplate,
} from "@/lib/workflowContracts";

interface RouteTemplatePickerProps {
    profile: TeamProfileTemplate;
    onRoutesChange?: (routes: string[]) => void;
    onBusModeChange?: (mode: BusExposureMode) => void;
}

function uniqueRoutes(routes: string[]): string[] {
    return Array.from(new Set(routes.map((r) => r.trim()).filter(Boolean)));
}

export default function RouteTemplatePicker({ profile, onRoutesChange, onBusModeChange }: RouteTemplatePickerProps) {
    const [mode, setMode] = useState<BusExposureMode>("basic");
    const [selectedTemplateId, setSelectedTemplateId] = useState<string>(ROUTE_TEMPLATE_OPTIONS[0].id);
    const [routes, setRoutes] = useState<string[]>(profile.suggestedRoutes);
    const [expertInput, setExpertInput] = useState<string>(profile.suggestedRoutes.join("\n"));

    useEffect(() => {
        setRoutes(profile.suggestedRoutes);
        setExpertInput(profile.suggestedRoutes.join("\n"));
    }, [profile.id, profile.suggestedRoutes]);

    useEffect(() => {
        onRoutesChange?.(routes);
    }, [onRoutesChange, routes]);

    useEffect(() => {
        onBusModeChange?.(mode);
    }, [onBusModeChange, mode]);

    const selectedTemplate = useMemo(
        () => ROUTE_TEMPLATE_OPTIONS.find((t) => t.id === selectedTemplateId) ?? ROUTE_TEMPLATE_OPTIONS[0],
        [selectedTemplateId]
    );

    const applyTemplate = () => {
        const next = uniqueRoutes([...routes, ...selectedTemplate.subjects]);
        setRoutes(next);
        setExpertInput(next.join("\n"));
    };

    const rollback = () => {
        setRoutes(profile.suggestedRoutes);
        setExpertInput(profile.suggestedRoutes.join("\n"));
    };

    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg p-3 space-y-3">
            <div>
                <p className="text-[11px] font-semibold text-cortex-text-main">NATS Route Exposure</p>
                <p className="text-[10px] text-cortex-text-muted mt-0.5">
                    Choose how much message bus detail to expose for this workflow.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-1.5">
                {BUS_EXPOSURE_MODES.map((m) => (
                    <button
                        key={m.id}
                        onClick={() => setMode(m.id)}
                        className={`rounded border px-2 py-1.5 text-[10px] font-mono text-left ${
                            mode === m.id
                                ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-primary"
                                : "border-cortex-border text-cortex-text-muted hover:text-cortex-text-main"
                        }`}
                    >
                        <div className="font-semibold">{m.label}</div>
                        <div className="mt-0.5">{m.description}</div>
                    </button>
                ))}
            </div>

            {mode === "guided" && (
                <div className="space-y-2">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                        <select
                            value={selectedTemplateId}
                            onChange={(e) => setSelectedTemplateId(e.target.value)}
                            className="bg-cortex-surface border border-cortex-border rounded px-2 py-1.5 text-[11px] font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary"
                        >
                            {ROUTE_TEMPLATE_OPTIONS.map((t) => (
                                <option key={t.id} value={t.id}>
                                    {t.label}
                                </option>
                            ))}
                        </select>
                        <button
                            onClick={applyTemplate}
                            className="px-2 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[11px] font-mono hover:bg-cortex-primary/10 inline-flex items-center justify-center gap-1"
                        >
                            <Wand2 className="w-3 h-3" />
                            Apply Template
                        </button>
                    </div>
                    <div className="rounded-md border border-cortex-warning/30 bg-cortex-warning/10 px-2.5 py-2">
                        <p className="text-[10px] font-semibold text-cortex-warning">Impact preview</p>
                        <p className="text-[10px] text-cortex-text-main mt-0.5">{selectedTemplate.impactPreview}</p>
                    </div>
                </div>
            )}

            {mode === "expert" && (
                <div className="space-y-2">
                    <textarea
                        value={expertInput}
                        onChange={(e) => {
                            setExpertInput(e.target.value);
                            setRoutes(uniqueRoutes(e.target.value.split("\n")));
                        }}
                        className="w-full min-h-[90px] rounded border border-cortex-border bg-cortex-surface p-2 text-[11px] font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary"
                        placeholder="One subject per line"
                    />
                    <p className="text-[10px] text-cortex-text-muted">
                        Expert mode edits raw subjects directly. Use with care.
                    </p>
                </div>
            )}

            <div className="rounded-md border border-cortex-border bg-cortex-surface px-2.5 py-2">
                <div className="flex items-center justify-between gap-2">
                    <p className="text-[10px] font-semibold text-cortex-text-main">
                        Active routes ({routes.length})
                    </p>
                    <button
                        onClick={rollback}
                        className="text-[10px] font-mono px-1.5 py-1 rounded border border-cortex-border hover:bg-cortex-border text-cortex-text-muted inline-flex items-center gap-1"
                    >
                        <RotateCcw className="w-3 h-3" />
                        Rollback
                    </button>
                </div>
                <div className="mt-1.5 flex flex-wrap gap-1">
                    {routes.map((route) => (
                        <span key={route} className="px-1.5 py-0.5 rounded border border-cortex-border text-[10px] font-mono text-cortex-text-muted">
                            {route}
                        </span>
                    ))}
                </div>
            </div>
        </div>
    );
}
