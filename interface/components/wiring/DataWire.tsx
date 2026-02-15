"use client";

import React from 'react';
import { getBezierPath, type EdgeProps } from 'reactflow';

const typeColors: Record<string, string> = {
    input: '#2ea043',
    output: '#58a6ff',
    governance: '#f85149',
};

const defaultColor = '#58a6ff';

export default function DataWire({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    data,
    markerEnd,
    style = {},
}: EdgeProps) {
    const [edgePath] = getBezierPath({
        sourceX,
        sourceY,
        sourcePosition,
        targetX,
        targetY,
        targetPosition,
    });

    const color = typeColors[data?.type] ?? defaultColor;

    return (
        <>
            {/* Invisible wide path for easier selection */}
            <path
                d={edgePath}
                fill="none"
                stroke="transparent"
                strokeWidth={20}
                className="react-flow__edge-interaction"
            />

            {/* Base path (faint glow) */}
            <path
                d={edgePath}
                fill="none"
                stroke={color}
                strokeWidth={3}
                strokeOpacity={0.15}
                markerEnd={markerEnd}
            />

            {/* Animated dashed path */}
            <path
                d={edgePath}
                fill="none"
                stroke={color}
                strokeWidth={1.5}
                strokeDasharray="6 4"
                strokeLinecap="round"
                markerEnd={markerEnd}
                style={{
                    ...style,
                    animation: `datawire-dash 1.5s linear infinite`,
                }}
            />

            {/* Animated circle traveling along the path */}
            <circle r={3} fill={color} opacity={0.9}>
                <animateMotion
                    dur="3s"
                    repeatCount="indefinite"
                    path={edgePath}
                />
            </circle>

            {/* Keyframe style injection (scoped once via ID) */}
            <style>
                {`
                    @keyframes datawire-dash {
                        to {
                            stroke-dashoffset: -20;
                        }
                    }
                `}
            </style>
        </>
    );
}

export const edgeTypes = {
    dataWire: DataWire,
};
