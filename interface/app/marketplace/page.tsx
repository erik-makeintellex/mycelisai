"use client";

import { useEffect, useState } from "react";
import { Box, Cloud, Network, Plug } from "lucide-react";

interface Template {
    id: string;
    name: string;
    type: string;
    image: string;
    config_schema: Record<string, any>; // Schema is dynamic JSON
    description?: string;
}

export default function MarketplacePage() {
    const [templates, setTemplates] = useState<Template[]>([]);
    const [selected, setSelected] = useState<Template | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch("/api/v1/registry/templates")
            .then(res => res.json())
            .then(data => {
                if (Array.isArray(data)) setTemplates(data);
            })
            .catch(err => console.error(err))
            .finally(() => setLoading(false));
    }, []);

    // Team Selection
    interface Team {
        id: string;
        name: string;
    }
    const [teams, setTeams] = useState<Team[]>([]);
    const [selectedTeamId, setSelectedTeamId] = useState<string>("");

    useEffect(() => {
        // Fetch Teams for dropdown
        fetch("/api/v1/teams")
            .then(res => res.json())
            .then(data => {
                // Handle different response structures if needed (array vs {teams: []})
                const list = Array.isArray(data) ? data : (data.teams || []);
                setTeams(list);
                if (list.length > 0) setSelectedTeamId(list[0].id);
            })
            .catch(err => console.error("Failed to fetch teams:", err));
    }, []);

    const [installing, setInstalling] = useState(false);

    const handleInstall = async () => {
        if (!selected || !selectedTeamId) return;
        setInstalling(true);

        try {
            // 2. Default Config (Mock Form)
            // Real implementation would gather state from <ConfigForm>
            // We just send empty or default values if schema allows
            const config: Record<string, unknown> = {};
            if (selected.config_schema?.properties?.city) config.city = "Seattle";
            if (selected.config_schema?.properties?.api_key) config.api_key = "123456";
            if (selected.config_schema?.properties?.webhook_url) config.webhook_url = "https://hooks.slack.com/services/xxx";
            if (selected.config_schema?.properties?.frequency) config.frequency = 101.5;

            const res = await fetch(`/api/v1/teams/${selectedTeamId}/connectors`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    template_id: selected.id,
                    name: `${selected.name} (Instance)`,
                    config: config
                })
            });

            if (!res.ok) throw new Error(await res.text());

            alert("Service Configured & Assigned Successfully!");
            setSelected(null);
        } catch (err: unknown) {
            alert("Configuration Failed: " + (err instanceof Error ? err.message : String(err)));
        } finally {
            setInstalling(false);
        }
    };

    return (
        <div className="h-full flex flex-col bg-zinc-50/50 p-8 overflow-y-auto">
            <header className="mb-8">
                <h1 className="text-2xl font-bold font-mono text-zinc-800">Service Targets</h1>
                <p className="text-zinc-500 mt-1">Configure and manage available Swarm Capabilities.</p>
            </header>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {templates.map(t => (
                    <div key={t.id} className="bg-white border border-zinc-200 rounded-lg p-6 hover:shadow-md transition-shadow cursor-pointer"
                        onClick={() => setSelected(t)}
                    >
                        <div className="flex items-center gap-4 mb-4">
                            <div className="w-12 h-12 bg-sky-50 text-sky-600 rounded-lg flex items-center justify-center">
                                {t.type === 'ingress' ? <Cloud size={24} /> : <Plug size={24} />}
                            </div>
                            <div>
                                <h3 className="font-semibold text-zinc-900">{t.name}</h3>
                                <span className="text-xs uppercase tracking-wider text-zinc-400 font-mono">{t.type}</span>
                            </div>
                        </div>
                        <p className="text-sm text-zinc-600 line-clamp-2 min-h-[40px]">
                            {t.description || "Standard connector for " + t.name}
                        </p>
                        <div className="mt-4 pt-4 border-t border-zinc-100 flex justify-end">
                            <button className="text-sm font-medium text-sky-600 hover:text-sky-700">Configure & Enable &rarr;</button>
                        </div>
                    </div>
                ))}

                {/* Loading State */}
                {loading && (
                    <div className="col-span-full py-12 text-center text-zinc-400">
                        Loading Services...
                    </div>
                )}

                {/* Empty State */}
                {!loading && templates.length === 0 && (
                    <div className="col-span-full py-12 text-center text-zinc-400 border-2 border-dashed border-zinc-200 rounded-lg">
                        No Service Targets available in the Registry.
                    </div>
                )}
            </div>

            {/* Modal */}
            {selected && (
                <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
                    onClick={() => setSelected(null)}
                >
                    <div className="bg-white rounded-xl shadow-xl p-8 max-w-lg w-full" onClick={e => e.stopPropagation()}>
                        <h2 className="text-xl font-bold mb-4">Configure {selected.name}</h2>
                        <div className="mb-4 text-xs text-zinc-500 bg-zinc-50 p-2 rounded">
                            Target Configuration Parameters:
                            <ul className="list-disc pl-4 mt-1 space-y-1">
                                {selected.config_schema?.properties ? (
                                    Object.entries(selected.config_schema.properties).map(([key, val]: [string, any]) => (
                                        <li key={key}>
                                            <span className="font-semibold">{val.title || key}</span>
                                            {val.default && <span className="text-zinc-400 ml-1">(Default: {val.default})</span>}
                                        </li>
                                    ))
                                ) : (
                                    <li>No configuration schema defined.</li>
                                )}
                            </ul>
                        </div>
                        <div className="mb-4">
                            <label className="block text-sm font-medium text-zinc-700 mb-1">Assign to Team</label>
                            <select
                                className="w-full border border-zinc-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black"
                                value={selectedTeamId}
                                onChange={e => setSelectedTeamId(e.target.value)}
                            >
                                <option value="" disabled>Select a Team...</option>
                                {teams.map(team => (
                                    <option key={team.id} value={team.id}>
                                        {team.name}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="mb-4 text-xs text-zinc-500 bg-zinc-50 p-2 rounded">
                            Target Configuration Parameters:
                            {/* ... list ... */}
                        </div>
                        <pre className="bg-zinc-100 p-4 rounded text-xs font-mono mb-4 text-zinc-600 overflow-auto max-h-40">
                            {JSON.stringify(selected.config_schema, null, 2)}
                        </pre>
                        <div className="flex justify-end gap-2">
                            <button className="px-4 py-2 text-sm text-zinc-600 hover:bg-zinc-100 rounded" onClick={() => setSelected(null)}>Cancel</button>
                            <button
                                className="px-4 py-2 text-sm bg-black text-white hover:bg-zinc-800 rounded disabled:opacity-50"
                                onClick={handleInstall}
                                disabled={installing || !selectedTeamId}
                            >
                                {installing ? "Configuring..." : "Enable Target"}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
