"use client";

import React, { useState, useCallback } from "react";
import { X, Loader2, Plus } from "lucide-react";
import type { MCPServer } from "@/store/useCortexStore";

// ── Props ─────────────────────────────────────────────────────

interface MCPInstallModalProps {
    isOpen: boolean;
    onClose: () => void;
    onInstall: (config: Partial<MCPServer>) => void;
}

// ── Parsers ───────────────────────────────────────────────────

/** "a, b, c" -> ["a", "b", "c"] */
function parseArgs(raw: string): string[] {
    return raw
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);
}

/** "KEY=value\nKEY2=value2" -> { KEY: "value", KEY2: "value2" } */
function parseKeyValue(raw: string): Record<string, string> {
    const result: Record<string, string> = {};
    for (const line of raw.split("\n")) {
        const trimmed = line.trim();
        if (!trimmed) continue;
        const eqIdx = trimmed.indexOf("=");
        if (eqIdx <= 0) continue;
        const key = trimmed.slice(0, eqIdx).trim();
        const value = trimmed.slice(eqIdx + 1).trim();
        if (key) result[key] = value;
    }
    return result;
}

// ── Component ─────────────────────────────────────────────────

export default function MCPInstallModal({ isOpen, onClose, onInstall }: MCPInstallModalProps) {
    const [name, setName] = useState("");
    const [transport, setTransport] = useState<"stdio" | "sse">("stdio");
    const [command, setCommand] = useState("");
    const [args, setArgs] = useState("");
    const [env, setEnv] = useState("");
    const [url, setUrl] = useState("");
    const [headers, setHeaders] = useState("");
    const [isInstalling, setIsInstalling] = useState(false);

    const resetForm = useCallback(() => {
        setName("");
        setTransport("stdio");
        setCommand("");
        setArgs("");
        setEnv("");
        setUrl("");
        setHeaders("");
        setIsInstalling(false);
    }, []);

    function handleClose() {
        if (isInstalling) return;
        resetForm();
        onClose();
    }

    async function handleInstall() {
        if (!name.trim()) return;
        setIsInstalling(true);

        const config: Partial<MCPServer> = {
            name: name.trim(),
            transport,
        };

        if (transport === "stdio") {
            if (command.trim()) config.command = command.trim();
            const parsedArgs = parseArgs(args);
            if (parsedArgs.length > 0) config.args = parsedArgs;
            const parsedEnv = parseKeyValue(env);
            if (Object.keys(parsedEnv).length > 0) config.env = parsedEnv;
        } else {
            if (url.trim()) config.url = url.trim();
            const parsedHeaders = parseKeyValue(headers);
            if (Object.keys(parsedHeaders).length > 0) config.headers = parsedHeaders;
        }

        try {
            await onInstall(config);
            resetForm();
            onClose();
        } catch {
            // Error handled by store
        } finally {
            setIsInstalling(false);
        }
    }

    if (!isOpen) return null;

    const inputClass =
        "w-full bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 text-sm font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary/60 focus:ring-1 focus:ring-cortex-primary/30 transition-colors";

    const textareaClass =
        "w-full bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 text-sm font-mono text-cortex-text-main placeholder:text-cortex-text-muted/50 focus:outline-none focus:border-cortex-primary/60 focus:ring-1 focus:ring-cortex-primary/30 transition-colors resize-none";

    return (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
            <div className="bg-cortex-surface rounded-xl border border-cortex-border shadow-2xl w-full max-w-lg">
                {/* Modal Header */}
                <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-4 rounded-t-xl">
                    <div className="flex items-center gap-2">
                        <Plus className="w-4 h-4 text-cortex-success" />
                        <span className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                            Install MCP Server
                        </span>
                    </div>
                    <button
                        onClick={handleClose}
                        disabled={isInstalling}
                        className="p-1 rounded-md hover:bg-cortex-border/50 text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50"
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>

                {/* Form Body */}
                <div className="p-5 space-y-4">
                    {/* Name */}
                    <div>
                        <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                            Server Name <span className="text-cortex-danger">*</span>
                        </label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="e.g. filesystem-server"
                            className={inputClass}
                            autoFocus
                        />
                    </div>

                    {/* Transport */}
                    <div>
                        <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                            Transport
                        </label>
                        <div className="flex gap-3">
                            {(["stdio", "sse"] as const).map((t) => (
                                <label
                                    key={t}
                                    className={`flex items-center gap-2 px-3 py-2 rounded-lg border cursor-pointer transition-all ${
                                        transport === t
                                            ? "bg-cortex-primary/10 border-cortex-primary/40 text-cortex-primary"
                                            : "bg-cortex-bg border-cortex-border text-cortex-text-muted hover:border-cortex-text-muted/50"
                                    }`}
                                >
                                    <input
                                        type="radio"
                                        name="transport"
                                        value={t}
                                        checked={transport === t}
                                        onChange={() => setTransport(t)}
                                        className="sr-only"
                                    />
                                    <span className={`w-3 h-3 rounded-full border-2 flex items-center justify-center ${
                                        transport === t
                                            ? "border-cortex-primary"
                                            : "border-cortex-border"
                                    }`}>
                                        {transport === t && (
                                            <span className="w-1.5 h-1.5 rounded-full bg-cortex-primary" />
                                        )}
                                    </span>
                                    <span className="text-xs font-mono font-bold uppercase">{t}</span>
                                </label>
                            ))}
                        </div>
                    </div>

                    {/* Conditional Fields: stdio */}
                    {transport === "stdio" && (
                        <>
                            <div>
                                <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                                    Command
                                </label>
                                <input
                                    type="text"
                                    value={command}
                                    onChange={(e) => setCommand(e.target.value)}
                                    placeholder="e.g. npx -y @modelcontextprotocol/server-fs"
                                    className={inputClass}
                                />
                            </div>
                            <div>
                                <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                                    Arguments <span className="text-cortex-text-muted/50 normal-case">(comma-separated)</span>
                                </label>
                                <input
                                    type="text"
                                    value={args}
                                    onChange={(e) => setArgs(e.target.value)}
                                    placeholder="e.g. /home/user/docs, --verbose"
                                    className={inputClass}
                                />
                            </div>
                            <div>
                                <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                                    Environment <span className="text-cortex-text-muted/50 normal-case">(KEY=value per line)</span>
                                </label>
                                <textarea
                                    value={env}
                                    onChange={(e) => setEnv(e.target.value)}
                                    placeholder={"NODE_ENV=production\nAPI_KEY=sk-..."}
                                    rows={3}
                                    className={textareaClass}
                                />
                            </div>
                        </>
                    )}

                    {/* Conditional Fields: sse */}
                    {transport === "sse" && (
                        <>
                            <div>
                                <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                                    URL
                                </label>
                                <input
                                    type="text"
                                    value={url}
                                    onChange={(e) => setUrl(e.target.value)}
                                    placeholder="e.g. http://localhost:3001/sse"
                                    className={inputClass}
                                />
                            </div>
                            <div>
                                <label className="block text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-1.5">
                                    Headers <span className="text-cortex-text-muted/50 normal-case">(KEY=value per line)</span>
                                </label>
                                <textarea
                                    value={headers}
                                    onChange={(e) => setHeaders(e.target.value)}
                                    placeholder={"Authorization=Bearer sk-...\nX-Custom=value"}
                                    rows={3}
                                    className={textareaClass}
                                />
                            </div>
                        </>
                    )}
                </div>

                {/* Footer Actions */}
                <div className="px-5 pb-5 flex items-center justify-end gap-3">
                    <button
                        onClick={handleClose}
                        disabled={isInstalling}
                        className="px-4 py-2 rounded-lg border border-cortex-border text-xs font-mono font-bold text-cortex-text-muted hover:bg-cortex-border/30 transition-colors disabled:opacity-50"
                    >
                        CANCEL
                    </button>
                    <button
                        onClick={handleInstall}
                        disabled={!name.trim() || isInstalling}
                        className="px-4 py-2 rounded-lg bg-cortex-success/10 border border-cortex-success/30 text-xs font-mono font-bold text-cortex-success hover:bg-cortex-success/20 transition-colors disabled:opacity-40 flex items-center gap-2"
                    >
                        {isInstalling ? (
                            <>
                                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                INSTALLING...
                            </>
                        ) : (
                            <>
                                <Plus className="w-3.5 h-3.5" />
                                INSTALL
                            </>
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
}
