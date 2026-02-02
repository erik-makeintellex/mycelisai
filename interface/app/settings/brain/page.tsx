"use client";

import { useState } from "react";
import { ModelHealth } from "@/components/settings/ModelHealth";

// Mock Data (matches brain.yaml default)
const INITIAL_CONFIG = {
    models: [
        { id: "local-qwen", name: "qwen2.5:7b" },
        { id: "cloud-gpt4", name: "gpt-4-turbo" },
    ],
    profiles: [
        { id: "coder", active_model: "local-qwen", temperature: 0.1 },
        { id: "chat", active_model: "cloud-gpt4", temperature: 0.7 },
    ]
};

export default function BrainSettings() {
    const [config, setConfig] = useState(INITIAL_CONFIG);

    // TODO: Implement GET /api/v1/config/brain and PATCH support

    return (
        <div className="p-8 max-w-4xl mx-auto font-mono">
            <h1 className="text-2xl font-bold mb-6 text-cyan-500">THE COGNITIVE MATRIX</h1>

            <div className="bg-slate-900 border border-slate-800 rounded-lg overflow-hidden">
                <table className="w-full text-left text-sm">
                    <thead className="bg-slate-950 text-slate-400">
                        <tr>
                            <th className="p-4">Profile</th>
                            <th className="p-4">Active Model</th>
                            <th className="p-4">Temperature</th>
                            <th className="p-4">Health</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800">
                        {config.profiles.map((profile) => (
                            <tr key={profile.id} className="hover:bg-slate-800/50 transition">
                                <td className="p-4 font-bold text-slate-200 uppercase">{profile.id}</td>
                                <td className="p-4">
                                    <select
                                        value={profile.active_model}
                                        onChange={(e) => {
                                            const newProfiles = config.profiles.map(p =>
                                                p.id === profile.id ? { ...p, active_model: e.target.value } : p
                                            );
                                            setConfig({ ...config, profiles: newProfiles });
                                        }}
                                        className="bg-slate-950 border border-slate-700 rounded px-2 py-1 text-slate-300 focus:border-cyan-500 outline-none"
                                    >
                                        {config.models.map(m => (
                                            <option key={m.id} value={m.id}>{m.name} ({m.id})</option>
                                        ))}
                                    </select>
                                </td>
                                <td className="p-4">
                                    <div className="flex items-center gap-2">
                                        <input
                                            type="range"
                                            min="0" max="1" step="0.1"
                                            value={profile.temperature}
                                            className="w-24 accent-cyan-500"
                                            readOnly // Readonly for MVP Demo
                                        />
                                        <span className="text-xs text-slate-500">{profile.temperature}</span>
                                    </div>
                                </td>
                                <td className="p-4">
                                    <ModelHealth modelId={profile.active_model} />
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="mt-6 flex justify-end">
                <button className="px-4 py-2 bg-cyan-600 hover:bg-cyan-500 text-white font-bold rounded shadow-[0_0_15px_#06b6d4]">
                    SYNC MATRIX
                </button>
            </div>
        </div>
    );
}
