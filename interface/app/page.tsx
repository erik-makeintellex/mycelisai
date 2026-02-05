"use client"
import { GenesisTerminal } from "@/components/genesis/GenesisTerminal"

export default function CommandPage() {
  // In a real app, we check if missions exist.
  // For Phase 3 Verification, we mount Genesis directly.
  return <GenesisTerminal />
}
