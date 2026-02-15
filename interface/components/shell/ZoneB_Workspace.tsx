import React from 'react';

export function ZoneB({ children }: { children: React.ReactNode }) {
    return (
        <div className="flex-1 flex flex-col relative overflow-hidden bg-cortex-bg min-w-0">
            {/* Workspace Canvas */}
            <div className="flex-1 overflow-y-auto p-0 scrollbar-thin scrollbar-thumb-cortex-border relative">
                {children}
            </div>
        </div>
    );
}
