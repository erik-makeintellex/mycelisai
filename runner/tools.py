import inspect
from typing import Callable, Dict, Any, List
from pydantic import create_model

class ToolRegistry:
    def __init__(self):
        self._tools: Dict[str, Callable] = {}
        self._definitions: List[Dict[str, Any]] = []

    def register(self, func: Callable):
        """Register a function as a tool."""
        name = func.__name__
        self._tools[name] = func
        
        # Generate JSON Schema for the function
        # We assume clean python typing
        # Simplified generation for V1
        schema = self._generate_schema(func)
        self._definitions.append({
            "type": "function",
            "function": schema
        })
        return func

    def _generate_schema(self, func: Callable) -> Dict[str, Any]:
        """Generate OpenAI/Ollama compatible function schema from docstring and type hints."""
        sig = inspect.signature(func)
        doc = inspect.getdoc(func) or "No description."
        
        parameters = {
            "type": "object",
            "properties": {},
            "required": []
        }
        
        for name, param in sig.parameters.items():
            if name == "self": continue
            
            # Helper to map python types to JSON types
            type_map = {
                str: "string",
                int: "integer",
                float: "number",
                bool: "boolean",
                list: "array",
                dict: "object"
            }
            
            param_type = type_map.get(param.annotation, "string")
            parameters["properties"][name] = {
                "type": param_type,
                "description": f"Parameter {name}" # Ideally parse from docstring
            }
            if param.default == inspect.Parameter.empty:
                parameters["required"].append(name)
                
        return {
            "name": func.__name__,
            "description": doc,
            "parameters": parameters
        }

    @property
    def definitions(self) -> List[Dict[str, Any]]:
        return self._definitions
        
    async def execute(self, name: str, arguments: Dict[str, Any]) -> Any:
        if name not in self._tools:
            raise ValueError(f"Tool {name} not found")
        
        func = self._tools[name]
        try:
             # Check if async
            if inspect.iscoroutinefunction(func):
                return await func(**arguments)
            else:
                return func(**arguments)
        except Exception as e:
            return f"Error executing tool {name}: {str(e)}"
