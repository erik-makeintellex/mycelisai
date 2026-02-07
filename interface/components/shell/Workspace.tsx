import React from 'react';

export function Workspace({ children }: { children: React.ReactNode }) {
    return (
        <div className="flex-1 flex flex-col relative overflow-hidden bg-white min-w-0">
            {/* Workspace Canvas */}
            <div className="flex-1 overflow-y-auto p-0 scrollbar-thin scrollbar-thumb-zinc-200 relative">
                {children}
            </div>
        </div>
    );
}
