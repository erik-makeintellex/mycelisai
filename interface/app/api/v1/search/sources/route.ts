import { proxyBackendGetRequest, proxyBackendPostRequest } from '../../../_chatProxy';

export async function GET(req: Request) {
    return proxyBackendGetRequest(req, {
        targetLabel: 'Search sources',
        path: '/api/v1/search/sources',
    });
}

export async function POST(req: Request) {
    return proxyBackendPostRequest(req, {
        targetLabel: 'Search sources',
        path: '/api/v1/search/sources',
    });
}
