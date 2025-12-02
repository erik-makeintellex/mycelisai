import { useState, useEffect, useRef } from 'react';
import { API_BASE_URL } from '@/config';

interface EventMessage {
    id: string;
    type: 'event';
    event_type: string;
    payload: any;
    source: string;
    timestamp: number;
}

interface StreamStats {
    eventsPerSecond: number;
    totalEvents: number;
    isConnected: boolean;
}

export function useEventStream(channel: string) {
    const [events, setEvents] = useState<EventMessage[]>([]);
    const [stats, setStats] = useState<StreamStats>({
        eventsPerSecond: 0,
        totalEvents: 0,
        isConnected: false,
    });

    useEffect(() => {
        const eventSource = new EventSource(`${API_BASE_URL}/stream/${channel}`);

        eventSource.onopen = () => {
            setStats(prev => ({ ...prev, isConnected: true }));
        };

        eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                const newEvent: EventMessage = {
                    ...data,
                    timestamp: data.timestamp || Date.now() / 1000
                };

                setEvents(prev => {
                    const updated = [newEvent, ...prev].slice(0, 50); // Keep last 50 events
                    return updated;
                });

                setStats(prev => ({
                    ...prev,
                    totalEvents: prev.totalEvents + 1
                }));

            } catch (err) {
                // eslint-disable-next-line no-console
                console.error('Error parsing event data:', err);
            }
        };

        eventSource.onerror = (err) => {
            // eslint-disable-next-line no-console
            console.warn('EventSource connection issue (retrying...):', err);
            setStats(prev => ({ ...prev, isConnected: false }));
            // Do NOT close; let EventSource auto-retry
            // eventSource.close(); 
        };

        // Calculate events per second
        const interval = setInterval(() => {
            setStats(prev => {
                // This is a simple approximation. For true rate, we'd need a sliding window.
                // For now, let's just reset a counter or something? 
                // Actually, let's just calculate it based on recent events in the last second.
                // But we don't have the history easily accessible in this closure without refs.
                return prev;
            });
        }, 1000);

        return () => {
            eventSource.close();
            clearInterval(interval);
        };
    }, [channel]);

    // Separate effect for rate calculation to avoid complex dependency chains
    useEffect(() => {
        const interval = setInterval(() => {
            const now = Date.now() / 1000;
            const recentEvents = events.filter(e => e.timestamp > now - 1);
            setStats(prev => ({
                ...prev,
                eventsPerSecond: recentEvents.length
            }));
        }, 1000);
        return () => clearInterval(interval);
    }, [events]);

    return { events, stats };
}
