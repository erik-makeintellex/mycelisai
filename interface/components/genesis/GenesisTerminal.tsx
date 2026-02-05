"use client";
import React, { useState, useEffect } from "react";
import { Terminal, Bot, Server, ArrowRight, Activity, ShieldCheck } from "lucide-react";

interface Node {
    id: string;
    status: string;
    capabilities: string[];
}

export const GenesisTerminal = () => {
    const [state, setState] = useState<"welcome" | "negotiation">("welcome");
    const [messages, setMessages] = useState<{ role: string, content: string }[]>([
        { role: 'system', content: 'Connection Established. Waiting for input...' }
    ]);
    const [pendingNodes, setPendingNodes] = useState<Node[]>([]);

    // Poll for ghost nodes
    useEffect(() => {
        if (state === "negotiation") {
            const interval = setInterval(async () => {
                try {
                    const res = await fetch('/api/v1/nodes/pending');
                    if (res.ok) {
                        const data = await res.json();
                        setPendingNodes(data || []);
                    }
                } catch (e) {
                    console.error("Poll failed", e);
                }
            }, 2000);
            return () => clearInterval(interval);
        }
    }, [state]);

    if (state === "welcome") {
        return (
            <div className="h-full flex flex-col items-center justify-center bg-zinc-50 p-6">
                <div className="max-w-4xl w-full space-y-8">
                    <div className="text-center space-y-2">
                        <h1 className="text-4xl font-bold tracking-tight text-slate-900">Mycelis Cortex V6.1</h1>
                        <p className="text-zinc-500">Initialize the Recursive Swarm Operating System.</p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {/* Option 1: The Architect */}
                        <button onClick={() => setState("negotiation")} className="relative group p-6 bg-white border border-zinc-200 rounded-xl shadow-sm hover:shadow-md hover:border-indigo-500 transition-all text-left">
                            <div className="absolute top-0 left-0 w-full h-1 bg-indigo-500 rounded-t-xl opacity-0 group-hover:opacity-100 transition-opacity" />
                            <Bot className="w-8 h-8 text-indigo-600 mb-4" />
                            <h3 className="font-bold text-slate-900 mb-2">The Architect</h3>
                            <p className="text-sm text-zinc-500">Manual construction. Chat with the Consul to build from scratch.</p>
                            <ArrowRight className="absolute bottom-6 right-6 w-5 h-5 text-indigo-400 opacity-0 group-hover:opacity-100 transition-all transform group-hover:translate-x-1" />
                        </button>

                        {/* Option 2: The Commander */}
                        <button className="relative group p-6 bg-white border border-zinc-200 rounded-xl shadow-sm hover:shadow-md hover:border-emerald-500 transition-all text-left">
                            <div className="absolute top-0 left-0 w-full h-1 bg-emerald-500 rounded-t-xl opacity-0 group-hover:opacity-100 transition-opacity" />
                            <Server className="w-8 h-8 text-emerald-600 mb-4" />
                            <h3 className="font-bold text-slate-900 mb-2">The Commander</h3>
                            <p className="text-sm text-zinc-500">Select a specialized Archetype (IoT Swarm, SaaS Team).</p>
                        </button>

                        {/* Option 3: The Explorer */}
                        <button className="relative group p-6 bg-white border border-zinc-200 rounded-xl shadow-sm hover:shadow-md hover:border-amber-500 transition-all text-left">
                            <div className="absolute top-0 left-0 w-full h-1 bg-amber-500 rounded-t-xl opacity-0 group-hover:opacity-100 transition-opacity" />
                            <Activity className="w-8 h-8 text-amber-600 mb-4" />
                            <h3 className="font-bold text-slate-900 mb-2">The Explorer</h3>
                            <p className="text-sm text-zinc-500">Spin up virtual agents to test capabilities safely.</p>
                        </button>
                    </div>
                </div>
            </div>
        )
    }

    // State 2: Negotiation
    return (
        <div className="h-full flex flex-col md:flex-row divide-y md:divide-y-0 md:divide-x divide-zinc-200">
            {/* Left: Chat */}
            <div className="flex-1 flex flex-col bg-white">
                <div className="p-4 border-b border-zinc-100 flex justify-between items-center">
                    <span className="text-xs font-bold text-zinc-400 uppercase tracking-wider">Consular Uplink</span>
                    <span className="text-[10px] bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded-full font-mono">ONLINE</span>
                </div>
                <div className="flex-1 p-4 space-y-4 overflow-y-auto">
                    {messages.map((m, i) => (
                        <div key={i} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                            <div className={`max-w-[80%] rounded-lg px-4 py-2 text-sm ${m.role === 'user' ? 'bg-zinc-900 text-white' : 'bg-zinc-100 text-zinc-800'}`}>
                                {m.content}
                            </div>
                        </div>
                    ))}
                    {pendingNodes.length > 0 && messages.length === 1 && (
                        <div className="flex justify-start">
                            <div className="max-w-[80%] rounded-lg px-4 py-2 text-sm bg-indigo-50 border border-indigo-100 text-indigo-900">
                                I have detected {pendingNodes.length} new hardware node(s). Would you like to embrace them into the Swarm?
                            </div>
                        </div>
                    )}
                </div>
                <div className="p-4 border-t border-zinc-100">
                    <div className="flex gap-2">
                        <input type="text" placeholder="Type your directive..." className="flex-1 bg-zinc-50 border border-zinc-200 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                        <button className="bg-zinc-900 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-zinc-800">Send</button>
                    </div>
                </div>
            </div>

            {/* Right: Graph / Resources */}
            <div className="w-full md:w-80 bg-zinc-50 flex flex-col">
                <div className="p-4 border-b border-zinc-200">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-zinc-500">Resource Graph</h3>
                </div>
                <div className="flex-1 p-4 space-y-2">
                    {pendingNodes.map(node => (
                        <div key={node.id} className="bg-white border border-zinc-200 rounded-lg p-3 shadow-sm flex items-center gap-3 animate-pulse">
                            <div className="w-8 h-8 bg-zinc-100 rounded flex items-center justify-center">
                                <Terminal size={14} className="text-zinc-600" />
                            </div>
                            <div className="flex-1 min-w-0">
                                <div className="flex justify-between items-center mb-0.5">
                                    <p className="text-sm font-medium text-slate-900 truncate">{node.id}</p>
                                    <span className="text-[10px] bg-amber-100 text-amber-700 px-1.5 rounded font-mono">GHOST</span>
                                </div>
                                <div className="flex flex-wrap gap-1">
                                    {node.capabilities.map(cap => (
                                        <span key={cap} className="text-[9px] bg-zinc-100 text-zinc-500 px-1 rounded">{cap}</span>
                                    ))}
                                </div>
                            </div>
                        </div>
                    ))}
                    {pendingNodes.length === 0 && (
                        <div className="text-center py-8 text-zinc-400 text-sm">
                            <p>Scanning for hardware...</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
