import httpx
import json
import logging
from typing import List, Dict, Any, Optional
from .client import LLMClient, LLMResponse, ToolCall

log = logging.getLogger("runner.llm.ollama")

class OllamaClient(LLMClient):
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")
        
    async def generate(self, 
                       model: str, 
                       messages: List[Dict[str, Any]], 
                       system_prompt: Optional[str] = None,
                       tools: Optional[List[Dict[str, Any]]] = None) -> LLMResponse:
        
        # Prepare Ollama Request
        url = f"{self.base_url}/api/chat"
        
        # Construct payload
        # Note: Ollama expects 'tools' key for native function calling (supported in newer versions usually)
        # We will try native 'tools' param first.
        
        payload = {
            "model": model,
            "messages": messages,
            "stream": False # Streaming complicates tool parsing, sticking to non-stream for now or we build a stream parser later
        }
        
        if system_prompt:
             # Prepend or override system prompt
             # Ideally the messages list already contains a system message if needed, 
             # but if provided separately we prepend it.
             payload["messages"] = [{"role": "system", "content": system_prompt}] + payload["messages"]

        if tools:
            payload["tools"] = tools
            
        log.info("ollama_request", model=model, tools_count=len(tools) if tools else 0)

        async with httpx.AsyncClient(timeout=60.0) as client:
            try:
                resp = await client.post(url, json=payload)
                resp.raise_for_status()
                data = resp.json()
                
                message = data.get("message", {})
                content = message.get("content", "")
                tool_calls_raw = message.get("tool_calls", [])
                
                native_tool_calls = []
                if tool_calls_raw:
                    for tc in tool_calls_raw:
                        # Ollama format: {'function': {'name': '...', 'arguments': {...}}}
                        func = tc.get("function", {})
                        native_tool_calls.append(ToolCall(
                            id=f"call_{len(native_tool_calls)}", # Ollama doesn't always give IDs
                            name=func.get("name"),
                            arguments=func.get("arguments", {})
                        ))
                        
                return LLMResponse(
                    content=content,
                    tool_calls=native_tool_calls if native_tool_calls else None
                )
                
            except Exception as e:
                log.error("ollama_generation_failed", error=str(e))
                # Fallback? No, just raise for now
                raise e
