import type React from "react";
import { Database, Route, ShieldCheck } from "lucide-react";
import type { CapabilityManifest } from "@/store/useCortexStore";

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
    const availableCount = capabilities.filter((capability) => isCapabilityAvailable(capability)).length;
    const mutatingCount = capabilities.filter((capability) => capability.writes && capability.writes.length > 0).length;
    const availableCapabilities = capabilities.filter(isCapabilityAvailable);
    const repairCapabilities = capabilities.filter((capability) => !isCapabilityAvailable(capability));

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="flex items-start gap-3">
                    <div className="mt-0.5 rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-2">
                        <ShieldCheck className="h-4 w-4 text-cortex-primary" />
                    </div>
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Capability overview
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                            Can use, needs repair, and can request
                        </p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                            {usingFallback
                                ? "Capability API data is unavailable, so this view is derived from MCP servers and search status."
                                : "Capability manifests are loaded from /api/v1/capabilities."}
                        </p>
                    </div>
                </div>
                <div className="flex flex-wrap gap-2">
                    <SummaryChip label={`${availableCount} can use now`} />
                    <SummaryChip label={`${repairCapabilities.length} need repair`} />
                    <SummaryChip label={`${mutatingCount} can write`} />
                </div>
            </div>

            {error && (
                <div className="mt-3 rounded-lg border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2">
                    <p className="text-[10px] font-mono text-cortex-warning">{error}</p>
                </div>
            )}

            {isLoading && capabilities.length === 0 ? (
                <p className="mt-4 text-xs text-cortex-text-muted">Loading capability registry...</p>
            ) : capabilities.length === 0 ? (
                <p className="mt-4 text-xs text-cortex-text-muted">
                    No capabilities are visible yet.
                </p>
            ) : (
                <div className="mt-4 grid gap-3 lg:grid-cols-2">
                    <CapabilitySection title="Can use now" count={availableCapabilities.length}>
                        {availableCapabilities.map((capability) => (
                            <CapabilityCard key={capability.id} capability={capability} />
                        ))}
                    </CapabilitySection>
                    <CapabilitySection title="Needs repair" count={repairCapabilities.length}>
                        {repairCapabilities.map((capability) => (
                            <CapabilityCard key={capability.id} capability={capability} />
                        ))}
                    </CapabilitySection>
                    <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Can request/add
                        </p>
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                            Use Add MCP to request another approved tool server, then return here to confirm availability and repair guidance.
                        </p>
                    </div>
                </div>
            )}
        </div>
    );
}

function CapabilitySection({ children, count, title }: { children: React.ReactNode; count: number; title: string }) {
    return (
        <section className="rounded-lg border border-cortex-border bg-cortex-surface/60 p-3">
            <div className="mb-2 flex items-center justify-between gap-2">
                <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">{title}</p>
                <SummaryChip label={`${count} item${count === 1 ? "" : "s"}`} />
            </div>
            {count > 0 ? <div className="grid gap-2">{children}</div> : (
                <p className="text-xs text-cortex-text-muted">No capabilities in this state.</p>
            )}
        </section>
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
                        <span className="rounded border border-cortex-border bg-cortex-surface px-1.5 py-0.5 text-[9px] font-mono uppercase text-cortex-text-muted">
                            {capability.id}
                        </span>
                    </div>
                    {capability.description && (
                        <p className="mt-1 text-xs leading-5 text-cortex-text-muted">{capability.description}</p>
                    )}
                </div>
                <div className="flex flex-wrap gap-1.5">
                    <CapabilityBadge tone={available ? "success" : "warning"} label={capability.availability_status ?? (available ? "available" : "needs attention")} />
                    <CapabilityBadge label={capability.category} />
                    <CapabilityBadge tone={riskTone(capability.risk)} label={`risk ${capability.risk}`} />
                    <CapabilityBadge tone={approvalTone(capability.approval)} label={`approval ${capability.approval}`} />
                </div>
            </div>

            <details className="mt-2 rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2" open={!available}>
                <summary className="cursor-pointer text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted">
                    Details and binding
                </summary>
                <div className="mt-2 grid gap-2">
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

function CapabilityDetail({ icon, label, value }: { icon?: React.ReactNode; label: string; value: string }) {
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
