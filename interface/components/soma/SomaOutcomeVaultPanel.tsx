"use client";

import Link from "next/link";
import { FolderOpen, Radio } from "lucide-react";
import type { ReactNode } from "react";

export function SomaOutcomeVaultPanel({
  output,
  activeWork,
  hasOutputs,
  operationCount,
}: {
  output?: ReactNode;
  activeWork?: ReactNode;
  hasOutputs: boolean;
  operationCount: number;
}) {
  return (
    <aside
      className="flex min-h-[620px] min-w-0 flex-col overflow-hidden rounded-[3rem] border border-[#E5E7EB] bg-[#FFFFFF] shadow-sm"
      aria-label="Outcomes and Vault"
      data-testid="soma-outcome-vault"
    >
      <div className="border-b border-[#E5E7EB] px-6 py-5">
        <h2 className="text-xl font-semibold tracking-tight text-[#111827]">Outcomes & Vault</h2>
      </div>
      <div className="min-h-0 flex-1 space-y-7 overflow-y-auto px-6 py-6">
        <section>
          <div className="mb-3 flex items-center justify-between gap-3">
            <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-[#4B5563]">Running in background</h3>
            {operationCount > 0 ? (
              <span className="rounded-full bg-[#ECFDF5] px-2 py-0.5 text-xs font-semibold text-[#15803D]">
                {operationCount}
              </span>
            ) : null}
          </div>
          <div className="space-y-3">
            {operationCount > 0 && activeWork ? (
              <div className="rounded-xl border border-[#E5E7EB] bg-[#FFFFFF] p-3 text-[#111827]">
                {activeWork}
              </div>
            ) : (
              <div className="rounded-xl border border-[#D1D5DB] bg-[#FFFFFF] px-4 py-3">
                <div className="flex items-center gap-2 font-semibold text-[#111827]">
                  <Radio className="h-4 w-4 text-[#9CA3AF]" />
                  No background work running
                </div>
                <p className="mt-1 text-sm leading-5 text-[#6B7280]">
                  Quick actions and approved Soma work will appear here while they run.
                </p>
              </div>
            )}
          </div>
        </section>
        <section>
          <div className="mb-3 flex items-center justify-between gap-3">
            <h3 className="text-xs font-bold uppercase tracking-[0.12em] text-[#4B5563]">Recent deliverables</h3>
            <Link href="/resources?tab=workspace" className="text-xs font-semibold text-[#2563EB] hover:underline">
              Open vault
            </Link>
          </div>
          {hasOutputs && output ? (
            <div className="rounded-xl border border-[#E5E7EB] bg-[#F9FAFB] p-3 text-[#111827]">
              {output}
            </div>
          ) : (
            <div className="space-y-3">
              <div className="flex items-center justify-between gap-4 rounded-xl border border-[#D1D5DB] bg-[#FFFFFF] px-4 py-3">
                <div>
                  <div className="font-semibold text-[#111827]">No retained deliverables yet</div>
                  <div className="text-sm text-[#6B7280]">Ask Soma to create or review something and save the output.</div>
                </div>
                <FolderOpen className="h-5 w-5 shrink-0 text-[#9CA3AF]" />
              </div>
            </div>
          )}
        </section>
      </div>
    </aside>
  );
}
