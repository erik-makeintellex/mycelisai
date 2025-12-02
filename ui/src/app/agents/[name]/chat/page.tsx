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
                    setAgentConfig(agent);
                } else {
                    console.error(`Agent ${agentName} not found in config`);
                }
            })
            .catch(err => {
                console.error("Failed to fetch agent config:", err);
            });
    }, [agentName]);

    const [isTyping, setIsTyping] = useState(false);

    // Subscribe to agent's output channel (or default to chat.user.user)
    const outputChannel = agentConfig?.messaging?.outputs?.[0] || 'chat.user.user';
    const { events: streamMessages } = useEventStream(outputChannel);

    useEffect(() => {
        if (streamMessages && streamMessages.length > 0) {
            // Defensive check added to prevent runtime error
            // streamMessages are already parsed objects from useEventStream
            const lastMsg = streamMessages[0]; // useEventStream returns newest first!

            // Check if we already processed this message
            // We need a way to track processed messages to avoid loops if we add to 'messages' state
            // But 'messages' state is local.

            try {
                // Filter for messages from THIS agent
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const msgData = lastMsg as any;
                if (msgData.sender === agentName) {
                    setIsTyping(false); // Stop typing indicator
                    const newMsg: Message = {
                        id: msgData.id,
                        sender: msgData.sender,
                        content: msgData.payload?.content || msgData.content || JSON.stringify(msgData.payload),
                        timestamp: new Date().toISOString(),
                        type: 'agent'
                    };
                    // Avoid duplicates (simple check)
                    setMessages(prev => {
                        if (prev.some(m => m.id === newMsg.id)) return prev;
                        return [...prev, newMsg];
                    });
                }
            } catch (e) {
                console.error("Error processing message", e);
            }
        }
    }, [streamMessages, agentName]);

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
            <div className="w-1/3 border-r border-gray-800 p-6 overflow-y-auto">
                <h2 className="text-xl font-bold mb-4 text-emerald-400">Context Inspector</h2>
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

            {/* Main Chat Area */}
            <div className="flex-1 flex flex-col">
                <header className="h-16 border-b border-gray-800 flex items-center px-6 bg-gray-900/50 backdrop-blur">
                    <h1 className="text-lg font-semibold">Chat with <span className="text-emerald-400">{agentName}</span></h1>
                </header>

                <div className="flex-1 overflow-y-auto p-6 space-y-4">
                    {messages.length === 0 && (
                        <div className="text-center text-gray-600 mt-20">
                            <p>Start a conversation with {agentName}.</p>
                            <p className="text-sm">Ask about its specific task or capabilities.</p>
                        </div>
                    )}
                    {messages.map((msg) => (
                        <div key={msg.id} className={`flex ${msg.type === 'user' ? 'justify-end' : 'justify-start'}`}>
                            <div className={`max-w-2xl px-4 py-3 rounded-2xl ${msg.type === 'user'
                                ? 'bg-emerald-600 text-white rounded-br-none'
                                : 'bg-gray-800 text-gray-200 rounded-bl-none'
                                }`}>
                                <div className="text-xs opacity-50 mb-1">{msg.sender}</div>
                                <div className="whitespace-pre-wrap">{msg.content}</div>
                            </div>
                        </div>
                    ))}
                    {isTyping && (
                        <div className="flex justify-start">
                            <div className="bg-gray-800 text-gray-400 rounded-2xl rounded-bl-none px-4 py-3 flex items-center gap-2">
                                <div className="w-2 h-2 bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                                <div className="w-2 h-2 bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                                <div className="w-2 h-2 bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                            </div>
                        </div>
                    )}
                    <div ref={messagesEndRef} />
                </div>

                <div className="p-4 border-t border-gray-800 bg-gray-900">
                    <div className="flex gap-4 max-w-4xl mx-auto">
                        <input
                            type="text"
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                            placeholder={`Message ${agentName}...`}
                            disabled={isTyping}
                            className="flex-1 bg-gray-800 border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-100 disabled:opacity-50"
                        />
                        <button
                            onClick={sendMessage}
                            disabled={isTyping || !input.trim()}
                            className="bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-3 rounded-lg font-medium transition-colors"
                        >
                            {isTyping ? '...' : 'Send'}
                        </button>
                    </div>
                    {/* Debug Info */}
                    <div className="mt-4 p-2 bg-black/50 text-xs font-mono text-gray-500 overflow-x-auto">
                        <p>Agent: {agentName}</p>
                        <p>Channel: {outputChannel}</p>
                        <p>API: {API_BASE_URL}</p>
                        <p>Events: {streamMessages.length}</p>
                        <p>Last Msg: {streamMessages[0] ? JSON.stringify(streamMessages[0]) : 'None'}</p>
                    </div>
                </div>
            </div>
        </div>
    );
}
