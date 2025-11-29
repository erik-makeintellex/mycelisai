'use client';

import { useState, useEffect } from 'react';
import { API_BASE_URL } from '@/config';

interface AIModel {
    id: string;
    name: string;
    provider: string;
}

interface AgentFormProps {
    initialData?: any;
    onSuccess?: () => void;
}

export default function AgentForm({ initialData, onSuccess }: AgentFormProps) {
    const [name, setName] = useState('');
    const [language, setLanguage] = useState('python');
    const [capabilities, setCapabilities] = useState('');

    // New State Fields
    const [inputs, setInputs] = useState('');
    const [outputs, setOutputs] = useState('');
    const [replicas, setReplicas] = useState(1);
    const [backend, setBackend] = useState('');
    const [promptConfig, setPromptConfig] = useState('{}');
    const [availableModels, setAvailableModels] = useState<AIModel[]>([]);

    useEffect(() => {
        if (initialData) {
            setName(initialData.name);
            setLanguage(initialData.languages[0] || 'python');
            setCapabilities(initialData.capabilities.join(', '));
            setInputs(initialData.messaging?.inputs?.join(', ') || '');
            setOutputs(initialData.messaging?.outputs?.join(', ') || '');
            setReplicas(initialData.deployment?.replicas || 1);
            setBackend(initialData.backend || '');
            setPromptConfig(JSON.stringify(initialData.prompt_config || {}, null, 2));
        }
    }, [initialData]);

    useEffect(() => {
        const fetchModels = async () => {
            console.log("Fetching models from:", `${API_BASE_URL}/models`);
            try {
                const res = await fetch(`${API_BASE_URL}/models`);
                const data = await res.json();
                setAvailableModels(data);
                if (data.length > 0 && !backend && !initialData) {
                    setBackend(data[0].id);
                }
            } catch (error) {
                // eslint-disable-next-line no-console
                console.error(error);
            }
        };
        fetchModels();
    }, [backend, initialData]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        let parsedPromptConfig = {};
        try {
            parsedPromptConfig = JSON.parse(promptConfig);
        } catch (e) {
            alert('Invalid JSON for Prompt Config');
            return;
        }

        const agent = {
            name,
            languages: [language],
            capabilities: capabilities.split(',').map(c => c.trim()).filter(Boolean),
            prompt_config: parsedPromptConfig,
            messaging: {
                inputs: inputs.split(',').map(s => s.trim()).filter(Boolean),
                outputs: outputs.split(',').map(s => s.trim()).filter(Boolean)
            },
            deployment: {
                replicas: Number(replicas),
                constraints: {}
            },
            backend
        };

        // eslint-disable-next-line no-console
        console.log('Registering agent:', agent);

        try {
            const res = await fetch(`${API_BASE_URL}/agents/register`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(agent)
            });
            if (res.ok) {
                alert(initialData ? 'Agent updated successfully!' : 'Agent registered successfully!');
                if (onSuccess) onSuccess();
                if (!initialData) {
                    // Reset form only if creating new
                    setName('');
                    setCapabilities('');
                    setInputs('');
                    setOutputs('');
                    setPromptConfig('{}');
                }
            } else {
                alert('Failed to register agent');
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
            alert('Error connecting to API');
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-6 p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Agent Name</label>
                    <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all placeholder:text-zinc-600"
                        required
                        placeholder="e.g., image-processor-v1"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Primary Language</label>
                    <select
                        value={language}
                        onChange={(e) => setLanguage(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all"
                    >
                        <option value="python">Python</option>
                        <option value="javascript">JavaScript</option>
                        <option value="go">Go</option>
                        <option value="rust">Rust</option>
                    </select>
                </div>
            </div>

            <div>
                <label className="block text-sm font-medium mb-1 text-zinc-400">Capabilities (comma separated)</label>
                <input
                    type="text"
                    value={capabilities}
                    onChange={(e) => setCapabilities(e.target.value)}
                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all placeholder:text-zinc-600"
                    placeholder="image-processing, nlp, sensor-reading"
                />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Input Channels</label>
                    <input
                        type="text"
                        value={inputs}
                        onChange={(e) => setInputs(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all placeholder:text-zinc-600"
                        placeholder="sensors.temp, video.stream"
                    />
                    <p className="text-xs text-zinc-500 mt-1">NATS subjects this agent listens to.</p>
                </div>
                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Output Channels</label>
                    <input
                        type="text"
                        value={outputs}
                        onChange={(e) => setOutputs(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all placeholder:text-zinc-600"
                        placeholder="processed.data, alerts"
                    />
                    <p className="text-xs text-zinc-500 mt-1">NATS subjects this agent publishes to.</p>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Replicas</label>
                    <input
                        type="number"
                        min="1"
                        value={replicas}
                        onChange={(e) => setReplicas(Number(e.target.value))}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all"
                    />
                    <p className="text-xs text-zinc-500 mt-1">Number of parallel instances.</p>
                </div>

                <div>
                    <label className="block text-sm font-medium mb-1 text-zinc-400">Backend Service</label>
                    <select
                        value={backend}
                        onChange={(e) => setBackend(e.target.value)}
                        className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all"
                    >
                        {availableModels.length > 0 ? (
                            availableModels.map(model => (
                                <option key={model.id} value={model.id}>
                                    {model.name} ({model.provider})
                                </option>
                            ))
                        ) : (
                            <option value="" disabled>No models available</option>
                        )}
                    </select>
                </div>
            </div>

            <div>
                <label className="block text-sm font-medium mb-1 text-zinc-400">Prompt Configuration (JSON)</label>
                <textarea
                    value={promptConfig}
                    onChange={(e) => setPromptConfig(e.target.value)}
                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 font-mono text-sm focus:ring-2 focus:ring-zinc-600 focus:border-zinc-600 outline-none transition-all h-32 placeholder:text-zinc-600"
                    placeholder='{"system_prompt": "You are a helpful assistant..."}'
                />
                <p className="text-xs text-zinc-500 mt-1">
                    JSON configuration for the agent's persona and instructions. Must be valid JSON.
                </p>
            </div>

            <button
                type="submit"
                className="w-full bg-zinc-100 text-zinc-900 p-3 rounded-lg hover:bg-white transition-colors font-semibold shadow-lg shadow-zinc-900/20"
            >
                {initialData ? 'Update Agent' : 'Create Agent'}
            </button>
        </form>
    );
}
