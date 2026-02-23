"use client";

import { AlertTriangle } from 'lucide-react';

interface DegradedStateProps {
    title: string;
    reason: string;
    unavailable?: string[];
    available?: string[];
    action?: string;
}

export default function DegradedState({ title, reason, unavailable, available, action }: DegradedStateProps) {
    return (
        <div className="flex flex-col items-center justify-center py-20 px-6">
            <div className="max-w-md w-full bg-cortex-surface border border-cortex-border rounded-xl p-8 text-center">
                <AlertTriangle className="w-12 h-12 text-cortex-warning mx-auto mb-4" />
                <h2 className="text-lg font-semibold text-cortex-text-main mb-2">{title}</h2>
                <p className="text-sm text-cortex-text-muted mb-4">{reason}</p>

                {unavailable && unavailable.length > 0 && (
                    <div className="text-left mb-4">
                        <p className="text-xs font-mono uppercase text-cortex-text-muted mb-1">Unavailable</p>
                        <ul className="space-y-1">
                            {unavailable.map((item) => (
                                <li key={item} className="text-xs text-cortex-danger flex items-center gap-2">
                                    <span className="w-1.5 h-1.5 rounded-full bg-cortex-danger flex-shrink-0" />
                                    {item}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}

                {available && available.length > 0 && (
                    <div className="text-left mb-4">
                        <p className="text-xs font-mono uppercase text-cortex-text-muted mb-1">Still Working</p>
                        <ul className="space-y-1">
                            {available.map((item) => (
                                <li key={item} className="text-xs text-cortex-success flex items-center gap-2">
                                    <span className="w-1.5 h-1.5 rounded-full bg-cortex-success flex-shrink-0" />
                                    {item}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}

                {action && (
                    <p className="text-xs text-cortex-primary font-medium mt-4">{action}</p>
                )}
            </div>
        </div>
    );
}
