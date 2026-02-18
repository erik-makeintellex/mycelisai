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
            // 2. Extract Config from Form
            const config: Record<string, unknown> = {};
            const formElements = document.querySelectorAll(`[name]`) as NodeListOf<HTMLInputElement | HTMLSelectElement>;

            formElements.forEach(el => {
                if (el.closest('.fixed')) { // Ensure we only get inputs from the modal
                    const key = el.getAttribute('name');
                    if (key && selected.config_schema?.properties?.[key]) {
                        const schema = selected.config_schema.properties[key];
                        let value: any = el.value;

                        // Type coercion
                        if (schema.type === 'number' || schema.type === 'integer') {
                            value = Number(value);
                        } else if (schema.type === 'boolean') {
                            value = value === 'true';
                        }

                        if (value !== "" && value !== null) {
                            config[key] = value;
                        }
                    }
                }
            });

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
                        {/* Dynamic Configuration Form */}
                        <div className="mb-6 space-y-4 max-h-[60vh] overflow-y-auto pr-2">
                            <div className="p-3 bg-zinc-50 rounded text-xs text-zinc-500 mb-2">
                                Configure the {selected.name} target before enabling.
                            </div>

                            {/* Team Selection */}
                            <div>
                                <label className="block text-xs font-medium text-zinc-700 mb-1 uppercase tracking-wider">Assigned Team</label>
                                <select
                                    className="w-full bg-white border border-zinc-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black transition-shadow"
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

                            <hr className="border-zinc-100 my-4" />

                            {selected.config_schema?.properties ? (
                                Object.entries(selected.config_schema.properties).map(([key, val]: [string, any]) => (
                                    <div key={key}>
                                        <label className="block text-xs font-medium text-zinc-700 mb-1">
                                            {val.title || key}
                                            {val.description && <span className="ml-2 font-normal text-zinc-400">- {val.description}</span>}
                                        </label>

                                        {val.type === 'boolean' ? (
                                            <select
                                                name={key}
                                                className="w-full bg-white border border-zinc-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black"
                                                defaultValue={val.default?.toString() || "false"}
                                            >
                                                <option value="true">True</option>
                                                <option value="false">False</option>
                                            </select>
                                        ) : val.enum ? (
                                            <select
                                                name={key}
                                                className="w-full bg-white border border-zinc-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black"
                                                defaultValue={val.default || ""}
                                            >
                                                {val.enum.map((opt: string) => (
                                                    <option key={opt} value={opt}>{opt}</option>
                                                ))}
                                            </select>
                                        ) : (
                                            <input
                                                type={val.type === 'number' || val.type === 'integer' ? 'number' : 'text'}
                                                name={key}
                                                className="w-full bg-white border border-zinc-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-black"
                                                placeholder={val.default ? `Default: ${val.default}` : `Enter ${key}...`}
                                                defaultValue={val.default}
                                            />
                                        )}
                                    </div>
                                ))
                            ) : (
                                <div className="text-center py-8 text-zinc-400 italic text-sm">
                                    No configuration required for this target.
                                </div>
                            )}
                        </div>
                        {/* Debug info removed for cleanliness */}
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
