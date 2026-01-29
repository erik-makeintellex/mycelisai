from typing import List, Optional
from pydantic import BaseModel
import time

class ContextMessage(BaseModel):
    role: str
    content: str
    timestamp: float = 0.0

class AgentContext:
    def __init__(self, max_history: int = 20, system_prompt: Optional[str] = None):
        self.max_history = max_history
        self._system_prompt = system_prompt
        self._history: List[ContextMessage] = []

    def set_system_prompt(self, prompt: str):
        self._system_prompt = prompt

    def add_message(self, role: str, content: str):
        msg = ContextMessage(role=role, content=content, timestamp=time.time())
        self._history.append(msg)
        
        # Trim history if needed
        # We might want to keep the system prompt virtual/separate (handled in LLMClient)
        if len(self._history) > self.max_history:
            self._history = self._history[-self.max_history:]

    def clear_history(self):
        self._history = []

    @property
    def history(self) -> List[ContextMessage]:
        return self._history

    def get_prompt(self) -> str:
        """Helper for simple completion APIs"""
        prompt = ""
        if self._system_prompt:
            prompt += f"System: {self._system_prompt}\n\n"
        
        for msg in self._history:
            prompt += f"{msg.role.capitalize()}: {msg.content}\n"
            
        return prompt
