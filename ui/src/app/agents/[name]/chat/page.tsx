'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams } from 'next/navigation';
import { API_BASE_URL } from '@/config';
import useEventStream from '@/hooks/useEventStream';

interface Message {
    id: string;
    sender: string;
    content: string;
    timestamp: string;
    type: 'user' | 'agent';
}

export default function AgentChatPage() {
    const params = useParams();
    const agentName = params.name as string;
    const [messages, setMessages] = useState<Message[]>([]);
    const [input, setInput] = useState('');
    const [agentConfig, setAgentConfig] = useState<any>(null);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    // Fetch Agent Config for Context View
    useEffect(() => {
        fetch(`${API_BASE_URL}/agents`)
            .then(res => res.json())
            .then(data => {
                const agent = data.find((a: any) => a.name === agentName);
                setAgentConfig(agent);
            });
    }, [agentName]);

    // Subscribe to chat stream (listening for replies to 'user')
    // In a real app, we'd use a unique session ID. Here we use 'user'.
    const { messages: streamMessages } = useEventStream('chat.user.user');

    useEffect(() => {
        if (streamMessages.length > 0) {
            const lastMsg = streamMessages[streamMessages.length - 1];
            try {
                const data = JSON.parse(lastMsg.data);
                // Filter for messages from THIS agent
                if (data.sender === agentName) {
                    const newMsg: Message = {
                        id: data.id,
                        sender: data.sender,
                        content: data.content,
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
                console.error("Error parsing message", e);
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

        try {
            await fetch(`${API_BASE_URL}/agents/${agentName}/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content: userMsg.content, id: userMsg.id, source_agent_id: 'user', type: 'text' })
            });
        } catch (e) {
            console.error("Failed to send message", e);
        }
    };

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

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
                            className="flex-1 bg-gray-800 border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-100"
                        />
                        <button
                            onClick={sendMessage}
                            className="bg-emerald-600 hover:bg-emerald-500 text-white px-6 py-3 rounded-lg font-medium transition-colors"
                        >
                            Send
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
