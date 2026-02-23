/**
 * OpsWidget Registry
 *
 * A lightweight plugin API for OpsOverview dashboard sections.
 * Register a widget once; OpsOverview renders from the registry automatically.
 *
 * Usage — adding a new widget:
 *   1. Create your React component (no special base class needed).
 *   2. Call registerOpsWidget() once at module level in OpsOverview.tsx.
 *   3. Done — no modifications to the OpsOverview render logic needed.
 *
 * Layout types:
 *   'grid'      — Placed in the responsive auto-fit grid at the top (compact cards).
 *   'fullWidth' — Rendered full-width below the grid (for list-heavy sections).
 */

import type { ComponentType } from "react";

export type WidgetLayout = "grid" | "fullWidth";

export interface OpsWidgetConfig {
    /** Unique identifier — prevents double-registration on HMR. */
    id: string;
    /**
     * Render order (ascending). Built-in widgets use multiples of 10
     * so third-party widgets can slot between them (e.g. order: 15).
     */
    order: number;
    /** Controls where in the layout the widget renders. */
    layout: WidgetLayout;
    /** The React component to render. Must handle its own data fetching. */
    Component: ComponentType;
}

const _registry: OpsWidgetConfig[] = [];

/**
 * Register an OpsOverview widget.
 * Safe to call multiple times with the same id (idempotent — HMR safe).
 */
export function registerOpsWidget(config: OpsWidgetConfig): void {
    if (_registry.some((w) => w.id === config.id)) return;
    _registry.push(config);
    _registry.sort((a, b) => a.order - b.order);
}

/**
 * Returns a sorted snapshot of all registered widgets.
 * Called by OpsOverview at render time.
 */
export function getOpsWidgets(): OpsWidgetConfig[] {
    return [..._registry];
}

/**
 * Removes a widget by id (useful in tests or for conditional features).
 */
export function unregisterOpsWidget(id: string): void {
    const idx = _registry.findIndex((w) => w.id === id);
    if (idx !== -1) _registry.splice(idx, 1);
}
