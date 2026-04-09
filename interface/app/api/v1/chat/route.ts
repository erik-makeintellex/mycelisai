import { proxyChatRequest } from '../../_chatProxy';

export async function POST(req: Request) {
    return proxyChatRequest(req, {
        targetLabel: 'Soma',
        path: '/api/v1/chat',
    });
}
