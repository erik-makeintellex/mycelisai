"use client";

import React, { createContext, useContext, useEffect, useState } from "react";

// Types matching Backend LogEntry / Signal
export interface Signal {
    type: string;
    source?: string;
    level?: string;
    message?: string;
    timestamp?: string;
    payload?: any;
    topic?: string;
}

interface SignalContextType {
    isConnected: boolean;
    signals: Signal[];
    clearSignals: () => void;
}

const SignalContext = createContext<SignalContextType>({
    isConnected: false,
    signals: [],
    clearSignals: () => { },
});

export const useSignalStream = () => useContext(SignalContext);

export function SignalProvider({ children }: { children: React.ReactNode }) {
    const [isConnected, setIsConnected] = useState(false);
    const [signals, setSignals] = useState<Signal[]>([]);

    useEffect(() => {
        const eventSource = new EventSource("/api/v1/stream");

        eventSource.onopen = () => {
            console.log("🟢 Signal Stream Connected");
            setIsConnected(true);
        };

        eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                setSignals(prev => [data, ...prev].slice(0, 100)); // Keep last 100
            } catch (e) {
                console.error("Signal Parse Error", e);
            }
        };

        eventSource.onerror = (err) => {
            console.error("🔴 Signal Stream Error", err);
            setIsConnected(false);
            eventSource.close();
        };

        return () => {
            eventSource.close();
        };
    }, []);

    const clearSignals = () => setSignals([]);

    return (
        <SignalContext.Provider value={{ isConnected, signals, clearSignals }}>
            {children}
        </SignalContext.Provider>
    );
}
