"use client";

import React, { useState } from "react";
import { AlertTriangle, Globe } from "lucide-react";

interface Props {
    provider: { id: string; data_boundary: string };
    onConfirm: () => void;
    onCancel: () => void;
}

export default function RemoteEnableModal({ provider, onConfirm, onCancel }: Props) {
    const [acknowledged, setAcknowledged] = useState(false);

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-cortex-surface border border-cortex-border rounded-lg w-full max-w-md p-6 space-y-4 shadow-2xl">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-amber-400/10 flex items-center justify-center">
                        <Globe className="w-5 h-5 text-amber-400" />
                    </div>
                    <div>
                        <h3 className="text-cortex-text-main font-bold text-sm">Enable Remote Provider</h3>
                        <p className="text-cortex-text-muted text-xs">
                            Enabling <strong className="text-cortex-text-main">{provider.id}</strong> will send data outside your environment.
                        </p>
                    </div>
                </div>

                <div className="p-3 rounded border border-amber-400/20 bg-amber-400/5 flex items-start gap-2">
                    <AlertTriangle className="w-4 h-4 text-amber-400 flex-shrink-0 mt-0.5" />
                    <p className="text-xs text-amber-400/90">
                        Data processed by this provider will leave your tenant boundary. Ensure this complies with your organization&apos;s data governance policies.
                    </p>
                </div>

                <label className="flex items-center gap-2 cursor-pointer">
                    <input
                        type="checkbox"
                        checked={acknowledged}
                        onChange={(e) => setAcknowledged(e.target.checked)}
                        className="rounded border-cortex-border bg-cortex-bg text-cortex-primary focus:ring-cortex-primary"
                    />
                    <span className="text-xs text-cortex-text-main">I understand data leaves tenant boundary</span>
                </label>

                <div className="flex items-center gap-2 justify-end">
                    <button
                        onClick={onCancel}
                        className="px-4 py-2 rounded border border-cortex-border text-cortex-text-muted text-xs font-mono hover:bg-cortex-border transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={onConfirm}
                        disabled={!acknowledged}
                        className="px-4 py-2 rounded bg-amber-400/20 border border-amber-400/40 text-amber-400 text-xs font-mono hover:bg-amber-400/30 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                        Enable Provider
                    </button>
                </div>
            </div>
        </div>
    );
}
