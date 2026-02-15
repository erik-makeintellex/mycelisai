-- 1. Hardware Target: HackRF
INSERT INTO connector_templates (name, type, image, config_schema, topic_template)
VALUES (
        'HackRF Listener',
        'ingress',
        'mycelis/hackrf:latest',
        '{
        "type": "object",
        "properties": {
            "frequency": { 
                "type": "string", 
                "title": "Center Frequency (Hz)", 
                "default": "100000000",
                "description": "Target frequency to tune the radio to."
            },
            "bandwidth": { 
                "type": "string", 
                "title": "Bandwidth (Hz)", 
                "default": "2000000",
                "description": "Sampling bandwidth."
            },
            "gain": { 
                "type": "integer", 
                "title": "VGA Gain (dB)", 
                "default": 30,
                "minimum": 0,
                "maximum": 62
            }
        },
        "required": ["frequency"]
    }',
        'swarm.radio.hackrf'
    ) ON CONFLICT DO NOTHING;
-- 2. API Target: HTTP Poller
INSERT INTO connector_templates (name, type, image, config_schema, topic_template)
VALUES (
        'REST API Poller',
        'ingress',
        'mycelis/http-poller:latest',
        '{
        "type": "object",
        "properties": {
            "url": { 
                "type": "string", 
                "title": "Target URL", 
                "description": "HTTPS Endpoint to poll."
            },
            "interval": { 
                "type": "integer", 
                "title": "Interval (seconds)", 
                "default": 60,
                "minimum": 1
            },
            "method": { 
                "type": "string", 
                "title": "HTTP Method", 
                "default": "GET",
                "enum": ["GET", "POST"]
            }
        },
        "required": ["url"]
    }',
        'swarm.api.poller'
    ) ON CONFLICT DO NOTHING;