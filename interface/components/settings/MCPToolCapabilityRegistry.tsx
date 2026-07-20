"use client";

import { useMemo, useState, type ReactNode } from "react";
import { Database, Route, ShieldCheck } from "lucide-react";
import type { CapabilityManifest } from "@/store/useCortexStore";
import {
    CAPABILITY_ORIGINS,
    capabilityOrigin,
    capabilityOriginLabel,
    type CapabilityOrigin,
} from "./MCPToolCapabilityOrigin";

type OriginFilter = "all" | CapabilityOrigin;
const PAGE_SIZE = 12;

export function CapabilityRegistryPanel({
    capabilities,
    isLoading,
    error,
    usingFallback,
}: {
    capabilities: CapabilityManifest[];
    isLoading: boolean;
    error: string | null;
    usingFallback: boolean;
}) {
    const [originFilter, setOriginFilter] = useState<OriginFilter>("all");
    const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);
    const filteredCapabilities = useMemo(
        () => originFilter === "all"
            ? capabilities
            : capabilities.filter((capability) => capabilityOrigin(capability) === originFilter),
        [capabilities, originFilter],
    );
    const availableCount = capabilities.filter(isCapabilityAvailable).length;
    const repairCount = capabilities.length - availableCount;
    const mutatingCount = capabilities.filter((capability) => capability.writes && capability.writes.length > 0).length;
    const availableCapabilities = filteredCapabilities.filter(isCapabilityAvailable);
    const repairCapabilities = filteredCapabilities.filter((capability) => !isCapabilityAvailable(capability));
    const orderedCapabilities = [...repairCapabilities, ...availableCapabilities];
    const visibleCapabilities = orderedCapabilities.slice(0, visibleCount);
    const remainingCount = orderedCapabilities.length - visibleCapabilities.length;

    function selectOrigin(next: OriginFilter) {
        setOriginFilter(next);
        setVisibleCount(PAGE_SIZE);
    }

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex items-start gap-3">
                    <div className="mt-0.5 rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-2">
                        <ShieldCheck className="h-4 w-4 text-cortex-primary" />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Capability catalog
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                            What Soma can use right now
                        </p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                            {usingFallback
                                ? "Using connected tools and search status because the capability registry is unavailable."
                                : "Use this first for readiness. Configure access or inspect raw bindings only when needed."}
                        </p>
                    </div>
                </div>
                <div className="flex flex-wrap gap-2">
                    <SummaryChip label={`${availableCount} can use now`} />
                    <SummaryChip label={`${repairCount} need repair`} />
                    <SummaryChip label={`${mutatingCount} can write`} />
                </div>
            </div>

            {error && (
                <div className="mt-3 rounded-lg border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2">
                    <p className="text-[10px] font-mono text-cortex-warning">{error}</p>
                </div>
            )}

            <div className="mt-4 flex gap-2 overflow-x-auto pb-1" aria-label="Capability origin filters">
                <OriginButton
                    active={originFilter === "all"}
                    count={capabilities.length}
                    label="All"
                    onClick={() => selectOrigin("all")}
                />
                {CAPABILITY_ORIGINS.map((origin) => (
                    <OriginButton
                        key={origin.id}
                        active={originFilter === origin.id}
                        count={capabilities.filter((capability) => capabilityOrigin(capability) === origin.id).length}
                        label={origin.label}
                        onClick={() => selectOrigin(origin.id)}
                    />
                ))}
            </div>

            {isLoading && capabilities.length === 0 ? (
                <p className="mt-4 text-xs text-cortex-text-muted">Loading capability registry...</p>
            ) : capabilities.length === 0 ? (
                <p className="mt-4 text-xs text-cortex-text-muted">
                    No capabilities are visible yet.
                </p>
            ) : (
                <div className="mt-4">
                    <p className="mb-2 text-[11px] leading-5 text-cortex-text-muted">
                        {originFilter === "all"
                            ? "Filter by origin to see whether a capability is part of Mycelis, exposed by its host, or supplied through MCP."
                            : CAPABILITY_ORIGINS.find((origin) => origin.id === originFilter)?.summary}
                    </p>
                    <div className="max-h-[min(58vh,38rem)] overflow-y-auto rounded-lg border border-cortex-border bg-cortex-bg/40 p-2" data-testid="capability-catalog-list">
                        <div className="grid gap-2">
                            {visibleCapabilities.map((capability) => (
                                <CapabilityCard key={capability.id} capability={capability} />
                            ))}
                            {visibleCapabilities.length === 0 ? (
                                <p className="px-2 py-4 text-xs text-cortex-text-muted">No capabilities have this origin.</p>
                            ) : null}
                        </div>
                        {remainingCount > 0 ? (
                            <button
                                type="button"
                                onClick={() => setVisibleCount((current) => current + PAGE_SIZE)}
                                className="mt-2 w-full rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2 text-xs font-semibold text-cortex-text-muted transition hover:text-cortex-text-main"
                            >
                                Show {Math.min(PAGE_SIZE, remainingCount)} more
                            </button>
                        ) : null}
                    </div>
                </div>
            )}
        </div>
    );
}

function OriginButton({
    active,
    count,
    label,
    onClick,
}: {
    active: boolean;
    count: number;
    label: string;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`inline-flex shrink-0 items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-semibold transition ${active
                ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                : "border-cortex-border bg-cortex-bg text-cortex-text-muted hover:text-cortex-text-main"}`}
        >
            {label}
            <span className="font-mono text-[10px]">{count}</span>
        </button>
    );
}

function CapabilityCard({ capability }: { capability: CapabilityManifest }) {
    const available = isCapabilityAvailable(capability);
    const writes = capability.writes ?? [];
    const outputs = capability.outputs ?? [];
    const roles = capability.allowed_roles ?? [];
    const binding = [capability.provider, capability.bound_server_name ?? capability.bound_server_id ?? capability.server_or_package, capability.bound_tool_name ?? capability.bound_tool_id]
        .filter(Boolean)
        .join(" / ") || capability.source;

    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
            <div className="flex flex-col gap-2">
                <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                        <span className={`h-2 w-2 rounded-full ${available ? "bg-cortex-success" : "bg-cortex-warning"}`} />
                        <p className="text-sm font-semibold text-cortex-text-main">{capability.name}</p>
                        <CapabilityBadge label={capabilityOriginLabel(capability)} />
                    </div>
                    {capability.description && (
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">{capability.description}</p>
                    )}
                </div>
                {!available && capability.fallback_behavior && (
                    <p className="rounded-lg border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2 text-xs leading-5 text-cortex-text-main">
                        {capability.fallback_behavior}
                    </p>
                )}
                <div className="flex flex-wrap gap-1.5">
                    <CapabilityBadge tone={available ? "success" : "warning"} label={capability.availability_status ?? (available ? "available" : "needs attention")} />
                    <CapabilityBadge label={capability.category} />
                    <CapabilityBadge tone={riskTone(capability.risk)} label={`risk ${capability.risk}`} />
                    <CapabilityBadge tone={approvalTone(capability.approval)} label={`approval ${capability.approval}`} />
                </div>
            </div>

            <details className="mt-2 rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2">
                <summary className="cursor-pointer text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                    Inspect capability details
                </summary>
                <div className="mt-2 grid gap-2">
                    <CapabilityDetail label="Capability ref" value={capability.id} />
                    <CapabilityDetail icon={<Route className="h-3.5 w-3.5" />} label="Outputs" value={outputs.length ? outputs.join(", ") : "ToolResult"} />
                    <CapabilityDetail icon={<Database className="h-3.5 w-3.5" />} label="Writes" value={writes.length ? writes.join(", ") : "Managed Exchange, run evidence"} />
                    <CapabilityDetail icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Audit" value={capability.audit ?? "required"} />
                    <CapabilityDetail label="Soma use" value={roles.length === 0 || roles.includes("soma") ? "allowed" : roles.join(", ")} />
                    <CapabilityDetail label="Fallback" value={capability.fallback_behavior ?? "Report a capability blocker and keep the run recoverable."} />
                    <CapabilityDetail label="Binding" value={binding} />
                </div>
            </details>
        </div>
    );
}

function SummaryChip({ label }: { label: string }) {
    return (
        <span className="rounded-full border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono uppercase text-cortex-text-muted">
            {label}
        </span>
    );
}

function CapabilityBadge({ label, tone = "neutral" }: { label: string; tone?: "neutral" | "success" | "warning" | "danger" }) {
    const className = tone === "success"
        ? "border-cortex-success/25 bg-cortex-success/10 text-cortex-success"
        : tone === "warning"
        ? "border-cortex-warning/25 bg-cortex-warning/10 text-cortex-warning"
        : tone === "danger"
        ? "border-cortex-danger/25 bg-cortex-danger/10 text-cortex-danger"
        : "border-cortex-border bg-cortex-surface text-cortex-text-muted";
    return (
        <span className={`rounded border px-1.5 py-0.5 text-[9px] font-mono uppercase ${className}`}>
            {label}
        </span>
    );
}

function CapabilityDetail({ icon, label, value }: { icon?: ReactNode; label: string; value: string }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2">
            <p className="flex items-center gap-1 text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                {icon}
                {label}
            </p>
            <p className="mt-1 break-words text-[11px] leading-4 text-cortex-text-main">{value}</p>
        </div>
    );
}

function isCapabilityAvailable(capability: CapabilityManifest): boolean {
    const status = capability.availability_status?.toLowerCase();
    if (!status) return true;
    return status === "available" || status === "connected" || status === "ready" || status === "online";
}

function riskTone(risk: string): "neutral" | "success" | "warning" | "danger" {
    if (risk === "high") return "danger";
    if (risk === "medium") return "warning";
    if (risk === "low") return "success";
    return "neutral";
}

function approvalTone(approval: string): "neutral" | "success" | "warning" | "danger" {
    if (approval === "required") return "warning";
    if (approval === "none") return "success";
    return "neutral";
}
