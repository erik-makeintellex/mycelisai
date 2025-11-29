'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams } from 'next/navigation';
import { API_BASE_URL } from '@/config';
import { useEventStream } from '@/hooks/useEventStream';

interface Message {
    id: string;
    sender: string;
    content: string;
    timestamp: string;
    type: 'user' | 'agent' | 'system';
}

interface Team {
    id: string;
    name: string;
    inter_comm_channel?: string;
}

export default function TeamChatPage() {
    const params = useParams();
    const teamId = params.id as string;
    const [team, setTeam] = useState<Team | null>(null);
    const [messages, setMessages] = useState<Message[]>([]);
    const [input, setInput] = useState('');
    const messagesEndRef = useRef<HTMLDivElement>(null);

    // Fetch Team Details
    useEffect(() => {
        fetch(`${API_BASE_URL}/teams`)
            .then(res => res.json())
            .then(data => {
                const found = data.find((t: any) => t.id === teamId);
                setTeam(found);
            });
    }, [teamId]);

    // Subscribe to team channel
    // We listen to 'team.{id}.>' to get all events, including chat
    const channel = `team.${teamId}`;
    const { events: streamEvents } = useEventStream(channel);

    useEffect(() => {
        if (streamEvents && streamEvents.length > 0) {
            const lastEvent = streamEvents[0];

            // We only care about 'chat' intent or specific event types if we want to filter
            // For now, let's display everything that looks like a message

            try {
                // Check for duplicates
                setMessages(prev => {
                    if (prev.some(m => m.id === lastEvent.id)) return prev;

                    let content = '';
                    let type: 'user' | 'agent' | 'system' = 'agent';

                    if (typeof lastEvent.payload === 'string') {
                        content = lastEvent.payload;
                    } else if (lastEvent.payload?.content) {
                        content = lastEvent.payload.content;
                    } else {
                        content = JSON.stringify(lastEvent.payload);
                        type = 'system';
                    }

                    // If source is 'user', mark as user type
                    if (lastEvent.source === 'user') type = 'user';

                    const ts = lastEvent.timestamp ? new Date(lastEvent.timestamp * 1000) : new Date();
                    const isoTs = !isNaN(ts.getTime()) ? ts.toISOString() : new Date().toISOString();

                    return [...prev, {
                        id: lastEvent.id,
                        sender: lastEvent.source,
                        content: content,
                        timestamp: isoTs,
                        type: type
                    }];
                });
            } catch (e) {
                console.error("Error processing team event", e);
            }
        }
    }, [streamEvents]);

    const sendMessage = async () => {
        if (!input.trim() || !team) return;

        // Optimistic update
        const userMsg: Message = {
            id: `local-${Date.now()}`,
            sender: 'user',
            content: input,
            timestamp: new Date().toISOString(),
            type: 'user'
        };

        // We don't add to local state immediately because the stream will echo it back
        // But for better UX we could, and then dedup based on ID.
        // Let's rely on the stream for now to ensure it actually went through.

        try {
            // Publish to the team channel
            // We use the ingest endpoint or a specific chat endpoint if we had one for teams
            // Let's use ingest for now, targeting the team channel

            await fetch(`${API_BASE_URL}/ingest/${team.inter_comm_channel || `team.${teamId}.chat`}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    id: `msg-${Date.now()}`,
                    source_agent_id: 'user',
                    type: 'event', // or 'chat'
                    payload: { content: input, intent: 'chat' }
                })
            });

            setInput('');
        } catch (e) {
            console.error("Failed to send message", e);
        }
    };

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

    if (!team) return <div className="p-6 text-zinc-500">Loading team...</div>;

    return (
        <div className="flex flex-col h-[calc(100vh-6rem)] bg-zinc-900 rounded-xl border border-zinc-800 overflow-hidden">
            <div className="p-4 border-b border-zinc-800 bg-zinc-900/50 flex justify-between items-center">
                <div>
                    <h1 className="text-lg font-bold text-zinc-100">Team Chat: <span className="text-blue-400">{team.name}</span></h1>
                    <p className="text-xs text-zinc-500 font-mono">Channel: {team.inter_comm_channel || `team.${teamId}.chat`}</p>
                </div>
                <div className="text-xs text-zinc-500">
                    {messages.length} messages
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {messages.length === 0 && (
                    <div className="text-center text-zinc-600 mt-10 italic">
                        No messages yet. Start the conversation!
                    </div>
                )}
                {messages.map((msg) => (
                    <div key={msg.id} className={`flex ${msg.type === 'user' ? 'justify-end' : 'justify-start'}`}>
                        <div className={`max-w-[80%] px-4 py-3 rounded-2xl ${msg.type === 'user'
                            ? 'bg-blue-600 text-white rounded-br-none'
                            : msg.type === 'system'
                                ? 'bg-zinc-800 text-zinc-400 font-mono text-xs border border-zinc-700'
                                : 'bg-zinc-800 text-zinc-200 rounded-bl-none'
                            }`}>
                            <div className="flex justify-between items-baseline gap-2 mb-1 opacity-50 text-xs">
                                <span className="font-semibold">{msg.sender}</span>
                                <span>{new Date(msg.timestamp).toLocaleTimeString()}</span>
                            </div>
                            <div className="whitespace-pre-wrap break-words">{msg.content}</div>
                        </div>
                    </div>
                ))}
                <div ref={messagesEndRef} />
            </div>

            <div className="p-4 border-t border-zinc-800 bg-zinc-900">
                <div className="flex gap-2">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                        placeholder={`Message ${team.name}...`}
                        className="flex-1 bg-zinc-950 border border-zinc-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500 text-zinc-100"
                    />
                    <button
                        onClick={sendMessage}
                        disabled={!input.trim()}
                        className="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-3 rounded-lg font-medium transition-colors"
                    >
                        Send
                    </button>
                </div>
            </div>
        </div>
    );
}
