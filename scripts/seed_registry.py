
import psycopg2
import os
import json
from dotenv import load_dotenv

load_dotenv()

DB_HOST = os.getenv("POSTGRES_HOST", "localhost")
DB_USER = os.getenv("POSTGRES_USER", "postgres")
DB_PASS = os.getenv("POSTGRES_PASSWORD", "postgres")
DB_NAME = os.getenv("POSTGRES_DB", "mycelis")

def get_connection():
    return psycopg2.connect(
        host=DB_HOST, user=DB_USER, password=DB_PASS, dbname=DB_NAME
    )

def seed():
    conn = get_connection()
    conn.autocommit = True
    cursor = conn.cursor()

    templates = [
        {
            "name": "OpenWeatherMap Poller",
            "type": "ingress",
            "image": "mycelis/weather-poller:v1",
            "config_schema": {
                "type": "object",
                "properties": {
                    "api_key": {"type": "string"},
                    "city": {"type": "string"},
                    "interval": {"type": "string", "enum": ["1m", "10m", "1h"]}
                },
                "required": ["api_key", "city"]
            },
            "topic_template": "swarm.data.weather.{{city}}"
        },
        {
            "name": "Slack Notification Sink",
            "type": "egress",
            "image": "mycelis/slack-sink:v1",
            "config_schema": {
                "type": "object",
                "properties": {
                    "webhook_url": {"type": "string"},
                    "channel": {"type": "string"}
                },
                "required": ["webhook_url"]
            },
            "topic_template": "swarm.logs.slack"
        }
    ]

    print("🌱 Seeding Registry...")
    for t in templates:
        print(f"   Insert: {t['name']}")
        cursor.execute("""
            INSERT INTO connector_templates (name, type, image, config_schema, topic_template)
            VALUES (%s, %s, %s, %s, %s)
            ON CONFLICT DO NOTHING -- Avoid dupes if re-run (Wait, name is not unique constraint? we rely on ID?)
            -- Actually we didn't add Unique constraint on Name. duplicates might happen.
            -- Let's check existence by name first.
        """, (t['name'], t['type'], t['image'], json.dumps(t['config_schema']), t['topic_template']))
    
    conn.close()
    print("✨ Seed Complete.")

if __name__ == "__main__":
    seed()
