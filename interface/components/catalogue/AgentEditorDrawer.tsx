"use client";

import React, { useCallback, useState } from "react";
import { Copy, Plus, Save, Trash2, X } from "lucide-react";
import type { AgentContextBinding, CatalogueAgent } from "@/store/useCortexStore";
import { TagInput } from "@/components/common/TagInput";

interface AgentEditorDrawerProps {
  agent: CatalogueAgent | null;
  onClose: () => void;
  onSave: (data: Partial<CatalogueAgent>) => void;
  onDuplicate?: (agent: CatalogueAgent) => void;
}

type EditorTab = "profile" | "access" | "quality";
const ROLES = ["cognitive", "sensory", "actuation", "ledger", "researcher", "analyst", "builder", "media_creator", "reviewer", "coordinator"];
const CONTEXT_KINDS = ["public_web", "approved_local_data", "approved_mount", "private_api", "deployment_context", "outcome_sources", "outcome_outputs", "team_workspace", "run_proof"];

export default function AgentEditorDrawer(props: AgentEditorDrawerProps) {
  return <AgentEditorForm key={props.agent?.id ?? "new-profile"} {...props} />;
}

function AgentEditorForm({ agent, onClose, onSave, onDuplicate }: AgentEditorDrawerProps) {
  const readOnly = Boolean(agent?.locked || agent?.source === "built_in");
  const [tab, setTab] = useState<EditorTab>("profile");
  const [name, setName] = useState(agent?.name ?? "");
  const [description, setDescription] = useState(agent?.description ?? "");
  const [role, setRole] = useState(agent?.role ?? "builder");
  const [systemPrompt, setSystemPrompt] = useState(agent?.system_prompt ?? "");
  const [model, setModel] = useState(agent?.model ?? "");
  const [capabilities, setCapabilities] = useState<string[]>(agent?.capability_refs?.length ? [...agent.capability_refs] : [...(agent?.tools ?? [])]);
  const [contextBindings, setContextBindings] = useState<AgentContextBinding[]>(() => [...(agent?.context_bindings ?? [])]);
  const [selection, setSelection] = useState(agent?.usage_policy?.selection ?? "soma_or_manual");
  const [scope, setScope] = useState(agent?.usage_policy?.scope ?? "workspace");
  const [inputs, setInputs] = useState<string[]>(() => [...(agent?.inputs ?? [])]);
  const [outputs, setOutputs] = useState<string[]>(() => [...(agent?.outputs ?? [])]);
  const [verificationStrategy, setVerificationStrategy] = useState(agent?.verification_strategy ?? "semantic");
  const [verificationRubric, setVerificationRubric] = useState(() => agent?.verification_rubric.join(", ") ?? "");
  const [validationCommand, setValidationCommand] = useState(agent?.validation_command ?? "");

  const inputClass = "w-full rounded-md border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main outline-none transition-colors placeholder:text-cortex-text-muted/60 focus:border-cortex-primary disabled:cursor-not-allowed disabled:opacity-70";
  const labelClass = "mb-1.5 block text-xs font-semibold text-cortex-text-main";

  const save = useCallback(() => {
    onSave({
      name: name.trim(), description: description.trim() || undefined, role,
      system_prompt: systemPrompt.trim() || undefined, model: model.trim() || undefined,
      tools: capabilities, capability_refs: capabilities, context_bindings: contextBindings,
      usage_policy: { selection, scope }, inputs, outputs,
      verification_strategy: verificationStrategy,
      verification_rubric: verificationRubric.split(",").map((item) => item.trim()).filter(Boolean),
      validation_command: validationCommand.trim() || undefined,
    });
  }, [name, description, role, systemPrompt, model, capabilities, contextBindings, selection, scope, inputs, outputs, verificationStrategy, verificationRubric, validationCommand, onSave]);

  const updateBinding = (index: number, patch: Partial<AgentContextBinding>) => {
    setContextBindings((current) => current.map((binding, bindingIndex) => bindingIndex === index ? { ...binding, ...patch } : binding));
  };

  return (
    <div className="absolute inset-0 z-40 flex justify-end bg-black/35" role="dialog" aria-modal="true" aria-label={agent ? `${agent.name} profile` : "New worker profile"}>
      <section className="flex h-full w-full max-w-xl flex-col border-l border-cortex-border bg-cortex-surface shadow-2xl">
        <header className="flex items-start justify-between gap-3 border-b border-cortex-border px-5 py-4">
          <div>
            <h2 className="text-base font-semibold text-cortex-text-main">{agent ? agent.name : "New worker profile"}</h2>
            <p className="mt-1 text-xs text-cortex-text-muted">{readOnly ? "Ready-made profile. Copy it to make changes." : "Define what this teammate does and what it may use."}</p>
          </div>
          <button type="button" aria-label="Close profile" onClick={onClose} className="rounded p-1.5 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-text-main"><X className="h-4 w-4" /></button>
        </header>

        <nav className="flex gap-1 border-b border-cortex-border px-5 py-2" role="tablist" aria-label="Profile settings">
          {([['profile', 'Profile'], ['access', 'Access & context'], ['quality', 'Quality']] as const).map(([value, label]) => (
            <button key={value} type="button" role="tab" aria-selected={tab === value} onClick={() => setTab(value)}
              className={`rounded px-3 py-2 text-xs font-semibold ${tab === value ? "bg-cortex-primary/15 text-cortex-primary" : "text-cortex-text-muted hover:text-cortex-text-main"}`}>{label}</button>
          ))}
        </nav>

        <div className="flex-1 overflow-y-auto p-5">
          {tab === "profile" && <div className="space-y-4">
            <Field label="Name"><input disabled={readOnly} value={name} onChange={(event) => setName(event.target.value)} className={inputClass} /></Field>
            <Field label="Purpose"><textarea disabled={readOnly} value={description} onChange={(event) => setDescription(event.target.value)} rows={2} className={`${inputClass} resize-y`} /></Field>
            <Field label="Role"><select disabled={readOnly} value={role} onChange={(event) => setRole(event.target.value)} className={inputClass}>{ROLES.map((value) => <option key={value} value={value}>{value.replaceAll("_", " ")}</option>)}</select></Field>
            <Field label="Instructions"><textarea disabled={readOnly} value={systemPrompt} onChange={(event) => setSystemPrompt(event.target.value)} rows={6} className={`${inputClass} resize-y`} /></Field>
            <Field label="Model override"><input disabled={readOnly} value={model} onChange={(event) => setModel(event.target.value)} placeholder="Uses workspace AI engine" className={inputClass} /></Field>
          </div>}

          {tab === "access" && <div className="space-y-5">
            <section>
              <h3 className={labelClass}>Capabilities</h3>
              {readOnly ? <TokenList values={capabilities} empty="No capabilities configured" /> : <TagInput label="Capability or tool references" value={capabilities} onChange={setCapabilities} />}
            </section>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Who may select it"><select disabled={readOnly} value={selection} onChange={(event) => setSelection(event.target.value)} className={inputClass}><option value="soma_or_manual">Soma or operator</option><option value="suggested">Soma when relevant</option><option value="soma">Soma</option><option value="manual">Operator</option><option value="automatic">Policy automation</option></select></Field>
              <Field label="Default scope"><select disabled={readOnly} value={scope} onChange={(event) => setScope(event.target.value)} className={inputClass}><option value="workspace">Workspace</option><option value="outcome">Outcome</option><option value="team">Team</option></select></Field>
            </div>
            <section>
              <div className="mb-2 flex items-center justify-between"><h3 className={labelClass}>Context sources</h3>{!readOnly && <button type="button" onClick={() => setContextBindings((current) => [...current, { kind: "approved_local_data", access: "read" }])} className="inline-flex items-center gap-1 text-xs font-semibold text-cortex-primary"><Plus className="h-3.5 w-3.5" /> Add source</button>}</div>
              <div className="space-y-2">
                {contextBindings.length === 0 && <p className="text-xs text-cortex-text-muted">No context sources configured.</p>}
                {contextBindings.map((binding, index) => <div key={`${binding.kind}-${index}`} className="grid grid-cols-[1fr_1fr_7rem_2rem] gap-2">
                  <select disabled={readOnly} aria-label={`Context type ${index + 1}`} value={binding.kind} onChange={(event) => updateBinding(index, { kind: event.target.value })} className={inputClass}>{CONTEXT_KINDS.map((kind) => <option key={kind} value={kind}>{kind.replaceAll("_", " ")}</option>)}</select>
                  <input disabled={readOnly} aria-label={`Context reference ${index + 1}`} value={binding.ref ?? ""} onChange={(event) => updateBinding(index, { ref: event.target.value || undefined })} placeholder="Any configured source" className={inputClass} />
                  <select disabled={readOnly} aria-label={`Context access ${index + 1}`} value={binding.access ?? "read"} onChange={(event) => updateBinding(index, { access: event.target.value })} className={inputClass}><option value="search">Search</option><option value="read">Read</option><option value="write">Write</option><option value="read_write">Read/write</option></select>
                  <button type="button" disabled={readOnly} aria-label={`Remove context source ${index + 1}`} onClick={() => setContextBindings((current) => current.filter((_, bindingIndex) => bindingIndex !== index))} className="rounded text-cortex-text-muted hover:bg-cortex-danger/15 hover:text-cortex-danger disabled:opacity-0"><Trash2 className="mx-auto h-4 w-4" /></button>
                </div>)}
              </div>
            </section>
            {!readOnly && <details className="rounded-md border border-cortex-border p-3"><summary className="cursor-pointer text-xs font-semibold text-cortex-text-main">Advanced message channels</summary><div className="mt-4 space-y-4"><TagInput label="Inputs" value={inputs} onChange={setInputs} /><TagInput label="Outputs" value={outputs} onChange={setOutputs} /></div></details>}
          </div>}

          {tab === "quality" && <div className="space-y-4">
            <Field label="Verification"><select disabled={readOnly} value={verificationStrategy} onChange={(event) => setVerificationStrategy(event.target.value)} className={inputClass}><option value="semantic">Review against criteria</option><option value="empirical">Run validation</option><option value="none">No extra check</option></select></Field>
            <Field label="Review criteria"><textarea disabled={readOnly} value={verificationRubric} onChange={(event) => setVerificationRubric(event.target.value)} rows={4} placeholder="Accuracy, evidence quality, usable output" className={`${inputClass} resize-y`} /></Field>
            <Field label="Validation command"><input disabled={readOnly} value={validationCommand} onChange={(event) => setValidationCommand(event.target.value)} placeholder="Optional governed command" className={inputClass} /></Field>
          </div>}
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-cortex-border px-5 py-4">
          <button type="button" onClick={onClose} className="rounded-md border border-cortex-border px-4 py-2 text-xs font-semibold text-cortex-text-muted hover:text-cortex-text-main">Cancel</button>
          {readOnly && agent && onDuplicate ? <button type="button" onClick={() => onDuplicate(agent)} className="inline-flex items-center gap-2 rounded-md border border-cortex-primary/40 bg-cortex-primary/10 px-4 py-2 text-xs font-semibold text-cortex-primary"><Copy className="h-4 w-4" /> Copy profile</button> : <button type="button" onClick={save} disabled={!name.trim()} className="inline-flex items-center gap-2 rounded-md border border-cortex-success/40 bg-cortex-success/10 px-4 py-2 text-xs font-semibold text-cortex-success disabled:opacity-40"><Save className="h-4 w-4" /> {agent ? "Save changes" : "Create profile"}</button>}
        </footer>
      </section>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) { return <label><span className="mb-1.5 block text-xs font-semibold text-cortex-text-main">{label}</span>{children}</label>; }
function TokenList({ values, empty }: { values: string[]; empty: string }) { return values.length ? <div className="flex flex-wrap gap-2">{values.map((value) => <span key={value} className="rounded border border-cortex-border bg-cortex-bg px-2 py-1 text-xs text-cortex-text-muted">{value}</span>)}</div> : <p className="text-xs text-cortex-text-muted">{empty}</p>; }
