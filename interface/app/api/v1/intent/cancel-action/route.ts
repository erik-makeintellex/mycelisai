import { proxyBackendPostRequest } from '../../../_chatProxy';

export async function POST(req: Request) {
    return proxyBackendPostRequest(req, {
        targetLabel: 'Soma proposal cancellation',
        path: '/api/v1/intent/cancel-action',
    });
}
