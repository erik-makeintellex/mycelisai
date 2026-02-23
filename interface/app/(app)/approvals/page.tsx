import { redirect } from 'next/navigation';

export default function ApprovalsRedirect() {
    redirect('/automations?tab=approvals');
}
