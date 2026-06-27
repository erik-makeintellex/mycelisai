import { proxyBackendGetRequest, proxyBackendPostRequest } from '../../_chatProxy';

export async function GET(req: Request) {
    return proxyBackendGetRequest(req, {
        targetLabel: 'Soma quick actions',
        path: '/api/v1/conversation-templates',
    });
}

export async function POST(req: Request) {
    return proxyBackendPostRequest(req, {
        targetLabel: 'Soma quick actions',
        path: '/api/v1/conversation-templates',
    });
}
