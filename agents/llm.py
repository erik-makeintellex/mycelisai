import abc
import json
import logging
from typing import List, Dict, Any, Optional
import aiohttp

logger = logging.getLogger(__name__)

class LLMClient(abc.ABC):
    @abc.abstractmethod
    async def generate(self, messages: List[Dict[str, str]], tools: Optional[List[Dict[str, Any]]] = None) -> str:
        """
        Generate a response from the LLM.
        
        Args:
            messages: List of message dicts (role, content).
            tools: Optional list of tool definitions.
            
        Returns:
            The generated text response.
        """
        pass

class OllamaClient(LLMClient):
    def __init__(self, model: str = "llama3", base_url: str = "http://localhost:11434"):
        self.model = model
        self.base_url = base_url

    async def generate(self, messages: List[Dict[str, str]], tools: Optional[List[Dict[str, Any]]] = None) -> str:
        url = f"{self.base_url}/api/chat"
        
        # Convert standard messages to Ollama format if needed (Ollama uses role/content standard)
        payload = {
            "model": self.model,
            "messages": messages,
            "stream": False
        }

        if tools:
            # Note: Ollama tool support varies by model/version. 
            # For now, we'll assume the model supports it or ignore if not strict.
            # This is a placeholder for actual tool integration.
            pass

        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(url, json=payload) as response:
                    if response.status != 200:
                        text = await response.text()
                        logger.error(f"Ollama API error: {response.status} - {text}")
                        return "Error: Failed to call LLM."
                    
                    data = await response.json()
                    return data.get("message", {}).get("content", "")
        except Exception as e:
            logger.error(f"Failed to connect to Ollama: {e}")
            return f"Error: {str(e)}"

class MockLLMClient(LLMClient):
    async def generate(self, messages: List[Dict[str, str]], tools: Optional[List[Dict[str, Any]]] = None) -> str:
        return "This is a mock response from the agent."
