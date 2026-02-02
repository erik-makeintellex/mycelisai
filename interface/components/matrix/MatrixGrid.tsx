"use client";

import { useState } from "react";
import { DataTable } from "./data-table";
import { columns, CognitiveProfile } from "./columns";
import { BrainCircuit } from "lucide-react";

// Mock Data for MVP
const MOCK_DATA: CognitiveProfile[] = [
    {
        id: "coder",
        activeModel: "qwen2.5:7b",
        provider: "ollama",
        status: "online",
        costPerTk: "0.00",
        temperature: 0.1
    },
    {
        id: "chat",
        activeModel: "gpt-4-turbo",
        provider: "openai",
        status: "online",
        costPerTk: "$0.03",
        temperature: 0.7
    },
    {
        id: "logic",
        activeModel: "deepseek-r1",
        provider: "openai",
        status: "offline",
        costPerTk: "$0.005",
        temperature: 0.0
    }
]

export function MatrixGrid() {
    const [data] = useState<CognitiveProfile[]>(MOCK_DATA);

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                        <BrainCircuit className="text-emerald-500" />
                        Cognitive Matrix
                    </h2>
                    <p className="text-zinc-500">Manage neural resources and model routing.</p>
                </div>
            </div>

            <DataTable columns={columns} data={data} />
        </div>
    )
}
