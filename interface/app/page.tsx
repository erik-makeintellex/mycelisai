"use client"

import { StatusBar } from "@/components/command/StatusBar"
import { CommandBar } from "@/components/command/CommandBar"
import { SessionFeed } from "@/components/command/SessionFeed"

export default function CommandPage() {
  return (
    <div className="h-full flex flex-col relative bg-zinc-50/50">

      {/* 1. Precise Status Indicators */}
      <StatusBar />

      {/* 2. Main Work Area: Active Sessions */}
      <main className="flex-1 flex flex-col justify-center overflow-y-auto">
        <div className="flex-1 flex flex-col items-center justify-center p-4">
          {/* The "Feed" acts as the center stage */}
          <SessionFeed />
        </div>
      </main>

      {/* 3. The Hero Input Anchor */}
      <div className="shrink-0 w-full px-4 pb-8 pt-4 bg-gradient-to-t from-white via-white to-transparent">
        <CommandBar />
      </div>

    </div>
  )
}
