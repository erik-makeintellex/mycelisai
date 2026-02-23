import { redirect } from 'next/navigation';

export default function ArchitectRedirect() {
    redirect('/automations?tab=wiring');
}
