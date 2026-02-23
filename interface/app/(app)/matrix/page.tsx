import { redirect } from 'next/navigation';

export default function MatrixRedirect() {
    redirect('/system?tab=matrix');
}
