from fastapi import APIRouter, Request
from shared.schemas import TextMessage, AgentMessage, MessageType
from nats.js import JetStreamContext
from datetime import datetime

router = APIRouter(tags=["chat"])

from fastapi import APIRouter, Request, Body
from shared.schemas import TextMessage, AgentMessage, MessageType

@router.post("/agents/{agent_name}/chat")
async def chat_with_agent(request: Request, agent_name: str, message: TextMessage):
    """Send a direct message to an agent."""
    try:
        js: JetStreamContext = request.app.state.js
        
        # Construct AgentMessage
        # We use "user" as the sender for now, or could use a session ID
        sender_id = "user" 
        
        agent_msg = AgentMessage(
            id=f"msg-{datetime.utcnow().timestamp()}",
            source_agent_id="api",
            sender=sender_id,
            recipient=agent_name,
            content=message.content,
            intent="chat",
            type=MessageType.TEXT
        )
        
        # Ensure stream exists
        try:
            await js.add_stream(name="chat-agent", subjects=["chat.agent.>"])
        except Exception as e:
            # print(f"Stream creation error: {e}")
            pass

        # Publish to chat.agent.{name}
        await js.publish(f"chat.agent.{agent_name}", agent_msg.model_dump_json().encode())
        
        return {"status": "sent", "id": agent_msg.id}
    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"status": "error", "detail": str(e), "traceback": traceback.format_exc()}
