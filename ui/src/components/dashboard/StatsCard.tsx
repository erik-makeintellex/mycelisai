'use client';

import { motion } from 'framer-motion';
import { LucideIcon } from 'lucide-react';

interface StatsCardProps {
    title: string;
    value: string | number;
    icon?: LucideIcon;
    delay?: number;
}

export default function StatsCard({ title, value, icon: Icon, delay = 0 }: StatsCardProps) {
    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay }}
            whileHover={{ y: -4 }}
            className="p-6 border border-[--border-subtle] rounded-xl bg-gradient-to-br from-[--bg-secondary] to-[--bg-primary] shadow-lg hover:shadow-[0_8px_30px_var(--accent-glow)] transition-all duration-300"
        >
            <div className="flex items-center justify-between mb-2">
                <h3 className="text-sm font-medium text-[--text-secondary] uppercase tracking-wider">
                    {title}
                </h3>
                {Icon && <Icon className="h-5 w-5 text-[--accent-info]" />}
            </div>
            <p className="text-4xl font-bold text-[--text-primary] font-mono tracking-tight">
                {value}
            </p>
        </motion.div>
    );
}
