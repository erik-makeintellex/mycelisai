import { proxyBackendGetRequest } from '../../../_chatProxy';

export async function GET(req: Request) {
    return proxyBackendGetRequest(req, {
        targetLabel: 'Search status',
        path: '/api/v1/search/status',
    });
}
