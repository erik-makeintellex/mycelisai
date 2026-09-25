"use client";

import { useState, useCallback } from "react";
import { Loader2, Zap } from "lucide-react";
import type { TriggerRuleCreate } from "@/store/useCortexStore";

// Event types available for triggering (from protocol/events.go)
const EVENT_PATTERNS = [
    { value: "mission.completed", label: "Mission Completed" },
    { value: "mission.failed", label: "Mission Failed" },
    { value: "mission.started", label: "Mission Started" },
    { value: "tool.completed", label: "Tool Completed" },
    { value: "tool.failed", label: "Tool Failed" },
    { value: "artifact.created", label: "Artifact Created" },
    { value: "memory.stored", label: "Memory Stored" },
    { value: "team.spawned", label: "Team Spawned" },
    { value: "team.stopped", label: "Team Stopped" },
    { value: "agent.started", label: "Agent Started" },
    { value: "agent.stopped", label: "Agent Stopped" },
];

/** New-trigger-rule creation form, extracted from TriggerRulesTab to keep that file under its line cap. */
export default function CreateRuleForm({
    onCreate,
    onCancel,
}: {
    onCreate: (rule: TriggerRuleCreate) => Promise<void>;
    onCancel: () => void;
}) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [eventPattern, setEventPattern] = useState("mission.completed");
    const [targetMissionId, setTargetMissionId] = useState("");
    const [mode, setMode] = useState<"propose" | "auto_execute">("propose");
    const [cooldown, setCooldown] = useState(60);
    const [maxDepth, setMaxDepth] = useState(5);
    const [maxActiveRuns, setMaxActiveRuns] = useState(3);
    const [saving, setSaving] = useState(false);

    const handleSubmit = useCallback(async () => {
        if (!name.trim() || !targetMissionId.trim()) return;
        setSaving(true);
        await onCreate({
            name: name.trim(),
            description: description.trim() || undefined,
            event_pattern: eventPattern,
            target_mission_id: targetMissionId.trim(),
            mode,
            cooldown_seconds: cooldown,
            max_depth: maxDepth,
            max_active_runs: maxActiveRuns,
            is_active: true,
        });
        setSaving(false);
    }, [name, description, eventPattern, targetMissionId, mode, cooldown, maxDepth, maxActiveRuns, onCreate]);

    return (
        <div className="rounded-lg border border-cortex-primary/30 bg-cortex-surface p-4 space-y-3">
            <h3 className="text-xs font-semibold text-cortex-primary uppercase tracking-wider">
                New Trigger Rule
            </h3>

            {/* Name */}
            <div>
                <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                    Name
                </label>
                <input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. Auto-archive on completion"
                    className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                />
            </div>

            {/* Description */}
            <div>
                <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                    Description (optional)
                </label>
                <input
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="What this rule does"
                    className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                />
            </div>

            <div className="grid grid-cols-2 gap-3">
                {/* Event Pattern */}
                <div>
                    <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                        When (event)
                    </label>
                    <select
                        value={eventPattern}
                        onChange={(e) => setEventPattern(e.target.value)}
                        className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    >
                        {EVENT_PATTERNS.map((ep) => (
                            <option key={ep.value} value={ep.value}>
                                {ep.label}
                            </option>
                        ))}
                    </select>
                </div>

                {/* Target Mission */}
                <div>
                    <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                        Target Mission ID
                    </label>
                    <input
                        value={targetMissionId}
                        onChange={(e) => setTargetMissionId(e.target.value)}
                        placeholder="UUID of mission to launch"
                        className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    />
                </div>
            </div>

            <div className="grid grid-cols-4 gap-3">
                {/* Mode */}
                <div>
                    <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                        Mode
                    </label>
                    <select
                        value={mode}
                        onChange={(e) =>
                            setMode(e.target.value as "propose" | "auto_execute")
                        }
                        className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    >
                        <option value="propose">Propose</option>
                        <option value="auto_execute">Auto Execute</option>
                    </select>
                </div>

                {/* Cooldown */}
                <div>
                    <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                        Cooldown (s)
                    </label>
                    <input
                        type="number"
                        min={0}
                        value={cooldown}
                        onChange={(e) => setCooldown(Number(e.target.value))}
                        className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    />
                </div>

                {/* Max Depth */}
                <div>
                    <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                        Max Depth
                    </label>
                    <input
                        type="number"
                        min={1}
                        max={10}
                        value={maxDepth}
                        onChange={(e) => setMaxDepth(Number(e.target.value))}
                        className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    />
                </div>

                {/* Max Active Runs */}
                <div>
                    <label className="text-[10px] font-medium text-cortex-text-muted uppercase tracking-wider">
                        Max Runs
                    </label>
                    <input
                        type="number"
                        min={1}
                        value={maxActiveRuns}
                        onChange={(e) => setMaxActiveRuns(Number(e.target.value))}
                        className="mt-1 w-full px-3 py-1.5 text-sm rounded-md bg-cortex-bg border border-cortex-border text-cortex-text-main focus:outline-none focus:ring-1 focus:ring-cortex-primary"
                    />
                </div>
            </div>

            {mode === "auto_execute" && (
                <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-amber-500/5 border border-amber-500/20">
                    <Zap size={14} className="text-amber-400 flex-shrink-0" />
                    <span className="text-xs text-amber-300">
                        Auto-execute fires without human approval. Ensure guards are properly configured.
                    </span>
                </div>
            )}

            {/* Actions */}
            <div className="flex items-center justify-end gap-2 pt-1">
                <button
                    onClick={onCancel}
                    className="px-3 py-1.5 text-xs font-medium rounded-md text-cortex-text-muted hover:bg-cortex-border transition-colors"
                >
                    Cancel
                </button>
                <button
                    onClick={handleSubmit}
                    disabled={saving || !name.trim() || !targetMissionId.trim()}
                    className="px-3 py-1.5 text-xs font-medium rounded-md bg-cortex-primary text-cortex-bg hover:bg-cortex-primary/80 disabled:opacity-40 disabled:cursor-not-allowed transition-colors flex items-center gap-1.5"
                >
                    {saving && <Loader2 size={12} className="animate-spin" />}
                    Create Rule
                </button>
            </div>
        </div>
    );
}
