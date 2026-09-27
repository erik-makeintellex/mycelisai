"use client";

import { useEffect, useState } from "react";
import { Loader2, Plus, Save, Settings } from "lucide-react";
import {
  useCortexStore,
  type PolicyConfig,
  type PolicyGroup,
  type PolicyRule,
} from "@/store/useCortexStore";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import PolicyGroupCard from "@/components/automations/ApprovalsPolicyGroupCard";

// Split out of ApprovalsTab.tsx to stay within the file's line cap.

export default function PolicyTab() {
  const policyConfig = useCortexStore((s) => s.policyConfig);
  const policyError = useCortexStore((s) => s.policyError);
  const isFetchingPolicy = useCortexStore((s) => s.isFetchingPolicy);
  const fetchPolicy = useCortexStore((s) => s.fetchPolicy);
  const updatePolicy = useCortexStore((s) => s.updatePolicy);

  const [draft, setDraft] = useState<PolicyConfig | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    fetchPolicy();
  }, [fetchPolicy]);

  // Never invent a policy: a fake "DENY" default here would look real and could be saved over the actual policy.
  const config = draft ?? policyConfig;

  if (isFetchingPolicy && !draft && !policyConfig) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 aria-hidden="true" size={24} className="text-cortex-primary animate-spin" />
      </div>
    );
  }

  if (!config) {
    // No fake DENY-all editor on a 403/503; show the blocker instead.
    return (
      <div className="p-6 max-w-4xl mx-auto">
        <InlineBlockerNotice code={policyError?.code ?? "governance_policy_unavailable"} httpStatus={policyError?.httpStatus} onRetry={fetchPolicy} />
      </div>
    );
  }

  // `config` narrows non-null from here on (a `const`); no per-mutator guards needed.
  const handleSave = async () => {
    setIsSaving(true);
    await updatePolicy(config);
    setIsSaving(false);
  };

  const addGroup = () => {
    const newGroup: PolicyGroup = {
      name: `group-${config.groups.length + 1}`,
      description: "",
      targets: [],
      rules: [],
    };
    setDraft({ ...config, groups: [...config.groups, newGroup] });
  };

  const removeGroup = (idx: number) => {
    setDraft({ ...config, groups: config.groups.filter((_, i) => i !== idx) });
  };

  const updateGroup = (idx: number, patch: Partial<PolicyGroup>) => {
    setDraft({
      ...config,
      groups: config.groups.map((g, i) => (i === idx ? { ...g, ...patch } : g)),
    });
  };

  const addRule = (groupIdx: number) => {
    const newRule: PolicyRule = {
      intent: "*",
      condition: "always",
      action: "REQUIRE_APPROVAL",
    };
    const groups = config.groups.map((g, i) =>
      i === groupIdx ? { ...g, rules: [...g.rules, newRule] } : g,
    );
    setDraft({ ...config, groups });
  };

  const removeRule = (groupIdx: number, ruleIdx: number) => {
    const groups = config.groups.map((g, gi) =>
      gi === groupIdx
        ? { ...g, rules: g.rules.filter((_, ri) => ri !== ruleIdx) }
        : g,
    );
    setDraft({ ...config, groups });
  };

  const updateRule = (
    groupIdx: number,
    ruleIdx: number,
    patch: Partial<PolicyRule>,
  ) => {
    const groups = config.groups.map((g, gi) =>
      gi === groupIdx
        ? {
            ...g,
            rules: g.rules.map((r, ri) =>
              ri === ruleIdx ? { ...r, ...patch } : r,
            ),
          }
        : g,
    );
    setDraft({ ...config, groups });
  };

  const setDefaultAction = (action: string) => {
    setDraft({ ...config, defaults: { default_action: action } });
  };

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-4">
      <div className="flex items-center justify-between">
        <button
          onClick={addGroup}
          className="px-3 py-1.5 text-xs font-medium text-cortex-primary border border-cortex-primary/30 hover:bg-cortex-primary/10 rounded-lg flex items-center gap-1.5 transition-colors"
        >
          <Plus size={14} />
          Add Group
        </button>
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="px-4 py-1.5 text-xs font-medium text-cortex-bg bg-cortex-primary hover:bg-cortex-primary/90 rounded-lg flex items-center gap-1.5 shadow-sm transition-colors disabled:opacity-50"
        >
          {isSaving ? (
            <Loader2 size={14} className="animate-spin" />
          ) : (
            <Save size={14} />
          )}
          Save Policy
        </button>
      </div>

      {config.groups.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-cortex-text-muted">
          <Settings size={40} className="mb-3 opacity-40" />
          <p className="text-sm">
            No policy groups configured. Add one to get started.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {config.groups.map((group, gi) => (
            <PolicyGroupCard
              key={gi}
              group={group}
              onUpdate={(patch) => updateGroup(gi, patch)}
              onRemove={() => removeGroup(gi)}
              onAddRule={() => addRule(gi)}
              onRemoveRule={(ri) => removeRule(gi, ri)}
              onUpdateRule={(ri, patch) => updateRule(gi, ri, patch)}
            />
          ))}
        </div>
      )}

      <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-sm font-semibold text-cortex-text-main">
              Default Action
            </h3>
            <p className="text-xs text-cortex-text-muted mt-0.5">
              Applied when no policy group rule matches
            </p>
          </div>
          <select
            value={config.defaults.default_action}
            onChange={(e) => setDefaultAction(e.target.value)}
            className="bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono rounded-lg px-3 py-1.5 focus:outline-none focus:border-cortex-primary"
          >
            <option value="ALLOW">ALLOW</option>
            <option value="DENY">DENY</option>
            <option value="REQUIRE_APPROVAL">REQUIRE_APPROVAL</option>
          </select>
        </div>
      </div>
    </div>
  );
}
