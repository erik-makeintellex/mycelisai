const axios = require('axios');

// Configuration
const API_URL = process.env.API_URL || 'http://localhost:8000';
const CHANNEL = 'external.node.events';
const SERVICE_NAME = 'node-service-01';

async function sendEvent(content) {
    const payload = {
        id: `evt-${Date.now()}`,
        source_agent_id: SERVICE_NAME,
        type: 'event',
        payload: {
            content: content,
            uptime: process.uptime()
        },
        timestamp: new Date().toISOString()
    };

    try {
        console.log(`Sending event to ${CHANNEL}...`);
        const res = await axios.post(`${API_URL}/ingest/${CHANNEL}`, payload);
        console.log(`✅ Success: ${res.status} - ${res.data.id}`);
    } catch (error) {
        console.error(`❌ Error: ${error.message}`);
        if (error.response) {
            console.error(error.response.data);
        }
    }
}

// Main Loop
async function main() {
    console.log(`Starting ${SERVICE_NAME}...`);

    // Simulate periodic events
    setInterval(() => {
        sendEvent('Hello from Node.js!');
    }, 5000);

    // Send one immediately
    await sendEvent('Node.js Service Started');
}

main();
