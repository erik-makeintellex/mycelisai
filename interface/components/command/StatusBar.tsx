"use client"

import { Wifi, Cpu, Layers } from "lucide-react"

export function StatusBar() {
    return (
        <div className="fixed top-4 right-8 flex items-center gap-6 z-20 bg-white/50 backdrop-blur-sm pl-4 py-2 rounded-l-lg border-l border-b border-t border-zinc-200/50 shadow-sm">
            <StatusItem
                label="UPLINK"
                value="ONLINE"
                color="text-emerald-600"
                icon={<Wifi size={14} />}
            />
            <div className="w-px h-4 bg-zinc-200" />
            <StatusItem
                label="BRAIN"
                value="HYBRID"
                color="text-sky-600"
                icon={<Cpu size={14} />}
            />
            <div className="w-px h-4 bg-zinc-200" />
            <StatusItem
                label="CAPACITY"
                value="4/12"
                color="text-slate-700"
                icon={<Layers size={14} />}
            />
        </div>
    )
}

function StatusItem({ label, value, color, icon }: { label: string, value: string, color: string, icon: React.ReactNode }) {
    return (
        <div className="flex flex-col items-end">
            <span className="text-[10px] font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-1.5">
                {label}
            </span>
            <span className={`text-xs font-mono font-semibold ${color} flex items-center gap-1.5 mt-0.5`}>
                {icon}
                {value}
            </span>
        </div>
    )
}
