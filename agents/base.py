import asyncio
import json
import logging
from typing import Optional, Callable, Awaitable
from datetime import datetime
import nats
from nats.js import JetStreamContext
from nats.aio.client import Client as NATS

from shared.schemas import AgentConfig, BaseMessage, EventMessage, MessageType, AgentMessage, ToolCallMessage, ToolResultMessage
from agents.llm import LLMClient, OllamaClient, MockLLMClient
from agents.memory import AgentContext

logger = logging.getLogger(__name__)

class BaseAgent:
    def __init__(self, config: AgentConfig, nats_url: str = "nats://localhost:4222"):
        self.config = config
        self.nats_url = nats_url
        self.nc: Optional[NATS] = None
        self.js: Optional[JetStreamContext] = None
        
        # Brain Components
        self.llm: Optional[LLMClient] = None
        self.memory: Optional[AgentContext] = None
        
        self._setup_brain()

    def _setup_brain(self):
        # Initialize LLM Client based on config
        # Defaulting to Ollama for now, but could switch based on self.config.backend
        if self.config.backend == "ollama":
            # Use the configured host IP for Ollama, or override with env var
            import os
            host = os.getenv("OLLAMA_HOST", self.config.host)
            ollama_url = f"http://{host}:11434"
            self.llm = OllamaClient(model="llama3", base_url=ollama_url) 
        else:
            self.llm = MockLLMClient()
            
        # Initialize Memory
        system_prompt = self.config.prompt_config.get("system_prompt", "You are a helpful AI agent.")
        
        # Append Tool Instructions
        system_prompt += "\n\nAvailable Tools:\n- mqtt_publish {\"topic\": str, \"payload\": str}\n\nTo use a tool, your response must contain the following line exactly:\nTOOL: tool_name {\"argument\": \"value\"}\nDo not simulate the tool execution. Just output the TOOL line and stop."
        
        self.memory = AgentContext(system_prompt=system_prompt)

    async def connect(self):
        self.nc = await nats.connect(self.nats_url)
        self.js = self.nc.jetstream()
        logger.info(f"Agent {self.config.name} connected to NATS at {self.nats_url}")

    async def disconnect(self):
        if self.nc:
            await self.nc.close()
            logger.info(f"Agent {self.config.name} disconnected")

    async def publish(self, subject: str, message: BaseMessage):
        if not self.js:
            raise RuntimeError("Not connected to NATS JetStream")
        
        payload = message.model_dump_json().encode()
        await self.js.publish(subject, payload)
        logger.info(f"Published message to {subject}: {message.id}")

    async def subscribe(self, subject: str, callback: Callable[[BaseMessage], Awaitable[None]]):
        if not self.js:
            raise RuntimeError("Not connected to NATS JetStream")

        async def msg_handler(msg):
            try:
                data = json.loads(msg.data.decode())
                # Basic type inference, could be improved
                if data.get("type") == MessageType.EVENT:
                    message = EventMessage(**data)
                elif data.get("sender") and data.get("recipient"): # Simple heuristic for AgentMessage
                    message = AgentMessage(**data)
                elif "result" in data and "call_id" in data:
                    message = ToolResultMessage(**data)
                elif "tool_name" in data and "call_id" in data:
                    message = ToolCallMessage(**data)
                else:
                    message = BaseMessage(**data)
                
                await callback(message)
                await msg.ack()
            except Exception as e:
                logger.error(f"Error processing message: {e}")
                # Optionally nak or term based on error type

        await self.js.subscribe(subject, cb=msg_handler)
        logger.info(f"Agent {self.config.name} subscribed to {subject}")

    async def run(self):
        # Override this method to implement agent logic
        await self.connect()
        
        # Auto-subscribe to configured input channels
        # Auto-subscribe to configured input channels
        for channel in self.config.messaging.inputs:
            await self.subscribe(channel, self.handle_message)
            
        # Subscribe to inter-agent communication channel if configured
        # Assuming the config object passed has this field (it's in the Team schema, but might need to be passed to agent config)
        # For now, let's assume the agent config has an 'inter_comm_channel' field or we derive it.
        # Subscribe to direct chat channel
        await self.subscribe(f"chat.agent.{self.config.name}", self.handle_agent_message)

        # Subscribe to inter-agent communication channel if configured (Legacy/Team)
        if hasattr(self.config, 'inter_comm_channel') and self.config.inter_comm_channel:
             await self.subscribe(self.config.inter_comm_channel, self.handle_agent_message)

        # Subscribe to tool results
        await self.subscribe(f"mcp.result.{self.config.name}", self.handle_tool_result)

        # Keep alive
        while True:
            await asyncio.sleep(1)

    async def handle_message(self, message: BaseMessage):
        """Default message handler, override in subclasses"""
        logger.info(f"Agent {self.config.name} received message: {message.id}")

    async def handle_agent_message(self, message: AgentMessage):
        """Handler for inter-agent messages"""
        # Accept if it's for me OR if it's a broadcast (and I'm not the sender)
        if (message.recipient == self.config.name or message.recipient == "all") and message.sender != self.config.name:
            logger.info(f"Agent {self.config.name} received direct message from {message.sender}: {message.content}")
            
            # 1. Update Memory
            self.memory.add_message("user", f"{message.sender}: {message.content}")
            
            # 2. Generate Response
            response_text = await self.llm.generate(self.memory.get_messages())
            
            # 3. Update Memory with own response
            self.memory.add_message("assistant", response_text)

            # Check for Tool Calls (Simple heuristic)
            # Format: TOOL: tool_name {"arg": "val"}
            if "TOOL:" in response_text:
                try:
                    parts = response_text.split("TOOL:", 1)[1].strip()
                    tool_name, json_args = parts.split(" ", 1)
                    args = json.loads(json_args)
                    logger.info(f"Agent {self.config.name} deciding to call tool: {tool_name}")
                    await self.call_tool(tool_name, args)
                    return # Stop here, wait for result
                except Exception as e:
                    logger.error(f"Failed to parse tool call: {e}")
            
            # 4. Reply
            logger.info(f"Agent {self.config.name} replying: {response_text}")
            await self.send_agent_message(recipient=message.sender, content=response_text, intent="reply")
        else:
            # Ignore messages not meant for this agent
            pass

    async def send_agent_message(self, recipient: str, content: str, intent: str = "inform"):
        """Send a message to another agent or user"""
        
        # Determine topic based on recipient
        if recipient == "all":
             # Fallback to team channel if available
             topic = self.config.inter_comm_channel if hasattr(self.config, 'inter_comm_channel') and self.config.inter_comm_channel else "chat.broadcast"
        elif recipient.startswith("user"):
             topic = f"chat.user.{recipient}" # Reply to user
        else:
             topic = f"chat.agent.{recipient}" # Reply to agent

        msg = AgentMessage(
            id=f"msg-{datetime.utcnow().timestamp()}",
            source_agent_id=self.config.name,
            sender=self.config.name,
            recipient=recipient,
            content=content,
            intent=intent,
            type=MessageType.TEXT
        )
        await self.publish(topic, msg)

    async def call_tool(self, tool_name: str, arguments: dict):
        """Call an external tool via MCP Bridge"""
        call_id = f"call-{datetime.utcnow().timestamp()}"
        msg = ToolCallMessage(
            id=f"msg-{datetime.utcnow().timestamp()}",
            source_agent_id=self.config.name,
            tool_name=tool_name,
            arguments=arguments,
            call_id=call_id
        )
        # Publish to mcp.call.<tool_name> (or just mcp.call.generic)
        await self.publish(f"mcp.call.{tool_name}", msg)
        return call_id

    async def handle_tool_result(self, message: ToolResultMessage):
        """Handle result from MCP Bridge"""
        logger.info(f"Agent {self.config.name} received tool result: {message.result}")
        
        # Add result to memory
        self.memory.add_message("user", f"Tool Result ({message.call_id}): {json.dumps(message.result)}")
        
        # Trigger another think cycle?
        # For now, just log it. In a real loop, we'd call the LLM again.
        response_text = await self.llm.generate(self.memory.get_messages())
        self.memory.add_message("assistant", response_text)
        
        # If we have a pending conversation, reply to it? 
        # This is where state management gets tricky. For now, let's just announce the result to the team.
        if hasattr(self.config, 'inter_comm_channel') and self.config.inter_comm_channel:
             await self.send_agent_message("all", f"I completed the action: {response_text}")
