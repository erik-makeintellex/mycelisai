"use client";

import React, { useMemo, useState } from "react";
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import { ChevronUp, ChevronDown } from "lucide-react";
import type { MycelisChartSpec } from "./types";

interface DataTableProps {
  spec: MycelisChartSpec;
  compact?: boolean;
  className?: string;
}

const PAGE_SIZE = 100;

export default function DataTable({
  spec,
  compact = false,
  className,
}: DataTableProps) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [page, setPage] = useState(0);

  // Auto-detect columns from first data row
  const columns = useMemo<ColumnDef<Record<string, unknown>>[]>(() => {
    if (!spec.data.length) return [];
    return Object.keys(spec.data[0]).map((key) => ({
      accessorKey: key,
      header: key,
      cell: (info) => {
        const val = info.getValue();
        if (val == null) return "";
        if (typeof val === "number") return val.toLocaleString();
        return String(val);
      },
    }));
  }, [spec.data]);

  const displayData = useMemo(() => {
    if (compact) return spec.data.slice(0, 5);
    return spec.data.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);
  }, [spec.data, compact, page]);

  const totalPages = Math.ceil(spec.data.length / PAGE_SIZE);

  const table = useReactTable({
    data: displayData,
    columns,
    state: { sorting },
    onSortingChange: compact ? undefined : setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: compact ? undefined : getSortedRowModel(),
  });

  return (
    <div className={`overflow-auto ${className ?? ""}`}>
      <table className="w-full text-xs font-mono border-collapse">
        <thead className="bg-cortex-bg text-cortex-text-muted border-b border-cortex-border sticky top-0">
          {table.getHeaderGroups().map((hg) => (
            <tr key={hg.id}>
              {hg.headers.map((header) => (
                <th
                  key={header.id}
                  className={`px-2 py-1.5 text-left whitespace-nowrap ${
                    !compact ? "cursor-pointer select-none hover:text-cortex-text-main" : ""
                  }`}
                  style={{ fontSize: compact ? "9px" : "11px" }}
                  onClick={
                    compact
                      ? undefined
                      : header.column.getToggleSortingHandler()
                  }
                >
                  <div className="flex items-center gap-1">
                    {flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
                    {!compact &&
                      (header.column.getIsSorted() === "asc" ? (
                        <ChevronUp className="w-3 h-3" />
                      ) : header.column.getIsSorted() === "desc" ? (
                        <ChevronDown className="w-3 h-3" />
                      ) : null)}
                  </div>
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody className="text-cortex-text-main">
          {table.getRowModel().rows.map((row) => (
            <tr
              key={row.id}
              className="border-b border-cortex-border/30 hover:bg-cortex-bg/50"
            >
              {row.getVisibleCells().map((cell) => (
                <td
                  key={cell.id}
                  className="px-2 py-1 whitespace-nowrap"
                  style={{ fontSize: compact ? "9px" : "11px" }}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      {/* Pagination — only for non-compact with >1 page */}
      {!compact && totalPages > 1 && (
        <div className="flex items-center justify-between px-2 py-1.5 bg-cortex-bg border-t border-cortex-border">
          <span className="text-[9px] font-mono text-cortex-text-muted">
            {spec.data.length} rows &middot; Page {page + 1}/{totalPages}
          </span>
          <div className="flex gap-1">
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="px-2 py-0.5 text-[9px] font-mono rounded bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main disabled:opacity-30 border border-cortex-border"
            >
              Prev
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="px-2 py-0.5 text-[9px] font-mono rounded bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main disabled:opacity-30 border border-cortex-border"
            >
              Next
            </button>
          </div>
        </div>
      )}

      {/* Compact truncation indicator */}
      {compact && spec.data.length > 5 && (
        <div className="text-center py-0.5 text-[8px] font-mono text-cortex-text-muted/40">
          +{spec.data.length - 5} more rows
        </div>
      )}
    </div>
  );
}
