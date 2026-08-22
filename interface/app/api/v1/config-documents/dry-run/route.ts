import { proxyBackendPostRequest } from '../../../_chatProxy';

export async function POST(req: Request) {
    return proxyBackendPostRequest(req, {
        targetLabel: 'Configuration preview',
        path: '/api/v1/config-documents/dry-run',
    });
}
