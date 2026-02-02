"use client"

import { useState, useEffect } from "react"
import { ShieldAlert, CheckCircle, XCircle, Clock } from "lucide-react"

// Types matching the Go backend
interface ApprovalRequest {
    request_id: string
    original_message: {
        id: string
        source_agent_id: string
        team_id: string
        timestamp: string
        event?: {
            event_type: string
            data: any
        }
        tool_call?: {
            tool_name: string // If tool call
        }
    }
    reason: string
    expires_at: string
}

export default function InboxPage() {
    const [requests, setRequests] = useState<ApprovalRequest[]>([])
    const [loading, setLoading] = useState(true)

    const fetchRequests = async () => {
        try {
            const res = await fetch("/admin/approvals") // Proxied via Next.js or direct? Assuming proxy setting exists or same host dev
            if (res.ok) {
                const data = await res.json()
                setRequests(data)
            }
        } catch (err) {
            console.error("Failed to fetch approvals", err)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchRequests()
        const interval = setInterval(fetchRequests, 5000)
        return () => clearInterval(interval)
    }, [])

    const handleResolve = async (id: string, action: "APPROVE" | "DENY") => {
        try {
            const res = await fetch(`/admin/approvals/${id}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ action }),
            })
            if (res.ok) {
                // Optimistic update
                setRequests((prev) => prev.filter((r) => r.request_id !== id))
            }
        } catch (err) {
            console.error("Failed to resolve", err)
        }
    }

    return (
        <div className="p-8 space-y-6">
            <header className="flex items-center justify-between">
                <h1 className="text-2xl font-bold flex items-center gap-2">
                    <ShieldAlert className="w-6 h-6 text-amber-500" />
                    Governance Inbox
                </h1>
                <span className="text-sm text-muted-foreground">
                    {requests.length} Pending Decisions
                </span>
            </header>

            {loading ? (
                <div className="text-center py-12 text-gray-500">Loading Matrix...</div>
            ) : requests.length === 0 ? (
                <div className="text-center py-20 bg-slate-900/50 rounded-lg border border-slate-800">
                    <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-4" />
                    <h2 className="text-xl font-semibold">All Systems Nominal</h2>
                    <p className="text-slate-400">No pending governance requests.</p>
                </div>
            ) : (
                <div className="grid gap-4">
                    {requests.map((req) => (
                        <div
                            key={req.request_id}
                            className="bg-slate-900 border border-slate-700 p-4 rounded-lg flex items-start justify-between shadow-lg"
                        >
                            <div className="space-y-2">
                                <div className="flex items-center gap-2">
                                    <span className="bg-amber-500/20 text-amber-500 px-2 py-1 rounded text-xs font-mono border border-amber-500/30">
                                        {req.reason}
                                    </span>
                                    <span className="text-sm text-slate-400 font-mono">
                                        {req.request_id}
                                    </span>
                                </div>

                                <h3 className="text-lg font-bold text-slate-100">
                                    {req.original_message.source_agent_id || "Unknown Agent"}
                                </h3>

                                <div className="text-sm text-slate-300">
                                    <span className="font-semibold text-slate-400">Intent:</span>{" "}
                                    {req.original_message.event?.event_type ||
                                        req.original_message.tool_call?.tool_name ||
                                        "Unknown Activity"}
                                </div>

                                {/* Payload Preview */}
                                <div className="text-xs font-mono bg-black/50 p-2 rounded overflow-hidden max-w-xl text-slate-400">
                                    {JSON.stringify(req.original_message.event?.data || {}, null, 2).slice(0, 200)}
                                </div>
                            </div>

                            <div className="flex flex-col gap-2">
                                <button
                                    onClick={() => handleResolve(req.request_id, "APPROVE")}
                                    className="bg-green-600 hover:bg-green-500 text-white px-4 py-2 rounded flex items-center gap-2 text-sm font-medium transition-colors"
                                >
                                    <CheckCircle className="w-4 h-4" />
                                    Approve
                                </button>
                                <button
                                    onClick={() => handleResolve(req.request_id, "DENY")}
                                    className="bg-red-900/50 hover:bg-red-900 border border-red-800 text-red-200 px-4 py-2 rounded flex items-center gap-2 text-sm font-medium transition-colors"
                                >
                                    <XCircle className="w-4 h-4" />
                                    Deny
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}
