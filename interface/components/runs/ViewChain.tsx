"use client";

import React, { useEffect, useMemo, useState } from "react";
import { AlertTriangle, Loader2, GitBranch, RefreshCw } from "lucide-react";
import type { MissionRun, RunChainResponse } from "@/types/events";
import RunChainNode, { type ChainNode } from "./RunChainNode";

function buildTree(runs: MissionRun[]): ChainNode[] {
    const nodeMap = new Map<string, ChainNode>();
    const childMap = new Map<string, ChainNode[]>();

    for (const run of runs) {
        const node: ChainNode = { run, children: [] };
        nodeMap.set(run.id, node);
        childMap.set(run.id, []);
    }

    for (const run of runs) {
        if (!run.parent_run_id) continue;
        const parent = nodeMap.get(run.parent_run_id);
        const child = nodeMap.get(run.id);
        if (!parent || !child) continue;
        const siblings = childMap.get(parent.run.id) ?? [];
        siblings.push(child);
        childMap.set(parent.run.id, siblings);
    }

    const roots = runs
        .filter((run) => !run.parent_run_id || !nodeMap.has(run.parent_run_id))
        .map((run) => nodeMap.get(run.id)!)
        .filter(Boolean);

    const sortNodes = (nodes: ChainNode[]) =>
        [...nodes].sort((a, b) => {
            const depthDiff = a.run.run_depth - b.run.run_depth;
            if (depthDiff !== 0) return depthDiff;
            return new Date(a.run.started_at).getTime() - new Date(b.run.started_at).getTime();
        });

    const attachChildren = (node: ChainNode): ChainNode => ({
        run: node.run,
        children: sortNodes(childMap.get(node.run.id) ?? []).map(attachChildren),
    });

    const orderedRoots = sortNodes(roots);
    return orderedRoots.map(attachChildren);
}

interface Props {
    runId: string;
}

export default function ViewChain({ runId }: Props) {
    const [chain, setChain] = useState<MissionRun[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchChain = async () => {
        try {
            const res = await fetch(`/api/v1/runs/${runId}/chain`);
            if (!res.ok) {
                setError(`Failed to load causal chain (${res.status})`);
                return;
            }
            const body = await res.json() as RunChainResponse | { data?: RunChainResponse };
            const payload = ("data" in body && body.data ? body.data : body) as RunChainResponse;
            setChain(payload.chain ?? []);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Network error");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchChain();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [runId]);

    const tree = useMemo(() => buildTree(chain), [chain]);
    const totalRuns = chain.length;

    return (
        <div className="space-y-4">
            <div className="flex items-start justify-between gap-3">
                <div className="space-y-1">
                    <div className="flex items-center gap-2">
                        <GitBranch className="h-4 w-4 text-cortex-primary" />
                        <h1 className="text-sm font-mono font-bold text-cortex-text-main uppercase tracking-widest">
                            Causal Chain
                        </h1>
                    </div>
                    <p className="text-[10px] font-mono text-cortex-text-muted/70">
                        Mission-linked run lineage for this execution path.
                    </p>
                </div>

                <button
                    type="button"
                    onClick={fetchChain}
                    className="flex items-center gap-1.5 rounded px-2 py-1 text-[10px] font-mono text-cortex-text-muted hover:text-cortex-primary transition-colors"
                >
                    <RefreshCw className="h-3 w-3" />
                    Refresh
                </button>
            </div>

            {loading && (
                <div className="flex items-center justify-center py-12 gap-2 text-cortex-text-muted">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    <span className="text-sm font-mono">Loading causal chain...</span>
                </div>
            )}

            {error && !loading && (
                <div className="rounded-md border border-cortex-danger/30 bg-cortex-danger/10 px-4 py-3 text-center">
                    <div className="flex items-center justify-center gap-2 text-cortex-danger">
                        <AlertTriangle className="h-4 w-4" />
                        <p className="text-sm font-mono">{error}</p>
                    </div>
                    <button
                        type="button"
                        onClick={fetchChain}
                        className="mt-2 text-[10px] font-mono text-cortex-primary hover:underline"
                    >
                        Retry
                    </button>
                </div>
            )}

            {!loading && !error && totalRuns === 0 && (
                <div className="flex flex-col items-center justify-center py-12 gap-2 text-cortex-text-muted">
                    <GitBranch className="h-8 w-8 opacity-20" />
                    <p className="text-sm font-mono">No chain data yet</p>
                    <p className="text-[10px] font-mono opacity-60">This mission has no run lineage to display.</p>
                </div>
            )}

            {!loading && !error && tree.length > 0 && (
                <div className="space-y-3">
                    <div className="flex items-center gap-2 text-[9px] font-mono text-cortex-text-muted/70">
                        <span className="rounded-full border border-cortex-border px-2 py-0.5 uppercase tracking-widest">
                            {totalRuns} runs
                        </span>
                        <span>Most recent runs for the mission are grouped by parent link and depth.</span>
                    </div>

                    <div className="space-y-2">
                        {tree.map((node) => (
                            <RunChainNode key={node.run.id} run={node.run} children={node.children} />
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
}
