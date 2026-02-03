"use client"

import { useState } from "react"
import { X, Cpu, Zap, FileJson, CheckCircle } from "lucide-react"

interface ForgeProps {
    isOpen: boolean
    onClose: () => void
}

type Step = "INTENT" | "BLUEPRINT" | "REVIEW" | "DEPLOYING"

export function Forge({ isOpen, onClose }: ForgeProps) {
    const [step, setStep] = useState<Step>("INTENT")
    const [intent, setIntent] = useState("")

    if (!isOpen) return null

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 backdrop-blur-sm">
            <div className="w-[600px] bg-[rgb(var(--surface))] border border-[rgb(var(--border))] rounded-xl shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">

                {/* Header */}
                <div className="h-14 border-b border-[rgb(var(--border))] flex items-center px-6 justify-between bg-zinc-50/50">
                    <div className="flex items-center gap-2">
                        <div className="w-8 h-8 bg-black text-white rounded-lg flex items-center justify-center">
                            <Zap size={16} fill="white" />
                        </div>
                        <div>
                            <h2 className="font-semibold text-sm">The Forge</h2>
                            <p className="text-xs text-zinc-500">Agent Provisioning Wizard</p>
                        </div>
                    </div>
                    <button onClick={onClose} className="p-2 hover:bg-zinc-100 rounded-md text-zinc-500">
                        <X size={16} />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 min-h-[300px]">
                    {step === "INTENT" && (
                        <div className="space-y-4">
                            <div className="space-y-2">
                                <label className="text-sm font-medium">What is the Prime Directive?</label>
                                <textarea
                                    className="w-full h-32 p-3 bg-zinc-50 border border-[rgb(var(--border))] rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-zinc-900/10 text-sm"
                                    placeholder="e.g., Use the Twitter API to find unhappy customers and escalate them to Jira..."
                                    value={intent}
                                    onChange={(e) => setIntent(e.target.value)}
                                    autoFocus
                                />
                            </div>
                            <div className="flex justify-end pt-4">
                                <button
                                    onClick={() => setStep("BLUEPRINT")}
                                    className="px-4 py-2 bg-black text-white rounded-lg text-sm font-medium hover:bg-zinc-800 transition-colors"
                                >
                                    Generate Blueprint
                                </button>
                            </div>
                        </div>
                    )}

                    {step === "BLUEPRINT" && (
                        <div className="space-y-4 flex flex-col items-center justify-center h-full py-10">
                            <div className="relative">
                                <div className="absolute inset-0 bg-sky-500/20 blur-xl rounded-full animate-pulse" />
                                <Cpu size={48} className="text-sky-600 relative z-10 animate-bounce" />
                            </div>
                            <p className="text-sm font-medium text-zinc-700 mt-4">Architecting Solution...</p>
                            <p className="text-xs text-zinc-500">Analyzing tools, permissions, and models.</p>

                            {/* Mock Transition */}
                            <button
                                onClick={() => setStep("REVIEW")}
                                className="mt-8 text-xs underline text-zinc-400 hover:text-zinc-600"
                            >
                                (Dev: Skip to Review)
                            </button>
                        </div>
                    )}

                    {step === "REVIEW" && (
                        <div className="space-y-4">
                            <div className="p-4 bg-zinc-50 border border-[rgb(var(--border))] rounded-lg space-y-3">
                                <div className="flex items-center gap-3 pb-3 border-b border-zinc-200">
                                    <FileJson size={16} className="text-zinc-500" />
                                    <span className="text-sm font-mono font-semibold">manifest.yaml</span>
                                </div>
                                <pre className="text-xs font-mono text-zinc-600 overflow-x-auto">
                                    {`name: customer-support-bot
role: responder
model: qwen2.5-coder:7b
tools:
  - twitter.search
  - jira.create_ticket
policy:
  max_spend: $5.00/day`}
                                </pre>
                            </div>
                            <div className="flex justify-end gap-2 pt-2">
                                <button
                                    onClick={() => setStep("INTENT")}
                                    className="px-4 py-2 text-zinc-600 text-sm font-medium hover:bg-zinc-100 rounded-lg"
                                >
                                    Back
                                </button>
                                <button
                                    onClick={() => {
                                        setStep("DEPLOYING")
                                        setTimeout(onClose, 2000)
                                    }}
                                    className="px-4 py-2 bg-[rgb(var(--status-stable))] text-white rounded-lg text-sm font-medium hover:brightness-110 transition-colors flex items-center gap-2"
                                >
                                    <CheckCircle size={14} />
                                    Deploy Agent
                                </button>
                            </div>
                        </div>
                    )}

                    {step === "DEPLOYING" && (
                        <div className="flex flex-col items-center justify-center h-full py-10 space-y-3">
                            <div className="w-12 h-12 border-4 border-emerald-500 border-t-transparent rounded-full animate-spin" />
                            <p className="text-sm font-medium text-emerald-700">Fabricating Container...</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
