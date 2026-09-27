"use client";

import { useState } from "react";
import { ChevronDown, ChevronRight, Plus, Trash2 } from "lucide-react";
import type { PolicyGroup, PolicyRule } from "@/store/useCortexStore";

// Split out of ApprovalsPolicyTab.tsx to stay within the line cap.

export default function PolicyGroupCard({
  group,
  onUpdate,
  onRemove,
  onAddRule,
  onRemoveRule,
  onUpdateRule,
}: {
  group: PolicyGroup;
  onUpdate: (patch: Partial<PolicyGroup>) => void;
  onRemove: () => void;
  onAddRule: () => void;
  onRemoveRule: (ruleIdx: number) => void;
  onUpdateRule: (ruleIdx: number, patch: Partial<PolicyRule>) => void;
}) {
  const [expanded, setExpanded] = useState(true);

  return (
    <div className="bg-cortex-surface border border-cortex-border rounded-xl overflow-hidden">
      <div
        className="px-4 py-3 flex items-center justify-between cursor-pointer hover:bg-cortex-bg/30 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2 min-w-0">
          {expanded ? (
            <ChevronDown
              size={14}
              className="text-cortex-text-muted flex-shrink-0"
            />
          ) : (
            <ChevronRight
              size={14}
              className="text-cortex-text-muted flex-shrink-0"
            />
          )}
          <div className="min-w-0">
            <h3 className="text-sm font-semibold text-cortex-text-main truncate">
              {group.name || "Untitled Group"}
            </h3>
            <p className="text-[10px] text-cortex-text-muted font-mono truncate">
              {group.targets.length > 0
                ? group.targets.join(", ")
                : "no targets"}
              {" -- "}
              {group.rules.length} rule{group.rules.length !== 1 ? "s" : ""}
            </p>
          </div>
        </div>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onRemove();
          }}
          className="p-1 text-cortex-text-muted hover:text-cortex-danger rounded transition-colors"
          title="Delete group"
        >
          <Trash2 size={14} />
        </button>
      </div>

      {expanded && (
        <div className="border-t border-cortex-border p-4 space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">
                Name
              </label>
              <input
                type="text"
                value={group.name}
                onChange={(e) => onUpdate({ name: e.target.value })}
                className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
              />
            </div>
            <div>
              <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">
                Targets (comma-separated)
              </label>
              <input
                type="text"
                value={group.targets.join(", ")}
                onChange={(e) =>
                  onUpdate({
                    targets: e.target.value
                      .split(",")
                      .map((t) => t.trim())
                      .filter(Boolean),
                  })
                }
                className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
                placeholder="team-*, agent-recon"
              />
            </div>
          </div>
          <div>
            <label className="text-[10px] font-mono uppercase text-cortex-text-muted block mb-1">
              Description
            </label>
            <input
              type="text"
              value={group.description}
              onChange={(e) => onUpdate({ description: e.target.value })}
              className="w-full bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
              placeholder="What this policy group controls..."
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-[10px] font-mono uppercase text-cortex-text-muted">
                Rules
              </span>
              <button
                onClick={onAddRule}
                className="text-[10px] font-medium text-cortex-primary hover:text-cortex-primary/80 flex items-center gap-1 transition-colors"
              >
                <Plus size={10} />
                Add Rule
              </button>
            </div>

            {group.rules.length === 0 ? (
              <p className="text-xs text-cortex-text-muted italic py-2">
                No rules. Add one to define behavior.
              </p>
            ) : (
              <div className="space-y-2">
                <div className="grid grid-cols-[1fr_1fr_140px_28px] gap-2 text-[9px] font-mono uppercase text-cortex-text-muted px-1">
                  <span>Intent Pattern</span>
                  <span>Condition</span>
                  <span>Action</span>
                  <span />
                </div>
                {group.rules.map((rule, ri) => (
                  <RuleRow
                    key={ri}
                    rule={rule}
                    onUpdate={(patch) => onUpdateRule(ri, patch)}
                    onRemove={() => onRemoveRule(ri)}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// ── Rule Row ────────────────────────────────────────────

const ACTION_COLORS: Record<string, string> = {
  ALLOW: "text-cortex-success border-cortex-success/30",
  DENY: "text-cortex-danger border-cortex-danger/30",
  REQUIRE_APPROVAL: "text-cortex-warning border-cortex-warning/30",
};

function RuleRow({
  rule,
  onUpdate,
  onRemove,
}: {
  rule: PolicyRule;
  onUpdate: (patch: Partial<PolicyRule>) => void;
  onRemove: () => void;
}) {
  return (
    <div className="grid grid-cols-[1fr_1fr_140px_28px] gap-2 items-center">
      <input
        type="text"
        value={rule.intent}
        onChange={(e) => onUpdate({ intent: e.target.value })}
        className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded px-2 py-1 focus:outline-none focus:border-cortex-primary"
        placeholder="file.write.*"
      />
      <input
        type="text"
        value={rule.condition}
        onChange={(e) => onUpdate({ condition: e.target.value })}
        className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded px-2 py-1 focus:outline-none focus:border-cortex-primary"
        placeholder="trust < 0.7"
      />
      <select
        value={rule.action}
        onChange={(e) =>
          onUpdate({ action: e.target.value as PolicyRule["action"] })
        }
        className={`bg-cortex-bg border text-xs font-mono font-semibold rounded px-2 py-1 focus:outline-none ${ACTION_COLORS[rule.action] ?? "text-cortex-text-main border-cortex-border"}`}
      >
        <option value="ALLOW">ALLOW</option>
        <option value="DENY">DENY</option>
        <option value="REQUIRE_APPROVAL">REQUIRE_APPROVAL</option>
      </select>
      <button
        onClick={onRemove}
        className="p-1 text-cortex-text-muted hover:text-cortex-danger rounded transition-colors"
        title="Delete rule"
      >
        <Trash2 size={12} />
      </button>
    </div>
  );
}
