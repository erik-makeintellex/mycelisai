"use client"

import { Play, Pause, ExternalLink } from "lucide-react"

interface ActiveSession {
    id: string
    agent: string
    intent: string
    status: "running" | "drafting" | "paused"
    output: string
}

export function SessionFeed() {
    // Mock Data for "Active State"
    const sessions: ActiveSession[] = [
        { id: "0xf1", agent: "@architect", intent: "provision", status: "drafting", output: "Mapping 'Market Watcher' to MCP Tools..." },
        { id: "0xa4", agent: "@sentry", intent: "scan", status: "running", output: "Parsed 450 logs. No anomalies found." },
        { id: "0xb2", agent: "@coder", intent: "refactor", status: "paused", output: "Wait: User Approval required for 'fs.write'." },
    ]

    return (
        <div className="flex flex-col flex-1 max-w-5xl mx-auto w-full px-8 py-12">

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 pb-4 border-b border-zinc-200 text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                <div className="col-span-2">Agent</div>
                <div className="col-span-1">Intent</div>
                <div className="col-span-2">Status</div>
                <div className="col-span-6">Output Stream</div>
                <div className="col-span-1 text-right">Action</div>
            </div>

            {/* List */}
            <div className="space-y-1 mt-2">
                {sessions.map((session) => (
                    <div key={session.id} className="grid grid-cols-12 gap-4 py-4 items-center border-b border-zinc-100 hover:bg-zinc-50 group transition-colors rounded-sm px-2 -mx-2">
                        {/* Agent */}
                        <div className="col-span-2 font-mono text-sm font-semibold text-slate-900">
                            {session.agent}
                        </div>

                        {/* Intent */}
                        <div className="col-span-1 text-xs text-zinc-500 font-mono">
                            {session.intent}
                        </div>

                        {/* Status */}
                        <div className="col-span-2 flex items-center gap-2">
                            <StatusIndicator status={session.status} />
                        </div>

                        {/* Output */}
                        <div className="col-span-6 font-mono text-xs text-slate-600 truncate opacity-80 group-hover:opacity-100">
                            {session.output}
                        </div>

                        {/* Action */}
                        <div className="col-span-1 text-right">
                            <button className="p-1 hover:bg-zinc-200 rounded text-zinc-400 hover:text-zinc-600 transition-colors">
                                <ExternalLink size={14} />
                            </button>
                        </div>
                    </div>
                ))}
            </div>

            {/* Empty Space Filler (Optional visual balance) */}
            <div className="flex-1 min-h-[200px] flex items-end justify-center pb-12 opacity-30">
                {/* Can put a watermark or subtle graphic here if needed */}
            </div>
        </div>
    )
}

function StatusIndicator({ status }: { status: ActiveSession['status'] }) {
    if (status === 'running') {
        return (
            <span className="flex items-center gap-1.5 text-xs font-medium text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-100">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                Running
            </span>
        )
    }
    if (status === 'drafting') {
        return (
            <span className="flex items-center gap-1.5 text-xs font-medium text-sky-600 bg-sky-50 px-2 py-0.5 rounded-full border border-sky-100">
                <span className="w-1.5 h-1.5 rounded-full bg-sky-500" />
                Drafting
            </span>
        )
    }
    if (status === 'paused') {
        return (
            <span className="flex items-center gap-1.5 text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-100">
                <Pause size={8} fill="currentColor" />
                Paused
            </span>
        )
    }
    return null
}
