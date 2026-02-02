"use client"

import { ColumnDef } from "@tanstack/react-table"
import { MoreHorizontal, Cpu, Cloud, Activity } from "lucide-react"

export type CognitiveProfile = {
    id: string
    activeModel: string
    provider: "ollama" | "openai"
    status: "online" | "offline" | "degraded"
    costPerTk: string
    temperature: number
}

export const columns: ColumnDef<CognitiveProfile>[] = [
    {
        accessorKey: "id",
        header: "Profile ID",
        cell: ({ row }) => {
            return (
                <div className="flex items-center gap-2 font-mono text-emerald-500">
                    <BrainIcon id={row.getValue("id")} />
                    <span className="uppercase">{row.getValue("id")}</span>
                </div>
            )
        }
    },
    {
        accessorKey: "activeModel",
        header: "Active Model",
        cell: ({ row }) => {
            const provider = row.original.provider;
            return (
                <div className="flex items-center gap-2">
                    {provider === "ollama" ? <Cpu className="w-4 h-4 text-orange-500" /> : <Cloud className="w-4 h-4 text-blue-500" />}
                    <span className="text-zinc-300">{row.getValue("activeModel")}</span>
                </div>
            )
        }
    },
    {
        accessorKey: "status",
        header: "Health",
        cell: ({ row }) => {
            const status = row.original.status
            return (
                <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${status === "online" ? "bg-emerald-500 animate-pulse" : "bg-red-500"}`} />
                    <span className="capitalize text-xs text-zinc-500">{status}</span>
                </div>
            )
        }
    },
    {
        accessorKey: "temperature",
        header: "Temp",
        cell: ({ row }) => <span className="font-mono text-zinc-400">{row.getValue("temperature")}</span>
    },
]

function BrainIcon({ id }: { id: string }) {
    // Just a visual helper
    return <Activity className="w-4 h-4" />
}
