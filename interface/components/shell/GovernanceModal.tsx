"use client";

import React, { useState } from 'react';
import { X, CheckCircle2, XCircle, FileCheck, ShieldCheck, AlertTriangle } from 'lucide-react';
import { useCortexStore } from '@/store/useCortexStore';

export default function GovernanceModal() {
    const selectedArtifact = useCortexStore((s) => s.selectedArtifact);
    const selectArtifact = useCortexStore((s) => s.selectArtifact);
    const approveArtifact = useCortexStore((s) => s.approveArtifact);
    const rejectArtifact = useCortexStore((s) => s.rejectArtifact);
    const [rejectReason, setRejectReason] = useState('');
    const [showRejectInput, setShowRejectInput] = useState(false);

    if (!selectedArtifact) return null;

    const proof = selectedArtifact.proof;

    const handleApprove = () => {
        approveArtifact(selectedArtifact.id);
    };

    const handleReject = () => {
        if (!showRejectInput) {
            setShowRejectInput(true);
            return;
        }
        rejectArtifact(selectedArtifact.id, rejectReason || 'No reason provided');
        setRejectReason('');
        setShowRejectInput(false);
    };

    const handleClose = () => {
        selectArtifact(null);
        setShowRejectInput(false);
        setRejectReason('');
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            {/* Backdrop */}
            <div
                className="absolute inset-0 backdrop-blur-md bg-cortex-bg/80"
                onClick={handleClose}
            />

            {/* Modal */}
            <div className="relative z-10 w-[90vw] max-w-5xl max-h-[85vh] bg-cortex-surface border border-cortex-border rounded-2xl shadow-2xl flex flex-col overflow-hidden">
                {/* Header */}
                <div className="flex items-center gap-3 px-6 py-4 border-b border-cortex-border flex-shrink-0">
                    <ShieldCheck className="w-5 h-5 text-cortex-primary" />
                    <div className="flex-1">
                        <h2 className="text-sm font-bold text-cortex-text-main font-mono uppercase tracking-wider">
                            Governance Review
                        </h2>
                        <p className="text-[10px] font-mono text-cortex-text-muted mt-0.5">
                            Agent: <span className="text-cortex-info">{selectedArtifact.source}</span>
                            {' '}&middot;{' '}
                            {new Date(selectedArtifact.timestamp).toLocaleString()}
                        </p>
                    </div>
                    <button
                        onClick={handleClose}
                        className="p-1.5 rounded-lg text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg transition-colors"
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>

                {/* Two-column body */}
                <div className="flex-1 grid grid-cols-2 min-h-0 divide-x divide-cortex-border">
                    {/* Left: The Output */}
                    <div className="flex flex-col min-h-0">
                        <div className="px-5 py-2.5 border-b border-cortex-border/50 flex items-center gap-2 flex-shrink-0">
                            <FileCheck className="w-3.5 h-3.5 text-cortex-info" />
                            <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wide">
                                Agent Output
                            </span>
                            <span className="ml-auto text-[9px] font-mono text-cortex-text-muted/60 uppercase">
                                {selectedArtifact.payload.content_type}
                            </span>
                        </div>
                        <div className="flex-1 overflow-y-auto p-5 scrollbar-thin scrollbar-thumb-cortex-border">
                            {selectedArtifact.payload.title && (
                                <h3 className="text-sm font-bold text-cortex-text-main font-mono mb-3">
                                    {selectedArtifact.payload.title}
                                </h3>
                            )}
                            <pre className="text-[11px] font-mono text-cortex-text-main/90 leading-relaxed whitespace-pre-wrap break-words">
                                {selectedArtifact.payload.content}
                            </pre>
                        </div>
                    </div>

                    {/* Right: Proof of Work */}
                    <div className="flex flex-col min-h-0">
                        <div className="px-5 py-2.5 border-b border-cortex-border/50 flex items-center gap-2 flex-shrink-0">
                            <ShieldCheck className="w-3.5 h-3.5 text-cortex-success" />
                            <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wide">
                                Proof of Work
                            </span>
                            {proof && (
                                <span className={`ml-auto text-[9px] font-mono font-bold uppercase px-1.5 py-0.5 rounded ${
                                    proof.pass
                                        ? 'bg-cortex-success/10 text-cortex-success'
                                        : 'bg-cortex-danger/10 text-cortex-danger'
                                }`}>
                                    {proof.pass ? 'PASSED' : 'FAILED'}
                                </span>
                            )}
                        </div>
                        <div className="flex-1 overflow-y-auto p-5 scrollbar-thin scrollbar-thumb-cortex-border">
                            {proof ? (
                                <>
                                    {/* Method */}
                                    <div className="mb-4">
                                        <span className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider">
                                            Verification Method
                                        </span>
                                        <p className="text-xs font-mono text-cortex-primary mt-1 capitalize">
                                            {proof.method}
                                        </p>
                                    </div>

                                    {/* Rubric Score */}
                                    <div className="mb-4">
                                        <span className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider">
                                            Rubric Score
                                        </span>
                                        <p className="text-xs font-mono text-cortex-text-main mt-1">
                                            {proof.rubric_score}
                                        </p>
                                    </div>

                                    {/* Logs */}
                                    <div>
                                        <span className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider">
                                            Verification Logs
                                        </span>
                                        <pre className="mt-2 p-3 rounded-lg bg-cortex-bg border border-cortex-border text-[10px] font-mono text-cortex-text-main/80 leading-relaxed whitespace-pre-wrap break-words max-h-64 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border">
                                            {proof.logs}
                                        </pre>
                                    </div>
                                </>
                            ) : (
                                <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                                    <AlertTriangle className="w-8 h-8 mb-2 opacity-30" />
                                    <p className="text-[10px] font-mono">No proof of work provided</p>
                                    <p className="text-[9px] font-mono opacity-60 mt-1">
                                        This agent did not submit verification evidence
                                    </p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Footer: Action Buttons */}
                <div className="px-6 py-4 border-t border-cortex-border flex items-center gap-4 flex-shrink-0">
                    {showRejectInput && (
                        <input
                            type="text"
                            value={rejectReason}
                            onChange={(e) => setRejectReason(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleReject()}
                            placeholder="Reason for rejection..."
                            autoFocus
                            className="flex-1 bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2.5 text-sm text-cortex-text-main placeholder-cortex-text-muted/50 font-mono focus:outline-none focus:border-cortex-danger focus:ring-1 focus:ring-cortex-danger/30"
                        />
                    )}

                    <div className={`flex items-center gap-3 ${showRejectInput ? '' : 'ml-auto'}`}>
                        <button
                            onClick={handleReject}
                            className="flex items-center gap-2.5 px-6 py-3 rounded-xl font-mono text-sm font-bold uppercase tracking-wider transition-all duration-200 bg-red-500/10 text-red-500 border border-red-500 hover:bg-red-500/20 hover:shadow-[0_0_15px_rgba(234,84,85,0.2)]"
                        >
                            <XCircle className="w-4 h-4" />
                            Reject & Rework
                        </button>

                        <button
                            onClick={handleApprove}
                            className="flex items-center gap-2.5 px-8 py-3 rounded-xl font-mono text-sm font-bold uppercase tracking-wider transition-all duration-200 bg-cortex-success text-cortex-bg hover:shadow-[0_0_25px_rgba(16,185,129,0.4)] hover:bg-cortex-success/90"
                        >
                            <CheckCircle2 className="w-4 h-4" />
                            Approve & Dispatch
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
