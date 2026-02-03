"use client"

import { Check, X, AlertCircle, FileCode } from "lucide-react"

interface Decision {
    id: string
    title: string
    agent: string
    risk: "low" | "medium" | "high" | "critical"
    description: string
    diff?: string
    timestamp: string
}

export function DecisionCard({ decision, onResolve }: { decision: Decision, onResolve: (id: string, approved: boolean) => void }) {
    return (
        <div className="bg-white border border-zinc-200 rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow">
            {/* Header */}
            <div className="px-4 py-3 border-b border-zinc-100 flex items-start justify-between bg-zinc-50/50">
                <div className="flex items-center gap-3">
                    <div className={`p-1.5 rounded-md border ${decision.risk === 'critical' ? 'bg-rose-50 border-rose-200 text-rose-600' :
                            decision.risk === 'high' ? 'bg-amber-50 border-amber-200 text-amber-600' :
                                'bg-blue-50 border-blue-200 text-blue-600'
                        }`}>
                        <AlertCircle size={16} />
                    </div>
                    <div>
                        <h3 className="text-sm font-semibold text-slate-900">{decision.title}</h3>
                        <p className="text-xs text-zinc-500">Requested by <span className="font-mono text-zinc-700">{decision.agent}</span> • {decision.timestamp}</p>
                    </div>
                </div>

                {/* Risk Badge */}
                <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${decision.risk === 'critical' ? 'bg-rose-100 text-rose-700' :
                        decision.risk === 'high' ? 'bg-amber-100 text-amber-700' :
                            'bg-blue-100 text-blue-700'
                    }`}>
                    {decision.risk} Risk
                </span>
            </div>

            {/* Content / Diff */}
            <div className="p-4">
                <p className="text-sm text-slate-600 mb-3">{decision.description}</p>

                {decision.diff && (
                    <div className="bg-zinc-900 rounded-md p-3 overflow-x-auto relative group">
                        <div className="absolute top-2 right-2 text-zinc-500 opacity-0 group-hover:opacity-100 transition-opacity">
                            <FileCode size={14} />
                        </div>
                        <pre className="text-xs font-mono text-zinc-300 whitespace-pre leading-relaxed">
                            {decision.diff}
                        </pre>
                    </div>
                )}
            </div>

            {/* Actions */}
            <div className="px-4 py-3 bg-zinc-50 border-t border-zinc-200 flex justify-end gap-3">
                <button
                    onClick={() => onResolve(decision.id, false)}
                    className="px-3 py-1.5 text-xs font-medium text-slate-600 hover:text-slate-900 hover:bg-zinc-200 rounded flex items-center gap-1.5 transition-colors"
                >
                    <X size={14} />
                    Reject
                </button>
                <button
                    onClick={() => onResolve(decision.id, true)}
                    className="px-3 py-1.5 text-xs font-medium text-white bg-slate-900 hover:bg-slate-800 rounded flex items-center gap-1.5 shadow-sm transition-colors"
                >
                    <Check size={14} />
                    Approve Request
                </button>
            </div>
        </div>
    )
}
