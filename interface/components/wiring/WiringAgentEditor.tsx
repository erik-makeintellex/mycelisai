"use client";

import React, { useState, useEffect, useCallback } from 'react';
import { X, Save, Trash2 } from 'lucide-react';
import type { AgentManifest } from '@/store/useCortexStore';
import type { MissionStatus } from '@/store/useCortexStore';

// ── Tag Input (inline chips + text input) ──────────────────────
function TagInput({
    label,
    value,
    onChange,
}: {
    label: string;
    value: string[];
    onChange: (tags: string[]) => void;
}) {
    const [input, setInput] = useState('');

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            const newTags = input
                .split(',')
                .map((t) => t.trim())
                .filter((t) => t.length > 0 && !value.includes(t));
            if (newTags.length > 0) {
                onChange([...value, ...newTags]);
            }
            setInput('');
        }
    };

    const removeTag = (tag: string) => {
        onChange(value.filter((t) => t !== tag));
    };

    return (
        <div>
            <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                {label}
            </label>
            {value.length > 0 && (
                <div className="flex flex-wrap gap-1 mb-1.5">
                    {value.map((tag) => (
                        <span
                            key={tag}
                            className="inline-flex items-center gap-1 text-[10px] font-mono px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30"
                        >
                            {tag}
                            <button
                                type="button"
                                onClick={() => removeTag(tag)}
                                className="hover:text-cortex-danger transition-colors"
                            >
                                <X className="w-2.5 h-2.5" />
                            </button>
                        </span>
                    ))}
                </div>
            )}
            <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type + Enter (comma-separated)"
                className="w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary transition-colors"
            />
        </div>
    );
}

// ── Drawer Props ───────────────────────────────────────────────
interface WiringAgentEditorProps {
    teamIdx: number;
    agentIdx: number;
    agent: AgentManifest;
    missionStatus: MissionStatus;
    onClose: () => void;
    onSave: (teamIdx: number, agentIdx: number, updates: Partial<AgentManifest>) => void;
    onDelete: (teamIdx: number, agentIdx: number) => void;
}

const ROLES = ['cognitive', 'sensory', 'actuation', 'ledger'] as const;

export default function WiringAgentEditor({
    teamIdx,
    agentIdx,
    agent,
    missionStatus,
    onClose,
    onSave,
    onDelete,
}: WiringAgentEditorProps) {
    const [agentId, setAgentId] = useState('');
    const [role, setRole] = useState<string>('cognitive');
    const [systemPrompt, setSystemPrompt] = useState('');
    const [model, setModel] = useState('');
    const [tools, setTools] = useState<string[]>([]);
    const [inputs, setInputs] = useState<string[]>([]);
    const [outputs, setOutputs] = useState<string[]>([]);

    // Pre-populate from agent
    useEffect(() => {
        setAgentId(agent.id);
        setRole(agent.role);
        setSystemPrompt(agent.system_prompt ?? '');
        setModel(agent.model ?? '');
        setTools([...(agent.tools ?? [])]);
        setInputs([...(agent.inputs ?? [])]);
        setOutputs([...(agent.outputs ?? [])]);
    }, [agent]);

    const handleSave = useCallback(() => {
        onSave(teamIdx, agentIdx, {
            id: agentId,
            role,
            system_prompt: systemPrompt || undefined,
            model: model || undefined,
            tools: tools.length > 0 ? tools : undefined,
            inputs: inputs.length > 0 ? inputs : undefined,
            outputs: outputs.length > 0 ? outputs : undefined,
        });
    }, [teamIdx, agentIdx, agentId, role, systemPrompt, model, tools, inputs, outputs, onSave]);

    const [confirmDelete, setConfirmDelete] = useState(false);

    const handleDelete = useCallback(() => {
        if (confirmDelete) {
            onDelete(teamIdx, agentIdx);
            setConfirmDelete(false);
        } else {
            setConfirmDelete(true);
            setTimeout(() => setConfirmDelete(false), 3000);
        }
    }, [confirmDelete, teamIdx, agentIdx, onDelete]);

    const isActive = missionStatus === 'active';
    const inputClasses =
        'w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary transition-colors';
    const selectClasses =
        'w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary transition-colors appearance-none';

    return (
        <div className="absolute right-0 top-0 bottom-0 w-96 z-40 bg-cortex-surface border-l border-cortex-border shadow-2xl flex flex-col">
            {/* Header */}
            <div className="h-12 flex items-center justify-between px-4 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-2">
                    <span className="text-xs font-mono font-bold text-cortex-text-main uppercase">
                        Edit Agent
                    </span>
                    {isActive && (
                        <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-success/15 text-cortex-success border border-cortex-success/30">
                            Active
                        </span>
                    )}
                </div>
                <button
                    onClick={onClose}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Form body */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {/* Agent ID */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Agent ID
                    </label>
                    <input
                        type="text"
                        value={agentId}
                        onChange={(e) => setAgentId(e.target.value)}
                        placeholder="agent-name"
                        disabled={isActive}
                        className={`${inputClasses} ${isActive ? 'opacity-50 cursor-not-allowed' : ''}`}
                    />
                    {isActive && (
                        <p className="text-[9px] font-mono text-cortex-text-muted/60 mt-0.5">
                            Agent ID is read-only for active missions
                        </p>
                    )}
                </div>

                {/* Role */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Role
                    </label>
                    <select
                        value={role}
                        onChange={(e) => setRole(e.target.value)}
                        className={selectClasses}
                    >
                        {ROLES.map((r) => (
                            <option key={r} value={r}>
                                {r}
                            </option>
                        ))}
                    </select>
                </div>

                {/* System Prompt */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        System Prompt
                    </label>
                    <textarea
                        value={systemPrompt}
                        onChange={(e) => setSystemPrompt(e.target.value)}
                        placeholder="You are a..."
                        rows={6}
                        className={`${inputClasses} resize-none`}
                    />
                </div>

                {/* Model */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Model
                    </label>
                    <input
                        type="text"
                        value={model}
                        onChange={(e) => setModel(e.target.value)}
                        placeholder="qwen2.5-coder:7b-instruct"
                        className={inputClasses}
                    />
                </div>

                {/* Tools */}
                <TagInput label="Tools" value={tools} onChange={setTools} />

                {/* Inputs (NATS topics) */}
                <TagInput label="Inputs (NATS Topics)" value={inputs} onChange={setInputs} />

                {/* Outputs (NATS topics) */}
                <TagInput label="Outputs (NATS Topics)" value={outputs} onChange={setOutputs} />
            </div>

            {/* Footer actions */}
            <div className="flex items-center gap-2 px-4 py-3 border-t border-cortex-border flex-shrink-0">
                <button
                    onClick={handleDelete}
                    className={`flex items-center justify-center gap-1.5 px-3 py-2 rounded text-xs font-mono transition-colors ${
                        confirmDelete
                            ? 'bg-cortex-danger/30 border border-cortex-danger text-cortex-danger animate-pulse'
                            : 'bg-cortex-danger/10 border border-cortex-danger/30 text-cortex-danger hover:bg-cortex-danger/20'
                    }`}
                >
                    <Trash2 className="w-3.5 h-3.5" />
                    {confirmDelete && <span>Confirm?</span>}
                </button>
                <button
                    onClick={onClose}
                    className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded text-xs font-mono bg-cortex-bg border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:border-cortex-text-muted transition-colors"
                >
                    Cancel
                </button>
                <button
                    onClick={handleSave}
                    disabled={!agentId.trim()}
                    className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded text-xs font-mono bg-cortex-primary/15 border border-cortex-primary/30 text-cortex-primary hover:bg-cortex-primary/25 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >
                    <Save className="w-3.5 h-3.5" />
                    Save
                </button>
            </div>
        </div>
    );
}
