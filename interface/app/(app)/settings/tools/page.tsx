"use client";

import { redirect } from "next/navigation";

export default function ToolsRoute() {
    redirect("/settings?tab=tools");
}
