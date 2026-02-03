"use client"

import { Search, Wifi } from "lucide-react"

export function Header() {
    return (
        <header className="h-14 flex items-center px-6 border-b border-zinc-200 bg-white/80 backdrop-blur-md sticky top-0 z-40">
            {/* 1. Breadcrumbs */}
            <div className="flex items-center text-sm font-medium text-slate-500 w-1/3">
                <span className="text-slate-900 cursor-pointer hover:underline">Mycelis</span>
                <span className="mx-2 text-zinc-300">/</span>
                <span className="text-slate-900 cursor-pointer hover:underline">System Workspace</span>
                <span className="mx-2 text-zinc-300">/</span>
                <span className="text-slate-500">Mission Control</span>
            </div>

            {/* 2. Omni-Command (Center) */}
            <div className="flex-1 flex justify-center">
                <div className="relative w-full max-w-md group">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400 group-hover:text-zinc-600 transition-colors" size={14} />
                    <input
                        type="text"
                        placeholder="Command or Search (Ctrl+K)..."
                        className="w-full h-9 pl-9 pr-4 bg-zinc-50 border border-zinc-200 rounded-md text-sm text-slate-900 placeholder:text-zinc-400 focus:outline-none focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500 transition-all font-sans"
                    />
                    <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
                        <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] font-mono font-medium text-zinc-500 bg-white border border-zinc-200 rounded shadow-sm">
                            Ctrl
                        </kbd>
                        <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] font-mono font-medium text-zinc-500 bg-white border border-zinc-200 rounded shadow-sm">
                            K
                        </kbd>
                    </div>
                </div>
            </div>

            {/* 3. System Vitality (Right) */}
            <div className="w-1/3 flex justify-end items-center gap-4">
                <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                    <span className="text-xs font-medium text-slate-600">NATS Connected</span>
                </div>
                <div className="h-4 w-px bg-zinc-200" />
                <div className="flex items-center gap-1 text-emerald-600">
                    <Wifi size={14} />
                    <span className="text-xs font-mono font-medium">12ms</span>
                </div>
            </div>
        </header>
    )
}
