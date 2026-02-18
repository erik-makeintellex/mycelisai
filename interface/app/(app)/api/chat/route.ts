// Proxy to Core backend for chat inference.
// In local dev, /api/v1/chat is rewritten by next.config.mjs.
// This route handles /api/chat as a secondary path.
export async function POST(req: Request) {
    const { messages } = await req.json();

    const host = process.env.MYCELIS_API_HOST ?? "localhost";
    const port = process.env.MYCELIS_API_PORT ?? "8081";

    try {
        const response = await fetch(`http://${host}:${port}/api/v1/chat`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ messages }),
        });

        if (!response.ok) {
            return new Response("Cognitive Uplink Invalid", { status: 503 });
        }

        return new Response(response.body);
    } catch {
        return new Response("Core Unreachable", { status: 503 });
    }
}
