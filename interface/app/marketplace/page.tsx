"use client";

import { useEffect, useState } from "react";
import { Box, Cloud, Network, Plug } from "lucide-react";

interface Template {
    id: string;
    name: string;
    type: string;
    image: string;
    config_schema: any;
    description?: string;
}

export default function MarketplacePage() {
    const [templates, setTemplates] = useState<Template[]>([]);
    const [selected, setSelected] = useState<Template | null>(null);

    useEffect(() => {
        fetch("/api/v1/registry/templates")
            .then(res => res.json())
            .then(data => {
                // Handle null case or empty
                if (Array.isArray(data)) setTemplates(data);
            })
            .catch(err => console.error(err));
    }, []);

    // Mock Loading
    // const templates = [
    //     { id: '1', name: 'Weather Poller', type: 'ingress', description: 'Fetch weather data from OpenWeatherMap', image: 'mycelis/weather', config_schema: {} },
    //     { id: '2', name: 'Slack Bot', type: 'egress', description: 'Post messages to Slack channel', image: 'mycelis/slack', config_schema: {} },
    // ];

    const [installing, setInstalling] = useState(false);

    const handleInstall = async () => {
        if (!selected) return;
        setInstalling(true);

        try {
            // 1. Get Default Team (TODO: Context)
            const teamRes = await fetch("/api/v1/teams");
            const teamData = await teamRes.json();
            const teamID = (teamData.teams && teamData.teams[0]?.id) || (Array.isArray(teamData) ? teamData[0]?.id : null);

            if (!teamID) throw new Error("No team found");

            // 2. Default Config (Mock Form)
            // Real implementation would gather state from <ConfigForm>
            // We just send empty or default values if schema allows
            const config: any = {};
            if (selected.config_schema?.properties?.city) config.city = "Seattle";
            if (selected.config_schema?.properties?.api_key) config.api_key = "123456";
            if (selected.config_schema?.properties?.webhook_url) config.webhook_url = "https://hooks.slack.com/services/xxx";

            const res = await fetch(`/api/v1/teams/${teamID}/connectors`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    template_id: selected.id,
                    name: `${selected.name} (Instance)`,
                    config: config
                })
            });

            if (!res.ok) throw new Error(await res.text());

            alert("Installed Successfully!");
            setSelected(null);
        } catch (err: any) {
            alert("Installation Failed: " + err.message);
        } finally {
            setInstalling(false);
        }
    };

    return (
        <div className="h-full flex flex-col bg-zinc-50/50 p-8 overflow-y-auto">
            <header className="mb-8">
                <h1 className="text-2xl font-bold font-mono text-zinc-800">Marketplace</h1>
                <p className="text-zinc-500 mt-1">Install capabilities, connectors, and intelligence blueprints.</p>
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
                            <button className="text-sm font-medium text-sky-600 hover:text-sky-700">Configure & Install &rarr;</button>
                        </div>
                    </div>
                ))}

                {/* Loading State */}
                {templates.length === 0 && (
                    <div className="col-span-full py-12 text-center text-zinc-400">
                        Loading Registry...
                    </div>
                )}
            </div>

            {/* Modal */}
            {selected && (
                <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
                    onClick={() => setSelected(null)}
                >
                    <div className="bg-white rounded-xl shadow-xl p-8 max-w-lg w-full" onClick={e => e.stopPropagation()}>
                        <h2 className="text-xl font-bold mb-4">Install {selected.name}</h2>
                        <div className="mb-4 text-xs text-zinc-500 bg-zinc-50 p-2 rounded">
                            Config Form (Auto-Filled for Demo):
                            <ul className="list-disc pl-4 mt-1">
                                <li>City: Seattle</li>
                                <li>API Key: ******</li>
                            </ul>
                        </div>
                        <pre className="bg-zinc-100 p-4 rounded text-xs font-mono mb-4 text-zinc-600 overflow-auto max-h-40">
                            {JSON.stringify(selected.config_schema, null, 2)}
                        </pre>
                        <div className="flex justify-end gap-2">
                            <button className="px-4 py-2 text-sm text-zinc-600 hover:bg-zinc-100 rounded" onClick={() => setSelected(null)}>Cancel</button>
                            <button
                                className="px-4 py-2 text-sm bg-black text-white hover:bg-zinc-800 rounded disabled:opacity-50"
                                onClick={handleInstall}
                                disabled={installing}
                            >
                                {installing ? "Installing..." : "Install"}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
