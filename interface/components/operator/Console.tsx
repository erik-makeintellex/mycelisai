"use client"

import { useState, useRef, useEffect } from "react"
import { useChat } from "@ai-sdk/react"
import { Terminal, X, ChevronUp, Bot, User as UserIcon, Send } from "lucide-react"

export function Console() {
    const [isExpanded, setIsExpanded] = useState(false)
    const { messages, input, handleInputChange, handleSubmit, isLoading } = useChat()
    const messagesEndRef = useRef<HTMLDivElement>(null)

    // Auto-scroll
    useEffect(() => {
        if (isExpanded) {
            messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
        }
    }, [messages, isExpanded])

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
                    <span className="font-mono">Operator Console</span>
                    {isLoading && <span className="text-xs text-zinc-400 animate-pulse">Processing...</span>}
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
                <>
                    {/* Chat History */}
                    <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-zinc-50/50">
                        {messages.length === 0 && (
                            <div className="h-full flex flex-col items-center justify-center text-zinc-400 text-sm">
                                <Bot size={32} className="mb-2 opacity-20" />
                                <p>Detailed instructions or standard chat.</p>
                            </div>
                        )}

                        {messages.map((m: any) => (
                            <div key={m.id} className={`flex items-start gap-3 ${m.role === 'user' ? 'flex-row-reverse' : ''}`}>
                                <div className={`
                        w-8 h-8 rounded-full flex items-center justify-center shrink-0 border
                        ${m.role === 'user' ? 'bg-white border-zinc-200' : 'bg-sky-100 border-sky-200 text-sky-700'}
                    `}>
                                    {m.role === 'user' ? <UserIcon size={14} className="text-zinc-500" /> : <Bot size={14} />}
                                </div>
                                <div className={`
                        max-w-[80%] rounded-lg px-4 py-2 text-sm shadow-sm border
                        ${m.role === 'user'
                                        ? 'bg-white border-zinc-200 text-slate-800'
                                        : 'bg-white border-zinc-200 text-slate-800' // Keeping it clean/uniform for now
                                    }
                    `}>
                                    {m.content}
                                </div>
                            </div>
                        ))}
                        <div ref={messagesEndRef} />
                    </div>

                    {/* Input Area */}
                    <div className="p-4 border-t border-zinc-200 bg-white">
                        <form onSubmit={handleSubmit} className="flex gap-2">
                            <input
                                className="flex-1 bg-zinc-50 border border-zinc-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500/20 focus:border-sky-500 font-mono"
                                value={input}
                                onChange={handleInputChange}
                                placeholder="Instructions..."
                                autoFocus
                            />
                            <button
                                type="submit"
                                disabled={isLoading || !input.trim()}
                                className="bg-black text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-zinc-800 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                            >
                                <Send size={14} />
                                Run
                            </button>
                        </form>
                    </div>
                </>
            )}
        </div>
    )
}
