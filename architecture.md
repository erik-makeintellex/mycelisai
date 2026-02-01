# System Architecture & Design Specification

> **Target Audience**: AI Agents, System Architects
> **Purpose**: Definitive source of truth for system components, data flows, and infrastructure interfaces.

## 1. System Overview

> 🔙 **Navigation**: [Back to README](README.md)

**Mycelis AI** is an Event-Driven, Agentic Service Network. It decouples "Thinking" (Agents/Runners) from "Serving" (API/UI) using an asynchronous message bus (NATS).

### Core Philosophy
-   **Event-First**: All generic state changes are broadcast as events.
-   **Persistence-Layered**: Instant operational state (NATS) vs. Long-term memory (Postgres).
-   **Agentic-Decoupling**: Agents are stateless logic units; Context is hydrated from the platform.

## 2. Component Manifest

| Component | Service Name | Path | Type | Responsibility |
| :--- | :--- | :--- | :--- | :--- |
| **Gateway** | `api` | `/api` | FastAPI | REST endpoints, SSE Streaming, Auth, DB Access (Write-Through). |
| **Frontend** | `ui` | `/ui` | Next.js | User Interface, Event Consumption (via SSE), Chat. |
| **Broker** | `nats` | - | NATS JetStream | Event Bus, Stream persistence (Message Queue). |
| **Brain** | `runner` | `/runner` | Python/LangChain | Agent Runtime. Subscribes to tasks, calls LLM, publishes results. |
| **Memory** | `memory-mcp` | `/services/memory_mcp` | FastMCP | Semantic retrieval, Embedding generation, Unstructured storage. |
| **Database** | `database` | - | PostgreSQL | Relational Source of Truth (Schemas: `shared/db.py`). |

## 3. Data Flow Architectures

### 3.1. The "Chat Loop" (Synchronous-Style Interaction)
This flow describes how a user message becomes an agent reply.

1.  **Ingest**: UI `POST /agents/{name}/chat` -> API.
2.  **Persist (User)**: API saves `MessageDB` (User) to Postgres.
3.  **Publish**: API publishes `AgentMessage` to NATS subject `chat.agent.{name}`.
4.  **Route**: NATS streams message to `runner` (subscribed to `chat.agent.>`).
5.  **Process**:
    -   Runner receives message.
    -   **Hydrate**: Runner calls API `GET /conversations/{id}/history` to load Context.
    -   **Think**: Runner calls LLM (Ollama/OpenAI).
    -   **Act**: Runner executes Tools (if any).
6.  **Reply**: Runner publishes result to NATS subject `chat.agent.{name}.reply`.
7.  **Persist (Agent)**: API (subscribed to `chat.agent.>`) receives reply -> Saves `MessageDB` (Agent).
8.  **Stream**: API pushes event via SSE (`/stream/{channel}`) -> UI displays message.

### 3.2. stream_provisioning (Infrastructure)
-   **Dynamic Streams**: Both API and Runner use `add_stream` (idempotent) or `update_stream` to ensure NATS streams exist before publishing.
-   **Resilience**: Logic handles `Stream Name in Use` errors by looking up existing streams by Subject.

## 4. Database Schema (Source: `shared/db.py`)

### Core Tables
-   `agents`: Configuration/Identity of computational entities.
-   `conversations`: Threads of interaction (`id`, `user_id`, `created_at`).
-   **`messages`**: The atomic unit of history (`id`, `conversation_id`, `sender`, `content`, `timestamp`).
-   **`memories`**: extracted insights (`agent_id`, `content`, `tags`). *Supports cascading deletes.*

## 5. Event Schema (NATS)

All messages on NATS follow the `AgentMessage` Pydantic model (`shared/schemas.py`).

```json
{
  "id": "uuid",
  "source_agent_id": "sender_name",
  "recipient": "target_agent",
  "content": "Text payload",
  "type": "text|event|tool_call",
  "timestamp": "iso-date",
  "payload": { "additional": "metadata" }
}
```

## 6. Infrastructure & Ports

| Service | Internal Port | Host Port (Dev) | Protocol |
| :--- | :--- | :--- | :--- |
| **API** | 8000 | 8000 | HTTP/Websocket |
| **UI** | 3000 | 3000 | HTTP |
| **NATS** | 4222 | 4222 | TCP |
| **Postgres** | 5432 | 5432 | TCP |
| **Ollama** | 11434 | 11434 | HTTP (LLM Backend) |

## 7. Developer Workflow (for Agents)

To modify this system:
1.  **Schema Change**: Edit `shared/db.py`. Restart `api` to auto-migrate.
2.  **New Event**: Define in `shared/schemas.py`. Update `api` (publisher) and `runner` (consumer).
3.  **New Service**: Create folder in `services/`. Add `Dockerfile`. Update `Makefile` to include in dev loop.

---
*Generated for Machine Readability by Antigravity.*
