import AccessDenied from "@/components/access/AccessDenied";

export default function AccessDeniedPage() {
    return (
        <AccessDenied
            requiredAccess="Workspace member, administrator, or owner"
            supportHref="/settings?tab=users"
        />
    );
}
