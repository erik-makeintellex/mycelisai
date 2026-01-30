'use client';

import { useState, useEffect } from 'react';
import { API_BASE_URL } from '@/config';

interface AIModel {
    id: string;
    name: string;
    provider: string;
    context_window: number;
    description?: string;
}

export default function ModelsPage() {
    const [models, setModels] = useState<AIModel[]>([]);
    const [newModel, setNewModel] = useState({
        id: '',
        name: '',
        provider: 'openai',
        context_window: 128000,
        description: ''
    });

    useEffect(() => {
        const fetchModels = async () => {
            try {
                const res = await fetch(`${API_BASE_URL}/models`);
                const data = await res.json();
                setModels(data);
            } catch (error) {
                // eslint-disable-next-line no-console
                console.error(error);
            }
        };
        fetchModels();
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const res = await fetch(`${API_BASE_URL}/models`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(newModel)
            });
            if (res.ok) {
                setModels([...models, newModel as AIModel]);
                setNewModel({ id: '', name: '', provider: 'openai', context_window: 128000, description: '' });
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    return (
        <div className="space-y-8">
            <div>
                <h1 className="text-3xl font-bold text-zinc-100">Model Registry</h1>
                <p className="text-zinc-400 mt-2">
                    Register external LLM providers (OpenAI, Anthropic) or local models (Ollama, vLLM) for your agents to use.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <div>
                    <h2 className="text-xl font-semibold mb-4 text-zinc-300">Add New Model</h2>
                    <form onSubmit={handleSubmit} className="space-y-4 p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                        <div>
                            <label className="block text-sm font-medium mb-1 text-zinc-400">Model ID</label>
                            <input
                                type="text"
                                value={newModel.id}
                                onChange={(e) => setNewModel({ ...newModel, id: e.target.value })}
                                className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                placeholder="e.g., gpt-4-turbo"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1 text-zinc-400">Display Name</label>
                            <input
                                type="text"
                                value={newModel.name}
                                onChange={(e) => setNewModel({ ...newModel, name: e.target.value })}
                                className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                placeholder="e.g., GPT-4 Turbo"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1 text-zinc-400">Provider</label>
                            <select
                                value={newModel.provider}
                                onChange={(e) => setNewModel({ ...newModel, provider: e.target.value })}
                                className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                            >
                                <option value="openai">OpenAI</option>
                                <option value="anthropic">Anthropic</option>
                                <option value="ollama">Ollama</option>
                                <option value="vllm">vLLM</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1 text-zinc-400">Context Window</label>
                            <input
                                type="number"
                                value={newModel.context_window}
                                onChange={(e) => setNewModel({ ...newModel, context_window: Number(e.target.value) })}
                                className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                            />
                        </div>
                        <button
                            type="submit"
                            className="w-full bg-zinc-100 text-zinc-900 p-2 rounded-lg hover:bg-white transition-colors font-semibold"
                        >
                            Register Model
                        </button>
                    </form>
                </div>

                <div>
                    <h2 className="text-xl font-semibold mb-4 text-zinc-300">Available Models</h2>
                    <div className="space-y-4">
                        {models.map(model => (
                            <div key={model.id} className="p-4 border border-zinc-700 rounded-xl bg-zinc-800 flex justify-between items-center shadow-sm">
                                <div>
                                    <h3 className="font-medium text-zinc-200">{model.name}</h3>
                                    <p className="text-sm text-zinc-500">{model.provider} • {model.context_window.toLocaleString()} ctx</p>
                                </div>
                                <div className="px-2 py-1 bg-zinc-800 rounded text-xs text-zinc-400 font-mono">
                                    {model.id}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
