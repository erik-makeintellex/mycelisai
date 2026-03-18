"use client";

import { use } from "react";
import OrganizationContextShell from "@/components/organizations/OrganizationContextShell";

export default function OrganizationPage({
    params,
}: {
    params: Promise<{ id: string }>;
}) {
    const { id } = use(params);
    return <OrganizationContextShell organizationId={id} />;
}
