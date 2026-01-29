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
            pass

        # Save to DB (Persistent History)
        # Use a agent-specific conversation for V1 isolation
        conversation_id = f"session-{agent_name}"
        
        from shared.db import get_db, MessageDB, ConversationDB
        from sqlalchemy import select

        async for session in get_db():
             # Check for thread
             conn_stmt = select(ConversationDB).where(ConversationDB.id == conversation_id)
             res = await session.execute(conn_stmt)
             conversation = res.scalars().first()
             
             if not conversation:
                 conversation = ConversationDB(id=conversation_id, title=f"Chat with {agent_name}", user_id="user")
                 session.add(conversation)
                 session.commit() # Commit to ensure existence
             
             # Save Message
             db_msg = MessageDB(
                 id=agent_msg.id,
                 conversation_id=conversation_id,
                 sender="user",
                 role="user",
                 content=message.content,
                 type="text",
                 timestamp=datetime.fromtimestamp(float(agent_msg.id.split('-')[1])) if '-' in agent_msg.id else datetime.utcnow()
             )
             session.add(db_msg)
             await session.commit() 
             break # Close session loop 

        await js.publish(f"chat.agent.{agent_name}", agent_msg.model_dump_json().encode())
        
        return {"status": "sent", "id": agent_msg.id}
    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"status": "error", "detail": str(e), "traceback": traceback.format_exc()}
