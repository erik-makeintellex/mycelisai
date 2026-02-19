"use client";

import React, { useEffect, useState } from "react";
import {
  MapContainer,
  TileLayer,
  Marker,
  Popup,
  GeoJSON,
} from "react-leaflet";
import L from "leaflet";
import type { MycelisChartSpec } from "./types";

// Fix Leaflet default marker icon paths (broken by bundlers)
const defaultIcon = L.icon({
  iconUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png",
  iconRetinaUrl:
    "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png",
  shadowUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png",
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

interface MapRendererProps {
  spec: MycelisChartSpec;
  compact?: boolean;
  className?: string;
}

const DARK_TILES =
  "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png";
const TILE_ATTRIBUTION =
  '&copy; <a href="https://carto.com/">CARTO</a>';

export default function MapRenderer({
  spec,
  compact = false,
  className,
}: MapRendererProps) {
  const geo = spec.geo;
  if (!geo) return null;

  const center: [number, number] = geo.center ?? [39.0, -98.0];
  const zoom = geo.zoom ?? 4;
  const height = compact ? 120 : spec.height ?? 400;

  if (geo.type === "points") {
    return (
      <PointsMap
        spec={spec}
        center={center}
        zoom={zoom}
        height={height}
        compact={compact}
        className={className}
      />
    );
  }

  if (geo.type === "choropleth" && geo.geojson_url) {
    return (
      <ChoroplethMap
        spec={spec}
        center={center}
        zoom={zoom}
        height={height}
        compact={compact}
        className={className}
      />
    );
  }

  return (
    <div className="text-[9px] font-mono text-cortex-text-muted/50 p-2">
      Unsupported geo type
    </div>
  );
}

// ── Points Map ──────────────────────────────────────

function PointsMap({
  spec,
  center,
  zoom,
  height,
  compact,
  className,
}: {
  spec: MycelisChartSpec;
  center: [number, number];
  zoom: number;
  height: number;
  compact: boolean;
  className?: string;
}) {
  const geo = spec.geo!;
  const latField = geo.lat_field ?? "lat";
  const lonField = geo.lon_field ?? "lon";

  return (
    <div className={className} style={{ height }}>
      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: "100%", width: "100%", borderRadius: "0.5rem" }}
        zoomControl={!compact}
        scrollWheelZoom={!compact}
        dragging={!compact}
      >
        <TileLayer url={DARK_TILES} attribution={TILE_ATTRIBUTION} />
        {spec.data.map((row, i) => {
          const lat = Number(row[latField]);
          const lon = Number(row[lonField]);
          if (isNaN(lat) || isNaN(lon)) return null;

          return (
            <Marker key={i} position={[lat, lon]} icon={defaultIcon}>
              {!compact && (
                <Popup>
                  <div className="text-xs font-mono space-y-0.5">
                    {Object.entries(row).map(([k, v]) => (
                      <div key={k}>
                        <span className="font-bold">{k}:</span> {String(v)}
                      </div>
                    ))}
                  </div>
                </Popup>
              )}
            </Marker>
          );
        })}
      </MapContainer>
    </div>
  );
}

// ── Choropleth Map ──────────────────────────────────

function ChoroplethMap({
  spec,
  center,
  zoom,
  height,
  compact,
  className,
}: {
  spec: MycelisChartSpec;
  center: [number, number];
  zoom: number;
  height: number;
  compact: boolean;
  className?: string;
}) {
  const geo = spec.geo!;
  const [geoData, setGeoData] = useState<GeoJSON.FeatureCollection | null>(
    null,
  );

  useEffect(() => {
    if (!geo.geojson_url) return;
    fetch(geo.geojson_url)
      .then((r) => r.json())
      .then(setGeoData)
      .catch((err) => console.error("[MapRenderer] GeoJSON fetch failed:", err));
  }, [geo.geojson_url]);

  const valueField = geo.value_field ?? "value";

  // Build a lookup from data rows by matching key
  const valueLookup = new Map<string, number>();
  for (const row of spec.data) {
    const key = String(row[spec.x ?? "name"] ?? row[spec.label ?? "id"] ?? "");
    const val = Number(row[valueField]);
    if (key && !isNaN(val)) valueLookup.set(key, val);
  }

  const maxVal = Math.max(...valueLookup.values(), 1);

  const getStyle = (feature?: GeoJSON.Feature) => {
    const name = feature?.properties?.name ?? feature?.properties?.NAME ?? "";
    const val = valueLookup.get(name) ?? 0;
    const intensity = Math.min(val / maxVal, 1);
    return {
      fillColor: `rgba(6, 182, 212, ${0.15 + intensity * 0.7})`,
      fillOpacity: 0.8,
      color: "#27272a",
      weight: 1,
    };
  };

  if (!geoData) {
    return (
      <div
        className={`flex items-center justify-center ${className ?? ""}`}
        style={{ height }}
      >
        <span className="text-[9px] font-mono text-cortex-text-muted/50 animate-pulse">
          Loading boundaries...
        </span>
      </div>
    );
  }

  return (
    <div className={className} style={{ height }}>
      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: "100%", width: "100%", borderRadius: "0.5rem" }}
        zoomControl={!compact}
        scrollWheelZoom={!compact}
        dragging={!compact}
      >
        <TileLayer url={DARK_TILES} attribution={TILE_ATTRIBUTION} />
        <GeoJSON data={geoData} style={getStyle} />
      </MapContainer>
    </div>
  );
}
