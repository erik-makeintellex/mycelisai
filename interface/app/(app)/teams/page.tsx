import { redirect } from 'next/navigation';

export default function TeamsRedirect() {
    redirect('/automations?tab=teams');
}
