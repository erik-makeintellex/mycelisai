import { Envelope } from './types/protocol';

export const MOCK_STREAM: Envelope[] = [
    {
        type: 'metric',
        source: 'system',
        timestamp: new Date().toISOString(),
        content: {
            label: 'Network Latency',
            value: '24ms',
            unit: '',
            status: 'nominal'
        }
    },
    {
        type: 'thought',
        source: 'architect-01',
        timestamp: new Date().toISOString(),
        content: {
            summary: 'Analyzing network topology',
            detail: 'Detected 3 disconnected nodes in the provisioner subnet. Recommending bridge activation.',
            model: 'qwen2.5-coder:7b'
        }
    },
    {
        type: 'governance',
        source: 'gatekeeper',
        timestamp: new Date().toISOString(),
        content: {
            request_id: 'req-123',
            agent_id: 'provisioner-01',
            description: 'Requesting egress access to 142.251.34.206 (google.com) for connectivity check.',
            action: 'NET_OPEN',
            status: 'pending'
        }
    },
    {
        type: 'artifact',
        source: 'coder-01',
        timestamp: new Date().toISOString(),
        content: {
            id: 'file-456',
            mime_type: 'text/markdown',
            title: 'TopologyRequest.md',
            preview: '# Topology Update\nRequesting bridge activation...',
            uri: '#'
        }
    }
];
