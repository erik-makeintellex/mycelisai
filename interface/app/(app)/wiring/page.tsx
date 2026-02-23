import { redirect } from 'next/navigation';

export default function WiringRedirect() {
    redirect('/automations?tab=wiring');
}
