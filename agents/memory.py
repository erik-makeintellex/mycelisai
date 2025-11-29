from typing import List, Dict, Any
from collections import deque

class AgentContext:
    def __init__(self, system_prompt: str, max_history: int = 10):
        self.system_prompt = system_prompt
        self.max_history = max_history
        self.history: deque = deque(maxlen=max_history)

    def add_message(self, role: str, content: str):
        self.history.append({"role": role, "content": content})

    def get_messages(self) -> List[Dict[str, str]]:
        messages = [{"role": "system", "content": self.system_prompt}]
        messages.extend(list(self.history))
        return messages

    def clear_history(self):
        self.history.clear()
