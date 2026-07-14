"use client";

import { type ReactNode, useEffect, useState } from "react";
import { GitBranch, MessageSquareMore, PackageSearch, RefreshCw } from "lucide-react";

type ExchangeChannel = {
    id: string;
    name: string;
    type: string;
    schema_id: string;
    visibility: string;
    sensitivity_class?: string;
    owner: string;
};

type ExchangeThread = {
    id: string;
    channel_name?: string;
    thread_type: string;
    title: string;
    status: string;
    participants: string[];
    allowed_reviewers?: string[];
};

type ExchangeItem = {
    id: string;
    channel_name?: string;
    schema_id: string;
    summary: string;
    created_by: string;
    created_at: string;
    sensitivity_class?: string;
    trust_class?: string;
    capability_id?: string;
    review_required?: boolean;
};

export default function ExchangeInspector() {
    const [channels, setChannels] = useState<ExchangeChannel[]>([]);
    const [threads, setThreads] = useState<ExchangeThread[]>([]);
    const [items, setItems] = useState<ExchangeItem[]>([]);
    const [activeView, setActiveView] = useState<"handoffs" | "threads" | "lanes">("handoffs");
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const load = async () => {
        setLoading(true);
        try {
            const [channelRes, threadRes, itemRes] = await Promise.all([
                fetch("/api/v1/exchange/channels"),
                fetch("/api/v1/exchange/threads?limit=12"),
                fetch("/api/v1/exchange/items?limit=12"),
            ]);
            if (!channelRes.ok || !threadRes.ok || !itemRes.ok) {
                throw new Error("Managed exchange is unavailable.");
            }
            const [channelData, threadData, itemData] = await Promise.all([
                channelRes.json(),
                threadRes.json(),
                itemRes.json(),
            ]);
            setChannels(Array.isArray(channelData) ? channelData : []);
            setThreads(Array.isArray(threadData) ? threadData : []);
            setItems(Array.isArray(itemData) ? itemData : []);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Managed exchange failed to load.");
            setChannels([]);
            setThreads([]);
            setItems([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-6 flex-shrink-0">
                <div>
                    <h2 className="text-xs font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                        Team handoffs
                    </h2>
                    <p className="text-[11px] text-cortex-text-muted mt-1">
                        Review evidence moving between Soma, teams, tools, and retained outputs.
                    </p>
                </div>
                <button
                    onClick={load}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-cortex-primary/10 border border-cortex-primary/30 text-xs font-mono font-bold text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                >
                    <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
                    REFRESH
                </button>
            </div>

            <div className="flex-1 overflow-y-auto">
                {error ? (
                    <div className="m-6 rounded-xl border border-cortex-danger/30 bg-cortex-danger/5 p-4">
                        <p className="text-sm text-cortex-danger">Managed exchange unavailable</p>
                        <p className="text-xs text-cortex-text-muted mt-1">{error}</p>
                    </div>
                ) : (
                    <div className="mx-auto flex max-w-5xl flex-col gap-4 p-6">
                        <div className="grid gap-3 md:grid-cols-3">
                            <SummaryButton active={activeView === "handoffs"} count={items.length} icon={<PackageSearch className="h-4 w-4" />} label="Recent handoffs" onClick={() => setActiveView("handoffs")} />
                            <SummaryButton active={activeView === "threads"} count={threads.length} icon={<MessageSquareMore className="h-4 w-4" />} label="Work threads" onClick={() => setActiveView("threads")} />
                            <SummaryButton active={activeView === "lanes"} count={channels.length} icon={<GitBranch className="h-4 w-4" />} label="Source lanes" onClick={() => setActiveView("lanes")} />
                        </div>
                        {activeView === "handoffs" && (
                            <Section title="Recent handoffs" subtitle="Normalized outputs and evidence Soma or another team can use next." empty="No handoffs have been published yet." loading={loading}>
                                {items.map((item) => <ExchangeItemCard key={item.id} item={item} />)}
                            </Section>
                        )}
                        {activeView === "threads" && (
                            <Section title="Work threads" subtitle="Planning, review, escalation, and learning conversations grouped for audit." empty="No threads created yet." loading={loading}>
                                {threads.map((thread) => <ThreadCard key={thread.id} thread={thread} />)}
                            </Section>
                        )}
                        {activeView === "lanes" && (
                            <Section title="Source lanes" subtitle="Advanced source channels that explain where handoffs are allowed to move." empty="No source lanes registered yet." loading={loading}>
                                {channels.map((channel) => <ChannelCard key={channel.id} channel={channel} />)}
                            </Section>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}

function Section({
    title,
    subtitle,
    empty,
    loading,
    children,
}: {
    title: string;
    subtitle: string;
    empty: string;
    loading: boolean;
    children: ReactNode;
}) {
    const count = Array.isArray(children) ? children.length : 0;

    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg/60">
            <div className="border-b border-cortex-border px-4 py-3">
                <h3 className="text-sm font-semibold text-cortex-text-main">{title}</h3>
                <p className="text-[11px] text-cortex-text-muted mt-1">{subtitle}</p>
            </div>
            <div className="p-4 space-y-3">
                {loading ? (
                    <p className="text-xs font-mono text-cortex-text-muted animate-pulse">Loading…</p>
                ) : count === 0 ? (
                    <p className="text-xs text-cortex-text-muted">{empty}</p>
                ) : children}
            </div>
        </div>
    );
}

function SummaryButton({ active, count, icon, label, onClick }: { active: boolean; count: number; icon: ReactNode; label: string; onClick: () => void }) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`rounded-xl border px-4 py-3 text-left transition ${active ? "border-cortex-primary/50 bg-cortex-primary/10" : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/30"}`}
        >
            <span className="flex items-center justify-between gap-3">
                <span className="flex items-center gap-2 text-sm font-semibold text-cortex-text-main">{icon}{label}</span>
                <span className="rounded-full border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono text-cortex-text-muted">{count}</span>
            </span>
        </button>
    );
}

function ExchangeItemCard({ item }: { item: ExchangeItem }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-3 space-y-1">
            <div className="flex items-start justify-between gap-2">
                <span className="text-sm font-semibold text-cortex-text-main">{item.summary || "Untitled handoff"}</span>
                {item.review_required && <span className="rounded bg-cortex-warning/15 px-1.5 py-0.5 text-[10px] font-mono uppercase text-cortex-warning">Review</span>}
            </div>
            <p className="text-xs text-cortex-text-muted">{item.channel_name ?? "Unassigned lane"} · {item.created_by}</p>
            <p className="text-[11px] text-cortex-text-muted">
                {item.schema_id} · {item.trust_class ?? "trusted_internal"} · {item.sensitivity_class ?? "role_scoped"}
            </p>
            <p className="text-[10px] text-cortex-text-muted">{formatTimestamp(item.created_at)}</p>
        </div>
    );
}

function ThreadCard({ thread }: { thread: ExchangeThread }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-3 space-y-1">
            <div className="flex items-start justify-between gap-2">
                <span className="text-sm font-semibold text-cortex-text-main">{thread.title}</span>
                <span className="rounded bg-cortex-info/15 px-1.5 py-0.5 text-[10px] font-mono uppercase text-cortex-info">{thread.status}</span>
            </div>
            <p className="text-xs text-cortex-text-muted">{thread.channel_name ?? "Unassigned lane"} · {thread.thread_type}</p>
            {thread.participants?.length > 0 && <p className="text-[11px] text-cortex-text-muted">{thread.participants.join(", ")}</p>}
            {thread.allowed_reviewers?.length ? <p className="text-[10px] text-cortex-text-muted">Reviewers {thread.allowed_reviewers.join(", ")}</p> : null}
        </div>
    );
}

function ChannelCard({ channel }: { channel: ExchangeChannel }) {
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-3 space-y-1">
            <div className="flex items-start justify-between gap-2">
                <span className="text-sm font-semibold text-cortex-text-main break-all">{channel.name}</span>
                <span className="rounded bg-cortex-primary/15 px-1.5 py-0.5 text-[10px] font-mono uppercase text-cortex-primary">{channel.type}</span>
            </div>
            <p className="text-xs text-cortex-text-muted">{channel.schema_id} · owner {channel.owner} · {channel.visibility}</p>
            {channel.sensitivity_class ? <p className="text-[10px] text-cortex-text-muted">Sensitivity {channel.sensitivity_class}</p> : null}
        </div>
    );
}

function formatTimestamp(value: string): string {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString();
}
