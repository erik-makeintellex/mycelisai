"use client";

import { useEffect, useState } from "react";
import { Check, Copy, ShieldAlert, X } from "lucide-react";

type ConfigDocumentExportPayload = {
  content?: string;
  redaction_applied?: boolean;
};

type DialogState = "loading" | "ready" | "forbidden" | "not_found" | "unavailable" | "error";

function stateForStatus(status: number): DialogState | null {
  if (status === 401 || status === 403) return "forbidden";
  if (status === 404) return "not_found";
  if (status === 503) return "unavailable";
  if (status >= 400) return "error";
  return null;
}

export default function ConfigDocumentViewDialog({
  recordId,
  onClose,
}: {
  recordId: string;
  onClose: () => void;
}) {
  const [state, setState] = useState<DialogState>("loading");
  const [payload, setPayload] = useState<ConfigDocumentExportPayload | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setState("loading");
    setPayload(null);
    fetch(`/api/v1/config-documents/${encodeURIComponent(recordId)}/export?format=yaml`, {
      cache: "no-store",
    })
      .then(async (res) => {
        if (cancelled) return;
        const mapped = stateForStatus(res.status);
        if (mapped) {
          setState(mapped);
          return;
        }
        const envelope = await res.json().catch(() => null);
        const data = envelope?.ok ? envelope.data : null;
        if (!data || typeof data.content !== "string") {
          setState("error");
          return;
        }
        setPayload(data);
        setState("ready");
      })
      .catch(() => {
        if (!cancelled) setState("error");
      });
    return () => {
      cancelled = true;
    };
  }, [recordId]);

  const copyContent = async () => {
    if (!payload?.content) return;
    try {
      await navigator.clipboard.writeText(payload.content);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      role="dialog"
      aria-modal="true"
      aria-label="View config"
    >
      <div className="flex max-h-[80vh] w-full max-w-2xl flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-panel">
        <div className="flex items-center justify-between border-b border-cortex-border px-4 py-3">
          <h2 className="text-sm font-semibold text-cortex-text-main">View config</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-md p-1 text-cortex-text-muted hover:bg-cortex-bg"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
          {state === "loading" ? (
            <p className="text-sm text-cortex-text-muted">Loading...</p>
          ) : state === "forbidden" ? (
            <p className="flex items-center gap-2 text-sm text-cortex-danger">
              <ShieldAlert className="h-4 w-4" /> You do not have access to this configuration.
            </p>
          ) : state === "not_found" ? (
            <p className="text-sm text-cortex-danger">This configuration could not be found.</p>
          ) : state === "unavailable" ? (
            <p className="text-sm text-cortex-warning">
              Configuration storage is unavailable right now. Try again shortly.
            </p>
          ) : state === "error" ? (
            <p className="text-sm text-cortex-danger">This configuration could not be loaded.</p>
          ) : (
            <div className="space-y-3">
              {payload?.redaction_applied ? (
                <p className="rounded-md border border-cortex-warning/40 bg-cortex-warning/10 px-3 py-2 text-xs text-cortex-warning">
                  Some values were hidden.
                </p>
              ) : null}
              <pre className="max-h-[50vh] overflow-auto whitespace-pre-wrap break-all rounded-md border border-cortex-border bg-cortex-bg p-3 font-mono text-xs text-cortex-text-main">
                {payload?.content}
              </pre>
            </div>
          )}
        </div>

        <div className="flex items-center justify-end gap-2 border-t border-cortex-border px-4 py-3">
          <button
            type="button"
            onClick={() => void copyContent()}
            disabled={state !== "ready" || !payload?.content}
            className="inline-flex items-center gap-1.5 rounded-md border border-cortex-border px-3 py-1.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-success/50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
            {copied ? "Copied" : "Copy"}
          </button>
        </div>
      </div>
    </div>
  );
}
