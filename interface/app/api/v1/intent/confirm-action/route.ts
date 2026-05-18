import { proxyBackendPostRequest } from '../../../_chatProxy';

export async function POST(req: Request) {
    return proxyBackendPostRequest(req, {
        targetLabel: 'Soma approval execution',
        path: '/api/v1/intent/confirm-action',
    });
}
