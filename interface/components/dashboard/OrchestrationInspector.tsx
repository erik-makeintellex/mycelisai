"use client";

import React from "react";
import { X, Brain, Shield, Eye, FileText, Fingerprint, Clock } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { brainDisplayName, brainLocationLabel, toolLabel, sourceNodeLabel, MODE_LABELS } from "@/lib/labels";

export default function OrchestrationInspector() {
    const msg = useCortexStore((s) => s.inspectedMessage);
    const isOpen = useCortexStore((s) => s.isInspectorOpen);
    const setInspected = useCortexStore((s) => s.setInspectedMessage);

    if (!isOpen || !msg) return null;

    const modeInfo = MODE_LABELS[msg.mode || "answer"] || MODE_LABELS.answer;

    return (
        <div className="fixed inset-y-0 right-0 w-96 bg-cortex-surface border-l border-cortex-border z-50 flex flex-col shadow-2xl">
            {/* Header */}
            <div className="h-12 border-b border-cortex-border flex items-center justify-between px-4 flex-shrink-0">
                <div className="flex items-center gap-2">
                    <Eye className="w-4 h-4 text-cortex-primary" />
                    <span className="font-mono text-sm font-bold text-cortex-text-main">INSPECTOR</span>
                </div>
                <button
                    onClick={() => setInspected(null)}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 text-xs font-mono">
                {/* Execution Mode */}
                <InspectorSection icon={Shield} title="Execution">
                    <Row label="Mode" value={modeInfo.label} valueClass={modeInfo.color} />
                    <Row label="Template" value={msg.template_id || "chat-to-answer"} />
                    <Row label="Source" value={sourceNodeLabel(msg.source_node || "admin")} />
                    {msg.trust_score != null && (
                        <Row label="Confidence" value={`${(msg.trust_score * 100).toFixed(0)}%`} />
                    )}
                </InspectorSection>

                {/* Brain Provenance */}
                {msg.brain && (
                    <InspectorSection icon={Brain} title="Brain Provenance">
                        <Row label="Provider" value={brainDisplayName(msg.brain.provider_id)} />
                        <Row label="Model" value={msg.brain.model_id || "\u2014"} />
                        <Row label="Location" value={brainLocationLabel(msg.brain.location)} />
                        <Row label="Data Boundary" value={msg.brain.data_boundary === "leaves_org" ? "Leaves Organization" : "Local Only"} />
                        {msg.brain.tokens_used != null && msg.brain.tokens_used > 0 && (
                            <Row label="Tokens" value={msg.brain.tokens_used.toLocaleString()} />
                        )}
                    </InspectorSection>
                )}

                {/* Governance Provenance */}
                {msg.provenance && (
                    <InspectorSection icon={Fingerprint} title="Governance Proof">
                        <Row label="Intent" value={msg.provenance.resolved_intent} />
                        <Row label="Permission" value={msg.provenance.permission_check} />
                        <Row label="Policy" value={msg.provenance.policy_decision} />
                        <Row label="Audit ID" value={msg.provenance.audit_event_id || "\u2014"} mono />
                        {msg.provenance.consult_chain && msg.provenance.consult_chain.length > 0 && (
                            <div className="mt-1">
                                <span className="text-cortex-text-muted">Consult Chain:</span>
                                <div className="mt-1 flex flex-wrap gap-1">
                                    {msg.provenance.consult_chain.map((c, i) => (
                                        <span key={i} className="px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary text-[10px]">
                                            {c}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        )}
                    </InspectorSection>
                )}

                {/* Tools Used */}
                {msg.tools_used && msg.tools_used.length > 0 && (
                    <InspectorSection icon={FileText} title="Tools Invoked">
                        <div className="flex flex-wrap gap-1">
                            {msg.tools_used.map((t) => (
                                <span key={t} className="px-2 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary text-[10px] border border-cortex-primary/20">
                                    {toolLabel(t)}
                                </span>
                            ))}
                        </div>
                    </InspectorSection>
                )}

                {/* Timestamp */}
                {msg.timestamp && (
                    <InspectorSection icon={Clock} title="Timing">
                        <Row label="Timestamp" value={new Date(msg.timestamp).toLocaleString()} />
                    </InspectorSection>
                )}

                {/* Proposal details */}
                {msg.proposal && (
                    <InspectorSection icon={Shield} title="Proposal Details">
                        <Row label="Intent" value={msg.proposal.intent} />
                        <Row label="Teams" value={String(msg.proposal.teams)} />
                        <Row label="Agents" value={String(msg.proposal.agents)} />
                        <Row label="Risk" value={msg.proposal.risk_level?.toUpperCase() || "LOW"} />
                        <Row label="Proof ID" value={msg.proposal.intent_proof_id || "\u2014"} mono />
                    </InspectorSection>
                )}
            </div>
        </div>
    );
}

function InspectorSection({ icon: Icon, title, children }: {
    icon: React.ElementType;
    title: string;
    children: React.ReactNode;
}) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg/50 overflow-hidden">
            <div className="px-3 py-2 border-b border-cortex-border bg-cortex-surface/50 flex items-center gap-2">
                <Icon className="w-3.5 h-3.5 text-cortex-primary" />
                <span className="text-cortex-text-main font-bold text-[11px] tracking-wider uppercase">{title}</span>
            </div>
            <div className="px-3 py-2.5 space-y-1.5">
                {children}
            </div>
        </div>
    );
}

function Row({ label, value, valueClass, mono }: {
    label: string;
    value: string;
    valueClass?: string;
    mono?: boolean;
}) {
    return (
        <div className="flex items-center justify-between gap-2">
            <span className="text-cortex-text-muted">{label}</span>
            <span className={`${valueClass || "text-cortex-text-main"} ${mono ? "text-[10px] break-all" : ""}`}>{value}</span>
        </div>
    );
}
