"use client";

import { toolLabel, toolOrigin } from "@/lib/labels";

export default function MissionControlToolsUsed({ tools }: { tools?: string[] }) {
    if (!tools?.length) return null;
    return (
        <div className="mt-0.5 flex flex-wrap gap-1 px-1">
            {tools.map((tool, index) => {
                const origin = toolOrigin(tool);
                return (
                    <span
                        key={`${tool}-${index}`}
                        className={`flex items-center gap-1 rounded border px-1.5 py-0.5 font-mono text-[7px] ${
                            origin === "external"
                                ? "border-amber-400/20 bg-amber-400/10 text-amber-400"
                                : origin === "sandboxed"
                                    ? "border-cortex-success/20 bg-cortex-success/10 text-cortex-success"
                                    : "border-cortex-primary/20 bg-cortex-primary/10 text-cortex-primary"
                        }`}
                        title={tool}
                    >
                        {toolLabel(tool)}
                        {origin === "external" && <span className="text-[6px] font-bold uppercase opacity-80">Ext</span>}
                        {origin === "sandboxed" && <span className="text-[6px] font-bold uppercase opacity-80">Box</span>}
                    </span>
                );
            })}
        </div>
    );
}
