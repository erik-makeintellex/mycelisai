"use client";

import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useCortexStore, type MCPServerWithTools } from "@/store/useCortexStore";
import { extractApiError, formatMCPToolResult, type ResourceCallRequest } from "@/lib/apiContracts";
import WorkspaceFolderAccessCard from "./WorkspaceFolderAccessCard";
import WorkspaceMCPRecoveryCard from "./WorkspaceMCPRecoveryCard";
import {
    WorkspaceBrowsePane,
    WorkspaceCreatePane,
    WorkspacePaneTabs,
    WorkspacePreviewPane,
} from "./WorkspaceExplorerPanes";
import { joinPath, normalizePath, parseListOutput } from "./WorkspaceExplorerUtils";

const WORKSPACE_ROOT_PATH = "workspace";
export type WorkspaceEntry = { name: string; path: string; type: "file" | "dir" };
export type WorkspacePane = "browse" | "preview" | "create";

export default function WorkspaceExplorer({ onOpenToolsTab }: { onOpenToolsTab: () => void }) {
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const isFetchingMCPServers = useCortexStore((s) => s.isFetchingMCPServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);

    const [currentPath, setCurrentPath] = useState(WORKSPACE_ROOT_PATH);
    const [entries, setEntries] = useState<WorkspaceEntry[]>([]);
    const [selectedFile, setSelectedFile] = useState<string | null>(null);
    const [preview, setPreview] = useState("");
    const [status, setStatus] = useState<string>("Ready");
    const [busy, setBusy] = useState(false);
    const [newDir, setNewDir] = useState("");
    const [newFile, setNewFile] = useState("");
    const [newFileContent, setNewFileContent] = useState("");
    const [activePane, setActivePane] = useState<WorkspacePane>("browse");

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
            const text = await callTool("list_directory", { path: currentPath });
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
            const text = await callTool("read_text_file", { path });
            setSelectedFile(path);
            setPreview(text);
            setActivePane("preview");
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
            await callTool("create_directory", { path });
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
            <WorkspaceMCPRecoveryCard
                title="Filesystem MCP not installed"
                detail="Output Files needs the filesystem connected tool before Mycelis can browse generated files here. Install it from Connected Tools, or view storage roots to confirm where generated content is mounted."
                onOpenToolsTab={onOpenToolsTab}
                onRefresh={fetchMCPServers}
            />
        );
    }

    if (filesystemServer.status !== "connected") {
        return (
            <WorkspaceMCPRecoveryCard
                title="Filesystem MCP not connected"
                detail={(
                    <>
                        Current status: <span className="font-mono">{filesystemServer.status}</span>. Reconnect it
                        from Connected Tools, then retry. View storage roots if you need to find generated output while
                        the MCP server is recovering.
                    </>
                )}
                onOpenToolsTab={onOpenToolsTab}
                onRefresh={fetchMCPServers}
            />
        );
    }

    return (
        <div className="flex h-full min-h-0 flex-col gap-3 p-4 sm:p-6">
            <WorkspaceFolderAccessCard currentPath={currentPath} onStatus={setStatus} />

            <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface">
                <div className="border-b border-cortex-border bg-cortex-bg/70 px-3 py-3">
                    <div className="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                        <div className="min-w-0">
                            <p className="text-xs font-semibold text-cortex-text-main">Workspace output access</p>
                            <p className="mt-1 text-[11px] leading-5 text-cortex-text-muted">
                                Start with retained files, then open a preview or create a small handoff artifact when needed.
                            </p>
                        </div>
                        <div className="flex min-w-0 items-center gap-2">
                            <span className="text-[11px] font-mono text-cortex-text-muted">Path</span>
                            <code className="max-w-full truncate rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-xs font-mono text-cortex-primary md:max-w-[22rem]">
                                {currentPath}
                            </code>
                        </div>
                    </div>

                    <WorkspacePaneTabs activePane={activePane} onSelect={setActivePane} />
                </div>

                <div className="min-h-0 flex-1 overflow-y-auto p-3">
                    {activePane === "browse" && (
                        <WorkspaceBrowsePane
                            currentPath={currentPath}
                            entries={entries}
                            busy={busy}
                            isFetchingMCPServers={isFetchingMCPServers}
                            onUp={() =>
                                setCurrentPath(
                                    currentPath === WORKSPACE_ROOT_PATH
                                        ? WORKSPACE_ROOT_PATH
                                        : normalizePath(`${currentPath}/..`),
                                )
                            }
                            onRefresh={refreshList}
                            onOpenFolder={setCurrentPath}
                            onOpenFile={openFile}
                        />
                    )}

                    {activePane === "preview" && (
                        <WorkspacePreviewPane
                            selectedFile={selectedFile}
                            preview={preview}
                            onPreviewChange={setPreview}
                        />
                    )}

                    {activePane === "create" && (
                        <WorkspaceCreatePane
                            newDir={newDir}
                            newFile={newFile}
                            newFileContent={newFileContent}
                            onNewDirChange={setNewDir}
                            onNewFileChange={setNewFile}
                            onNewFileContentChange={setNewFileContent}
                            onCreateDirectory={createDirectory}
                            onCreateFile={createFile}
                        />
                    )}
                </div>

                <div className="border-t border-cortex-border bg-cortex-bg px-3 py-2 text-[11px] font-mono text-cortex-text-muted">
                    <span className="block truncate">{status}</span>
                </div>
            </section>
        </div>
    );
}
