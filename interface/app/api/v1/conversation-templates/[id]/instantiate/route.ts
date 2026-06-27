import { proxyBackendPostRequest } from '../../../../_chatProxy';

export async function POST(req: Request, { params }: { params: Promise<{ id: string }> }) {
    const { id } = await params;
    return proxyBackendPostRequest(req, {
        targetLabel: 'Soma quick action',
        path: `/api/v1/conversation-templates/${encodeURIComponent(id)}/instantiate`,
    });
}
