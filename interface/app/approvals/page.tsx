"use client"

import { useState } from "react"
import { DecisionCard } from "@/components/approvals/DecisionCard"
import { CheckCircle2 } from "lucide-react"

export default function ApprovalsPage() {
    const [decisions, setDecisions] = useState([
        {
            id: "dec-1",
            title: "File Write Access",
            agent: "@coder",
            risk: "medium" as const,
            description: "Agent requesting to overwrite 'core/internal/router.go' with refactored logic.",
            diff: "- func Route(msg Message) {\n+ func Route(msg Message, context Context) {",
            timestamp: "2 mins ago"
        },
        {
            id: "dec-2",
            title: "External Network Request",
            agent: "@researcher",
            risk: "high" as const,
            description: "Requesting access to 'https://api.openai.com/v1/chat/completions'. Domain is not whitelisted.",
            timestamp: "15 mins ago"
        }
    ])

    const handleResolve = (id: string, approved: boolean) => {
        console.log(`Decision ${id} resolved: ${approved ? 'APPROVED' : 'REJECTED'}`)
        setDecisions(prev => prev.filter(d => d.id !== id))
    }

    if (decisions.length === 0) {
        return (
            <div className="h-full flex flex-col items-center justify-center text-zinc-400">
                <CheckCircle2 size={48} className="mb-4 text-emerald-500/50" />
                <h2 className="text-lg font-semibold text-slate-700">All Clear</h2>
                <p className="text-sm">No pending governance requests.</p>
            </div>
        )
    }

    return (
        <div className="p-8 max-w-3xl mx-auto space-y-6">
            <header className="flex items-end justify-between border-b border-zinc-200 pb-4">
                <div>
                    <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Approvals</h1>
                    <p className="text-zinc-500 mt-1">Governance Queue • {decisions.length} Pending</p>
                </div>
            </header>

            <div className="space-y-4">
                {decisions.map(d => (
                    <DecisionCard key={d.id} decision={d} onResolve={handleResolve} />
                ))}
            </div>
        </div>
    )
}
