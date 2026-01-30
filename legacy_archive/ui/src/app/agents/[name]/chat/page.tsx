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
    type: 'user' | 'agent';
}

export default function AgentChatPage() {
    const params = useParams();
    const agentName = decodeURIComponent(params.name as string);
    const [messages, setMessages] = useState<Message[]>([]);
    const [input, setInput] = useState('');
    const [agentConfig, setAgentConfig] = useState<any>(null);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    // Fetch Agent Config for Context View
    useEffect(() => {
        fetch(`${API_BASE_URL}/agents`)
            .then(res => {
                if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
                return res.json();
            })
            .then(data => {
                const agent = data.find((a: any) => a.name === agentName);
                if (agent) {
                    console.log("ChatPage: Loaded agent config:", agent);
                    setAgentConfig(agent);
                } else {
                    console.error(`Agent ${agentName} not found in config`);
                }
            })
            .catch(err => {
                console.error("Failed to fetch agent config:", err);
            });

        // Fetch Agent-Specific History
        fetch(`${API_BASE_URL}/conversations/session-${agentName}/history`)
            .then(res => {
                if (res.ok) return res.json();
                return [];
            })
            .then(data => {
                const mapped = data.map((m: any) => ({
                    id: m.id,
                    sender: m.sender === 'user' ? 'User' : agentName,
                    content: m.content,
                    timestamp: m.timestamp,
                    type: m.sender === 'user' ? 'user' : 'agent'
                }));
                setMessages(mapped);
            })
            .catch(err => console.error("Failed to load history:", err));

    }, [agentName]);

    const [isTyping, setIsTyping] = useState(false);
    const [isInspectorOpen, setIsInspectorOpen] = useState(true);
    const [inspectorState, setInspectorState] = useState<any>(null); // State for Inspector

    // Subscribe to agent's output channel (or default to chat.user.user)
    const outputChannel = agentConfig?.messaging?.outputs?.[0] || 'chat.user.user';
    const { events: streamMessages, stats } = useEventStream(outputChannel);

    useEffect(() => {
        if (!streamMessages || streamMessages.length === 0) return;

        // Listen for Debug Context
        const debugMsg = streamMessages.find(m =>
            (m.payload && m.payload.type === 'agent_context_state')
        );
        if (debugMsg) {
            setInspectorState(debugMsg.payload);
        }

        console.log("ChatPage processing stream update, count:", streamMessages.length);

        const newMessages: Message[] = [];

        streamMessages.forEach((msg) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const msgData = msg as any;

            // 1. Check if it's from this agent
            if (msgData.sender !== agentName) {
                // Ignore messages not from this agent
            } else {
                console.log(`[ChatDebug] Accepting message from ${msgData.sender}`);
            }

            // 2. Map to local Message format
            const chatMsg: Message = {
                id: msgData.id,
                sender: msgData.sender,
                content: msgData.payload?.content || msgData.content || JSON.stringify(msgData.payload),
                timestamp: msgData.timestamp ? new Date(msgData.timestamp * 1000).toISOString() : new Date().toISOString(),
                type: 'agent'
            };

            // 3. Collect if valid
            newMessages.push(chatMsg);
        });

        if (newMessages.length > 0) {
            setMessages(prev => {
                const existingIds = new Set(prev.map(m => m.id));
                const uniqueNewBatch: Message[] = [];
                const newIds = new Set<string>();

                for (const msg of newMessages) {
                    if (!existingIds.has(msg.id) && !newIds.has(msg.id)) {
                        uniqueNewBatch.push(msg);
                        newIds.add(msg.id);
                    }
                }

                if (uniqueNewBatch.length === 0) return prev;

                console.log("Adding new valid messages:", uniqueNewBatch.length);
                setIsTyping(false);

                return [...prev, ...uniqueNewBatch.reverse()];
            });
        }
    }, [streamMessages, agentName]);

    const handleClearMemory = async () => {
        if (!confirm("Are you sure? This will wipe the agent's memory and chat history.")) return;
        try {
            await fetch(`${API_BASE_URL}/agents/${agentName}/memory`, { method: 'DELETE' });
            setMessages([]);
            setInspectorState(null);
            alert("Memory Wiped!");
        } catch (e) {
            alert("Failed to wipe memory: " + e);
        }
    };

    const sendMessage = async () => {
        if (!input.trim()) return;

        const userMsg: Message = {
            id: `local-${Date.now()}`,
            sender: 'User',
            content: input,
            timestamp: new Date().toISOString(),
            type: 'user'
        };

        setMessages(prev => [...prev, userMsg]);
        setInput('');
        setIsTyping(true); // Start typing indicator

        try {
            // Determine target channel: use first input channel or default
            const targetChannel = agentConfig?.messaging?.inputs?.[0] || `chat.agent.${agentName}`;

            // Use generic ingest endpoint
            const res = await fetch(`${API_BASE_URL}/ingest/${targetChannel}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    id: userMsg.id,
                    source_agent_id: 'user',
                    type: 'event',
                    payload: {
                        content: userMsg.content,
                        intent: 'chat'
                    }
                })
            });

            if (!res.ok) {
                console.error("Failed to send message:", await res.text());
                setIsTyping(false); // Stop on error
            }
        } catch (e) {
            console.error("Network error sending message", e);
            setIsTyping(false); // Stop on error
        }
    };

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages, isTyping]);

    return (
        <div className="flex h-screen bg-gray-900 text-gray-100">
            {/* Sidebar: Context Inspector */}
            <div className={`${isInspectorOpen ? 'w-1/3' : 'w-12'} border-r border-gray-800 transition-all duration-300 flex flex-col`}>
                <div className="p-4 border-b border-gray-800 flex justify-between items-center">
                    {isInspectorOpen && <h2 className="text-xl font-bold text-emerald-400 whitespace-nowrap overflow-hidden">Context</h2>}
                    <button
                        onClick={() => setIsInspectorOpen(!isInspectorOpen)}
                        className="p-2 hover:bg-gray-800 rounded text-gray-400 hover:text-white"
                        title={isInspectorOpen ? "Collapse Sidebar" : "Expand Sidebar"}
                    >
                        {isInspectorOpen ? '«' : '»'}
                    </button>
                </div>

                {isInspectorOpen && (
                    <div className="p-6 overflow-y-auto flex-1">
                        {agentConfig ? (
                            <div className="space-y-6">
                                <div>
                                    <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">Identity</h3>
                                    <div className="mt-2 bg-gray-800 p-3 rounded text-sm">
                                        <p><span className="text-gray-400">Name:</span> {agentConfig.name}</p>
                                        <p><span className="text-gray-400">Backend:</span> {agentConfig.backend}</p>
                                        <p><span className="text-gray-400">Languages:</span> {agentConfig.languages.join(', ')}</p>
                                    </div>
                                </div>

                                <div>
                                    <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">System Prompt</h3>
                                    <div className="mt-2 bg-gray-800 p-3 rounded text-sm whitespace-pre-wrap font-mono text-xs text-gray-300">
                                        {agentConfig.prompt_config?.system_prompt || "No system prompt configured."}
                                    </div>
                                </div>

                                <div>
                                    <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">Capabilities</h3>
                                    <div className="mt-2 flex flex-wrap gap-2">
                                        {agentConfig.capabilities?.map((cap: string) => (
                                            <span key={cap} className="px-2 py-1 bg-blue-900/30 text-blue-400 text-xs rounded border border-blue-800">
                                                {cap}
                                            </span>
                                        ))}
                                    </div>
                                </div>

                                <div className="p-4 bg-yellow-900/20 border border-yellow-800/50 rounded text-xs text-yellow-200">
                                    <p>⚠️ Note: This view shows the <strong>static configuration</strong>. Real-time memory state is held within the agent process.</p>
                                </div>
                            </div>
                        ) : (
                            <p className="text-gray-500">Loading agent config...</p>
                        )}
                    </div>
                )}
            </div>

            {/* Main Chat Area */}
            <div className="flex-1 flex flex-col">
                <header className="h-16 border-b border-gray-800 flex items-center px-6 bg-gray-900/50 backdrop-blur">
                    <h1 className="text-lg font-semibold">Chat with <span className="text-emerald-400">{agentName}</span></h1>
                </header>

                <div className="flex-1 overflow-y-auto p-4 space-y-4 relative">
                    {/* DEBUG STATUS OVERLAY */}
                    <div className="flex flex-col flex-1 relative">
                        {/* Header */}
                        <div className="h-16 border-b border-token-border-light flex items-center justify-between px-6 bg-token-bg-secondary/50 backdrop-blur">
                            <div className="flex items-center gap-3">
                                <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                                <h1 className="font-display font-bold text-xl text-token-text-primary">{agentConfig?.name}</h1>
                                <span className="text-xs px-2 py-0.5 rounded-full bg-token-surface-highlight text-token-text-secondary border border-token-border-light">
                                    {agentConfig?.role}
                                </span>
                            </div>
                            <div className="flex items-center gap-4 text-xs font-mono text-token-text-tertiary">
                                <button
                                    onClick={() => setIsInspectorOpen(!isInspectorOpen)}
                                    className="hover:text-token-text-primary transition-colors flex items-center gap-1"
                                >
                                    {isInspectorOpen ? 'Hide' : 'Show'} Brain
                                </button>
                                <span title={stats.isConnected ? "Connected" : "Disconnected"} className={stats.isConnected ? "text-green-500" : "text-red-500"}>
                                    ●
                                </span>
                                <span>{stats.eventsPerSecond.toFixed(1)}/s</span>
                                <span>Total: {stats.totalEvents}</span>
                                <span>Msgs: {messages.length}</span>
                            </div>
                        </div>

                        {/* Messages */}
                        <div className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-thin scrollbar-thumb-token-border-light scrollbar-track-transparent">
                            {messages.map((msg) => (
                                <div key={msg.id} className={`flex ${msg.sender === agentName ? 'justify-start' : 'justify-end'}`}>
                                    <div className={`
                               max-w-[70%] rounded-2xl p-5 shadow-sm
                               ${msg.sender === agentName
                                            ? 'bg-token-surface-primary border border-token-border-light text-token-text-primary rounded-tl-sm'
                                            : 'bg-token-accent-primary text-white rounded-tr-sm shadow-glow-sm'}
                           `}>
                                        <div className="prose prose-invert max-w-none text-sm leading-relaxed">
                                            <p>{msg.content}</p>
                                        </div>
                                        <div className="mt-2 text-[10px] opacity-40 font-mono uppercase tracking-wider">
                                            {new Date(msg.timestamp).toLocaleTimeString()}
                                        </div>
                                    </div>
                                </div>
                            ))}
                            <div ref={messagesEndRef} />
                        </div>

                        {/* Input */}
                        <div className="p-6 bg-token-bg-primary border-t border-token-border-light">
                            <div className="max-w-4xl mx-auto relative group">
                                <input
                                    type="text"
                                    value={input}
                                    onChange={(e) => setInput(e.target.value)}
                                    onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
                                    placeholder={`Message ${agentConfig?.name}...`}
                                    className="w-full bg-token-bg-secondary border border-token-border-light rounded-xl px-4 py-4 pr-12 text-token-text-primary outline-none focus:border-token-accent-primary/50 focus:ring-1 focus:ring-token-accent-primary/50 transition-all placeholder:text-token-text-tertiary shadow-lg"
                                    disabled={isTyping}
                                />
                                <button
                                    onClick={sendMessage}
                                    disabled={isTyping || !input.trim()}
                                    className="bg-green-600 hover:bg-green-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-3 rounded-lg font-medium transition-colors"
                                >
                                    {isTyping ? '...' : 'Send'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Inspector Sidebar */}
            {isInspectorOpen && (
                <div className="w-96 border-l border-token-border-light bg-token-bg-secondary p-4 overflow-y-auto font-mono text-xs shadow-2xl z-20 transition-all">
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="font-bold text-token-text-primary">🧠 Brain Inspector</h2>
                        <button
                            onClick={handleClearMemory}
                            className="px-2 py-1 bg-red-900/20 text-red-400 border border-red-900/50 rounded hover:bg-red-900/40 transition-colors"
                        >
                            🛑 Wipe Memory
                        </button>
                    </div>

                    {inspectorState ? (
                        <div className="space-y-4">
                            <div>
                                <h3 className="text-token-text-tertiary mb-1 uppercase tracking-wider text-[10px]">Active System Prompt</h3>
                                <div className="p-2 bg-black/20 rounded border border-token-border-light/50 text-token-text-secondary whitespace-pre-wrap">
                                    {inspectorState.system_prompt}
                                </div>
                            </div>
                            <div>
                                <h3 className="text-token-text-tertiary mb-1 uppercase tracking-wider text-[10px]">Context History ({inspectorState.history?.length || 0})</h3>
                                <div className="space-y-2 max-h-[60vh] overflow-y-auto">
                                    {inspectorState.history?.map((m: any, i: number) => (
                                        <div key={i} className="p-2 bg-black/20 rounded border border-token-border-light/50">
                                            <span className={`text-[9px] uppercase font-bold ${m.role === 'assistant' ? 'text-blue-400' : 'text-green-400'}`}>
                                                {m.role}
                                            </span>
                                            <p className="text-token-text-secondary mt-1 whitespace-pre-wrap">{m.content}</p>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="text-token-text-tertiary italic p-4 text-center">
                            Waiting for agent thought process...
                            <br />
                            <span className="text-[10px] opacity-50">(Send a message to see context state)</span>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
