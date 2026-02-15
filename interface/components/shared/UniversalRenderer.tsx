"use client";

import React from "react";
import { AlertCircle, Terminal, Cpu, Database } from "lucide-react";

/**
 * UniversalRenderer
 * 
 * The "Visual Cortex" of the frontend. 
 * Takes an arbitrary payload (node, message, signal) and renders the appropriate semantic view.
 */

interface RenderProps {
    type: "text" | "agent" | "log" | "error" | "genesis";
    payload: any;
}

export function UniversalRenderer({ type, payload }: RenderProps) {
    switch (type) {
        case "text":
            return <TextRenderer content={payload} />;
        case "agent":
            return <AgentCard agent={payload} />;
        case "log":
            return <LogEntry entry={payload} />;
        case "error":
            return <ErrorDisplay error={payload} />;
        case "genesis":
            // Fallback for raw genesis data if not handled by GenesisTerminal
            return <pre className="p-4 bg-zinc-900 text-emerald-400 font-mono text-sm overflow-auto">{JSON.stringify(payload, null, 2)}</pre>;
        default:
            return (
                <div className="p-4 border border-dashed border-zinc-300 rounded bg-zinc-50 text-zinc-500 text-sm">
                    Unknown Render Type: {type}
                </div>
            );
    }
}

function TextRenderer({ content }: { content: string }) {
    return (
        <div className="prose prose-zinc max-w-none p-6">
            <p>{content}</p>
        </div>
    );
}

function AgentCard({ agent }: { agent: any }) {
    return (
        <div className="p-4 bg-white border border-zinc-200 rounded-lg shadow-sm flex items-start gap-4">
            <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${agent.status === 'alive' ? 'bg-emerald-100 text-emerald-600' : 'bg-zinc-100 text-zinc-400'}`}>
                <Cpu className="w-5 h-5" />
            </div>
            <div>
                <h3 className="font-semibold text-zinc-900">{agent.id}</h3>
                <p className="text-xs text-zinc-500 font-mono mt-1">{agent.type}</p>
                <div className="mt-2 flex gap-2">
                    <span className="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded bg-zinc-100 text-zinc-600 border border-zinc-200">
                        {agent.status}
                    </span>
                    {agent.last_seen && (
                        <span className="text-[10px] font-mono text-zinc-400 pt-0.5">
                            Last Seen: {new Date(agent.last_seen).toLocaleTimeString()}
                        </span>
                    )}
                </div>
            </div>
        </div>
    )
}

function LogEntry({ entry }: { entry: any }) {
    return (
        <div className="font-mono text-xs py-1 border-b border-zinc-100 last:border-0 flex gap-2">
            <span className="text-zinc-400 w-20 shrink-0">{new Date(entry.timestamp).toLocaleTimeString()}</span>
            <span className={`font-bold w-12 shrink-0 ${entry.level === 'ERROR' ? 'text-red-500' : 'text-indigo-500'}`}>{entry.level}</span>
            <span className="text-zinc-600 truncate">{entry.message}</span>
        </div>
    )
}

function ErrorDisplay({ error }: { error: any }) {
    return (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg flex gap-3 text-red-800">
            <AlertCircle className="w-5 h-5 shrink-0" />
            <div className="text-sm">
                <h4 className="font-bold">System Error</h4>
                <p className="mt-1">{error.message || JSON.stringify(error)}</p>
            </div>
        </div>
    )
}
