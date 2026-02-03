import { streamText } from 'ai';

// Mock response generation
export async function POST(req: Request) {
    const { messages } = await req.json();

    // Simple mock stream using basic response for now as we don't have OpenAI keys set up in this env yet
    // We'll simulate a stream response

    const response = "I am the Operator Console. I am listening.";

    // Create a stream of characters
    const stream = new ReadableStream({
        async start(controller) {
            const encoder = new TextEncoder();
            for (const char of response) {
                controller.enqueue(encoder.encode(char));
                await new Promise(r => setTimeout(r, 20)); // Simulate typing
            }
            controller.close();
        }
    });

    return new Response(stream);
}
