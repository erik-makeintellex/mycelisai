"use client";

import { useState } from "react";
import useSWR from "swr";
import { Copy, Server, Zap, Compass, Box } from "lucide-react";

interface Node {
    id: string;
    type: string;
    status: string;
    last_seen: string;
    specs: Record<string, unknown>;
}

const fetcher = (url: string) => fetch(url).then(r => r.json());

export default function GenesisTerminal() {
    // 1. Poll for Ghost Nodes
    const { data: nodes, error } = useSWR<Node[]>("/api/v1/nodes/pending", fetcher, { refreshInterval: 2000 });
    const [selectedMode, setSelectedMode] = useState<"manual" | "template" | "explorer">("explorer");

    return (
        <div className="flex flex-col h-full bg-zinc-950/50 backdrop-blur-sm p-4 gap-4">
            {/* Header: Pathfinder Deck */}
            <div className="flex gap-4 p-2 bg-black/20 rounded-lg border border-white/5">
                <DeckCard
                    title="Manual Entry"
                    icon={<Copy size={16} />}
                    active={selectedMode === "manual"}
                    onClick={() => setSelectedMode("manual")}
                />
                <DeckCard
                    title="Template Library"
                    icon={<Box size={16} />}
                    active={selectedMode === "template"}
                    onClick={() => setSelectedMode("template")}
                />
                <DeckCard
                    title="Node Explorer"
                    icon={<Compass size={16} />}
                    active={selectedMode === "explorer"}
                    onClick={() => setSelectedMode("explorer")}
                />
            </div>

            {/* Main Content Area */}
            <div className="flex-1 bg-zinc-950/80 rounded-xl border border-white/10 p-6 overflow-hidden flex flex-col relative">

                {/* EXPLORER MODE */}
                {selectedMode === "explorer" && (
                    <>
                        <div className="flex items-center justify-between mb-6">
                            <h2 className="text-sm font-mono text-zinc-400 uppercase tracking-widest flex items-center gap-2">
                                <Zap size={14} className={nodes && nodes.length > 0 ? "text-amber-400" : "text-zinc-600"} />
                                DETECTED SIGNALS ({nodes?.length || 0})
                            </h2>
                            <span className="text-xs text-zinc-600 font-mono flex items-center gap-2">
                                {error ? (
                                    <span className="text-red-500">UPLINK OFFLINE</span>
                                ) : (
                                    <>
                                        <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                                        SCANNING...
                                    </>
                                )}
                            </span>
                        </div>

                        {/* Grid */}
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 overflow-y-auto pr-2">
                            {nodes?.map(node => (
                                <div key={node.id} draggable className="
                                    group bg-zinc-900 border border-zinc-800 hover:border-cyan-500/50 
                                    p-4 rounded-lg cursor-grab active:cursor-grabbing transition-all
                                    hover:shadow-[0_0_15px_-5px_cyan] relative overflow-hidden
                                ">
                                    <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-cyan-500 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />

                                    <div className="flex items-center justify-between mb-3">
                                        <div className="p-2 bg-emerald-500/10 text-emerald-400 rounded">
                                            <Server size={18} />
                                        </div>
                                        <span className="text-[10px] bg-zinc-800 px-2 py-1 rounded text-zinc-400 font-mono border border-zinc-700">
                                            {node.type.toUpperCase()}
                                        </span>
                                    </div>
                                    <div className="font-mono text-xs text-zinc-300 break-all">
                                        {node.id}
                                    </div>
                                    <div className="mt-2 text-[10px] text-zinc-500 flex justify-between">
                                        <span>GHOST NODE</span>
                                        <span>{new Date(node.last_seen).toLocaleTimeString()}</span>
                                    </div>
                                </div>
                            ))}

                            {(!nodes || nodes.length === 0) && (
                                <div className="col-span-full h-64 flex flex-col items-center justify-center text-zinc-700 font-mono text-sm border-2 border-dashed border-zinc-800 rounded-lg gap-4">
                                    <div className="p-4 bg-zinc-900/50 rounded-full">
                                        <Zap size={32} className="text-zinc-800" />
                                    </div>
                                    <span>NO HARDWARE DETECTED</span>
                                </div>
                            )}
                        </div>
                    </>
                )}

                {/* MANUAL MODE */}
                {selectedMode === "manual" && (
                    <div className="flex flex-col items-center justify-center h-full text-zinc-500 font-mono gap-4">
                        <Copy size={48} className="text-zinc-800" />
                        <p>MANUAL INGESTION PROTOCOL</p>
                        <button className="px-4 py-2 bg-zinc-800 text-zinc-400 text-xs rounded border border-zinc-700 hover:text-white hover:border-zinc-500 transition-colors">
                            INITIATE FORM
                        </button>
                    </div>
                )}

                {/* TEMPLATE MODE */}
                {selectedMode === "template" && (
                    <div className="flex flex-col items-center justify-center h-full text-zinc-500 font-mono gap-4">
                        <Box size={48} className="text-zinc-800" />
                        <p>TEMPLATE LIBRARY UNLINKED</p>
                        <button className="px-4 py-2 bg-zinc-800 text-zinc-400 text-xs rounded border border-zinc-700 hover:text-white hover:border-zinc-500 transition-colors">
                            CONNECT REGISTRY
                        </button>
                    </div>
                )}

            </div>
        </div>
    );
}

function DeckCard({ title, icon, active, onClick }: { title: string, icon: React.ReactNode, active: boolean, onClick: () => void }) {
    return (
        <button
            onClick={onClick}
            className={`
                flex-1 flex items-center justify-center gap-3 py-3 rounded-md transition-all
                text-sm font-medium border
                ${active
                    ? "bg-zinc-800 text-white border-zinc-600 shadow-lg"
                    : "bg-transparent text-zinc-500 border-transparent hover:bg-white/5 hover:text-zinc-300"
                }
            `}
        >
            {icon}
            {title}
        </button>
    );
}
