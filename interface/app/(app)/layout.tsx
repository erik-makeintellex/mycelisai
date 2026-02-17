import { ShellLayout } from "@/components/shell/ShellLayout";

export default function AppLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <ShellLayout>
            {children}
        </ShellLayout>
    );
}
