import { proxyBackendDeleteRequest, proxyBackendPatchRequest } from '../../../../_chatProxy';

type RouteContext = {
    params: Promise<{ id: string }> | { id: string };
};

async function sourcePath(context: RouteContext) {
    const params = await context.params;
    return `/api/v1/search/sources/${encodeURIComponent(params.id)}`;
}

export async function PATCH(req: Request, context: RouteContext) {
    return proxyBackendPatchRequest(req, {
        targetLabel: 'Search source',
        path: await sourcePath(context),
    });
}

export async function DELETE(req: Request, context: RouteContext) {
    return proxyBackendDeleteRequest(req, {
        targetLabel: 'Search source',
        path: await sourcePath(context),
    });
}
