"use client";

import { useEffect, useState } from "react";
import { Cpu, Activity, Server, Shield, Terminal } from "lucide-react";

interface Agent {
    id: string;
    team: string;
    status: string;
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
                        setAgents([]);
                    }
                })
                .catch(err => console.error(err));
        };
        fetchAgents();
        const timer = setInterval(fetchAgents, 5000);
        return () => clearInterval(timer);
    }, []);

    return (
        <div className="p-4 bg-cortex-bg border border-cortex-border rounded-xl h-full flex flex-col">
            <h3 className="text-xs font-bold text-cortex-text-muted mb-4 flex items-center gap-2 font-mono uppercase tracking-wider">
                <Activity size={12} className="text-cortex-success" />
                ACTIVE_NODES ({agents.length})
            </h3>

            <div className="grid grid-cols-2 lg:grid-cols-3 gap-3">
                {agents.map((agent) => {
                    let Icon = Server;
                    let color = "text-cortex-text-muted";

                    if (agent.id.startsWith("cli")) { Icon = Terminal; color = "text-pink-500"; }
                    if (agent.id.startsWith("core")) { Icon = Shield; color = "text-cortex-success"; }
                    if (agent.id.startsWith("interface")) { Icon = Cpu; color = "text-cortex-info"; }

                    return (
                        <div key={agent.id} className="bg-cortex-surface border border-cortex-border p-3 rounded-lg flex items-center gap-3">
                            <div className={`p-2 rounded-md bg-cortex-bg/50 ${color}`}>
                                <Icon size={16} />
                            </div>
                            <div className="flex flex-col overflow-hidden">
                                <span className="text-xs font-bold text-cortex-text-main truncate font-mono" title={agent.id}>
                                    {agent.id}
                                </span>
                                <span className="text-[10px] text-cortex-text-muted uppercase tracking-wider font-mono">
                                    {agent.team} / {agent.status}
                                </span>
                            </div>
                        </div>
                    );
                })}

                {agents.length === 0 && (
                    <div className="col-span-3 text-center py-8 text-cortex-text-muted text-xs italic font-mono">
                        No active agents detected via SCIP.
                    </div>
                )}
            </div>
        </div>
    );
}
