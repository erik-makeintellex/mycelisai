import structlog
import logging
import sys
import os
from contextvars import ContextVar
from typing import Optional, Dict, Any, Union

# Context Variables
_trace_id_ctx = ContextVar("trace_id", default=None)
_agent_id_ctx = ContextVar("agent_id", default=None)

def add_context_vars(logger, method_name, event_dict):
    """
    Structlog processor to inject context variables (trace_id, agent_id).
    """
    trace_id = _trace_id_ctx.get()
    if trace_id:
        event_dict["trace_id"] = trace_id
    
    agent_id = _agent_id_ctx.get()
    if agent_id:
        event_dict.setdefault("agent", {})["id"] = agent_id
        
    return event_dict

def setup_logger(service_name: str, level: str = "INFO", json_logs: bool = False):
    """
    Configure structlog and standard logging.
    """
    # 1. Configure Standard Logging (for libraries like uvicorn)
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=level.upper(),
    )
    
    # Silence noisy libs
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)

    # 2. Configure Structlog Processors
    processors = [
        structlog.contextvars.merge_contextvars,
        add_context_vars,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
    ]

    # 3. Choose Renderer (JSON or Console)
    # Using JSON for robust processing as requested, but ConsoleRenderer is nicer for dev
    # We will stick to JSON if requested or if in "prod", otherwise nice console
    
    # User requested "agent focused but still human usable". 
    # JSON is best for agents/tools. Console is best for humans.
    # We'll default to JSON if json_logs=True (e.g. prod env var), else Console.
    
    if json_logs or os.getenv("LOG_JSON", "false").lower() == "true":
        processors.append(structlog.processors.JSONRenderer())
    else:
        processors.append(structlog.dev.ConsoleRenderer())

    structlog.configure(
        processors=processors,
        logger_factory=structlog.stdlib.LoggerFactory(),
        wrapper_class=structlog.stdlib.BoundLogger,
        cache_logger_on_first_use=True,
    )

    # Create logger with static context
    log = structlog.get_logger()
    log = log.bind(service=service_name)
    return log

# --- Context Helpers ---

def set_trace_id(trace_id: str):
    return _trace_id_ctx.set(trace_id)

def reset_trace_id(token):
    _trace_id_ctx.reset(token)

def set_agent_id(agent_id: str):
    return _agent_id_ctx.set(agent_id)

def get_logger(name: str):
    """
    Return a structlog logger. Name determines component.
    """
    return structlog.get_logger().bind(component=name)
