"use client"

import { ReactNode } from "react"
import { ArrowUpRight, ArrowDownRight } from "lucide-react"

interface MetricCardProps {
    label: string
    value: string
    subtext?: string
    subtextClass?: string
    trend?: "up" | "down" | "neutral"
    icon?: ReactNode
}

export function MetricCard({ label, value, subtext, subtextClass, trend, icon }: MetricCardProps) {
    return (
        <div className="bg-white border border-zinc-200 rounded-lg p-4 shadow-sm hover:shadow-md transition-shadow">
            <div className="flex justify-between items-start mb-2">
                <span className="text-xs font-medium text-slate-500 uppercase tracking-wide">{label}</span>
                {icon && <div className="text-zinc-400">{icon}</div>}
            </div>
            <div className="flex flex-col gap-1">
                <span className="text-2xl font-bold text-slate-900 tracking-tight">{value}</span>
                {subtext && (
                    <div className="flex items-center gap-1.5">
                        {trend === "up" && <ArrowUpRight size={12} className="text-emerald-600" />}
                        {trend === "down" && <ArrowDownRight size={12} className="text-rose-600" />}
                        <span className={`text-xs ${subtextClass || "text-zinc-500"}`}>{subtext}</span>
                    </div>
                )}
            </div>
        </div>
    )
}
