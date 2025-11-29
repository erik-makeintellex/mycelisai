'use client';

import { useState, useEffect } from 'react';
import { API_BASE_URL } from '@/config';

interface Service {
    id: string;
    name: string;
    type: string;
    config: Record<string, any>;
    status: string;
    description?: string;
}

export default function ServicesPage() {
    const [services, setServices] = useState<Service[]>([]);
    const [showAddModal, setShowAddModal] = useState(false);
    const [newService, setNewService] = useState({
        id: '',
        name: '',
        type: 'iot_device',
        config: '{}',
        description: ''
    });

    const fetchServices = async () => {
        try {
            const res = await fetch(`${API_BASE_URL}/services`);
            const data = await res.json();
            setServices(data);
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    useEffect(() => {
        fetchServices();
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            let parsedConfig = {};
            try {
                parsedConfig = JSON.parse(newService.config);
            } catch (e) {
                alert('Invalid JSON for Config');
                return;
            }

            const body = {
                ...newService,
                config: parsedConfig
            };

            const res = await fetch(`${API_BASE_URL}/services`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                setShowAddModal(false);
                setNewService({ id: '', name: '', type: 'iot_device', config: '{}', description: '' });
                fetchServices();
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this service?')) return;
        try {
            const res = await fetch(`${API_BASE_URL}/services/${id}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                fetchServices();
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error(error);
        }
    };

    return (
        <div className="space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-zinc-100">Service Registry</h1>
                    <p className="text-zinc-400 mt-2">
                        Register external services, IoT devices, and APIs.
                    </p>
                </div>
                <button
                    onClick={() => setShowAddModal(true)}
                    className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold shadow-lg shadow-zinc-900/20"
                >
                    + Register Service
                </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {services.length === 0 ? (
                    <div className="col-span-full p-8 text-center border border-zinc-700 rounded-xl bg-zinc-800/50">
                        <p className="text-zinc-500 italic">No services registered yet.</p>
                    </div>
                ) : (
                    services.map(service => (
                        <div key={service.id} className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg group relative">
                            <button
                                onClick={() => handleDelete(service.id)}
                                className="absolute top-4 right-4 text-zinc-600 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                            >
                                Delete
                            </button>
                            <div className="flex items-center gap-3 mb-4">
                                <div className={`w-2 h-2 rounded-full ${service.status === 'active' ? 'bg-emerald-500' : 'bg-zinc-500'}`}></div>
                                <span className="text-xs font-mono text-zinc-500 uppercase">{service.type}</span>
                            </div>
                            <h3 className="text-xl font-bold text-zinc-100 mb-2">{service.name}</h3>
                            <p className="text-sm text-zinc-400 mb-4 line-clamp-2">{service.description || 'No description'}</p>

                            <div className="bg-zinc-950 rounded p-3 font-mono text-xs text-zinc-500 overflow-hidden">
                                <div className="mb-1 text-zinc-600">ID: {service.id}</div>
                                <div className="truncate">Config: {JSON.stringify(service.config)}</div>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {showAddModal && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                    <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-[500px] shadow-2xl">
                        <h3 className="text-lg font-bold text-zinc-100 mb-4">Register New Service</h3>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Service Name</label>
                                <input
                                    type="text"
                                    value={newService.name}
                                    onChange={(e) => {
                                        const name = e.target.value;
                                        setNewService({
                                            ...newService,
                                            name,
                                            id: name.toLowerCase().replace(/\s+/g, '-')
                                        });
                                    }}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    required
                                    placeholder="e.g., Temperature Sensor A"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Service ID</label>
                                <input
                                    type="text"
                                    value={newService.id}
                                    onChange={(e) => setNewService({ ...newService, id: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none font-mono text-sm"
                                    required
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Type</label>
                                <select
                                    value={newService.type}
                                    onChange={(e) => setNewService({ ...newService, type: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                >
                                    <option value="iot_device">IoT Device</option>
                                    <option value="api">External API</option>
                                    <option value="database">Database</option>
                                    <option value="other">Other</option>
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Description</label>
                                <textarea
                                    value={newService.description}
                                    onChange={(e) => setNewService({ ...newService, description: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none"
                                    rows={2}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1 text-zinc-400">Configuration (JSON)</label>
                                <textarea
                                    value={newService.config}
                                    onChange={(e) => setNewService({ ...newService, config: e.target.value })}
                                    className="w-full p-2 border border-zinc-700 rounded-lg bg-zinc-950 text-zinc-200 focus:ring-2 focus:ring-zinc-600 outline-none font-mono text-sm h-24"
                                    placeholder='{"topic": "sensors/temp", "protocol": "mqtt"}'
                                />
                                <p className="text-xs text-zinc-500 mt-1">Connection details for this service.</p>
                            </div>
                            <div className="flex justify-end gap-2 mt-6">
                                <button
                                    type="button"
                                    onClick={() => setShowAddModal(false)}
                                    className="px-4 py-2 text-zinc-400 hover:text-zinc-200 transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    className="bg-zinc-100 text-zinc-900 px-4 py-2 rounded-lg hover:bg-white transition-colors font-semibold"
                                >
                                    Register
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
