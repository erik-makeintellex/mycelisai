import { proxyChatRequest } from '../../../../_chatProxy';

export async function POST(req: Request, context: { params: Promise<{ member: string }> }) {
    const { member } = await context.params;
    const targetLabel = member ? `Council member ${member}` : 'Council member';
    return proxyChatRequest(req, {
        targetLabel,
        path: `/api/v1/council/${member}/chat`,
    });
}
