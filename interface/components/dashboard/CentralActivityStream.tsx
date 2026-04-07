"use client";

import { useEffect, useMemo, useState } from "react";
import { ChevronDown, Filter, Radio } from "lucide-react";
import { useCortexStore, type StreamSignal } from "@/store/useCortexStore";
import { streamSignalToDetail } from "@/lib/signalNormalize";

type ActivityAspect =
    | "status"
    | "results"
    | "artifacts"
    | "tools"
    | "governance"
    | "errors"
    | "reasoning";

type FilterOption = {
    id: string;
    label: string;
    count: number;
};

const ASPECT_LABELS: Record<ActivityAspect, string> = {
    status: "Status",
    results: "Results",
    artifacts: "Artifacts",
    tools: "Tools / MCP",
    governance: "Governance",
    errors: "Errors",
    reasoning: "Reasoning",
};

function classifyAspect(signal: StreamSignal): ActivityAspect[] {
    const type = signal.type?.toLowerCase() ?? "";
    const message = signal.message?.toLowerCase() ?? "";
    const sourceKind = signal.source_kind?.toLowerCase() ?? "";
    const payloadKind = signal.payload_kind?.toLowerCase() ?? "";
    const channel = signal.source_channel?.toLowerCase() ?? signal.topic?.toLowerCase() ?? "";
    const aspects = new Set<ActivityAspect>();

    if (type.includes("heartbeat") || type.includes("connected") || channel.includes(".signal.status")) aspects.add("status");
    if (type.includes("artifact")) aspects.add("artifacts");
    if (type.includes("output") || type.includes("task_complete") || channel.includes(".signal.result")) aspects.add("results");
    if (type.includes("tool") || type.includes("actuation") || sourceKind === "mcp" || message.includes("mcp")) aspects.add("tools");
    if (type.includes("governance") || message.includes("approval") || message.includes("governed")) aspects.add("governance");
    if (type.includes("thought") || type.includes("cognitive") || payloadKind.includes("thought")) aspects.add("reasoning");
    if (type.includes("error") || signal.level?.toLowerCase() === "error") aspects.add("errors");

    if (aspects.size === 0) aspects.add("results");
    return Array.from(aspects);
}

function relativeTime(timestamp?: string): string {
    if (!timestamp) return "now";
    const diff = Date.now() - new Date(timestamp).getTime();
    if (!Number.isFinite(diff) || diff < 0) return "now";
    if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
    return `${Math.floor(diff / 86_400_000)}d ago`;
}

function toneClass(signal: StreamSignal): string {
    const aspects = classifyAspect(signal);
    if (aspects.includes("errors")) return "border-cortex-danger/30 bg-cortex-danger/5";
    if (aspects.includes("governance")) return "border-cortex-warning/30 bg-cortex-warning/5";
    if (aspects.includes("artifacts")) return "border-cortex-success/30 bg-cortex-success/5";
    return "border-cortex-border bg-cortex-surface";
}

export default function CentralActivityStream() {
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const selectSignalDetail = useCortexStore((s) => s.selectSignalDetail);
    const [selectedTeamIds, setSelectedTeamIds] = useState<string[]>([]);
    const [selectedAspects, setSelectedAspects] = useState<ActivityAspect[]>([]);

    useEffect(() => {
        if (teamsDetail.length === 0) {
            void fetchTeamsDetail();
        }
    }, [fetchTeamsDetail, teamsDetail.length]);

    const teamOptions = useMemo<FilterOption[]>(() => {
        const counts = new Map<string, number>();
        streamLogs.forEach((signal) => {
            const teamId = signal.team_id?.trim();
            if (!teamId) return;
            counts.set(teamId, (counts.get(teamId) ?? 0) + 1);
        });
        return Array.from(counts.entries())
            .map(([teamId, count]) => ({
                id: teamId,
                label: teamsDetail.find((team) => team.id === teamId)?.name ?? teamId,
                count,
            }))
            .sort((left, right) => left.label.localeCompare(right.label));
    }, [streamLogs, teamsDetail]);

    const aspectOptions = useMemo<FilterOption[]>(() => {
        const counts = new Map<ActivityAspect, number>();
        streamLogs.forEach((signal) => {
            classifyAspect(signal).forEach((aspect) => counts.set(aspect, (counts.get(aspect) ?? 0) + 1));
        });
        return (Object.keys(ASPECT_LABELS) as ActivityAspect[]).map((aspect) => ({
            id: aspect,
            label: ASPECT_LABELS[aspect],
            count: counts.get(aspect) ?? 0,
        }));
    }, [streamLogs]);

    const filteredLogs = useMemo(() => {
        return streamLogs.filter((signal) => {
            const teamMatch = selectedTeamIds.length === 0 || (signal.team_id ? selectedTeamIds.includes(signal.team_id) : false);
            const aspectKinds = classifyAspect(signal);
            const aspectMatch = selectedAspects.length === 0 || selectedAspects.some((aspect) => aspectKinds.includes(aspect));
            return teamMatch && aspectMatch;
        });
    }, [selectedAspects, selectedTeamIds, streamLogs]);

    return (
        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5" data-testid="central-activity-stream">
            <div className="flex items-start justify-between gap-3">
                <div>
                    <p className="text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">Live team interaction stream</p>
                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                        Watch the central stream of team activity from Soma home, then narrow it by team and output aspect without leaving the main admin surface.
                    </p>
                </div>
                <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                    {filteredLogs.length} visible
                </span>
            </div>

            <div className="mt-4 grid gap-3 md:grid-cols-2">
                <ChecklistDropdown
                    title="Teams"
                    summary={selectedTeamIds.length === 0 ? "All teams" : `${selectedTeamIds.length} selected`}
                    options={teamOptions}
                    selected={selectedTeamIds}
                    onToggle={(teamId) =>
                        setSelectedTeamIds((current) =>
                            current.includes(teamId) ? current.filter((item) => item !== teamId) : [...current, teamId],
                        )
                    }
                    onClear={() => setSelectedTeamIds([])}
                    emptyCopy="No team-tagged signals yet."
                    testId="activity-team-filter"
                />
                <ChecklistDropdown
                    title="Output aspects"
                    summary={selectedAspects.length === 0 ? "All aspects" : `${selectedAspects.length} selected`}
                    options={aspectOptions}
                    selected={selectedAspects}
                    onToggle={(aspectId) =>
                        setSelectedAspects((current) =>
                            current.includes(aspectId as ActivityAspect)
                                ? current.filter((item) => item !== aspectId)
                                : [...current, aspectId as ActivityAspect],
                        )
                    }
                    onClear={() => setSelectedAspects([])}
                    emptyCopy="No aspect categories available yet."
                    testId="activity-aspect-filter"
                />
            </div>

            <div className="mt-4 max-h-[420px] overflow-y-auto space-y-3 pr-1">
                {filteredLogs.length === 0 ? (
                    <div className="rounded-2xl border border-dashed border-cortex-border bg-cortex-bg px-4 py-6 text-center">
                        <Radio className="mx-auto h-8 w-8 text-cortex-text-muted/40" />
                        <p className="mt-3 text-sm text-cortex-text-muted">
                            No matching live stream entries yet. Leave filters broad or wait for the next team interaction.
                        </p>
                    </div>
                ) : (
                    filteredLogs.slice(0, 30).map((signal, index) => {
                        const aspects = classifyAspect(signal);
                        const teamLabel = signal.team_id
                            ? teamsDetail.find((team) => team.id === signal.team_id)?.name ?? signal.team_id
                            : "Soma / system";
                        return (
                            <button
                                key={`${signal.timestamp ?? "ts"}-${signal.source ?? "src"}-${index}`}
                                type="button"
                                onClick={() => selectSignalDetail(streamSignalToDetail(signal))}
                                className={`w-full rounded-2xl border px-4 py-3 text-left transition-colors hover:border-cortex-primary/25 ${toneClass(signal)}`}
                            >
                                <div className="flex items-start justify-between gap-3">
                                    <div className="min-w-0">
                                        <div className="flex flex-wrap items-center gap-2">
                                            <span className="text-sm font-semibold text-cortex-text-main">{teamLabel}</span>
                                            <span className="text-[11px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
                                                {signal.source ?? signal.source_kind ?? "system"}
                                            </span>
                                        </div>
                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                            {signal.message?.trim() || "No message body"}
                                        </p>
                                    </div>
                                    <span className="flex-shrink-0 text-[11px] font-mono text-cortex-text-muted">
                                        {relativeTime(signal.timestamp)}
                                    </span>
                                </div>
                                <div className="mt-3 flex flex-wrap gap-2">
                                    {aspects.map((aspect) => (
                                        <span
                                            key={aspect}
                                            className="rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-1 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted"
                                        >
                                            {ASPECT_LABELS[aspect]}
                                        </span>
                                    ))}
                                    {signal.run_id ? (
                                        <span className="rounded-full border border-cortex-border bg-cortex-bg px-2.5 py-1 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
                                            Run {signal.run_id}
                                        </span>
                                    ) : null}
                                </div>
                            </button>
                        );
                    })
                )}
            </div>
        </div>
    );
}

function ChecklistDropdown({
    title,
    summary,
    options,
    selected,
    onToggle,
    onClear,
    emptyCopy,
    testId,
}: {
    title: string;
    summary: string;
    options: FilterOption[];
    selected: string[];
    onToggle: (id: string) => void;
    onClear: () => void;
    emptyCopy: string;
    testId: string;
}) {
    return (
        <details className="rounded-2xl border border-cortex-border bg-cortex-bg" data-testid={testId}>
            <summary className="flex cursor-pointer list-none items-center justify-between gap-3 px-4 py-3">
                <div className="min-w-0">
                    <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">{title}</p>
                    <p className="mt-1 text-sm text-cortex-text-main">{summary}</p>
                </div>
                <div className="flex items-center gap-2 text-cortex-text-muted">
                    <Filter className="h-4 w-4" />
                    <ChevronDown className="h-4 w-4" />
                </div>
            </summary>
            <div className="border-t border-cortex-border px-4 py-3">
                <div className="mb-3 flex items-center justify-between gap-3">
                    <p className="text-xs text-cortex-text-muted">Select multiple filters to narrow the stream.</p>
                    <button type="button" onClick={onClear} className="text-[11px] font-semibold text-cortex-primary hover:underline">
                        Clear
                    </button>
                </div>
                {options.length === 0 ? (
                    <p className="text-sm text-cortex-text-muted">{emptyCopy}</p>
                ) : (
                    <div className="space-y-2">
                        {options.map((option) => (
                            <label key={option.id} className="flex items-center justify-between gap-3 rounded-xl border border-cortex-border bg-cortex-surface px-3 py-2">
                                <span className="flex items-center gap-2">
                                    <input
                                        type="checkbox"
                                        checked={selected.includes(option.id)}
                                        onChange={() => onToggle(option.id)}
                                    />
                                    <span className="text-sm text-cortex-text-main">{option.label}</span>
                                </span>
                                <span className="text-[11px] font-mono text-cortex-text-muted">{option.count}</span>
                            </label>
                        ))}
                    </div>
                )}
            </div>
        </details>
    );
}
