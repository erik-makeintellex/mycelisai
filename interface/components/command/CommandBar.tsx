"use client"

import { Terminal } from "lucide-react"

export function CommandBar() {
    return (
        <div className="w-full max-w-3xl mx-auto mb-12 relative z-20">
            <div className="relative group">
                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-500">
                    <Terminal size={18} />
                </div>
                <input
                    autoFocus
                    type="text"
                    placeholder="Ready for instruction. Type 'help' or describe a task..."
                    className="w-full h-14 pl-12 pr-4 bg-zinc-900 text-zinc-50 rounded-lg shadow-xl border border-zinc-800 focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-500 text-base font-mono placeholder:text-zinc-600 transition-all"
                    onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                            console.log('Command executed:', e.currentTarget.value)
                            e.currentTarget.value = ''
                        }
                    }}
                />
                <div className="absolute right-4 top-1/2 -translate-y-1/2 flex items-center gap-2">
                    <span className="text-[10px] uppercase text-zinc-600 font-bold bg-zinc-800 px-1.5 py-0.5 rounded border border-zinc-700">Enter</span>
                </div>
            </div>
        </div>
    )
}
