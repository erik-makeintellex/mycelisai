import { streamText } from 'ai';

// Mock response generation
export async function POST(req: Request) {
    const { messages } = await req.json();

    try {
        const response = await fetch("http://mycelis-core:8080/api/v1/chat", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ messages })
        });

        if (!response.ok) {
            return new Response("Cognitive Uplink Invalid", { status: 503 });
        }

        return new Response(response.body);
    } catch (e) {
        return new Response("Core Unreachable", { status: 503 });
    }
}
