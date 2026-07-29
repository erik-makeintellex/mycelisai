import { proxyBackendGetRequest } from '@/app/api/_chatProxy';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

export async function GET(request: Request) {
    return proxyBackendGetRequest(request, {
        targetLabel: 'Live workspace updates',
        path: '/api/v1/stream',
    });
}
