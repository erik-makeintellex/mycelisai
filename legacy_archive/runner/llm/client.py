from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional, Union
from pydantic import BaseModel

class ToolCall(BaseModel):
    id: str
    name: str
    arguments: Dict[str, Any]

class LLMResponse(BaseModel):
    content: Optional[str] = None
    tool_calls: Optional[List[ToolCall]] = None

class LLMClient(ABC):
    """Abstract base class for LLM providers."""
    
    @abstractmethod
    async def generate(self, 
                       model: str, 
                       messages: List[Dict[str, Any]], 
                       system_prompt: Optional[str] = None,
                       tools: Optional[List[Dict[str, Any]]] = None) -> LLMResponse:
        """
        Generate a response from the LLM.
        
        Args:
            model: Model identifier (e.g. 'llama3', 'gpt-4o')
            messages: List of message dicts {'role': 'user', 'content': '...'}
            system_prompt: Optional system instruction override
            tools: Optional list of JSON Schema tool definitions
            
        Returns:
            LLMResponse object containing content and/or tool_calls
        """
        pass
