"use client";

import React, { createContext, useContext, useEffect, useRef, useState } from "react";

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

const MAX_SIGNALS = 100;
const RECONNECT_BASE_MS = 2000;
const RECONNECT_MAX_MS = 30000;

export function SignalProvider({ children }: { children: React.ReactNode }) {
    const [isConnected, setIsConnected] = useState(false);
    const [signals, setSignals] = useState<Signal[]>([]);
    const retryRef = useRef(0);
    const esRef = useRef<EventSource | null>(null);
    const mountedRef = useRef(true);

    useEffect(() => {
        mountedRef.current = true;

        function connect() {
            if (!mountedRef.current) return;

            const eventSource = new EventSource("/api/v1/stream");
            esRef.current = eventSource;

            eventSource.onopen = () => {
                if (!mountedRef.current) return;
                retryRef.current = 0;
                setIsConnected(true);
            };

            eventSource.onmessage = (event) => {
                if (!mountedRef.current) return;
                try {
                    const data = JSON.parse(event.data);
                    setSignals(prev => [data, ...prev].slice(0, MAX_SIGNALS));
                } catch (e) {
                    console.error("Signal Parse Error", e);
                }
            };

            eventSource.onerror = () => {
                if (!mountedRef.current) return;
                setIsConnected(false);
                eventSource.close();
                esRef.current = null;

                // Exponential backoff reconnect
                const delay = Math.min(
                    RECONNECT_BASE_MS * Math.pow(2, retryRef.current),
                    RECONNECT_MAX_MS,
                );
                retryRef.current++;
                setTimeout(connect, delay);
            };
        }

        connect();

        return () => {
            mountedRef.current = false;
            if (esRef.current) {
                esRef.current.close();
                esRef.current = null;
            }
        };
    }, []);

    const clearSignals = () => setSignals([]);

    return (
        <SignalContext.Provider value={{ isConnected, signals, clearSignals }}>
            {children}
        </SignalContext.Provider>
    );
}
