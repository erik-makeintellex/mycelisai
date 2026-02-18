/** MycelisChartSpec — the visualization contract between agents and the frontend.
 *  Agents produce this as artifact content for type="chart".
 *  The frontend detects it and renders interactively via Observable Plot / react-leaflet. */
export interface MycelisChartSpec {
  // ── Required ──────────────────────────────
  version: string;
  chart_type: ChartType;
  title: string;
  data: Record<string, unknown>[];

  // ── Column Encoding ───────────────────────
  x?: string;
  y?: string;
  color?: string;
  size?: string;
  label?: string;

  // ── Optional Styling ──────────────────────
  x_label?: string;
  y_label?: string;
  color_scheme?: string;
  sort?: { field: string; order: "asc" | "desc" };
  width?: number;
  height?: number;

  // ── Geo-Specific (chart_type === "geo") ───
  geo?: GeoConfig;
}

export type ChartType =
  | "bar"
  | "line"
  | "area"
  | "dot"
  | "geo"
  | "table"
  | "waffle"
  | "tree";

export interface GeoConfig {
  type: "points" | "choropleth";
  lat_field?: string;
  lon_field?: string;
  value_field?: string;
  geojson_url?: string;
  center?: [number, number];
  zoom?: number;
}
