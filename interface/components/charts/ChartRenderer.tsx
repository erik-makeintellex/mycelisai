"use client";

import React, { useRef, useEffect } from "react";
import * as Plot from "@observablehq/plot";
import dynamic from "next/dynamic";
import type { MycelisChartSpec } from "./types";

const MapRenderer = dynamic(() => import("./MapRenderer"), { ssr: false });
const DataTable = dynamic(() => import("./DataTable"), { ssr: false });

interface ChartRendererProps {
  spec: MycelisChartSpec;
  className?: string;
  compact?: boolean;
}

export default function ChartRenderer({
  spec,
  className,
  compact = false,
}: ChartRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  // Geo → delegate to MapRenderer
  if (spec.chart_type === "geo" && spec.geo) {
    return <MapRenderer spec={spec} compact={compact} className={className} />;
  }

  // Table → delegate to DataTable
  if (spec.chart_type === "table") {
    return <DataTable spec={spec} compact={compact} className={className} />;
  }

  // All other chart types → Observable Plot
  return (
    <ObservablePlotChart
      spec={spec}
      compact={compact}
      className={className}
      containerRef={containerRef}
    />
  );
}

// ── Observable Plot rendering ────────────────────────────────

function ObservablePlotChart({
  spec,
  compact,
  className,
  containerRef,
}: {
  spec: MycelisChartSpec;
  compact: boolean;
  className?: string;
  containerRef: React.RefObject<HTMLDivElement | null>;
}) {
  useEffect(() => {
    if (!containerRef.current || !spec.data?.length) return;

    const container = containerRef.current;
    const width = compact
      ? 200
      : spec.width ?? container.clientWidth ?? 600;
    const height = compact ? 120 : spec.height ?? 400;

    const plotOptions = buildPlotOptions(spec, width, height, compact);

    // Dark theme injection
    plotOptions.style = {
      background: "transparent",
      color: "#CFD3EC",
      fontSize: compact ? "9px" : "11px",
    };

    try {
      const plot = Plot.plot(plotOptions);
      container.replaceChildren(plot);
    } catch (err) {
      console.error("[ChartRenderer] Plot render failed:", err);
      container.replaceChildren();
      const fallback = document.createElement("div");
      fallback.className =
        "text-[9px] font-mono text-cortex-text-muted/50 p-2";
      fallback.textContent = "Chart render error";
      container.appendChild(fallback);
    }

    return () => {
      container.replaceChildren();
    };
  }, [spec, compact, containerRef]);

  return <div className={`chart-container ${className ?? ""}`} ref={containerRef} />;
}

// ── Build Observable Plot options from MycelisChartSpec ──────

function buildPlotOptions(
  spec: MycelisChartSpec,
  width: number,
  height: number,
  compact: boolean,
): Plot.PlotOptions {
  const { chart_type, data, x, y, color, size, color_scheme, sort } = spec;

  const sortedData = sort
    ? [...data].sort((a, b) => {
        const aVal = a[sort.field];
        const bVal = b[sort.field];
        if (aVal == null || bVal == null) return 0;
        return sort.order === "asc"
          ? aVal > bVal
            ? 1
            : -1
          : aVal < bVal
            ? 1
            : -1;
      })
    : data;

  const defaultColor = "#7367F0"; // cortex-primary

  const base: Plot.PlotOptions = {
    width,
    height,
    x: compact ? { axis: null } : { label: spec.x_label ?? x ?? "" },
    y: compact ? { axis: null } : { label: spec.y_label ?? y ?? "" },
    color: color
      ? { legend: !compact, scheme: (color_scheme ?? "observable10") as Plot.ColorScheme }
      : undefined,
    marks: [],
  };

  switch (chart_type) {
    case "bar":
      base.marks = [
        Plot.barY(sortedData, {
          x: x!,
          y: y!,
          fill: color ?? defaultColor,
        }),
        Plot.ruleY([0]),
      ];
      break;

    case "line":
      base.marks = [
        Plot.lineY(sortedData, {
          x: x!,
          y: y!,
          stroke: color ?? defaultColor,
          strokeWidth: 2,
        }),
      ];
      break;

    case "area":
      base.marks = [
        Plot.areaY(sortedData, {
          x: x!,
          y: y!,
          fill: color ?? defaultColor,
          fillOpacity: 0.3,
        }),
        Plot.lineY(sortedData, {
          x: x!,
          y: y!,
          stroke: color ?? defaultColor,
          strokeWidth: 2,
        }),
      ];
      break;

    case "dot":
      base.marks = [
        Plot.dot(sortedData, {
          x: x!,
          y: y!,
          fill: color ?? defaultColor,
          r: size ? (d: Record<string, unknown>) => Number(d[size]) : 3,
        }),
      ];
      break;

    case "waffle":
      base.marks = [
        Plot.waffleY(sortedData, {
          x: x!,
          y: y!,
          fill: color ?? defaultColor,
        } as Record<string, unknown>),
      ];
      break;

    case "tree":
      base.marks = [
        Plot.tree(sortedData, {
          path: x,
          delimiter: "/",
          fill: color ?? defaultColor,
        } as Record<string, unknown>),
      ];
      base.axis = null;
      break;

    default:
      // Fallback to dot
      base.marks = [
        Plot.dot(sortedData, {
          x: x,
          y: y,
          fill: defaultColor,
        }),
      ];
  }

  return base;
}
