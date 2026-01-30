'use client';

import { motion } from 'framer-motion';
import { Activity } from 'lucide-react';

interface StatusPanelProps {
    health: {
        status: string;
        components: {
            database: string;
            nats: string;
            ollama?: string;
            api: string;
        };
    } | null;
    delay?: number;
}

export default function StatusPanel({ health, delay = 0 }: StatusPanelProps) {
    const getStatusColor = (status: string) => {
        if (status === 'connected' || status === 'healthy') return 'bg-[--accent-success]';
        if (status === 'degraded') return 'bg-[--accent-warn]';
        return 'bg-[--accent-error]';
    };

    const statusItems = [
        { label: 'API', key: 'api' as const },
        { label: 'Database', key: 'database' as const },
        { label: 'NATS', key: 'nats' as const },
        { label: 'Ollama', key: 'ollama' as const },
    ];

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay }}
            whileHover={{ y: -4 }}
            className="p-6 border border-[--border-subtle] rounded-xl bg-gradient-to-br from-[--bg-secondary] to-[--bg-primary] shadow-lg hover:shadow-[0_8px_30px_var(--accent-glow)] transition-all duration-300"
        >
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-[--text-secondary] uppercase tracking-wider">
                    System Status
                </h3>
                <Activity className="h-5 w-5 text-[--accent-info]" />
            </div>
            <div className="flex flex-col gap-3">
                {statusItems.map((item) => (
                    <div key={item.key} className="flex items-center justify-between">
                        <span className="text-[--text-secondary] text-sm font-medium">{item.label}</span>
                        <div className="flex items-center gap-2">
                            <div className={`h-2 w-2 rounded-full ${health ? getStatusColor(health.components[item.key] || 'unknown') : 'bg-[--text-secondary]'} shadow-[0_0_6px_currentColor]`}></div>
                            <span className="text-[--text-primary] text-sm font-mono">
                                {health?.components[item.key] || 'Checking...'}
                            </span>
                        </div>
                    </div>
                ))}
            </div>
        </motion.div>
    );
}
