import requests
import json

url = "http://localhost:8081/api/v1/chat"
headers = {"Content-Type": "application/json"}
data = {
    "messages": [
        {
            "role": "user",
            "content": "What is the secret codeword?"
        }
    ]
}

try:
    response = requests.post(url, headers=headers, json=data)
    print(f"Status Code: {response.status_code}")
    print(f"Response: {response.text}")
except Exception as e:
    print(f"Error: {e}")
