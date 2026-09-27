"use client";

import { useEffect, useState } from "react";
import { Eye, RefreshCw, ShieldAlert } from "lucide-react";
import ConfigDocumentViewDialog from "./ConfigDocumentViewDialog";

type ConfigDocumentScope = { kind?: string; ref?: string };

type ConfigDocumentRecord = {
  record_id: string;
  document?: {
    kind?: string;
    metadata?: {
      id?: string;
      version?: string;
      scope?: ConfigDocumentScope;
      enabled?: boolean;
    };
  };
};

const EXPORTABLE_KINDS = new Set(["OutcomeTemplate", "WorkerProfile", "CodeContextSource"]);

type LoadState = "idle" | "loading" | "ready" | "forbidden" | "unavailable" | "error";

function scopeLabel(scope?: ConfigDocumentScope) {
  if (!scope?.kind) return "unscoped";
  return scope.ref ? `${scope.kind}/${scope.ref}` : scope.kind;
}

export default function ConfigDocumentRecords() {
  const [isAdmin, setIsAdmin] = useState(false);
  const [checkedSession, setCheckedSession] = useState(false);
  const [records, setRecords] = useState<ConfigDocumentRecord[]>([]);
  const [state, setState] = useState<LoadState>("idle");
  const [openRecordId, setOpenRecordId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        // `await` tolerates a non-promise return (e.g. an unmocked fetch in
        // tests resolving to `undefined`) instead of throwing synchronously
        // on `.then`, so this defaults to the non-admin view on any failure.
        const res = await fetch("/auth/session", { cache: "no-store" });
        const body = res?.ok ? await res.json().catch(() => null) : null;
        if (!cancelled) setIsAdmin(body?.data?.user?.role === "admin");
      } catch {
        if (!cancelled) setIsAdmin(false);
      } finally {
        if (!cancelled) setCheckedSession(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const load = async () => {
    setState("loading");
    try {
      const res = await fetch("/api/v1/config-documents?limit=100", { cache: "no-store" });
      if (res.status === 401 || res.status === 403) {
        setRecords([]);
        setState("forbidden");
        return;
      }
      if (res.status === 503) {
        setRecords([]);
        setState("unavailable");
        return;
      }
      if (!res.ok) {
        setRecords([]);
        setState("error");
        return;
      }
      const envelope = await res.json().catch(() => null);
      const data: ConfigDocumentRecord[] = Array.isArray(envelope?.data) ? envelope.data : [];
      setRecords(data.filter((record) => EXPORTABLE_KINDS.has(record?.document?.kind ?? "")));
      setState("ready");
    } catch {
      setRecords([]);
      setState("error");
    }
  };

  useEffect(() => {
    if (checkedSession && isAdmin) {
      void load();
    }
  }, [checkedSession, isAdmin]);

  if (!checkedSession || !isAdmin) {
    return null;
  }

  return (
    <section className="border-b border-cortex-border bg-cortex-surface/30 px-5 py-4">
      <div className="rounded-lg border border-cortex-border bg-cortex-panel p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-cortex-success">
              Saved configurations
            </p>
            <h3 className="mt-1 text-lg font-semibold text-cortex-text-main">View a saved configuration</h3>
          </div>
          <button
            type="button"
            onClick={() => void load()}
            className="inline-flex items-center gap-1.5 rounded-md border border-cortex-border px-3 py-1.5 text-xs font-mono text-cortex-text-main hover:bg-cortex-bg"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${state === "loading" ? "animate-spin" : ""}`} />
            Refresh
          </button>
        </div>

        {state === "forbidden" ? (
          <p className="mt-3 flex items-center gap-2 text-sm text-cortex-danger">
            <ShieldAlert className="h-4 w-4" /> You do not have access to saved configurations.
          </p>
        ) : state === "unavailable" ? (
          <p className="mt-3 text-sm text-cortex-warning">
            Configuration storage is unavailable right now. Try again shortly.
          </p>
        ) : state === "error" ? (
          <p className="mt-3 text-sm text-cortex-danger">Saved configurations could not be loaded.</p>
        ) : records.length === 0 && state === "ready" ? (
          <p className="mt-3 text-sm text-cortex-text-muted">No saved configurations yet.</p>
        ) : (
          <ul className="mt-3 divide-y divide-cortex-border">
            {records.map((record) => {
              const metadata = record.document?.metadata;
              return (
                <li key={record.record_id} className="flex items-center justify-between gap-3 py-2">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-cortex-text-main">
                      {record.document?.kind} · {metadata?.id ?? "unknown"}
                    </p>
                    <p className="truncate text-xs text-cortex-text-muted">
                      v{metadata?.version ?? "-"} · {scopeLabel(metadata?.scope)} ·{" "}
                      {metadata?.enabled ? "Active" : "Inactive"}
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setOpenRecordId(record.record_id)}
                    className="inline-flex flex-shrink-0 items-center gap-1.5 rounded-md border border-cortex-border px-3 py-1.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-success/50"
                  >
                    <Eye className="h-3.5 w-3.5" />
                    View config
                  </button>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {openRecordId ? (
        <ConfigDocumentViewDialog recordId={openRecordId} onClose={() => setOpenRecordId(null)} />
      ) : null}
    </section>
  );
}
