from shared.schemas import TextMessage
import json

payload = {
    "content": "hello",
    "id": "1",
    "source_agent_id": "user",
    "type": "text"
}

try:
    msg = TextMessage(**payload)
    print("Validation Successful")
    print(msg.model_dump_json())
except Exception as e:
    print(f"Validation Failed: {e}")
