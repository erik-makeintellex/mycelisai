"use client";

import { useEffect, useState } from "react";
import { Cpu, Activity, Server, Shield, Terminal } from "lucide-react";

interface Agent {
    id: string;
    team: string; // e.g. "core", "cli"
    status: string; // "active"
}

export default function NetworkMap() {
    const [agents, setAgents] = useState<Agent[]>([]);

    useEffect(() => {
        const fetchAgents = () => {
            fetch("/agents")
                .then(res => res.json())
                .then(data => {
                    if (data.agents && Array.isArray(data.agents)) {
                        setAgents(data.agents);
                    } else if (data.agents === null) {
                        setAgents([])
                    }
                })
                .catch(err => console.error(err));
        };
        fetchAgents();
        const timer = setInterval(fetchAgents, 5000);
        return () => clearInterval(timer);
    }, []);

    return (
        <div className="p-4 bg-[#0b101a] border border-slate-800 rounded-xl h-full flex flex-col">
            <h3 className="text-xs font-bold text-slate-400 mb-4 flex items-center gap-2">
                <Activity size={12} className="text-emerald-500" />
                ACTIVE_NODES ({agents.length})
            </h3>

            <div className="grid grid-cols-2 lg:grid-cols-3 gap-3">
                {agents.map((agent) => {
                    // Icon logic based on ID prefix
                    // Default
                    let Icon = Server;
                    let color = "text-slate-400";

                    if (agent.id.startsWith("cli")) { Icon = Terminal; color = "text-pink-500"; }
                    if (agent.id.startsWith("core")) { Icon = Shield; color = "text-emerald-500"; }
                    if (agent.id.startsWith("interface")) { Icon = Cpu; color = "text-blue-500"; }

                    return (
                        <div key={agent.id} className="bg-slate-900 border border-slate-800 p-3 rounded-lg flex items-center gap-3">
                            <div className={`p-2 rounded-md bg-slate-800/50 ${color}`}>
                                <Icon size={16} />
                            </div>
                            <div className="flex flex-col overflow-hidden">
                                <span className="text-xs font-bold text-slate-200 truncate" title={agent.id}>
                                    {agent.id}
                                </span>
                                <span className="text-[10px] text-slate-500 uppercase tracking-wider">
                                    {agent.team} / {agent.status}
                                </span>
                            </div>
                        </div>
                    )
                })}

                {agents.length === 0 && (
                    <div className="col-span-3 text-center py-8 text-slate-600 text-xs italic">
                        No active agents detected via SCIP.
                    </div>
                )}
            </div>
        </div>
    );
}
