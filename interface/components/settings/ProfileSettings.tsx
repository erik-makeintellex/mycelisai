"use client";

import { useEffect, useState } from "react";
import { useCortexStore } from "@/store/useCortexStore";

export default function ProfileSettings() {
    const assistantName = useCortexStore((s) => s.assistantName);
    const updateAssistantName = useCortexStore((s) => s.updateAssistantName);
    const theme = useCortexStore((s) => s.theme);
    const updateTheme = useCortexStore((s) => s.updateTheme);
    const [nameDraft, setNameDraft] = useState(assistantName);
    const [isSavingName, setIsSavingName] = useState(false);
    const [saveMessage, setSaveMessage] = useState<string | null>(null);
    const [themeDraft, setThemeDraft] = useState(theme);
    const [isSavingTheme, setIsSavingTheme] = useState(false);
    const [themeMessage, setThemeMessage] = useState<string | null>(null);

    useEffect(() => {
        setNameDraft(assistantName);
    }, [assistantName]);

    useEffect(() => {
        setThemeDraft(theme);
    }, [theme]);

    const onSaveAssistantName = async () => {
        setIsSavingName(true);
        setSaveMessage(null);
        const ok = await updateAssistantName(nameDraft);
        setIsSavingName(false);
        setSaveMessage(ok ? "Saved" : "Failed to save");
    };

    const onSaveTheme = async () => {
        setIsSavingTheme(true);
        setThemeMessage(null);
        const ok = await updateTheme(themeDraft);
        setIsSavingTheme(false);
        setThemeMessage(ok ? "Saved" : "Failed to save");
    };

    return (
        <div className="space-y-6 max-w-lg">
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Identity</h3>
                <div className="space-y-2">
                    <label className="text-cortex-text-main text-sm block" htmlFor="assistant-name">
                        Assistant Name
                    </label>
                    <div className="flex items-center gap-2">
                        <input
                            id="assistant-name"
                            value={nameDraft}
                            onChange={(e) => setNameDraft(e.target.value)}
                            placeholder="Soma"
                            maxLength={48}
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-sm text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        />
                        <button
                            type="button"
                            onClick={onSaveAssistantName}
                            disabled={isSavingName || !nameDraft.trim()}
                            className="px-3 py-1 rounded border border-cortex-primary/40 text-cortex-primary text-sm font-mono hover:bg-cortex-primary/10 disabled:opacity-50"
                        >
                            {isSavingName ? "Saving..." : "Save"}
                        </button>
                    </div>
                    <p className="text-xs text-cortex-text-muted">
                        This updates how your orchestrator name is shown across chat, status, and workflow surfaces.
                    </p>
                    {saveMessage && <p className="text-xs font-mono text-cortex-text-muted">{saveMessage}</p>}
                </div>
            </div>
            <div className="p-6 rounded-lg border border-cortex-border bg-cortex-surface shadow-sm space-y-4">
                <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">Appearance</h3>
                <div className="space-y-2">
                    <label className="text-cortex-text-main text-sm block" htmlFor="theme-select">
                        Theme
                    </label>
                    <div className="flex items-center gap-2">
                        <select
                            id="theme-select"
                            value={themeDraft}
                            onChange={(e) => setThemeDraft(e.target.value as typeof themeDraft)}
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded px-2 py-1 text-sm text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                        >
                            <option value="aero-light">Aero Light</option>
                            <option value="midnight-cortex">Midnight Cortex</option>
                            <option value="system">System</option>
                        </select>
                        <button
                            type="button"
                            onClick={onSaveTheme}
                            disabled={isSavingTheme || themeDraft === theme}
                            className="px-3 py-1 rounded border border-cortex-primary/40 text-cortex-primary text-sm font-mono hover:bg-cortex-primary/10 disabled:opacity-50"
                        >
                            {isSavingTheme ? "Saving..." : "Save"}
                        </button>
                    </div>
                    <p className="text-xs text-cortex-text-muted">
                        Pick the product surface theme for your workspace. System follows your device preference automatically.
                    </p>
                    {themeMessage && <p className="text-xs font-mono text-cortex-text-muted">{themeMessage}</p>}
                </div>
            </div>
        </div>
    );
}
