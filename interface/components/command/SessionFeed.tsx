"use client"

import { useEffect, useState } from "react"
import { Pause, ExternalLink, Activity } from "lucide-react"

interface ActiveSession {
    id: string
    agent: string
    intent: string
    status: "running" | "drafting" | "paused"
    output: string
}

export function SessionFeed() {
    const [sessions, setSessions] = useState<ActiveSession[]>([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const fetchSessions = async () => {
            try {
                const res = await fetch('/agents')
                if (!res.ok) throw new Error('Failed to fetch')
                const data = await res.json()
                const agents = data.agents || []
                const mapped: ActiveSession[] = agents.map((a: Record<string, string>) => ({
                    id: a.id || a.agent_id || '',
                    agent: `@${a.role || a.id || 'unknown'}`,
                    intent: a.status === '1' ? 'idle' : a.status === '2' ? 'active' : 'offline',
                    status: a.status === '2' ? 'running' as const : a.status === '1' ? 'drafting' as const : 'paused' as const,
                    output: a.last_heartbeat ? `Last seen: ${new Date(a.last_heartbeat).toLocaleTimeString()}` : 'No activity',
                }))
                setSessions(mapped)
            } catch {
                setSessions([])
            } finally {
                setLoading(false)
            }
        }
        fetchSessions()
        const timer = setInterval(fetchSessions, 10000)
        return () => clearInterval(timer)
    }, [])

    return (
        <div className="flex flex-col flex-1 max-w-5xl mx-auto w-full px-8 py-12">

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 pb-4 border-b border-cortex-border text-xs font-semibold text-cortex-text-muted uppercase tracking-wider">
                <div className="col-span-2">Agent</div>
                <div className="col-span-1">Intent</div>
                <div className="col-span-2">Status</div>
                <div className="col-span-6">Output Stream</div>
                <div className="col-span-1 text-right">Action</div>
            </div>

            {/* List */}
            <div className="space-y-1 mt-2">
                {loading ? (
                    <div className="py-8 text-center text-cortex-text-muted text-xs font-mono">Loading agents...</div>
                ) : sessions.length === 0 ? (
                    <div className="py-12 flex flex-col items-center text-cortex-text-muted">
                        <Activity className="w-8 h-8 mb-2 opacity-20" />
                        <p className="text-xs font-mono">No active sessions</p>
                    </div>
                ) : (
                    sessions.map((session) => (
                        <div key={session.id} className="grid grid-cols-12 gap-4 py-4 items-center border-b border-cortex-border hover:bg-cortex-bg group transition-colors rounded-sm px-2 -mx-2">
                            <div className="col-span-2 font-mono text-sm font-semibold text-cortex-text-main">
                                {session.agent}
                            </div>
                            <div className="col-span-1 text-xs text-cortex-text-muted font-mono">
                                {session.intent}
                            </div>
                            <div className="col-span-2 flex items-center gap-2">
                                <StatusIndicator status={session.status} />
                            </div>
                            <div className="col-span-6 font-mono text-xs text-cortex-text-muted truncate opacity-80 group-hover:opacity-100">
                                {session.output}
                            </div>
                            <div className="col-span-1 text-right">
                                <button className="p-1 hover:bg-cortex-border rounded text-cortex-text-muted hover:text-cortex-text-main transition-colors">
                                    <ExternalLink size={14} />
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    )
}

function StatusIndicator({ status }: { status: ActiveSession['status'] }) {
    if (status === 'running') {
        return (
            <span className="flex items-center gap-1.5 text-xs font-medium text-cortex-success bg-cortex-success/10 px-2 py-0.5 rounded-full border border-cortex-success/20">
                <span className="w-1.5 h-1.5 rounded-full bg-cortex-success animate-pulse" />
                Running
            </span>
        )
    }
    if (status === 'drafting') {
        return (
            <span className="flex items-center gap-1.5 text-xs font-medium text-cortex-info bg-cortex-info/10 px-2 py-0.5 rounded-full border border-cortex-info/20">
                <span className="w-1.5 h-1.5 rounded-full bg-cortex-info" />
                Idle
            </span>
        )
    }
    if (status === 'paused') {
        return (
            <span className="flex items-center gap-1.5 text-xs font-medium text-cortex-warning bg-cortex-warning/10 px-2 py-0.5 rounded-full border border-cortex-warning/20">
                <Pause size={8} fill="currentColor" />
                Offline
            </span>
        )
    }
    return null
}
