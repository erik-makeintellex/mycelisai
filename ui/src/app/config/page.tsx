'use client';

import { useState, useEffect } from 'react';
import { API_BASE_URL } from '@/config';

interface SystemConfig {
    ollama_base_url: string;
    database_url: string;
    nats_url: string;
}

interface HealthStatus {
    status: string;
    components: {
        database: string;
        nats: string;
        api: string;
        ollama: string;
    };
    environment: {
        nats_url: string;
        database_url: string;
        backend_host: string;
    };
}

export default function ConfigPage() {
    const [config, setConfig] = useState<SystemConfig>({
        ollama_base_url: '',
        database_url: '',
        nats_url: ''
    });
    const [health, setHealth] = useState<HealthStatus | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState('');

    useEffect(() => {
        fetchHealth();
        fetchConfig();
    }, []);

    const fetchHealth = async () => {
        try {
            const response = await fetch(`${API_BASE_URL}/health`);
            const data = await response.json();
            setHealth(data);
        } catch (error) {
            console.error('Failed to fetch health:', error);
        }
    };

    const fetchConfig = async () => {
        try {
            const response = await fetch(`${API_BASE_URL}/config`);
            if (response.ok) {
                const data = await response.json();
                setConfig(data);
            }
        } catch (error) {
            console.error('Failed to fetch config:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        setMessage('');
        try {
            const response = await fetch(`${API_BASE_URL}/config`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            if (response.ok) {
                setMessage('Configuration saved successfully! Restart API to apply changes.');
                fetchHealth();
            } else {
                setMessage('Failed to save configuration');
            }
        } catch (error) {
            setMessage('Error saving configuration');
        } finally {
            setSaving(false);
        }
    };

    const getStatusColor = (status: string) => {
        if (status.includes('connected') || status === 'healthy') return 'text-green-400';
        if (status.includes('disconnected') || status.includes('error')) return 'text-red-400';
        return 'text-yellow-400';
    };

    return (
        <div className="space-y-8">
            <h1 className="text-3xl font-bold text-zinc-100">System Configuration</h1>

            {/* Health Status */}
            <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                <h2 className="text-xl font-semibold text-zinc-100 mb-4">System Health</h2>
                {health ? (
                    <div className="space-y-2">
                        <div className="flex justify-between items-center">
                            <span className="text-zinc-400">Overall Status:</span>
                            <span className={`font-semibold ${getStatusColor(health.status)}`}>
                                {health.status.toUpperCase()}
                            </span>
                        </div>
                        <div className="flex justify-between items-center">
                            <span className="text-zinc-400">Database:</span>
                            <span className={getStatusColor(health.components.database)}>
                                {health.components.database}
                            </span>
                        </div>
                        <div className="flex justify-between items-center">
                            <span className="text-zinc-400">NATS:</span>
                            <span className={getStatusColor(health.components.nats)}>
                                {health.components.nats}
                            </span>
                        </div>
                        <div className="flex justify-between items-center">
                            <span className="text-zinc-400">API:</span>
                            <span className={getStatusColor(health.components.api)}>
                                {health.components.api}
                            </span>
                        </div>
                        <div className="flex justify-between items-center">
                            <span className="text-zinc-400">Ollama:</span>
                            <span className={getStatusColor(health.components.ollama)}>
                                {health.components.ollama}
                            </span>
                        </div>
                    </div>
                ) : (
                    <p className="text-zinc-500">Loading health status...</p>
                )}
            </div>

            {/* Configuration Form */}
            <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                <h2 className="text-xl font-semibold text-zinc-100 mb-4">Configuration</h2>
                {loading ? (
                    <p className="text-zinc-500">Loading configuration...</p>
                ) : (
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-zinc-300 mb-2">
                                Ollama Base URL
                            </label>
                            <input
                                type="text"
                                value={config.ollama_base_url}
                                onChange={(e) => setConfig({ ...config, ollama_base_url: e.target.value })}
                                className="w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-zinc-100 focus:outline-none focus:ring-2 focus:ring-zinc-600"
                                placeholder="http://192.168.50.156:11434"
                            />
                            <p className="text-xs text-zinc-500 mt-1">
                                URL to your Ollama instance (e.g., http://host-ip:11434)
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-zinc-300 mb-2">
                                NATS URL
                            </label>
                            <input
                                type="text"
                                value={config.nats_url}
                                onChange={(e) => setConfig({ ...config, nats_url: e.target.value })}
                                className="w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-zinc-100 focus:outline-none focus:ring-2 focus:ring-zinc-600"
                                placeholder="nats://nats:4222"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-zinc-300 mb-2">
                                Database URL
                            </label>
                            <input
                                type="password"
                                value={config.database_url}
                                onChange={(e) => setConfig({ ...config, database_url: e.target.value })}
                                className="w-full px-4 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-zinc-100 focus:outline-none focus:ring-2 focus:ring-zinc-600"
                                placeholder="postgresql+asyncpg://..."
                            />
                        </div>

                        <button
                            onClick={handleSave}
                            disabled={saving}
                            className="bg-zinc-700 text-zinc-100 px-6 py-2 rounded-lg hover:bg-zinc-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {saving ? 'Saving...' : 'Save Configuration'}
                        </button>

                        {message && (
                            <p className={`text-sm ${message.includes('success') ? 'text-green-400' : 'text-red-400'}`}>
                                {message}
                            </p>
                        )}
                    </div>
                )}
            </div>

            {/* Current Environment */}
            {health && (
                <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
                    <h2 className="text-xl font-semibold text-zinc-100 mb-4">Current Environment</h2>
                    <div className="space-y-2 font-mono text-sm">
                        <p className="text-zinc-400">
                            API Endpoint: <span className="text-zinc-300">{API_BASE_URL}</span>
                        </p>
                        <p className="text-zinc-400">
                            NATS URL: <span className="text-zinc-300">{health.environment.nats_url}</span>
                        </p>
                        <p className="text-zinc-400">
                            Database: <span className="text-zinc-300">{health.environment.database_url}</span>
                        </p>
                    </div>
                </div>
            )}
        </div>
    );
}
