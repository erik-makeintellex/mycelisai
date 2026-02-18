"use client";

import React, { useState, useEffect, useCallback } from 'react';
import { X, Save } from 'lucide-react';
import type { CatalogueAgent } from '@/store/useCortexStore';

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
            {/* Rendered chips */}
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
            {/* Text input */}
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
interface AgentEditorDrawerProps {
    agent: CatalogueAgent | null;
    onClose: () => void;
    onSave: (data: Partial<CatalogueAgent>) => void;
}

const ROLES = ['cognitive', 'sensory', 'actuation', 'ledger'] as const;
const VERIFICATION_STRATEGIES = ['semantic', 'empirical', 'none'] as const;

export default function AgentEditorDrawer({ agent, onClose, onSave }: AgentEditorDrawerProps) {
    const [name, setName] = useState('');
    const [role, setRole] = useState<string>('cognitive');
    const [systemPrompt, setSystemPrompt] = useState('');
    const [model, setModel] = useState('');
    const [tools, setTools] = useState<string[]>([]);
    const [inputs, setInputs] = useState<string[]>([]);
    const [outputs, setOutputs] = useState<string[]>([]);
    const [verificationStrategy, setVerificationStrategy] = useState<string>('none');
    const [verificationRubric, setVerificationRubric] = useState('');
    const [validationCommand, setValidationCommand] = useState('');

    // Pre-populate in edit mode
    useEffect(() => {
        if (agent) {
            setName(agent.name);
            setRole(agent.role);
            setSystemPrompt(agent.system_prompt ?? '');
            setModel(agent.model ?? '');
            setTools([...agent.tools]);
            setInputs([...agent.inputs]);
            setOutputs([...agent.outputs]);
            setVerificationStrategy(agent.verification_strategy ?? 'none');
            setVerificationRubric(agent.verification_rubric.join(', '));
            setValidationCommand(agent.validation_command ?? '');
        } else {
            setName('');
            setRole('cognitive');
            setSystemPrompt('');
            setModel('');
            setTools([]);
            setInputs([]);
            setOutputs([]);
            setVerificationStrategy('none');
            setVerificationRubric('');
            setValidationCommand('');
        }
    }, [agent]);

    const handleSave = useCallback(() => {
        const rubricArray = verificationRubric
            .split(',')
            .map((s) => s.trim())
            .filter((s) => s.length > 0);

        onSave({
            name,
            role,
            system_prompt: systemPrompt || undefined,
            model: model || undefined,
            tools,
            inputs,
            outputs,
            verification_strategy: verificationStrategy || undefined,
            verification_rubric: rubricArray,
            validation_command: validationCommand || undefined,
        });
    }, [name, role, systemPrompt, model, tools, inputs, outputs, verificationStrategy, verificationRubric, validationCommand, onSave]);

    const isEditing = agent !== null;
    const inputClasses =
        'w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary transition-colors';
    const selectClasses =
        'w-full bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary transition-colors appearance-none';

    return (
        <div className="absolute right-0 top-0 bottom-0 w-96 z-40 bg-cortex-surface border-l border-cortex-border shadow-2xl flex flex-col">
            {/* Header */}
            <div className="h-12 flex items-center justify-between px-4 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <span className="text-xs font-mono font-bold text-cortex-text-main uppercase">
                    {isEditing ? 'Edit Agent' : 'New Agent'}
                </span>
                <button
                    onClick={onClose}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Form body */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {/* Name */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Name
                    </label>
                    <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="agent-name"
                        className={inputClasses}
                    />
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
                        rows={4}
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

                {/* Inputs */}
                <TagInput label="Inputs" value={inputs} onChange={setInputs} />

                {/* Outputs */}
                <TagInput label="Outputs" value={outputs} onChange={setOutputs} />

                {/* Verification Strategy */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Verification Strategy
                    </label>
                    <select
                        value={verificationStrategy}
                        onChange={(e) => setVerificationStrategy(e.target.value)}
                        className={selectClasses}
                    >
                        {VERIFICATION_STRATEGIES.map((s) => (
                            <option key={s} value={s}>
                                {s}
                            </option>
                        ))}
                    </select>
                </div>

                {/* Verification Rubric */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Verification Rubric
                    </label>
                    <textarea
                        value={verificationRubric}
                        onChange={(e) => setVerificationRubric(e.target.value)}
                        placeholder="criteria-1, criteria-2, criteria-3"
                        rows={2}
                        className={`${inputClasses} resize-none`}
                    />
                    <p className="text-[9px] font-mono text-cortex-text-muted/60 mt-0.5">
                        Comma-separated criteria
                    </p>
                </div>

                {/* Validation Command */}
                <div>
                    <label className="block text-[10px] font-mono uppercase text-cortex-text-muted mb-1">
                        Validation Command
                    </label>
                    <input
                        type="text"
                        value={validationCommand}
                        onChange={(e) => setValidationCommand(e.target.value)}
                        placeholder="go test ./..."
                        className={inputClasses}
                    />
                </div>
            </div>

            {/* Footer actions */}
            <div className="flex items-center gap-2 px-4 py-3 border-t border-cortex-border flex-shrink-0">
                <button
                    onClick={onClose}
                    className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded text-xs font-mono bg-cortex-bg border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:border-cortex-text-muted transition-colors"
                >
                    Cancel
                </button>
                <button
                    onClick={handleSave}
                    disabled={!name.trim()}
                    className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded text-xs font-mono bg-cortex-primary/15 border border-cortex-primary/30 text-cortex-primary hover:bg-cortex-primary/25 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >
                    <Save className="w-3.5 h-3.5" />
                    {isEditing ? 'Update' : 'Create'}
                </button>
            </div>
        </div>
    );
}
