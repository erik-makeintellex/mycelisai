"use client"

import { useState } from "react"
import { Terminal, X, ChevronUp } from "lucide-react"

export function Console() {
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleExpand = () => setIsExpanded(!isExpanded)

    return (
        <div
            className={`
    fixed bottom-0 left-64 right-0 bg-white border-t border-zinc-200 shadow-lg transition-all duration-300 ease-in-out z-50 flex flex-col
    ${isExpanded ? "h-[50vh]" : "h-12"}
  `}
        >
            {/* Header / Minimized State */}
            <div
                onClick={!isExpanded ? toggleExpand : undefined}
                className={`h-12 flex items-center px-4 justify-between cursor-pointer ${!isExpanded && "hover:bg-zinc-50"}`}
            >
                <div className="flex items-center gap-2 text-sm font-medium text-slate-700">
                    <Terminal size={16} className="text-sky-600" />
                    <span className="font-mono">Operator Console (Offline)</span>
                </div>

                {/* Toggle Button */}
                <button
                    onClick={(e) => { e.stopPropagation(); toggleExpand() }}
                    className="p-1.5 hover:bg-zinc-100 rounded text-zinc-400 hover:text-zinc-600"
                >
                    {isExpanded ? <X size={16} /> : <ChevronUp size={16} />}
                </button>
            </div>

            {/* Expanded Content */}
            {isExpanded && (
                <div className="flex-1 p-4 bg-zinc-50 flex items-center justify-center text-zinc-400">
                    Console currently disabled for maintenance.
                </div>
            )}
        </div>
    )
}
