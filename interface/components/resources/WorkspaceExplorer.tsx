"use client";

import React, { useCallback, useEffect, useMemo, useState } from "react";
import { AlertTriangle, ChevronUp, File, Folder, FolderPlus, RefreshCw, Save } from "lucide-react";
import { useCortexStore, type MCPServerWithTools } from "@/store/useCortexStore";
import { extractApiError, formatMCPToolResult, type ResourceCallRequest } from "@/lib/apiContracts";

type EntryType = "file" | "dir";

interface WorkspaceEntry {
    name: string;
    path: string;
    type: EntryType;
}

function normalizePath(path: string): string {
    const raw = (path || ".").replaceAll("\\", "/").trim();
    const parts = raw.split("/").filter(Boolean);
    const stack: string[] = [];
    for (const p of parts) {
        if (p === ".") continue;
        if (p === "..") {
            stack.pop();
            continue;
        }
        stack.push(p);
    }
    return stack.length === 0 ? "." : stack.join("/");
}

function joinPath(base: string, child: string): string {
    return normalizePath(base === "." ? child : `${base}/${child}`);
}

function parseListOutput(raw: string, currentPath: string): WorkspaceEntry[] {
    const text = raw.trim();
    if (!text) return [];

    try {
        const parsed = JSON.parse(text) as unknown;
        if (Array.isArray(parsed)) {
            return parsed
                .filter((v) => typeof v === "string")
                .map((name) => ({
                    name: name as string,
                    path: joinPath(currentPath, name as string),
                    type: (name as string).endsWith("/") ? "dir" : "file",
                }));
        }
        if (parsed && typeof parsed === "object" && Array.isArray((parsed as { entries?: unknown[] }).entries)) {
            return ((parsed as { entries: unknown[] }).entries ?? [])
                .filter((v) => v && typeof v === "object")
                .map((v) => {
                    const entry = v as Record<string, unknown>;
                    const name = String(entry.name ?? "");
                    const isDir = entry.type === "directory" || entry.type === "dir" || Boolean(entry.isDirectory);
                    return {
                        name,
                        path: joinPath(currentPath, name),
                        type: isDir ? "dir" : "file",
                    } satisfies WorkspaceEntry;
                })
                .filter((e) => e.name.length > 0);
        }
    } catch {
        // Non-JSON output, parse line format.
    }

    return text
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter(Boolean)
        .map((line) => {
            const dirTagged = line.startsWith("[DIR] ");
            const fileTagged = line.startsWith("[FILE] ");
            const cleaned = line
                .replace(/^\[DIR\]\s*/, "")
                .replace(/^\[FILE\]\s*/, "")
                .replace(/\s+\(directory\)$/i, "")
                .trim();
            const type: EntryType = dirTagged || cleaned.endsWith("/") ? "dir" : fileTagged ? "file" : "file";
            const name = cleaned.endsWith("/") ? cleaned.slice(0, -1) : cleaned;
            return { name, path: joinPath(currentPath, name), type } satisfies WorkspaceEntry;
        })
        .filter((entry) => entry.name.length > 0);
}

export default function WorkspaceExplorer({ onOpenToolsTab }: { onOpenToolsTab: () => void }) {
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const isFetchingMCPServers = useCortexStore((s) => s.isFetchingMCPServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);

    const [currentPath, setCurrentPath] = useState(".");
    const [entries, setEntries] = useState<WorkspaceEntry[]>([]);
    const [selectedFile, setSelectedFile] = useState<string | null>(null);
    const [preview, setPreview] = useState("");
    const [status, setStatus] = useState<string>("Ready");
    const [busy, setBusy] = useState(false);
    const [newDir, setNewDir] = useState("");
    const [newFile, setNewFile] = useState("");
    const [newFileContent, setNewFileContent] = useState("");

    useEffect(() => {
        fetchMCPServers();
    }, [fetchMCPServers]);

    const filesystemServer = useMemo<MCPServerWithTools | null>(() => {
        const match = mcpServers.find((s) => s.name === "filesystem");
        return match ?? null;
    }, [mcpServers]);

    const canBrowse = filesystemServer?.status === "connected";

    const callTool = useCallback(
        async (toolName: string, args: Record<string, unknown>): Promise<string> => {
            if (!filesystemServer) {
                throw new Error("filesystem MCP server not installed");
            }
            const body: ResourceCallRequest<Record<string, unknown>> = { arguments: args };
            const res = await fetch(`/api/v1/mcp/servers/${filesystemServer.id}/tools/${toolName}/call`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const payload = await res.json().catch(async () => ({ error: await res.text() }));
            if (!res.ok) {
                const err = extractApiError(payload) ?? (typeof (payload as { error?: unknown }).error === "string" ? (payload as { error?: string }).error : undefined);
                throw new Error(err || "tool call failed");
            }
            return formatMCPToolResult(payload);
        },
        [filesystemServer]
    );

    const refreshList = useCallback(async () => {
        if (!canBrowse) return;
        setBusy(true);
        setStatus(`Listing ${currentPath}...`);
        try {
            const text = await callTool("list_dir", { path: currentPath });
            const parsed = parseListOutput(text, currentPath);
            parsed.sort((a, b) => {
                if (a.type !== b.type) return a.type === "dir" ? -1 : 1;
                return a.name.localeCompare(b.name);
            });
            setEntries(parsed);
            setStatus(`Loaded ${parsed.length} entries`);
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "List failed");
            setEntries([]);
        } finally {
            setBusy(false);
        }
    }, [canBrowse, currentPath, callTool]);

    useEffect(() => {
        if (canBrowse) {
            refreshList();
        }
    }, [canBrowse, currentPath, refreshList]);

    const openFile = async (path: string) => {
        setBusy(true);
        setStatus(`Reading ${path}...`);
        try {
            const text = await callTool("read_file", { path });
            setSelectedFile(path);
            setPreview(text);
            setStatus(`Opened ${path}`);
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "Read failed");
        } finally {
            setBusy(false);
        }
    };

    const createDirectory = async () => {
        if (!newDir.trim()) return;
        setBusy(true);
        try {
            const path = joinPath(currentPath, newDir.trim());
            await callTool("create_dir", { path });
            setNewDir("");
            setStatus(`Created directory ${path}`);
            refreshList();
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "Create directory failed");
        } finally {
            setBusy(false);
        }
    };

    const createFile = async () => {
        if (!newFile.trim()) return;
        setBusy(true);
        try {
            const path = joinPath(currentPath, newFile.trim());
            await callTool("write_file", { path, content: newFileContent });
            setStatus(`Created file ${path}`);
            setNewFile("");
            setNewFileContent("");
            refreshList();
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "Create file failed");
        } finally {
            setBusy(false);
        }
    };

    if (!filesystemServer) {
        return (
            <div className="h-full p-6">
                <div className="rounded-xl border border-cortex-warning/30 bg-cortex-warning/10 p-5 max-w-3xl">
                    <div className="flex items-center gap-2 text-cortex-warning mb-2">
                        <AlertTriangle className="w-4 h-4" />
                        <h3 className="text-sm font-semibold">Filesystem MCP not installed</h3>
                    </div>
                    <p className="text-xs text-cortex-text-muted mb-3">
                        Workspace Explorer requires the `filesystem` MCP server. Install/connect it from Resources / MCP Tools.
                    </p>
                    <div className="flex gap-2">
                        <button onClick={onOpenToolsTab} className="px-3 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10">
                            Open MCP Tools
                        </button>
                        <button onClick={fetchMCPServers} className="px-3 py-1.5 rounded border border-cortex-border text-cortex-text-main text-xs font-mono hover:bg-cortex-border">
                            Refresh
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    if (filesystemServer.status !== "connected") {
        return (
            <div className="h-full p-6">
                <div className="rounded-xl border border-cortex-warning/30 bg-cortex-warning/10 p-5 max-w-3xl">
                    <div className="flex items-center gap-2 text-cortex-warning mb-2">
                        <AlertTriangle className="w-4 h-4" />
                        <h3 className="text-sm font-semibold">Filesystem MCP not connected</h3>
                    </div>
                    <p className="text-xs text-cortex-text-muted mb-3">
                        Current status: <span className="font-mono">{filesystemServer.status}</span>. Connect/reinstall from MCP Tools, then retry.
                    </p>
                    <div className="flex gap-2">
                        <button onClick={onOpenToolsTab} className="px-3 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10">
                            Open MCP Tools
                        </button>
                        <button onClick={fetchMCPServers} className="px-3 py-1.5 rounded border border-cortex-border text-cortex-text-main text-xs font-mono hover:bg-cortex-border">
                            Refresh
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="h-full grid grid-cols-12 gap-4 p-6">
            <section className="col-span-5 rounded-xl border border-cortex-border bg-cortex-surface p-3 flex flex-col min-h-0">
                <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs font-mono text-cortex-text-muted">Path:</span>
                    <code className="text-xs font-mono text-cortex-primary bg-cortex-bg px-2 py-1 rounded border border-cortex-border flex-1 truncate">
                        {currentPath}
                    </code>
                </div>
                <div className="flex gap-2 mb-3">
                    <button
                        onClick={() => setCurrentPath(normalizePath(`${currentPath}/..`))}
                        className="px-2 py-1 rounded border border-cortex-border text-xs font-mono text-cortex-text-main hover:bg-cortex-border inline-flex items-center gap-1"
                    >
                        <ChevronUp className="w-3 h-3" />
                        Up
                    </button>
                    <button
                        onClick={refreshList}
                        disabled={busy || isFetchingMCPServers}
                        className="px-2 py-1 rounded border border-cortex-border text-xs font-mono text-cortex-text-main hover:bg-cortex-border inline-flex items-center gap-1"
                    >
                        <RefreshCw className={`w-3 h-3 ${(busy || isFetchingMCPServers) ? "animate-spin" : ""}`} />
                        Refresh
                    </button>
                </div>

                <div className="flex-1 overflow-y-auto rounded border border-cortex-border/60 bg-cortex-bg">
                    {entries.length === 0 ? (
                        <div className="p-3 text-xs font-mono text-cortex-text-muted">No entries</div>
                    ) : (
                        entries.map((entry) => (
                            <button
                                key={`${entry.type}:${entry.path}`}
                                onClick={() => (entry.type === "dir" ? setCurrentPath(entry.path) : openFile(entry.path))}
                                className="w-full px-3 py-2 text-left border-b last:border-b-0 border-cortex-border/40 hover:bg-cortex-surface/70 transition-colors flex items-center gap-2"
                            >
                                {entry.type === "dir" ? <Folder className="w-3.5 h-3.5 text-cortex-warning" /> : <File className="w-3.5 h-3.5 text-cortex-primary" />}
                                <span className="text-xs font-mono text-cortex-text-main truncate">{entry.name}</span>
                            </button>
                        ))
                    )}
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2">
                    <div className="flex gap-2">
                        <input
                            value={newDir}
                            onChange={(e) => setNewDir(e.target.value)}
                            placeholder="new directory name"
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs font-mono text-cortex-text-main"
                        />
                        <button onClick={createDirectory} className="px-2 py-1 rounded border border-cortex-border text-xs font-mono text-cortex-text-main hover:bg-cortex-border inline-flex items-center gap-1">
                            <FolderPlus className="w-3 h-3" />
                            Create Dir
                        </button>
                    </div>
                    <div className="flex gap-2">
                        <input
                            value={newFile}
                            onChange={(e) => setNewFile(e.target.value)}
                            placeholder="new file name"
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-xs font-mono text-cortex-text-main"
                        />
                        <button onClick={createFile} className="px-2 py-1 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10 inline-flex items-center gap-1">
                            <Save className="w-3 h-3" />
                            Write File
                        </button>
                    </div>
                </div>
            </section>

            <section className="col-span-7 rounded-xl border border-cortex-border bg-cortex-surface p-3 flex flex-col min-h-0">
                <div className="mb-2">
                    <p className="text-xs font-mono text-cortex-text-muted">Preview</p>
                    <code className="text-[11px] font-mono text-cortex-primary truncate block">
                        {selectedFile ?? "(no file selected)"}
                    </code>
                </div>
                <textarea
                    value={preview}
                    onChange={(e) => setPreview(e.target.value)}
                    placeholder="Select a file to preview contents"
                    className="flex-1 w-full bg-cortex-bg border border-cortex-border rounded p-3 text-xs font-mono text-cortex-text-main resize-none"
                />
                <textarea
                    value={newFileContent}
                    onChange={(e) => setNewFileContent(e.target.value)}
                    placeholder="Optional content for new file"
                    className="mt-3 h-24 w-full bg-cortex-bg border border-cortex-border rounded p-2 text-xs font-mono text-cortex-text-main resize-y"
                />
                <div className="mt-2 text-[11px] font-mono text-cortex-text-muted truncate">{status}</div>
            </section>
        </div>
    );
}
