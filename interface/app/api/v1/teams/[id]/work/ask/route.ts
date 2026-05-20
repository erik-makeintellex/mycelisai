import { proxyBackendPostRequest } from "../../../../../_chatProxy";

export const maxDuration = 120;

export async function POST(req: Request, context: { params: Promise<{ id: string }> }) {
  const { id } = await context.params;
  const teamId = encodeURIComponent(id);
  return proxyBackendPostRequest(req, {
    targetLabel: `Team ${id} ask lane`,
    path: `/api/v1/teams/${teamId}/work/ask`,
  });
}
